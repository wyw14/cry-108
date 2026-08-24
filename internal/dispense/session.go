package dispense

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/wyw14/fabchem/internal/meter"
	"github.com/wyw14/fabchem/internal/model"
)

type State string

const (
	Requested  State = "requested"
	Flowing    State = "flowing"
	Settling   State = "settling"
	Delivered  State = "delivered"
	Incomplete State = "incomplete"
	Aborted    State = "aborted"
)

type Session struct {
	ID        string    `json:"id"`
	BlendID   string    `json:"blend_id"`
	ToolID    string    `json:"tool_id"`
	Target    float64   `json:"target"`
	Tolerance float64   `json:"tolerance"`
	Actual    float64   `json:"actual"`
	State     State     `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	ClosedAt  time.Time `json:"closed_at,omitempty"`
}

type Request struct {
	ID           string  `json:"id"`
	BlendID      string  `json:"blend_id"`
	ToolID       string  `json:"tool_id"`
	TargetLiters float64 `json:"target_liters"`
	Tolerance    float64 `json:"tolerance"`
}

func (r Request) Validate() error {
	if r.BlendID == "" || r.ToolID == "" || r.TargetLiters <= 0 || r.Tolerance <= 0 {
		return fmt.Errorf("dispense requires blend, tool, positive target, and tolerance")
	}
	return nil
}

func (s *Service) Start(ctx context.Context, id string) error {
	s.mu.Lock()
	session, err := s.ensureSession(id)
	if err == nil && session.State != Requested && session.State != Incomplete {
		err = fmt.Errorf("dispense %s cannot start from %s", id, session.State)
	}
	if err == nil {
		session.State = Flowing
	}
	s.mu.Unlock()
	if err != nil {
		return err
	}
	if err := s.valves.OpenOutlet(ctx, id); err != nil {
		return err
	}
	return s.recorder.Append(ctx, model.NewEvent("dispense.started", id, nil))
}

func (s *Service) AddMeasured(ctx context.Context, id string, liters float64) error {
	return s.meters.AddDispensed(ctx, id, liters)
}

func (s *Service) Close(ctx context.Context, id string, tailDelay time.Duration, tailVolume float64) (Session, error) {
	ack, err := s.valves.CloseOutlet(ctx, id)
	if err != nil {
		return Session{}, err
	}
	s.mu.Lock()
	session, err := s.ensureSession(id)
	if err == nil {
		session.State = Settling
		session.ClosedAt = ack.ClosedAt
	}
	s.mu.Unlock()
	if err != nil {
		return Session{}, err
	}
	settlement, err := s.meters.CloseAndSettle(ctx, id, 0, 0)
	if err != nil {
		return Session{}, err
	}
	return s.Finalize(ctx, settlement)
}

func (s *Service) Finalize(ctx context.Context, settlement meter.Settlement) (Session, error) {
	if !settlement.Settled {
		return Session{}, fmt.Errorf("dispense meter has not settled")
	}
	s.mu.Lock()
	session, err := s.ensureSession(settlement.DispenseID)
	if err == nil {
		session.Actual = settlement.Volume
		if math.Abs(session.Target-session.Actual) <= session.Tolerance {
			session.State = Delivered
		} else {
			session.State = Incomplete
		}
	}
	result := Session{}
	if session != nil {
		result = *session
	}
	s.mu.Unlock()
	if err != nil {
		return Session{}, err
	}
	err = s.recorder.Append(ctx, model.NewEvent("dispense.finalized", result.ID, map[string]any{"actual": result.Actual, "state": result.State}))
	return result, err
}
