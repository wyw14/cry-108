package returnline

import (
	"fmt"

	"github.com/wyw14/fabchem/internal/model"
)

type Destination struct {
	Name    string
	Classes []model.ChemicalClass
	Accepts []Disposition
}

func (d Destination) Allows(request Request) bool {
	classOK := false
	for _, class := range d.Classes {
		if model.Compatible(class, request.Class) {
			classOK = true
			break
		}
	}
	if !classOK {
		return false
	}
	for _, disposition := range d.Accepts {
		if disposition == request.Disposition {
			return true
		}
	}
	return false
}

func ResolveDestination(request Request, destinations []Destination) (Destination, error) {
	if err := request.Validate(); err != nil {
		return Destination{}, err
	}
	for _, destination := range destinations {
		if destination.Allows(request) {
			return destination, nil
		}
	}
	return Destination{}, fmt.Errorf("no return destination accepts %s/%s", request.Class, request.Disposition)
}

func DefaultDestinations() []Destination {
	return []Destination{
		{Name: "product-reclaim", Classes: []model.ChemicalClass{model.ClassAcid, model.ClassBase, model.ClassSolvent}, Accepts: []Disposition{Product}},
		{Name: "compatible-waste", Classes: []model.ChemicalClass{model.ClassAcid, model.ClassBase, model.ClassSolvent, model.ClassWater}, Accepts: []Disposition{Flush, Waste}},
	}
}
