package verifycase

import (
	"context"
	"testing"

	"github.com/wyw14/fabchem/internal/blend"
	"github.com/wyw14/fabchem/internal/meter"
	"github.com/wyw14/fabchem/internal/model"
)

func TestBlendReleaseUsesMatchingDensityAndFlow(t *testing.T) {
	ctx := context.Background()
	meters := meter.NewService(model.NopRecorder{})
	quality := blend.NewQualityReducer(meters, model.NopRecorder{})
	if err := meters.RecordFlowTotal(ctx, "C207", 100); err != nil {
		t.Fatal(err)
	}
	if err := meters.RecordDensity(ctx, "C207", 1.10); err != nil {
		t.Fatal(err)
	}
	if err := meters.RecordFlowTotal(ctx, "C208", 102); err != nil {
		t.Fatal(err)
	}
	result, err := quality.Evaluate(ctx, "C208", 1.10, 0.02)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Pending || result.Qualified {
		t.Fatalf("C208 combined C207 density with C208 flow: %+v", result)
	}
	if err := meters.RecordDensity(ctx, "C208", 1.15); err != nil {
		t.Fatal(err)
	}
	result, err = quality.Evaluate(ctx, "C208", 1.10, 0.02)
	if err != nil || result.Pending || result.Qualified || result.BlendID != "C208" {
		t.Fatalf("C208 did not evaluate its own measurements: result=%+v err=%v", result, err)
	}
}
