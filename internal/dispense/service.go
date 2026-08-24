package dispense

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/fabchem/internal/interlock"
	"github.com/wyw14/fabchem/internal/meter"
	"github.com/wyw14/fabchem/internal/model"
	"github.com/wyw14/fabchem/internal/valve"
)

type Service struct {
	mu       sync.Mutex
	sessions map[string]*Session
	valves   *valve.Controller
	meters   *meter.Service
	permits  *interlock.PermitBook
	recorder model.Recorder
}

func NewService(valves *valve.Controller, meters *meter.Service, permits *interlock.PermitBook, recorder model.Recorder) *Service {
	if recorder == nil {
		recorder = model.NopRecorder{}
	}
	return &Service{sessions: make(map[string]*Session), valves: valves, meters: meters, permits: permits, recorder: recorder}
}

func (s *Service) Create(ctx context.Context, request Request) (Session, error) {
	if err := request.Validate(); err != nil {
		return Session{}, err
	}
	if request.ID == "" {
		request.ID = uuid.NewString()
	}
	s.mu.Lock()
	if existing := s.sessions[request.ID]; existing != nil {
		result := *existing
		s.mu.Unlock()
		return result, nil
	}
	session := &Session{ID: request.ID, BlendID: request.BlendID, ToolID: request.ToolID, Target: request.TargetLiters, Tolerance: request.Tolerance, State: Requested, CreatedAt: time.Now().UTC()}
	s.sessions[session.ID] = session
	s.mu.Unlock()
	if err := s.meters.BeginSettlement(session.ID); err != nil {
		return Session{}, err
	}
	err := s.recorder.Append(ctx, model.NewEvent("dispense.requested", session.ID, map[string]any{"tool": session.ToolID, "target": session.Target}))
	return *session, err
}

func (s *Service) Session(id string) (Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if !ok {
		return Session{}, false
	}
	return *session, true
}

func (s *Service) List() []Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		result = append(result, *session)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.Before(result[j].CreatedAt) })
	return result
}

func (s *Service) ensureSession(id string) (*Session, error) {
	session := s.sessions[id]
	if session == nil {
		return nil, fmt.Errorf("dispense %s does not exist", id)
	}
	return session, nil
}
