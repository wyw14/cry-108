package nitrogen

import (
	"context"
	"fmt"

	"github.com/wyw14/fabchem/internal/source"
)

type PressureSample struct {
	TankID   string  `json:"tank_id"`
	Pressure float64 `json:"pressure"`
	Demand   float64 `json:"demand"`
}

func (s *Service) Observe(samples []PressureSample) error {
	for _, sample := range samples {
		if sample.TankID == "" || sample.Pressure < 0 || sample.Demand < 0 {
			return fmt.Errorf("invalid day tank pressure sample")
		}
		if err := s.source.UpdatePressure(sample.TankID, sample.Pressure); err != nil {
			return err
		}
		if err := s.source.RequestNitrogen(sample.TankID, sample.Demand); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Protect(ctx context.Context, samples []PressureSample) (Plan, error) {
	if err := s.Observe(samples); err != nil {
		return Plan{}, err
	}
	return s.Allocate(ctx)
}

func PlanProtects(tanks []source.DayTank, plan Plan) bool {
	minimums := make(map[string]float64, len(tanks))
	for _, tank := range tanks {
		minimums[tank.ID] = tank.MinimumPressure
	}
	for _, allocation := range plan.Allocations {
		if allocation.Predicted < minimums[allocation.TankID] && allocation.Flow > 0 {
			return false
		}
	}
	return true
}
