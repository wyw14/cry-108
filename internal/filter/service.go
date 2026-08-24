package filter

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
)

type Service struct {
	mu       sync.Mutex
	sessions map[string]*Session
	meters   *meter.Service
	permits  *interlock.PermitBook
	recorder model.Recorder
}

func NewService(meters *meter.Service, permits *interlock.PermitBook, recorder model.Recorder) *Service {
	if recorder == nil {
		recorder = model.NopRecorder{}
	}
	return &Service{sessions: make(map[string]*Session), meters: meters, permits: permits, recorder: recorder}
}

func (s *Service) Begin(ctx context.Context, filterID string, wettingTarget float64) (Session, error) {
	if filterID == "" || wettingTarget <= 0 {
		return Session{}, fmt.Errorf("filter identity and wetting target are required")
	}
	session := &Session{ID: uuid.NewString(), FilterID: filterID, WettingTarget: wettingTarget, Phase: Preparing, StartedAt: time.Now().UTC()}
	s.mu.Lock()
	s.sessions[session.ID] = session
	s.mu.Unlock()
	s.permits.Revoke("filter:" + filterID)
	err := s.recorder.Append(ctx, model.NewEvent("filter.preparation.started", session.ID, map[string]any{"filter": filterID, "target": wettingTarget}))
	return *session, err
}

func (s *Service) Session(sessionID string) (Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[sessionID]
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
	sort.Slice(result, func(i, j int) bool { return result[i].StartedAt.Before(result[j].StartedAt) })
	return result
}
