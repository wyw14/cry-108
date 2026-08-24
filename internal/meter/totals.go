package meter

import (
	"context"
	"fmt"

	"github.com/wyw14/fabchem/internal/model"
)

type BatchMeasurements struct {
	BlendID string  `json:"blend_id"`
	Flow    float64 `json:"flow"`
	Density float64 `json:"density"`
	Ready   bool    `json:"ready"`
}

func (s *Service) RecordFlowTotal(ctx context.Context, blendID string, liters float64) error {
	if blendID == "" || liters <= 0 {
		return fmt.Errorf("flow total requires batch and positive volume")
	}
	s.mu.Lock()
	s.flowTotals[blendID] = liters
	s.mu.Unlock()
	return s.record(ctx, model.NewEvent("blend.flow.total", blendID, map[string]any{"liters": liters}))
}

func (s *Service) RecordDensity(ctx context.Context, blendID string, density float64) error {
	if blendID == "" || density <= 0 {
		return fmt.Errorf("density requires batch and positive value")
	}
	s.mu.Lock()
	s.densities[blendID] = density
	s.mu.Unlock()
	return s.record(ctx, model.NewEvent("blend.density", blendID, map[string]any{"density": density}))
}

func (s *Service) Measurements(blendID string) BatchMeasurements {
	s.mu.Lock()
	defer s.mu.Unlock()
	flow, flowOK := s.flowTotals[blendID]
	density, densityOK := s.densities[blendID]
	return BatchMeasurements{BlendID: blendID, Flow: flow, Density: density, Ready: flowOK && densityOK}
}
