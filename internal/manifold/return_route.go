package manifold

import (
	"context"
	"fmt"

	"github.com/wyw14/fabchem/internal/model"
	"github.com/wyw14/fabchem/internal/returnline"
)

type ReturnRouter struct {
	compiler *Compiler
}

func NewReturnRouter(compiler *Compiler) *ReturnRouter {
	return &ReturnRouter{compiler: compiler}
}

func (r *ReturnRouter) Resolve(ctx context.Context, request returnline.Request) (model.Route, error) {
	if err := request.Validate(); err != nil {
		return model.Route{}, err
	}
	target := "product-reclaim"
	mode := "return"
	if request.Disposition == returnline.Flush || request.Disposition == returnline.Waste {
		target = "compatible-waste"
		mode = "flush"
	}
	if request.Disposition == returnline.Product && request.Class == model.ClassWater {
		return model.Route{}, fmt.Errorf("water cannot use product reclaim disposition")
	}
	return r.compiler.CompileAndCommit(ctx, RouteRequest{
		Class: request.Class, Source: request.Origin, Target: target, Mode: mode,
	})
}
