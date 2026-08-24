package valve

import (
	"context"
	"fmt"
	"time"
)

type CloseAck struct {
	DispenseID string    `json:"dispense_id"`
	ClosedAt   time.Time `json:"closed_at"`
}

func (c *Controller) OpenOutlet(ctx context.Context, dispenseID string) error {
	if dispenseID == "" {
		return fmt.Errorf("dispense identity is required")
	}
	return c.command(ctx, "outlet", Open, dispenseID)
}

func (c *Controller) CloseOutlet(ctx context.Context, dispenseID string) (CloseAck, error) {
	if dispenseID == "" {
		return CloseAck{}, fmt.Errorf("dispense identity is required")
	}
	if err := c.command(ctx, "outlet", Closed, dispenseID); err != nil {
		return CloseAck{}, err
	}
	return CloseAck{DispenseID: dispenseID, ClosedAt: time.Now().UTC()}, nil
}
