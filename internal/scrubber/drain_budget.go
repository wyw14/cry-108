package scrubber

import (
	"math"
	"sync"
)

type DrainBudget struct {
	mu       sync.Mutex
	capacity float64
	active   map[string]float64
}

func NewDrainBudget(capacity float64) *DrainBudget {
	return &DrainBudget{capacity: capacity, active: make(map[string]float64)}
}

func (b *DrainBudget) Grant(incidentID string, requested float64) float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	used := 0.0
	for _, value := range b.active {
		used += value
	}
	available := math.Max(0, b.capacity-used)
	granted := math.Min(math.Max(0, requested), available)
	b.active[incidentID] += granted
	return granted
}

func (b *DrainBudget) Release(incidentID string, volume float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	remaining := b.active[incidentID] - math.Max(0, volume)
	if remaining <= 0 {
		delete(b.active, incidentID)
		return
	}
	b.active[incidentID] = remaining
}

func (b *DrainBudget) Available() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	used := 0.0
	for _, value := range b.active {
		used += value
	}
	return math.Max(0, b.capacity-used)
}

func (b *DrainBudget) Snapshot() map[string]float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	result := make(map[string]float64, len(b.active))
	for incidentID, volume := range b.active {
		result[incidentID] = volume
	}
	return result
}
