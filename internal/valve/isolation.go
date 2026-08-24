package valve

import (
	"context"
	"fmt"
	"time"

	"github.com/wyw14/fabchem/internal/model"
)

type IsolationProof struct {
	IncidentID string    `json:"incident_id"`
	ValveID    string    `json:"valve_id"`
	ClosedAt   time.Time `json:"closed_at"`
	Confirmed  bool      `json:"confirmed"`
}

func (c *Controller) IsolateUpstream(ctx context.Context, incidentID string, delay time.Duration, fail bool) (IsolationProof, error) {
	if incidentID == "" {
		return IsolationProof{}, fmt.Errorf("incident identity is required")
	}
	if delay > 0 {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return IsolationProof{}, ctx.Err()
		case <-timer.C:
		}
	}
	if fail {
		return IsolationProof{IncidentID: incidentID, ValveID: "upstream-feed"}, fmt.Errorf("upstream isolation did not confirm closed")
	}
	if err := c.command(ctx, "upstream-feed", Closed, incidentID); err != nil {
		return IsolationProof{}, err
	}
	proof := IsolationProof{IncidentID: incidentID, ValveID: "upstream-feed", ClosedAt: time.Now().UTC(), Confirmed: true}
	_ = c.recorder.Append(ctx, model.NewEvent("isolation.confirmed", incidentID, map[string]any{"valve": proof.ValveID}))
	return proof, nil
}
