package verifycase

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/fabchem/internal/blend"
	"github.com/wyw14/fabchem/internal/meter"
	"github.com/wyw14/fabchem/internal/model"
	"github.com/wyw14/fabchem/internal/valve"
)

func TestAcidInjectionWaitsForDilutionWaterFlow(t *testing.T) {
	valves := valve.NewController(model.NopRecorder{})
	meters := meter.NewService(model.NopRecorder{})
	service := blend.NewService(valves, meters, model.NopRecorder{})
	batch, err := service.Create(context.Background(), blend.CreateRequest{Chemical: model.Catalog()["sulfuric"], TargetLiters: 100})
	if err != nil {
		t.Fatal(err)
	}
	started, err := service.Start(context.Background(), batch.ID, blend.StartOptions{WaterFeedbackDelay: 60 * time.Millisecond, MinimumWaterFlow: 8})
	if err != nil {
		t.Fatal(err)
	}
	if !started.ConcentrateStartedAfterWater() {
		t.Fatalf("acid started at %s before water flow at %s", started.AcidStartedAt, started.WaterFlowAt)
	}
	if started.AcidStartedAt.Sub(started.WaterAcceptedAt) < 50*time.Millisecond {
		t.Fatalf("acid command followed command acceptance instead of delayed physical flow")
	}
}
