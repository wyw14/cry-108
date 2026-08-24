package meter

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/fabchem/internal/model"
)

type Pulse struct {
	ID         string    `json:"id"`
	RouteID    string    `json:"route_id"`
	BranchID   string    `json:"branch_id"`
	Volume     float64   `json:"volume"`
	CapturedAt time.Time `json:"captured_at"`
}

func (s *Service) CapturePulse(ctx context.Context, routeID, branchID string, volume float64) (Pulse, error) {
	if routeID == "" || branchID == "" || volume <= 0 {
		return Pulse{}, fmt.Errorf("pulse requires physical route, branch, and positive volume")
	}
	pulse := Pulse{ID: uuid.NewString(), RouteID: routeID, BranchID: branchID, Volume: volume, CapturedAt: time.Now().UTC()}
	s.mu.Lock()
	s.pulses = append(s.pulses, pulse)
	s.mu.Unlock()
	err := s.record(ctx, model.NewEvent("meter.pulse", routeID, map[string]any{
		"pulse": pulse.ID, "branch": branchID, "volume": volume,
	}))
	return pulse, err
}

func DeliverPulses(pulses []Pulse, consume func(Pulse) error) error {
	for _, pulse := range pulses {
		if err := consume(pulse); err != nil {
			return err
		}
	}
	return nil
}
