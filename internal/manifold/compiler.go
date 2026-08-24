package manifold

import (
	"context"
	"fmt"

	"github.com/wyw14/fabchem/internal/interlock"
	"github.com/wyw14/fabchem/internal/model"
)

type Compiler struct {
	service *Service
	safety  *interlock.Compatibility
}

func NewCompiler(service *Service, safety *interlock.Compatibility) *Compiler {
	return &Compiler{service: service, safety: safety}
}

func (c *Compiler) CompileAndCommit(ctx context.Context, request RouteRequest) (model.Route, error) {
	route, err := NewRoute(request)
	if err != nil {
		return model.Route{}, err
	}
	if err := c.safety.CheckCandidate(route, c.service.Routes()); err != nil {
		return model.Route{}, err
	}
	if err := c.service.Commit(ctx, route); err != nil {
		return model.Route{}, fmt.Errorf("reserve complete route: %w", err)
	}
	route.Committed = true
	if err := c.safety.CheckCommitted(c.service.Routes()); err != nil {
		_ = c.service.Release(ctx, route.ID)
		return model.Route{}, err
	}
	return route, nil
}

func (c *Compiler) CompilePair(ctx context.Context, left, right RouteRequest) [2]error {
	start := make(chan struct{})
	errors := make(chan struct {
		index int
		err   error
	}, 2)
	requests := []RouteRequest{left, right}
	for index := range requests {
		go func(index int) {
			<-start
			_, err := c.CompileAndCommit(ctx, requests[index])
			errors <- struct {
				index int
				err   error
			}{index: index, err: err}
		}(index)
	}
	close(start)
	var result [2]error
	for range requests {
		item := <-errors
		result[item.index] = item.err
	}
	return result
}
