package filter

import (
	"fmt"
	"time"

	"github.com/wyw14/fabchem/internal/model"
)

type Phase string

const (
	Preparing Phase = "preparing"
	Ready     Phase = "ready"
	Rejected  Phase = "rejected"
)

type Session struct {
	ID            string      `json:"id"`
	FilterID      string      `json:"filter_id"`
	WettingTarget float64     `json:"wetting_target"`
	Pressure      model.Proof `json:"pressure"`
	Wetting       model.Proof `json:"wetting"`
	Debubble      model.Proof `json:"debubble"`
	Phase         Phase       `json:"phase"`
	StartedAt     time.Time   `json:"started_at"`
	ReadyAt       time.Time   `json:"ready_at,omitempty"`
}

func (s Session) ValidateProofs() error {
	if !model.SameSession(s.Pressure, s.Wetting, s.Debubble) {
		return fmt.Errorf("filter preparation proofs must belong to the same session")
	}
	if s.Pressure.Kind != "differential-pressure" || s.Wetting.Kind != "wetting" || s.Debubble.Kind != "no-bubbles" {
		return fmt.Errorf("filter preparation requires pressure, wetting, and no-bubbles proofs")
	}
	if s.Wetting.Value < s.WettingTarget {
		return fmt.Errorf("filter wetting volume is below target")
	}
	return nil
}

func (s Session) Available() bool { return s.Phase == Ready }
