package nitrogen

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/fabchem/internal/model"
	"github.com/wyw14/fabchem/internal/source"
	"github.com/wyw14/fabchem/internal/valve"
)

type Service struct {
	mu       sync.Mutex
	capacity float64
	source   *source.Service
	valves   *valve.Controller
	lastPlan Plan
	recorder model.Recorder
}

func NewService(capacity float64, sourceService *source.Service, valves *valve.Controller, recorder model.Recorder) *Service {
	if recorder == nil {
		recorder = model.NopRecorder{}
	}
	return &Service{capacity: capacity, source: sourceService, valves: valves, recorder: recorder}
}

func (s *Service) SetCapacity(value float64) error {
	if value <= 0 {
		return fmt.Errorf("nitrogen header capacity must be positive")
	}
	s.mu.Lock()
	s.capacity = value
	s.mu.Unlock()
	return nil
}

func (s *Service) Allocate(ctx context.Context) (Plan, error) {
	s.mu.Lock()
	capacity := s.capacity
	s.mu.Unlock()
	plan, err := BuildPlan(uuid.NewString(), capacity, s.source.DayTanks())
	if err != nil {
		return Plan{}, err
	}
	settings := make([]valve.NitrogenSetting, 0, len(plan.Allocations))
	for _, allocation := range plan.Allocations {
		settings = append(settings, valve.NitrogenSetting{TankID: allocation.TankID, Flow: allocation.Flow, Reason: allocation.Reason})
	}
	if err := s.valves.ApplyNitrogen(ctx, plan.ID, settings); err != nil {
		return Plan{}, err
	}
	s.mu.Lock()
	s.lastPlan = plan
	s.mu.Unlock()
	err = s.recorder.Append(ctx, model.Event{Kind: "nitrogen.plan", Aggregate: plan.ID, At: time.Now().UTC(), Data: map[string]any{"capacity": capacity, "allocated": plan.TotalAllocated}})
	return plan, err
}

func (s *Service) LastPlan() Plan {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := s.lastPlan
	result.Allocations = append([]Allocation(nil), result.Allocations...)
	sort.Slice(result.Allocations, func(i, j int) bool { return result.Allocations[i].TankID < result.Allocations[j].TankID })
	return result
}
