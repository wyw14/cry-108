package manifold

import (
	"context"
	"fmt"

	"github.com/wyw14/fabchem/internal/meter"
	"github.com/wyw14/fabchem/internal/source"
)

type PurgeCoordinator struct {
	sources *source.Service
	meters  *meter.Service
}

func NewPurgeCoordinator(sources *source.Service, meters *meter.Service) *PurgeCoordinator {
	return &PurgeCoordinator{sources: sources, meters: meters}
}

func (p *PurgeCoordinator) CaptureAndApply(ctx context.Context, routeID, branchID string, volume float64) (bool, error) {
	pulse, err := p.meters.CapturePulse(ctx, routeID, branchID, volume)
	if err != nil {
		return false, err
	}
	return p.sources.ApplyPurgePulse(ctx, pulse)
}

func (p *PurgeCoordinator) ApplyDelayed(ctx context.Context, pulses []meter.Pulse) error {
	return meter.DeliverPulses(pulses, func(pulse meter.Pulse) error {
		_, err := p.sources.ApplyPurgePulse(ctx, pulse)
		return err
	})
}

func (p *PurgeCoordinator) Available(branchID string) (bool, error) {
	branch, ok := p.sources.Branch(branchID)
	if !ok {
		return false, fmt.Errorf("source branch %s does not exist", branchID)
	}
	return branch.Available, nil
}
