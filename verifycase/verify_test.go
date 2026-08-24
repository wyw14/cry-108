package verifycase

import (
	"context"
	"testing"

	"github.com/wyw14/fabchem/internal/filter"
	"github.com/wyw14/fabchem/internal/interlock"
	"github.com/wyw14/fabchem/internal/meter"
	"github.com/wyw14/fabchem/internal/model"
)

func TestFilterReleaseRequiresWettingAndDebubble(t *testing.T) {
	ctx := context.Background()
	permits := interlock.NewPermitBook()
	service := filter.NewService(meter.NewService(model.NopRecorder{}), permits, model.NopRecorder{})
	session, err := service.Begin(ctx, "hf-filter", 12)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.ConfirmPressure(ctx, session.ID, 1.2, true); err != nil {
		t.Fatal(err)
	}
	if released, err := service.Release(ctx, session.ID); err == nil || released.Available() {
		t.Fatalf("pressure-only preparation was released: result=%+v err=%v", released, err)
	}
	if permits.Allowed("filter:hf-filter", "prepared") {
		t.Fatal("downstream permit opened without wetting and debubble proofs")
	}
}
