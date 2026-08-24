package verifycase

import (
	"context"
	"testing"

	"github.com/wyw14/fabchem/internal/meter"
	"github.com/wyw14/fabchem/internal/model"
	"github.com/wyw14/fabchem/internal/source"
)

func TestSourceSwitchCountsPurgeOnPhysicalBranch(t *testing.T) {
	ctx := context.Background()
	sources := source.NewService(model.NopRecorder{})
	chemical := model.Catalog()["hydrochloric"]
	for _, branch := range []source.Branch{
		{ID: "A", RouteID: "route-A", Chemical: chemical, PurgeTarget: 5},
		{ID: "B", RouteID: "route-B", Chemical: chemical, PurgeTarget: 5},
	} {
		if err := sources.AddBranch(branch); err != nil {
			t.Fatal(err)
		}
	}
	meters := meter.NewService(model.NopRecorder{})
	oldPulse, err := meters.CapturePulse(ctx, "route-A", "A", 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sources.SwitchTo(ctx, "B"); err != nil {
		t.Fatal(err)
	}
	if _, err := sources.ApplyPurgePulse(ctx, oldPulse); err != nil {
		t.Fatal(err)
	}
	branchB, _ := sources.Branch("B")
	if branchB.Available || branchB.Purged != 0 {
		t.Fatalf("new branch consumed delayed pulse from old physical route: %+v", branchB)
	}
	branchA, _ := sources.Branch("A")
	if branchA.Purged != 5 {
		t.Fatalf("old branch did not retain its physical pulse: %+v", branchA)
	}
}
