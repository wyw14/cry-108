package valve

import (
	"context"
	"fmt"

	"github.com/wyw14/fabchem/internal/model"
)

type ReservationConflict struct {
	ValveID string
	Owner   string
}

func (e ReservationConflict) Error() string {
	return fmt.Sprintf("valve %s is reserved by route %s", e.ValveID, e.Owner)
}

func (c *Controller) ReserveRoute(ctx context.Context, route model.Route) error {
	if err := route.Validate(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	valves := model.CanonicalValves(route.Valves)
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, valveID := range valves {
		if owner := c.reservations[valveID]; owner != "" && owner != route.ID {
			return ReservationConflict{ValveID: valveID, Owner: owner}
		}
	}
	for _, valveID := range valves {
		c.reservations[valveID] = route.ID
		c.states[valveID] = Open
	}
	c.routeClasses[route.ID] = route.Class
	c.commands = append(c.commands, Command{ValveID: "route:" + route.ID, State: Open, Session: route.ID})
	return nil
}

func (c *Controller) ReleaseRoute(ctx context.Context, routeID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for valveID, owner := range c.reservations {
		if owner == routeID {
			delete(c.reservations, valveID)
			c.states[valveID] = Closed
		}
	}
	delete(c.routeClasses, routeID)
	return nil
}

func (c *Controller) Reservations() map[string]string {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make(map[string]string, len(c.reservations))
	for valveID, owner := range c.reservations {
		result[valveID] = owner
	}
	return result
}

func (c *Controller) ReservedClass(routeID string) model.ChemicalClass {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.routeClasses[routeID]
}
