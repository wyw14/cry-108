package meter

import (
	"context"
	"sync"

	"github.com/wyw14/fabchem/internal/model"
)

type Service struct {
	mu            sync.Mutex
	flows         map[string]FlowProof
	pulses        []Pulse
	wetting       map[string]float64
	flowTotals    map[string]float64
	densities     map[string]float64
	latestFlow    string
	latestDensity string
	settlements   map[string]*Settlement
	recorder      model.Recorder
}

func NewService(recorder model.Recorder) *Service {
	if recorder == nil {
		recorder = model.NopRecorder{}
	}
	return &Service{
		flows:       make(map[string]FlowProof),
		wetting:     make(map[string]float64),
		flowTotals:  make(map[string]float64),
		densities:   make(map[string]float64),
		settlements: make(map[string]*Settlement),
		recorder:    recorder,
	}
}

func (s *Service) record(ctx context.Context, event model.Event) error {
	return s.recorder.Append(ctx, event)
}

func (s *Service) PulseHistory() []Pulse {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Pulse(nil), s.pulses...)
}
