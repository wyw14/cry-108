package interlock

import (
	"fmt"
	"sync"

	"github.com/wyw14/fabchem/internal/model"
)

type ExhaustPermit struct {
	mu        sync.Mutex
	sessionID string
	rpm       model.Proof
	draft     model.Proof
	available bool
}

func NewExhaustPermit() *ExhaustPermit { return &ExhaustPermit{} }

func (p *ExhaustPermit) Evaluate(rpm, draft model.Proof) (bool, error) {
	if rpm.SessionID == "" || rpm.SessionID != draft.SessionID {
		return false, fmt.Errorf("fan rpm and draft proof must belong to the same handover")
	}
	if rpm.Kind != "fan-rpm" || draft.Kind != "draft-pressure" {
		return false, fmt.Errorf("exhaust permit requires rpm and draft proofs")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sessionID = rpm.SessionID
	p.rpm = rpm
	p.draft = draft
	p.available = rpm.Valid && draft.Valid
	return p.available, nil
}

func (p *ExhaustPermit) Available() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.available
}

func (p *ExhaustPermit) SessionID() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.sessionID
}
