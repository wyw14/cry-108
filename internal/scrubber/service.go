package scrubber

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/wyw14/fabchem/internal/interlock"
	"github.com/wyw14/fabchem/internal/model"
)

type Service struct {
	mu          sync.Mutex
	handovers   map[string]*Handover
	drainBudget *DrainBudget
	exhaust     *interlock.ExhaustPermit
	recorder    model.Recorder
}

func NewService(capacity float64, exhaust *interlock.ExhaustPermit, recorder model.Recorder) *Service {
	if recorder == nil {
		recorder = model.NopRecorder{}
	}
	return &Service{handovers: make(map[string]*Handover), drainBudget: NewDrainBudget(capacity), exhaust: exhaust, recorder: recorder}
}

func (s *Service) Handover(id string) (Handover, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	handover, ok := s.handovers[id]
	if !ok {
		return Handover{}, false
	}
	return *handover, true
}

func (s *Service) Handovers() []Handover {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]Handover, 0, len(s.handovers))
	for _, handover := range s.handovers {
		result = append(result, *handover)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].StartedAt.Before(result[j].StartedAt) })
	return result
}

func (s *Service) Drain(ctx context.Context, incidentID string, requested float64) (float64, error) {
	if incidentID == "" {
		return 0, fmt.Errorf("incident identity is required for drain")
	}
	granted := s.drainBudget.Grant(incidentID, requested)
	if granted <= 0 {
		return 0, fmt.Errorf("scrubber has no drain capacity available")
	}
	err := s.recorder.Append(ctx, model.Event{Kind: "scrubber.drain", Aggregate: incidentID, At: time.Now().UTC(), Data: map[string]any{"requested": requested, "granted": granted}})
	return granted, err
}

func (s *Service) CompleteDrain(incidentID string, volume float64) {
	s.drainBudget.Release(incidentID, volume)
}

func (s *Service) DrainSnapshot() map[string]float64 { return s.drainBudget.Snapshot() }
