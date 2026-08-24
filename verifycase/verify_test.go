package verifycase

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/fabchem/internal/interlock"
	"github.com/wyw14/fabchem/internal/leak"
	"github.com/wyw14/fabchem/internal/model"
	"github.com/wyw14/fabchem/internal/scrubber"
	"github.com/wyw14/fabchem/internal/valve"
)

func TestLeakDrainWaitsForUpstreamIsolation(t *testing.T) {
	ctx := context.Background()
	recorder := model.NopRecorder{}
	detector := leak.NewDetector(recorder)
	valves := valve.NewController(recorder)
	scrubberService := scrubber.NewService(20, interlock.NewExhaustPermit(), recorder)
	sequence := leak.NewSequence(detector, interlock.NewIsolator(valves), scrubberService, recorder)
	incident, err := detector.Detect(ctx, "solvent-cabinet", "solvent", 2)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		_, _ = sequence.Contain(ctx, incident.ID, leak.ResponseOptions{IsolationDelay: 80 * time.Millisecond, ResidualVolume: 5})
		close(done)
	}()
	time.Sleep(20 * time.Millisecond)
	if len(scrubberService.DrainSnapshot()) != 0 {
		t.Fatal("accident drain opened while upstream isolation remained pending")
	}
	<-done
}
