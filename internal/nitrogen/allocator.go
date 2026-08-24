package nitrogen

import (
	"fmt"
	"math"
	"sort"

	"github.com/wyw14/fabchem/internal/source"
)

type Allocation struct {
	TankID    string  `json:"tank_id"`
	Flow      float64 `json:"flow"`
	Predicted float64 `json:"predicted_pressure"`
	Protected bool    `json:"protected"`
	Reason    string  `json:"reason"`
}

type Plan struct {
	ID             string       `json:"id"`
	Capacity       float64      `json:"capacity"`
	TotalAllocated float64      `json:"total_allocated"`
	Allocations    []Allocation `json:"allocations"`
}

func requiredFlow(tank source.DayTank) float64 {
	deficit := math.Max(0, tank.MinimumPressure-tank.Pressure)
	minimumProtection := deficit * 2
	if minimumProtection > tank.Demand {
		return tank.Demand
	}
	return minimumProtection
}

func BuildPlan(id string, capacity float64, tanks []source.DayTank) (Plan, error) {
	if id == "" || capacity <= 0 {
		return Plan{}, fmt.Errorf("nitrogen plan requires identity and positive capacity")
	}
	ordered := append([]source.DayTank(nil), tanks...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Hazard == ordered[j].Hazard {
			return ordered[i].Pressure < ordered[j].Pressure
		}
		return ordered[i].Hazard > ordered[j].Hazard
	})
	plan := Plan{ID: id, Capacity: capacity, Allocations: make([]Allocation, 0, len(ordered))}
	remaining := capacity
	for _, tank := range ordered {
		minimum := requiredFlow(tank)
		flow := math.Min(tank.Demand, remaining)
		if flow < minimum {
			flow = math.Min(minimum, remaining)
		}
		remaining -= flow
		predicted := tank.Pressure + flow/2
		protected := predicted >= tank.MinimumPressure || tank.Demand == 0
		reason := "capacity available"
		if flow < tank.Demand {
			reason = "capacity reduced by safety priority"
		}
		if !protected {
			reason = "header capacity cannot protect minimum pressure"
		}
		plan.Allocations = append(plan.Allocations, Allocation{TankID: tank.ID, Flow: flow, Predicted: predicted, Protected: protected, Reason: reason})
		plan.TotalAllocated += flow
	}
	return plan, nil
}
