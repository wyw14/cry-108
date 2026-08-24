package manifold

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/wyw14/fabchem/internal/model"
	"github.com/wyw14/fabchem/internal/valve"
)

type Service struct {
	mu       sync.Mutex
	routes   map[string]model.Route
	valves   *valve.Controller
	recorder model.Recorder
}

func NewService(valves *valve.Controller, recorder model.Recorder) *Service {
	if recorder == nil {
		recorder = model.NopRecorder{}
	}
	return &Service{routes: make(map[string]model.Route), valves: valves, recorder: recorder}
}

func (s *Service) Commit(ctx context.Context, route model.Route) error {
	if err := s.valves.ReserveRoute(ctx, route); err != nil {
		return err
	}
	route.Committed = true
	s.mu.Lock()
	s.routes[route.ID] = route
	s.mu.Unlock()
	if err := s.recorder.Append(ctx, model.Event{
		Kind: "manifold.route.committed", Aggregate: route.ID, At: time.Now().UTC(),
		Data: map[string]any{"class": route.Class, "source": route.Source, "target": route.Target},
	}); err != nil {
		_ = s.valves.ReleaseRoute(ctx, route.ID)
		s.mu.Lock()
		delete(s.routes, route.ID)
		s.mu.Unlock()
		return err
	}
	return nil
}

func (s *Service) Release(ctx context.Context, routeID string) error {
	s.mu.Lock()
	if _, exists := s.routes[routeID]; !exists {
		s.mu.Unlock()
		return fmt.Errorf("route %s does not exist", routeID)
	}
	delete(s.routes, routeID)
	s.mu.Unlock()
	if err := s.valves.ReleaseRoute(ctx, routeID); err != nil {
		return err
	}
	return s.recorder.Append(ctx, model.NewEvent("manifold.route.released", routeID, nil))
}

func (s *Service) Routes() []model.Route {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]model.Route, 0, len(s.routes))
	for _, route := range s.routes {
		result = append(result, route)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}
