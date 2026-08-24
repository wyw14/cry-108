package source

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/fabchem/internal/meter"
)

type Switch struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	RouteID   string    `json:"route_id"`
	StartedAt time.Time `json:"started_at"`
}

func (s *Service) SwitchTo(ctx context.Context, branchID string) (Switch, error) {
	s.mu.Lock()
	branch := s.branches[branchID]
	if branch == nil {
		s.mu.Unlock()
		return Switch{}, fmt.Errorf("source branch %s does not exist", branchID)
	}
	previous := s.active
	branch.Purged = 0
	branch.Available = false
	s.active = branchID
	switching := Switch{ID: uuid.NewString(), From: previous, To: branchID, RouteID: branch.RouteID, StartedAt: time.Now().UTC()}
	s.mu.Unlock()
	err := s.record(ctx, "source.switched", switching.ID, map[string]any{"from": previous, "to": branchID, "route": branch.RouteID})
	return switching, err
}

func (s *Service) ApplyPurgePulse(ctx context.Context, pulse meter.Pulse) (bool, error) {
	s.mu.Lock()
	// Credit the displacement to the physical path the pulse actually traversed
	// (pulse.BranchID), not to the currently active branch. During a source
	// switch the residual flow still travels the old branch's piping for a
	// short time; those pulses are labelled with the old branch and must only
	// advance that branch's purge accounting. The new branch becomes available
	// solely from volume that genuinely displaced fluid through its own path,
	// otherwise it would be reported purge-complete while still carrying the
	// previous source's tracer.
	branch := s.branches[pulse.BranchID]
	if branch == nil {
		s.mu.Unlock()
		return false, fmt.Errorf("pulse physical path %s/%s is not a configured source branch", pulse.RouteID, pulse.BranchID)
	}
	available := branch.AddPurge(pulse.Volume)
	remaining := branch.RemainingPurge()
	s.mu.Unlock()
	err := s.record(ctx, "source.purge", pulse.BranchID, map[string]any{
		"pulse": pulse.ID, "route": pulse.RouteID, "available": available, "remaining": remaining,
	})
	return available, err
}

func (s *Service) Branch(branchID string) (Branch, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	branch, ok := s.branches[branchID]
	if !ok {
		return Branch{}, false
	}
	return *branch, true
}
