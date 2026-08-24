package model

import (
	"context"
	"time"
)

type Event struct {
	ID        string         `json:"id"`
	Kind      string         `json:"kind"`
	Aggregate string         `json:"aggregate"`
	At        time.Time      `json:"at"`
	Data      map[string]any `json:"data,omitempty"`
}

type Recorder interface {
	Append(context.Context, Event) error
}

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now().UTC() }

type NopRecorder struct{}

func (NopRecorder) Append(context.Context, Event) error { return nil }

func NewEvent(kind, aggregate string, data map[string]any) Event {
	return Event{Kind: kind, Aggregate: aggregate, At: time.Now().UTC(), Data: data}
}
