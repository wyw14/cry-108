package dispense

import (
	"fmt"
	"sync"

	"github.com/wyw14/fabchem/internal/blend"
	"github.com/wyw14/fabchem/internal/interlock"
)

type ReleaseService struct {
	mu       sync.Mutex
	released map[string]bool
	permits  *interlock.PermitBook
}

func NewReleaseService(permits *interlock.PermitBook) *ReleaseService {
	return &ReleaseService{released: make(map[string]bool), permits: permits}
}

func (s *ReleaseService) Apply(result blend.QualityResult) (bool, error) {
	if result.BlendID == "" || result.Pending {
		return false, fmt.Errorf("blend quality is not complete")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !result.Qualified {
		delete(s.released, result.BlendID)
		return false, nil
	}
	if err := s.permits.Set("blend:"+result.BlendID, "quality", true); err != nil {
		return false, err
	}
	s.released[result.BlendID] = true
	return true, nil
}

func (s *ReleaseService) Released(blendID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.released[blendID]
}
