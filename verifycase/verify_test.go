package verifycase

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/fabchem/internal/dispense"
	"github.com/wyw14/fabchem/internal/interlock"
	"github.com/wyw14/fabchem/internal/meter"
	"github.com/wyw14/fabchem/internal/model"
	"github.com/wyw14/fabchem/internal/valve"
)

func TestDispenseCompletesAfterMeterTailSettles(t *testing.T) {
	ctx := context.Background()
	meters := meter.NewService(model.NopRecorder{})
	service := dispense.NewService(valve.NewController(model.NopRecorder{}), meters, interlock.NewPermitBook(), model.NopRecorder{})
	session, err := service.Create(ctx, dispense.Request{ID: "dose-1", BlendID: "C208", ToolID: "wet-tool", TargetLiters: 2, Tolerance: 0.02})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Start(ctx, session.ID); err != nil {
		t.Fatal(err)
	}
	if err := service.AddMeasured(ctx, session.ID, 1.72); err != nil {
		t.Fatal(err)
	}
	closed, err := service.Close(ctx, session.ID, 50*time.Millisecond, 0.28)
	if err != nil {
		t.Fatal(err)
	}
	if closed.State != dispense.Delivered || closed.Actual != 2 {
		t.Fatalf("dispense finalized before meter tail settled: %+v", closed)
	}
	retry, err := service.Create(ctx, dispense.Request{ID: session.ID, BlendID: "C208", ToolID: "wet-tool", TargetLiters: 2, Tolerance: 0.02})
	if err != nil || retry.State != dispense.Delivered {
		t.Fatalf("idempotent completed session changed: result=%+v err=%v", retry, err)
	}
}
