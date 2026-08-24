package verifycase

import (
	"context"
	"testing"

	"github.com/wyw14/fabchem/internal/interlock"
	"github.com/wyw14/fabchem/internal/model"
	"github.com/wyw14/fabchem/internal/scrubber"
)

func TestChemicalReleaseRequiresScrubberDraftProof(t *testing.T) {
	ctx := context.Background()
	exhaust := interlock.NewExhaustPermit()
	service := scrubber.NewService(20, exhaust, model.NopRecorder{})
	handover, err := service.BeginHandover(ctx, "fan-primary", "fan-standby")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.ApplyRPM(ctx, handover.ID, 1800, 1700); err != nil {
		t.Fatal(err)
	}
	if err := service.ApplyDraft(ctx, handover.ID, 0.1, 1.5); err != nil {
		t.Fatal(err)
	}
	completed, err := service.CompleteHandover(ctx, handover.ID)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Available || exhaust.Available() {
		t.Fatalf("chemical release opened on rpm without draft proof: %+v", completed)
	}
}
