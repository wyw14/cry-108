package manifold

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/wyw14/fabchem/internal/model"
)

type RouteRequest struct {
	Class  model.ChemicalClass `json:"class"`
	Source string              `json:"source"`
	Target string              `json:"target"`
	Mode   string              `json:"mode"`
}

func (r RouteRequest) Validate() error {
	if r.Source == "" || r.Target == "" {
		return fmt.Errorf("route source and target are required")
	}
	switch r.Class {
	case model.ClassAcid, model.ClassBase, model.ClassSolvent, model.ClassWater:
	default:
		return fmt.Errorf("route compatibility class is invalid")
	}
	switch r.Mode {
	case "supply", "return", "flush":
		return nil
	default:
		return fmt.Errorf("route mode must be supply, return, or flush")
	}
}

func routeValves(request RouteRequest) []string {
	normalize := func(value string) string {
		return strings.ToLower(strings.ReplaceAll(value, " ", "-"))
	}
	values := []string{"source:" + normalize(request.Source), "bridge:central", "target:" + normalize(request.Target)}
	if request.Mode == "return" {
		values = append(values, "return:header")
	}
	if request.Mode == "flush" {
		values = append(values, "waste:flush")
	}
	return model.CanonicalValves(values)
}

func NewRoute(request RouteRequest) (model.Route, error) {
	if err := request.Validate(); err != nil {
		return model.Route{}, err
	}
	return model.Route{
		ID: uuid.NewString(), Class: request.Class, Source: request.Source,
		Target: request.Target, Valves: routeValves(request),
	}, nil
}
