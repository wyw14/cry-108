package returnline

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/wyw14/fabchem/internal/model"
)

type Transfer struct {
	Request   Request   `json:"request"`
	Target    string    `json:"target"`
	StartedAt time.Time `json:"started_at"`
	Completed bool      `json:"completed"`
}

type Service struct {
	mu        sync.Mutex
	transfers map[string]Transfer
	recorder  model.Recorder
}

func NewService(recorder model.Recorder) *Service {
	if recorder == nil {
		recorder = model.NopRecorder{}
	}
	return &Service{transfers: make(map[string]Transfer), recorder: recorder}
}

func (s *Service) Start(ctx context.Context, request Request, target string) (Transfer, error) {
	if err := request.Validate(); err != nil {
		return Transfer{}, err
	}
	transfer := Transfer{Request: request, Target: target, StartedAt: time.Now().UTC()}
	s.mu.Lock()
	s.transfers[request.SessionID] = transfer
	s.mu.Unlock()
	err := s.recorder.Append(ctx, model.NewEvent("return.started", request.SessionID, map[string]any{"target": target, "disposition": request.Disposition}))
	return transfer, err
}

func (s *Service) Complete(ctx context.Context, sessionID string) (Transfer, error) {
	s.mu.Lock()
	transfer := s.transfers[sessionID]
	transfer.Completed = true
	s.transfers[sessionID] = transfer
	s.mu.Unlock()
	err := s.recorder.Append(ctx, model.NewEvent("return.completed", sessionID, map[string]any{"target": transfer.Target}))
	return transfer, err
}

func (s *Service) List() []Transfer {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]Transfer, 0, len(s.transfers))
	for _, transfer := range s.transfers {
		result = append(result, transfer)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].StartedAt.Before(result[j].StartedAt) })
	return result
}
