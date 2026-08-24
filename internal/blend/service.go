package blend

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/fabchem/internal/meter"
	"github.com/wyw14/fabchem/internal/model"
	"github.com/wyw14/fabchem/internal/valve"
)

type Service struct {
	mu       sync.Mutex
	batches  map[string]*Batch
	valves   *valve.Controller
	meters   *meter.Service
	recorder model.Recorder
}

func NewService(valves *valve.Controller, meters *meter.Service, recorder model.Recorder) *Service {
	if recorder == nil {
		recorder = model.NopRecorder{}
	}
	return &Service{batches: make(map[string]*Batch), valves: valves, meters: meters, recorder: recorder}
}

func (s *Service) Create(ctx context.Context, request CreateRequest) (Batch, error) {
	if err := request.Validate(); err != nil {
		return Batch{}, err
	}
	batch := &Batch{ID: uuid.NewString(), Chemical: request.Chemical, TargetLiters: request.TargetLiters, State: model.BatchPrepared, CreatedAt: time.Now().UTC()}
	s.mu.Lock()
	s.batches[batch.ID] = batch
	s.mu.Unlock()
	if err := s.recorder.Append(ctx, model.NewEvent("blend.created", batch.ID, map[string]any{"chemical": batch.Chemical.Name, "liters": batch.TargetLiters})); err != nil {
		return Batch{}, err
	}
	return *batch, nil
}

func (s *Service) Batch(batchID string) (Batch, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch, ok := s.batches[batchID]
	if !ok {
		return Batch{}, false
	}
	return *batch, true
}

func (s *Service) List() []Batch {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]Batch, 0, len(s.batches))
	for _, batch := range s.batches {
		result = append(result, *batch)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.Before(result[j].CreatedAt) })
	return result
}

func (s *Service) setState(ctx context.Context, batchID string, state model.BatchState) error {
	s.mu.Lock()
	batch := s.batches[batchID]
	if batch == nil {
		s.mu.Unlock()
		return fmt.Errorf("blend %s does not exist", batchID)
	}
	if err := model.RequireTransition(batch.State, state); err != nil {
		s.mu.Unlock()
		return err
	}
	batch.State = state
	s.mu.Unlock()
	return s.recorder.Append(ctx, model.NewEvent("blend.state", batchID, map[string]any{"state": state}))
}
