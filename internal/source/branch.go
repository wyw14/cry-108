package source

import (
	"fmt"

	"github.com/wyw14/fabchem/internal/model"
)

type Branch struct {
	ID          string         `json:"id"`
	Chemical    model.Chemical `json:"chemical"`
	RouteID     string         `json:"route_id"`
	PurgeTarget float64        `json:"purge_target"`
	Purged      float64        `json:"purged"`
	Available   bool           `json:"available"`
}

func (b Branch) Validate() error {
	if b.ID == "" || b.RouteID == "" {
		return fmt.Errorf("source branch identity and route are required")
	}
	if b.PurgeTarget <= 0 {
		return fmt.Errorf("source branch purge target must be positive")
	}
	return b.Chemical.Validate()
}

func (b *Branch) AddPurge(volume float64) bool {
	if volume <= 0 {
		return b.Available
	}
	b.Purged += volume
	if b.Purged >= b.PurgeTarget {
		b.Available = true
	}
	return b.Available
}

func (b Branch) RemainingPurge() float64 {
	remaining := b.PurgeTarget - b.Purged
	if remaining < 0 {
		return 0
	}
	return remaining
}
