package interlock

import (
	"fmt"

	"github.com/wyw14/fabchem/internal/model"
)

type Compatibility struct{}

func NewCompatibility() *Compatibility { return &Compatibility{} }

func sharesValve(left, right model.Route) bool {
	seen := make(map[string]bool, len(left.Valves))
	for _, valveID := range left.Valves {
		seen[valveID] = true
	}
	for _, valveID := range right.Valves {
		if seen[valveID] {
			return true
		}
	}
	return false
}

func (c *Compatibility) CheckCandidate(candidate model.Route, committed []model.Route) error {
	if err := candidate.Validate(); err != nil {
		return err
	}
	for _, active := range committed {
		if sharesValve(candidate, active) && !model.Compatible(candidate.Class, active.Class) {
			return fmt.Errorf("incompatible media connected at bridge segment")
		}
	}
	return nil
}

func (c *Compatibility) CheckCommitted(routes []model.Route) error {
	for left := range routes {
		for right := left + 1; right < len(routes); right++ {
			if sharesValve(routes[left], routes[right]) && !model.Compatible(routes[left].Class, routes[right].Class) {
				return fmt.Errorf("incompatible media connected at bridge segment")
			}
		}
	}
	return nil
}
