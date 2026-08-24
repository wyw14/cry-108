package meter

import (
	"context"
	"fmt"
	"time"

	"github.com/wyw14/fabchem/internal/model"
)

type Settlement struct {
	DispenseID string    `json:"dispense_id"`
	Volume     float64   `json:"volume"`
	Closed     bool      `json:"closed"`
	Settled    bool      `json:"settled"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (s *Service) BeginSettlement(dispenseID string) error {
	if dispenseID == "" {
		return fmt.Errorf("dispense identity is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.settlements[dispenseID]; exists {
		return nil
	}
	s.settlements[dispenseID] = &Settlement{DispenseID: dispenseID, UpdatedAt: time.Now().UTC()}
	return nil
}

func (s *Service) AddDispensed(ctx context.Context, dispenseID string, volume float64) error {
	if volume < 0 {
		return fmt.Errorf("dispensed volume cannot be negative")
	}
	if err := s.BeginSettlement(dispenseID); err != nil {
		return err
	}
	s.mu.Lock()
	settlement := s.settlements[dispenseID]
	settlement.Volume += volume
	settlement.UpdatedAt = time.Now().UTC()
	total := settlement.Volume
	s.mu.Unlock()
	return s.record(ctx, model.NewEvent("dispense.meter", dispenseID, map[string]any{"increment": volume, "total": total}))
}

func (s *Service) CloseAndSettle(ctx context.Context, dispenseID string, tailDelay time.Duration, tailVolume float64) (Settlement, error) {
	if err := s.BeginSettlement(dispenseID); err != nil {
		return Settlement{}, err
	}
	s.mu.Lock()
	s.settlements[dispenseID].Closed = true
	s.mu.Unlock()
	if tailDelay > 0 {
		timer := time.NewTimer(tailDelay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return Settlement{}, ctx.Err()
		case <-timer.C:
		}
	}
	if tailVolume > 0 {
		if err := s.AddDispensed(ctx, dispenseID, tailVolume); err != nil {
			return Settlement{}, err
		}
	}
	s.mu.Lock()
	settlement := s.settlements[dispenseID]
	settlement.Settled = true
	settlement.UpdatedAt = time.Now().UTC()
	result := *settlement
	s.mu.Unlock()
	err := s.record(ctx, model.NewEvent("dispense.settled", dispenseID, map[string]any{"volume": result.Volume}))
	return result, err
}

func (s *Service) FinalVolume(dispenseID string) (Settlement, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	settlement, ok := s.settlements[dispenseID]
	if !ok {
		return Settlement{}, false
	}
	return *settlement, true
}
