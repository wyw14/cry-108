package valve

import (
	"context"
	"fmt"
	"sort"
)

type NitrogenSetting struct {
	TankID string  `json:"tank_id"`
	Flow   float64 `json:"flow"`
	Reason string  `json:"reason"`
}

func (c *Controller) ApplyNitrogen(ctx context.Context, session string, settings []NitrogenSetting) error {
	ordered := append([]NitrogenSetting(nil), settings...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].TankID < ordered[j].TankID })
	for _, setting := range ordered {
		if setting.TankID == "" || setting.Flow < 0 {
			return fmt.Errorf("invalid nitrogen branch setting")
		}
		state := Closed
		if setting.Flow > 0 {
			state = Open
		}
		if err := c.command(ctx, "nitrogen:"+setting.TankID, state, session); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) NitrogenOpen(tankID string) bool {
	return c.State("nitrogen:"+tankID) == Open
}
