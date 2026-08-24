package blend

import (
	"context"
	"fmt"
	"time"
)

type StartOptions struct {
	WaterFeedbackDelay time.Duration
	MinimumWaterFlow   float64
}

func (s *Service) Start(ctx context.Context, batchID string, options StartOptions) (Batch, error) {
	if options.MinimumWaterFlow <= 0 {
		return Batch{}, fmt.Errorf("minimum dilution water flow must be positive")
	}
	if err := s.setState(ctx, batchID, "routing"); err != nil {
		return Batch{}, err
	}
	command, err := s.valves.OpenDilutionWater(ctx, batchID)
	if err != nil {
		return Batch{}, err
	}
	s.mu.Lock()
	s.batches[batchID].WaterAcceptedAt = command.AcceptedAt
	s.mu.Unlock()
	startedAt, err := s.valves.StartConcentrate(ctx, batchID)
	if err != nil {
		return Batch{}, err
	}
	proof, err := s.meters.EstablishDilutionFlow(ctx, batchID, options.WaterFeedbackDelay, options.MinimumWaterFlow)
	if err != nil {
		return Batch{}, err
	}
	if proof.BatchID != batchID || proof.LitersMinute < options.MinimumWaterFlow {
		return Batch{}, fmt.Errorf("dilution flow proof does not satisfy current batch")
	}
	s.mu.Lock()
	batch := s.batches[batchID]
	batch.WaterFlowAt = proof.Established
	batch.AcidStartedAt = startedAt
	result := *batch
	s.mu.Unlock()
	if err := s.setState(ctx, batchID, "conditioning"); err != nil {
		return Batch{}, err
	}
	result.State = "conditioning"
	return result, nil
}

func (s *Service) Stop(ctx context.Context, batchID string) error {
	return s.valves.StopBlend(ctx, batchID)
}
