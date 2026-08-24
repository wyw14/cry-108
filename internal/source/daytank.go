package source

import (
	"fmt"
	"sort"
)

type DayTank struct {
	ID              string  `json:"id"`
	Chemical        string  `json:"chemical"`
	Hazard          int     `json:"hazard"`
	MinimumPressure float64 `json:"minimum_pressure"`
	Pressure        float64 `json:"pressure"`
	Demand          float64 `json:"demand"`
}

func (t DayTank) Validate() error {
	if t.ID == "" || t.Chemical == "" {
		return fmt.Errorf("day tank identity and chemical are required")
	}
	if t.Hazard < 1 || t.Hazard > 5 || t.MinimumPressure <= 0 {
		return fmt.Errorf("day tank requires valid hazard and minimum pressure")
	}
	return nil
}

func (s *Service) UpdatePressure(tankID string, pressure float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tank := s.tanks[tankID]
	if tank == nil {
		return fmt.Errorf("day tank %s does not exist", tankID)
	}
	tank.Pressure = pressure
	return nil
}

func (s *Service) RequestNitrogen(tankID string, demand float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tank := s.tanks[tankID]
	if tank == nil || demand < 0 {
		return fmt.Errorf("invalid day tank nitrogen request")
	}
	tank.Demand = demand
	return nil
}

func (s *Service) DayTanks() []DayTank {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]DayTank, 0, len(s.tanks))
	for _, tank := range s.tanks {
		result = append(result, *tank)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Hazard == result[j].Hazard {
			return result[i].ID < result[j].ID
		}
		return result[i].Hazard > result[j].Hazard
	})
	return result
}
