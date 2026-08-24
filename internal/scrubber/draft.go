package scrubber

import (
	"fmt"
	"sync"
	"time"
)

type DraftSample struct {
	SessionID string    `json:"session_id"`
	Pressure  float64   `json:"pressure"`
	Observed  time.Time `json:"observed"`
}

type DraftObserver struct {
	mu      sync.Mutex
	samples map[string]DraftSample
}

func NewDraftObserver() *DraftObserver {
	return &DraftObserver{samples: make(map[string]DraftSample)}
}

func (o *DraftObserver) Update(sessionID string, pressure float64) (DraftSample, error) {
	if sessionID == "" || pressure < 0 {
		return DraftSample{}, fmt.Errorf("draft sample requires session and non-negative pressure")
	}
	sample := DraftSample{SessionID: sessionID, Pressure: pressure, Observed: time.Now().UTC()}
	o.mu.Lock()
	o.samples[sessionID] = sample
	o.mu.Unlock()
	return sample, nil
}

func (o *DraftObserver) Latest(sessionID string) (DraftSample, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	sample, ok := o.samples[sessionID]
	return sample, ok
}

func (o *DraftObserver) Valid(sessionID string, minimum float64) bool {
	sample, ok := o.Latest(sessionID)
	return ok && sample.Pressure >= minimum
}
