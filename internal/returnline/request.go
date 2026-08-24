package returnline

import (
	"fmt"

	"github.com/wyw14/fabchem/internal/model"
)

type Disposition string

const (
	Product Disposition = "product"
	Flush   Disposition = "flush"
	Waste   Disposition = "waste"
)

type Request struct {
	SessionID   string              `json:"session_id"`
	RouteID     string              `json:"route_id"`
	Origin      string              `json:"origin"`
	Class       model.ChemicalClass `json:"class"`
	Disposition Disposition         `json:"disposition"`
	Volume      float64             `json:"volume"`
}

func NewReturnRequest(sessionID, routeID, origin string, class model.ChemicalClass, disposition Disposition, volume float64) (Request, error) {
	request := Request{SessionID: sessionID, RouteID: routeID, Origin: origin, Class: class, Disposition: disposition, Volume: volume}
	if err := request.Validate(); err != nil {
		return Request{}, err
	}
	return request, nil
}

func (r Request) Validate() error {
	if r.SessionID == "" || r.RouteID == "" || r.Origin == "" {
		return fmt.Errorf("return request requires session, route, and origin")
	}
	if r.Volume <= 0 {
		return fmt.Errorf("return volume must be positive")
	}
	switch r.Disposition {
	case Product, Flush, Waste:
		return nil
	default:
		return fmt.Errorf("return disposition %q is invalid", r.Disposition)
	}
}
