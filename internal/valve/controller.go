package valve

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/wyw14/fabchem/internal/model"
)

type State string

const (
	Closed State = "closed"
	Open   State = "open"
)

type Command struct {
	ValveID  string    `json:"valve_id"`
	State    State     `json:"state"`
	Session  string    `json:"session"`
	IssuedAt time.Time `json:"issued_at"`
}

type Controller struct {
	mu           sync.Mutex
	states       map[string]State
	reservations map[string]string
	routeClasses map[string]model.ChemicalClass
	commands     []Command
	recorder     model.Recorder
}

func NewController(recorder model.Recorder) *Controller {
	if recorder == nil {
		recorder = model.NopRecorder{}
	}
	return &Controller{
		states:       make(map[string]State),
		reservations: make(map[string]string),
		routeClasses: make(map[string]model.ChemicalClass),
		recorder:     recorder,
	}
}

func (c *Controller) command(ctx context.Context, valveID string, state State, session string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	now := time.Now().UTC()
	c.mu.Lock()
	c.states[valveID] = state
	c.commands = append(c.commands, Command{ValveID: valveID, State: state, Session: session, IssuedAt: now})
	c.mu.Unlock()
	return c.recorder.Append(ctx, model.Event{
		Kind: "valve.command", Aggregate: session, At: now,
		Data: map[string]any{"valve": valveID, "state": state},
	})
}

func (c *Controller) State(valveID string) State {
	c.mu.Lock()
	defer c.mu.Unlock()
	if value, ok := c.states[valveID]; ok {
		return value
	}
	return Closed
}

func (c *Controller) Commands() []Command {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := append([]Command(nil), c.commands...)
	sort.Slice(result, func(i, j int) bool { return result[i].IssuedAt.Before(result[j].IssuedAt) })
	return result
}

func (c *Controller) CloseAll(ctx context.Context, session string, valves []string) error {
	for _, valveID := range model.CanonicalValves(valves) {
		if err := c.command(ctx, valveID, Closed, session); err != nil {
			return err
		}
	}
	return nil
}
