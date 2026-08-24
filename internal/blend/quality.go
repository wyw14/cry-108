package blend

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/wyw14/fabchem/internal/meter"
	"github.com/wyw14/fabchem/internal/model"
)

type QualityResult struct {
	BlendID   string    `json:"blend_id"`
	Density   float64   `json:"density"`
	Flow      float64   `json:"flow"`
	Qualified bool      `json:"qualified"`
	Pending   bool      `json:"pending"`
	Evaluated time.Time `json:"evaluated,omitempty"`
}

type QualityReducer struct {
	meters   *meter.Service
	recorder model.Recorder
}

func NewQualityReducer(meters *meter.Service, recorder model.Recorder) *QualityReducer {
	if recorder == nil {
		recorder = model.NopRecorder{}
	}
	return &QualityReducer{meters: meters, recorder: recorder}
}

func (q *QualityReducer) Evaluate(ctx context.Context, blendID string, targetDensity, tolerance float64) (QualityResult, error) {
	if blendID == "" || targetDensity <= 0 || tolerance <= 0 {
		return QualityResult{}, fmt.Errorf("quality evaluation requires batch, target, and tolerance")
	}
	measurements := q.meters.Measurements(blendID)
	if !measurements.Ready {
		return QualityResult{BlendID: blendID, Pending: true}, nil
	}
	result := QualityResult{
		BlendID: blendID, Density: measurements.Density, Flow: measurements.Flow,
		Qualified: math.Abs(measurements.Density-targetDensity) <= tolerance,
		Evaluated: time.Now().UTC(),
	}
	err := q.recorder.Append(ctx, model.NewEvent("blend.quality", blendID, map[string]any{
		"density": result.Density, "flow": result.Flow, "qualified": result.Qualified,
	}))
	return result, err
}

func (s *Service) ApplyQuality(ctx context.Context, result QualityResult) error {
	if result.Pending || result.BlendID == "" {
		return fmt.Errorf("cannot apply incomplete quality result")
	}
	s.mu.Lock()
	batch := s.batches[result.BlendID]
	if batch == nil {
		s.mu.Unlock()
		return fmt.Errorf("blend %s does not exist", result.BlendID)
	}
	batch.Quality = result
	s.mu.Unlock()
	if result.Qualified {
		return s.setState(ctx, result.BlendID, "available")
	}
	return s.setState(ctx, result.BlendID, "isolated")
}
