package interlock

import (
	"fmt"
	"sync"
)

type PermitBook struct {
	mu      sync.Mutex
	permits map[string]map[string]bool
}

func NewPermitBook() *PermitBook {
	return &PermitBook{permits: make(map[string]map[string]bool)}
}

func (b *PermitBook) Set(resource, proof string, valid bool) error {
	if resource == "" || proof == "" {
		return fmt.Errorf("permit resource and proof are required")
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.permits[resource] == nil {
		b.permits[resource] = make(map[string]bool)
	}
	b.permits[resource][proof] = valid
	return nil
}

func (b *PermitBook) Allowed(resource string, required ...string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(required) == 0 {
		return false
	}
	for _, proof := range required {
		if !b.permits[resource][proof] {
			return false
		}
	}
	return true
}

func (b *PermitBook) Revoke(resource string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.permits, resource)
}
