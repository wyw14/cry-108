package meter

import (
	"context"
	"fmt"

	"github.com/wyw14/fabchem/internal/model"
)

func (s *Service) AddWettingVolume(ctx context.Context, sessionID string, liters float64) (model.Proof, error) {
	if sessionID == "" || liters <= 0 {
		return model.Proof{}, fmt.Errorf("wetting session and positive volume are required")
	}
	s.mu.Lock()
	s.wetting[sessionID] += liters
	total := s.wetting[sessionID]
	s.mu.Unlock()
	err := s.record(ctx, model.NewEvent("filter.wetting", sessionID, map[string]any{"liters": liters, "total": total}))
	return model.Proof{SessionID: sessionID, Kind: "wetting", Value: total, Valid: true}, err
}

func (s *Service) WettingProof(sessionID string, minimum float64) model.Proof {
	s.mu.Lock()
	defer s.mu.Unlock()
	value := s.wetting[sessionID]
	return model.Proof{SessionID: sessionID, Kind: "wetting", Value: value, Valid: value >= minimum}
}
