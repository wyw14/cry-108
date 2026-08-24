package api

import (
	"fmt"
	"net/http"

	"github.com/wyw14/fabchem/internal/manifold"
	"github.com/wyw14/fabchem/internal/model"
)

type routePayload struct {
	Class  model.ChemicalClass `json:"class"`
	Source string              `json:"source"`
	Target string              `json:"target"`
	Mode   string              `json:"mode"`
}

func (p routePayload) request() manifold.RouteRequest {
	return manifold.RouteRequest{Class: p.Class, Source: p.Source, Target: p.Target, Mode: p.Mode}
}

func (s *Server) listRoutes(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{"items": s.runtime.Manifolds.Routes()})
}

func (s *Server) createRoute(writer http.ResponseWriter, request *http.Request) {
	var payload routePayload
	if err := readJSON(request, &payload); err != nil {
		s.problem(writer, http.StatusBadRequest, err)
		return
	}
	route, err := s.runtime.Routes.CompileAndCommit(request.Context(), payload.request())
	if err != nil {
		s.problem(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusCreated, route)
}

type routePairPayload struct {
	Left  routePayload `json:"left"`
	Right routePayload `json:"right"`
}

func (s *Server) createRoutePair(writer http.ResponseWriter, request *http.Request) {
	var payload routePairPayload
	if err := readJSON(request, &payload); err != nil {
		s.problem(writer, http.StatusBadRequest, err)
		return
	}
	errors := s.runtime.Routes.CompilePair(request.Context(), payload.Left.request(), payload.Right.request())
	results := make([]map[string]any, 2)
	for index, err := range errors {
		results[index] = map[string]any{"success": err == nil}
		if err != nil {
			results[index]["error"] = err.Error()
		}
	}
	if (errors[0] == nil) == (errors[1] == nil) && !model.Compatible(payload.Left.Class, payload.Right.Class) {
		s.problem(writer, http.StatusInternalServerError, fmt.Errorf("incompatible media connected at bridge segment"))
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"results": results})
}
