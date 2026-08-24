package valve

import (
	"context"
	"fmt"
	"time"
)

type WaterCommand struct {
	BatchID    string    `json:"batch_id"`
	AcceptedAt time.Time `json:"accepted_at"`
	ValveID    string    `json:"valve_id"`
}

func (c *Controller) OpenDilutionWater(ctx context.Context, batchID string) (WaterCommand, error) {
	if batchID == "" {
		return WaterCommand{}, fmt.Errorf("water command requires a batch")
	}
	accepted := time.Now().UTC()
	if err := c.command(ctx, "water-dilution", Open, batchID); err != nil {
		return WaterCommand{}, err
	}
	return WaterCommand{BatchID: batchID, AcceptedAt: accepted, ValveID: "water-dilution"}, nil
}

func (c *Controller) StartConcentrate(ctx context.Context, batchID string) (time.Time, error) {
	if batchID == "" {
		return time.Time{}, fmt.Errorf("concentrate command requires a batch")
	}
	if err := c.command(ctx, "concentrate-pump", Open, batchID); err != nil {
		return time.Time{}, err
	}
	return time.Now().UTC(), nil
}

func (c *Controller) StopBlend(ctx context.Context, batchID string) error {
	if err := c.command(ctx, "concentrate-pump", Closed, batchID); err != nil {
		return err
	}
	return c.command(ctx, "water-dilution", Closed, batchID)
}
