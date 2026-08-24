package journal

import (
	"context"
	"fmt"

	"github.com/wyw14/fabchem/internal/model"
)

type Reducer func(model.Event) error

func (s *Store) Replay(ctx context.Context, reduce Reducer) (int, error) {
	if reduce == nil {
		return 0, fmt.Errorf("journal reducer is required")
	}
	events, err := s.ReadAll(ctx)
	if err != nil {
		return 0, err
	}
	for index, event := range events {
		if err := reduce(event); err != nil {
			return index, fmt.Errorf("replay event %s: %w", event.ID, err)
		}
	}
	return len(events), nil
}

func (s *Store) Directory() string { return s.dir }
