package verifycase

import (
	"context"
	"testing"

	"github.com/wyw14/fabchem/internal/dispense"
	"github.com/wyw14/fabchem/internal/model"
	"github.com/wyw14/fabchem/internal/returnline"
)

func TestAbortFlushReturnsOnlyToWastePath(t *testing.T) {
	returns := returnline.NewService(model.NopRecorder{})
	finalizer := dispense.NewAbortFinalizer(returns)
	started, err := finalizer.StartFlush(context.Background(), dispense.Session{ID: "dispense-1", ToolID: "wet-tool"}, "route-1", 3)
	if err != nil {
		t.Fatal(err)
	}
	if started.Request.Disposition != returnline.Flush || started.Request.Class != model.ClassWater || started.Target != "compatible-waste" {
		t.Fatalf("flush identity was lost during abort cleanup: %+v", started)
	}
	if len(returns.List()) != 1 || returns.List()[0].Target != "compatible-waste" {
		t.Fatalf("flush was not routed to waste: %+v", returns.List())
	}
}
