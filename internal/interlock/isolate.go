package interlock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/fabchem/internal/valve"
)

type Isolator struct {
	mu      sync.Mutex
	valves  *valve.Controller
	latches map[string]string
}

func NewIsolator(valves *valve.Controller) *Isolator {
	return &Isolator{valves: valves, latches: make(map[string]string)}
}

func (i *Isolator) Isolate(ctx context.Context, incidentID, reason string, delay time.Duration, fail bool) (valve.IsolationProof, error) {
	if incidentID == "" || reason == "" {
		return valve.IsolationProof{}, fmt.Errorf("incident identity and isolation reason are required")
	}
	i.mu.Lock()
	i.latches[incidentID] = reason
	i.mu.Unlock()
	proof, err := i.valves.IsolateUpstream(ctx, incidentID, delay, fail)
	if err != nil {
		return proof, err
	}
	return proof, nil
}

func (i *Isolator) Latched(incidentID string) bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.latches[incidentID] != ""
}

func (i *Isolator) Reason(incidentID string) string {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.latches[incidentID]
}
