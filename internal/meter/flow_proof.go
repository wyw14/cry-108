package meter

import (
	"context"
	"fmt"
	"time"

	"github.com/wyw14/fabchem/internal/model"
)

type FlowProof struct {
	BatchID      string    `json:"batch_id"`
	LitersMinute float64   `json:"liters_minute"`
	Established  time.Time `json:"established"`
}

func (s *Service) EstablishDilutionFlow(ctx context.Context, batchID string, delay time.Duration, rate float64) (FlowProof, error) {
	if batchID == "" || rate <= 0 {
		return FlowProof{}, fmt.Errorf("batch and positive dilution flow are required")
	}
	if delay > 0 {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return FlowProof{}, ctx.Err()
		case <-timer.C:
		}
	}
	proof := FlowProof{BatchID: batchID, LitersMinute: rate, Established: time.Now().UTC()}
	s.mu.Lock()
	s.flows[batchID] = proof
	s.mu.Unlock()
	err := s.record(ctx, model.NewEvent("meter.flow.established", batchID, map[string]any{"rate": rate}))
	return proof, err
}

func (s *Service) DilutionFlow(batchID string) (FlowProof, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	proof, ok := s.flows[batchID]
	return proof, ok
}
