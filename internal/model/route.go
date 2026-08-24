package model

import (
	"fmt"
	"sort"
)

type Route struct {
	ID        string        `json:"id"`
	Class     ChemicalClass `json:"class"`
	Valves    []string      `json:"valves"`
	Source    string        `json:"source"`
	Target    string        `json:"target"`
	Committed bool          `json:"committed"`
}

func (r Route) Validate() error {
	if r.ID == "" || r.Source == "" || r.Target == "" {
		return fmt.Errorf("route identity, source, and target are required")
	}
	if len(r.Valves) == 0 {
		return fmt.Errorf("route requires at least one valve")
	}
	seen := make(map[string]bool, len(r.Valves))
	for _, valve := range r.Valves {
		if valve == "" || seen[valve] {
			return fmt.Errorf("route contains an empty or duplicate valve")
		}
		seen[valve] = true
	}
	return nil
}

func CanonicalValves(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
