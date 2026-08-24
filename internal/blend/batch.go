package blend

import (
	"fmt"
	"time"

	"github.com/wyw14/fabchem/internal/model"
)

type Batch struct {
	ID              string           `json:"id"`
	Chemical        model.Chemical   `json:"chemical"`
	TargetLiters    float64          `json:"target_liters"`
	State           model.BatchState `json:"state"`
	WaterAcceptedAt time.Time        `json:"water_accepted_at,omitempty"`
	WaterFlowAt     time.Time        `json:"water_flow_at,omitempty"`
	AcidStartedAt   time.Time        `json:"acid_started_at,omitempty"`
	Quality         QualityResult    `json:"quality"`
	CreatedAt       time.Time        `json:"created_at"`
}

type CreateRequest struct {
	Chemical     model.Chemical `json:"chemical"`
	TargetLiters float64        `json:"target_liters"`
}

func (r CreateRequest) Validate() error {
	if err := r.Chemical.Validate(); err != nil {
		return err
	}
	if r.TargetLiters <= 0 {
		return fmt.Errorf("blend target volume must be positive")
	}
	return nil
}

func (b Batch) ConcentrateStartedAfterWater() bool {
	return !b.WaterFlowAt.IsZero() && !b.AcidStartedAt.Before(b.WaterFlowAt)
}

func (b Batch) Complete() bool {
	return b.State == model.BatchClosed || b.State == model.BatchIsolated
}
