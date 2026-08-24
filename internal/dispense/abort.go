package dispense

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/wyw14/fabchem/internal/model"
	"github.com/wyw14/fabchem/internal/returnline"
)

type AbortFinalizer struct {
	returns *returnline.Service
}

func NewAbortFinalizer(returns *returnline.Service) *AbortFinalizer {
	return &AbortFinalizer{returns: returns}
}

type FlushStart struct {
	Request returnline.Request `json:"request"`
	Target  string             `json:"target"`
}

func (f *AbortFinalizer) StartFlush(ctx context.Context, session Session, routeID string, volume float64) (FlushStart, error) {
	if session.ID == "" || routeID == "" {
		return FlushStart{}, fmt.Errorf("aborted dispense and route are required")
	}
	request, err := returnline.NewReturnRequest(uuid.NewString(), routeID, session.ToolID, model.ClassWater, returnline.Flush, volume)
	if err != nil {
		return FlushStart{}, err
	}
	destination, err := returnline.ResolveDestination(request, returnline.DefaultDestinations())
	if err != nil {
		return FlushStart{}, err
	}
	if _, err := f.returns.Start(ctx, request, destination.Name); err != nil {
		return FlushStart{}, err
	}
	return FlushStart{Request: request, Target: destination.Name}, nil
}

func (s *Service) Abort(ctx context.Context, id string) (Session, error) {
	s.mu.Lock()
	session, err := s.ensureSession(id)
	if err == nil {
		session.State = Aborted
	}
	result := Session{}
	if session != nil {
		result = *session
	}
	s.mu.Unlock()
	if err != nil {
		return Session{}, err
	}
	err = s.recorder.Append(ctx, model.NewEvent("dispense.aborted", id, nil))
	return result, err
}
