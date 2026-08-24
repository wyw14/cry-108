package api

import (
	"context"
	"fmt"

	"github.com/wyw14/fabchem/internal/blend"
	"github.com/wyw14/fabchem/internal/dispense"
	"github.com/wyw14/fabchem/internal/filter"
	"github.com/wyw14/fabchem/internal/interlock"
	"github.com/wyw14/fabchem/internal/journal"
	"github.com/wyw14/fabchem/internal/leak"
	"github.com/wyw14/fabchem/internal/manifold"
	"github.com/wyw14/fabchem/internal/meter"
	"github.com/wyw14/fabchem/internal/model"
	"github.com/wyw14/fabchem/internal/nitrogen"
	"github.com/wyw14/fabchem/internal/returnline"
	"github.com/wyw14/fabchem/internal/scrubber"
	"github.com/wyw14/fabchem/internal/source"
	"github.com/wyw14/fabchem/internal/valve"
)

type Runtime struct {
	Journal        *journal.Store
	Valves         *valve.Controller
	Meters         *meter.Service
	Sources        *source.Service
	Blends         *blend.Service
	Quality        *blend.QualityReducer
	Manifolds      *manifold.Service
	Routes         *manifold.Compiler
	Returns        *returnline.Service
	ReturnRouter   *manifold.ReturnRouter
	Filters        *filter.Service
	Nitrogen       *nitrogen.Service
	Dispenses      *dispense.Service
	Releases       *dispense.ReleaseService
	DispensePermit *dispense.PermitService
	Leaks          *leak.Detector
	LeakSequence   *leak.Sequence
	Scrubber       *scrubber.Service
	Exhaust        *interlock.ExhaustPermit
	Permits        *interlock.PermitBook
}

func NewRuntime(dataDir string) (*Runtime, error) {
	store, err := journal.Open(dataDir)
	if err != nil {
		return nil, err
	}
	valves := valve.NewController(store)
	meters := meter.NewService(store)
	sources := source.NewService(store)
	permits := interlock.NewPermitBook()
	exhaust := interlock.NewExhaustPermit()
	manifolds := manifold.NewService(valves, store)
	routes := manifold.NewCompiler(manifolds, interlock.NewCompatibility())
	returns := returnline.NewService(store)
	scrubberService := scrubber.NewService(25, exhaust, store)
	leaks := leak.NewDetector(store)
	runtime := &Runtime{
		Journal: store, Valves: valves, Meters: meters, Sources: sources,
		Blends: blend.NewService(valves, meters, store), Quality: blend.NewQualityReducer(meters, store),
		Manifolds: manifolds, Routes: routes, Returns: returns,
		ReturnRouter: manifold.NewReturnRouter(routes), Filters: filter.NewService(meters, permits, store),
		Nitrogen:  nitrogen.NewService(18, sources, valves, store),
		Dispenses: dispense.NewService(valves, meters, permits, store), Releases: dispense.NewReleaseService(permits),
		Leaks: leaks, Scrubber: scrubberService, Exhaust: exhaust, Permits: permits,
	}
	runtime.DispensePermit = dispense.NewPermitService(permits, exhaust)
	runtime.LeakSequence = leak.NewSequence(leaks, interlock.NewIsolator(valves), scrubberService, store)
	if err := runtime.seed(); err != nil {
		return nil, err
	}
	return runtime, nil
}

func (r *Runtime) seed() error {
	catalog := model.Catalog()
	branches := []source.Branch{
		{ID: "acid-a", Chemical: catalog["hydrochloric"], RouteID: "route-acid-a", PurgeTarget: 5},
		{ID: "acid-b", Chemical: catalog["hydrochloric"], RouteID: "route-acid-b", PurgeTarget: 5},
	}
	for _, branch := range branches {
		if err := r.Sources.AddBranch(branch); err != nil {
			return err
		}
	}
	tanks := []source.DayTank{
		{ID: "nitric-far", Chemical: "nitric", Hazard: 5, MinimumPressure: 4, Pressure: 4, Demand: 8},
		{ID: "acid-mid", Chemical: "hydrochloric", Hazard: 4, MinimumPressure: 3, Pressure: 4, Demand: 6},
		{ID: "solvent-near", Chemical: "solvent", Hazard: 3, MinimumPressure: 2, Pressure: 4, Demand: 4},
	}
	for _, tank := range tanks {
		if err := r.Sources.AddDayTank(tank); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runtime) Snapshot(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	state := map[string]any{
		"blends": r.Blends.List(), "routes": r.Manifolds.Routes(),
		"dispenses": r.Dispenses.List(), "incidents": r.Leaks.List(),
	}
	if err := r.Journal.SaveSnapshot(state); err != nil {
		return fmt.Errorf("save runtime snapshot: %w", err)
	}
	return nil
}
