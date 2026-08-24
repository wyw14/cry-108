package source

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/wyw14/fabchem/internal/model"
)

type Service struct {
	mu       sync.Mutex
	branches map[string]*Branch
	tanks    map[string]*DayTank
	active   string
	recorder model.Recorder
}

func NewService(recorder model.Recorder) *Service {
	if recorder == nil {
		recorder = model.NopRecorder{}
	}
	return &Service{branches: make(map[string]*Branch), tanks: make(map[string]*DayTank), recorder: recorder}
}

func (s *Service) AddBranch(branch Branch) error {
	if err := branch.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.branches[branch.ID]; exists {
		return fmt.Errorf("source branch %s already exists", branch.ID)
	}
	copy := branch
	s.branches[branch.ID] = &copy
	if s.active == "" {
		s.active = branch.ID
	}
	return nil
}

func (s *Service) AddDayTank(tank DayTank) error {
	if err := tank.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	copy := tank
	s.tanks[tank.ID] = &copy
	return nil
}

func (s *Service) ActiveBranch() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}

func (s *Service) Branches() []Branch {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]Branch, 0, len(s.branches))
	for _, branch := range s.branches {
		result = append(result, *branch)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func (s *Service) record(ctx context.Context, kind, aggregate string, values map[string]any) error {
	return s.recorder.Append(ctx, model.Event{Kind: kind, Aggregate: aggregate, At: time.Now().UTC(), Data: values})
}
