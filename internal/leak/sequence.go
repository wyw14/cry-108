package leak

import (
	"context"
	"fmt"
	"time"

	"github.com/wyw14/fabchem/internal/interlock"
	"github.com/wyw14/fabchem/internal/model"
	"github.com/wyw14/fabchem/internal/scrubber"
)

type Sequence struct {
	detector *Detector
	isolate  *interlock.Isolator
	scrubber *scrubber.Service
	recorder model.Recorder
}

func NewSequence(detector *Detector, isolate *interlock.Isolator, scrubberService *scrubber.Service, recorder model.Recorder) *Sequence {
	if recorder == nil {
		recorder = model.NopRecorder{}
	}
	return &Sequence{detector: detector, isolate: isolate, scrubber: scrubberService, recorder: recorder}
}

type ResponseOptions struct {
	IsolationDelay time.Duration
	IsolationFails bool
	ResidualVolume float64
}

func (s *Sequence) Contain(ctx context.Context, incidentID string, options ResponseOptions) (Incident, error) {
	incident, ok := s.detector.Incident(incidentID)
	if !ok {
		return Incident{}, fmt.Errorf("incident %s does not exist", incidentID)
	}
	incident.State = Isolating
	s.detector.Update(incident)
	granted, err := s.scrubber.Drain(ctx, incident.ID, options.ResidualVolume)
	if err != nil {
		incident.Alarm = err.Error()
		s.detector.Update(incident)
		return incident, err
	}
	proof, err := s.isolate.Isolate(ctx, incident.ID, "cabinet leak", options.IsolationDelay, options.IsolationFails)
	if err != nil || !proof.Confirmed {
		incident.State = Unsecured
		incident.Alarm = "upstream isolation failed"
		s.detector.Update(incident)
		return incident, fmt.Errorf("upstream isolation remains pending: %w", err)
	}
	incident.IsolationConfirmed = true
	incident.State = Draining
	s.detector.Update(incident)
	incident.DrainedVolume = granted
	if granted < options.ResidualVolume {
		incident.Alarm = "high VOC loading while controlled drain is active"
		s.detector.Update(incident)
		return incident, fmt.Errorf("scrubber capacity below requested residual volume")
	}
	incident.State = Contained
	incident.ContainedAt = time.Now().UTC()
	s.detector.Update(incident)
	s.scrubber.CompleteDrain(incident.ID, granted)
	err = s.recorder.Append(ctx, model.NewEvent("leak.contained", incident.ID, map[string]any{"drained": granted}))
	return incident, err
}
