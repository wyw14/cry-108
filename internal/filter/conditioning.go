package filter

import (
	"context"
	"fmt"
	"time"

	"github.com/wyw14/fabchem/internal/model"
)

func (s *Service) ConfirmPressure(ctx context.Context, sessionID string, differential float64, stable bool) error {
	if differential < 0 {
		return fmt.Errorf("filter differential pressure cannot be negative")
	}
	s.mu.Lock()
	session := s.sessions[sessionID]
	if session == nil {
		s.mu.Unlock()
		return fmt.Errorf("filter preparation %s does not exist", sessionID)
	}
	session.Pressure = model.Proof{SessionID: sessionID, Kind: "differential-pressure", Value: differential, Valid: stable}
	s.mu.Unlock()
	return s.recorder.Append(ctx, model.NewEvent("filter.pressure", sessionID, map[string]any{"differential": differential, "stable": stable}))
}

func (s *Service) AddWetting(ctx context.Context, sessionID string, liters float64) error {
	proof, err := s.meters.AddWettingVolume(ctx, sessionID, liters)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	session := s.sessions[sessionID]
	if session == nil {
		return fmt.Errorf("filter preparation %s does not exist", sessionID)
	}
	proof.Valid = proof.Value >= session.WettingTarget
	session.Wetting = proof
	return nil
}

func (s *Service) ConfirmDebubble(ctx context.Context, sessionID string, noBubbles bool) error {
	s.mu.Lock()
	session := s.sessions[sessionID]
	if session == nil {
		s.mu.Unlock()
		return fmt.Errorf("filter preparation %s does not exist", sessionID)
	}
	session.Debubble = model.Proof{SessionID: sessionID, Kind: "no-bubbles", Value: 1, Valid: noBubbles}
	s.mu.Unlock()
	return s.recorder.Append(ctx, model.NewEvent("filter.debubble", sessionID, map[string]any{"clear": noBubbles}))
}

func (s *Service) Release(ctx context.Context, sessionID string) (Session, error) {
	s.mu.Lock()
	session := s.sessions[sessionID]
	if session == nil {
		s.mu.Unlock()
		return Session{}, fmt.Errorf("filter preparation %s does not exist", sessionID)
	}
	if !session.Pressure.Valid {
		s.mu.Unlock()
		return *session, fmt.Errorf("filter pressure is not stable")
	}
	session.Phase = Ready
	session.ReadyAt = time.Now().UTC()
	result := *session
	s.mu.Unlock()
	if err := s.permits.Set("filter:"+result.FilterID, "prepared", true); err != nil {
		return Session{}, err
	}
	err := s.recorder.Append(ctx, model.NewEvent("filter.available", sessionID, map[string]any{"filter": result.FilterID}))
	return result, err
}
