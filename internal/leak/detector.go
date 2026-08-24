package leak

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/fabchem/internal/model"
)

type Detector struct {
	mu       sync.Mutex
	active   map[string]Incident
	recorder model.Recorder
}

func NewDetector(recorder model.Recorder) *Detector {
	if recorder == nil {
		recorder = model.NopRecorder{}
	}
	return &Detector{active: make(map[string]Incident), recorder: recorder}
}

func (d *Detector) Detect(ctx context.Context, cabinet, chemical string, rate float64) (Incident, error) {
	if cabinet == "" || chemical == "" || rate <= 0 {
		return Incident{}, fmt.Errorf("leak detection requires cabinet, chemical, and positive rate")
	}
	incident := Incident{ID: uuid.NewString(), Cabinet: cabinet, Chemical: chemical, Rate: rate, State: Detected, StartedAt: time.Now().UTC()}
	d.mu.Lock()
	d.active[incident.ID] = incident
	d.mu.Unlock()
	err := d.recorder.Append(ctx, model.NewEvent("leak.detected", incident.ID, map[string]any{"cabinet": cabinet, "chemical": chemical, "rate": rate}))
	return incident, err
}

func (d *Detector) Update(incident Incident) {
	d.mu.Lock()
	d.active[incident.ID] = incident
	d.mu.Unlock()
}

func (d *Detector) Incident(id string) (Incident, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	incident, ok := d.active[id]
	return incident, ok
}

func (d *Detector) List() []Incident {
	d.mu.Lock()
	defer d.mu.Unlock()
	result := make([]Incident, 0, len(d.active))
	for _, incident := range d.active {
		result = append(result, incident)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].StartedAt.Before(result[j].StartedAt) })
	return result
}
