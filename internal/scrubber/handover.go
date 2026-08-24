package scrubber

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/fabchem/internal/model"
)

type Handover struct {
	ID          string      `json:"id"`
	FromFan     string      `json:"from_fan"`
	ToFan       string      `json:"to_fan"`
	RPM         model.Proof `json:"rpm"`
	Draft       model.Proof `json:"draft"`
	Available   bool        `json:"available"`
	StartedAt   time.Time   `json:"started_at"`
	CompletedAt time.Time   `json:"completed_at,omitempty"`
}

func (s *Service) BeginHandover(ctx context.Context, fromFan, toFan string) (Handover, error) {
	if fromFan == "" || toFan == "" || fromFan == toFan {
		return Handover{}, fmt.Errorf("scrubber handover requires distinct fan identities")
	}
	handover := &Handover{ID: uuid.NewString(), FromFan: fromFan, ToFan: toFan, StartedAt: time.Now().UTC()}
	s.mu.Lock()
	s.handovers[handover.ID] = handover
	s.mu.Unlock()
	err := s.recorder.Append(ctx, model.NewEvent("scrubber.handover.started", handover.ID, map[string]any{"from": fromFan, "to": toFan}))
	return *handover, err
}

func (s *Service) ApplyRPM(ctx context.Context, handoverID string, rpm, minimum float64) error {
	s.mu.Lock()
	handover := s.handovers[handoverID]
	if handover == nil {
		s.mu.Unlock()
		return fmt.Errorf("scrubber handover %s does not exist", handoverID)
	}
	handover.RPM = model.Proof{SessionID: handoverID, Kind: "fan-rpm", Value: rpm, Valid: rpm >= minimum}
	s.mu.Unlock()
	return s.recorder.Append(ctx, model.NewEvent("scrubber.rpm", handoverID, map[string]any{"rpm": rpm, "minimum": minimum}))
}

func (s *Service) ApplyDraft(ctx context.Context, handoverID string, pressure, minimum float64) error {
	s.mu.Lock()
	handover := s.handovers[handoverID]
	if handover == nil {
		s.mu.Unlock()
		return fmt.Errorf("scrubber handover %s does not exist", handoverID)
	}
	handover.Draft = model.Proof{SessionID: handoverID, Kind: "draft-pressure", Value: pressure, Valid: pressure >= minimum}
	s.mu.Unlock()
	return s.recorder.Append(ctx, model.NewEvent("scrubber.draft", handoverID, map[string]any{"pressure": pressure, "minimum": minimum}))
}

func (s *Service) CompleteHandover(ctx context.Context, handoverID string) (Handover, error) {
	s.mu.Lock()
	handover := s.handovers[handoverID]
	if handover == nil {
		s.mu.Unlock()
		return Handover{}, fmt.Errorf("scrubber handover %s does not exist", handoverID)
	}
	rpm, draft := handover.RPM, handover.Draft
	s.mu.Unlock()
	draft.Valid = true
	available, err := s.exhaust.Evaluate(rpm, draft)
	if err != nil {
		return Handover{}, err
	}
	s.mu.Lock()
	handover.Available = available
	handover.CompletedAt = time.Now().UTC()
	result := *handover
	s.mu.Unlock()
	err = s.recorder.Append(ctx, model.NewEvent("scrubber.handover.completed", handoverID, map[string]any{"available": available}))
	return result, err
}
