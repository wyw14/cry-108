package dispense

import (
	"fmt"

	"github.com/wyw14/fabchem/internal/interlock"
)

type PermitService struct {
	book    *interlock.PermitBook
	exhaust *interlock.ExhaustPermit
}

func NewPermitService(book *interlock.PermitBook, exhaust *interlock.ExhaustPermit) *PermitService {
	return &PermitService{book: book, exhaust: exhaust}
}

func (p *PermitService) Allow(blendID, filterID string) error {
	if blendID == "" || filterID == "" {
		return fmt.Errorf("blend and filter identities are required")
	}
	if !p.book.Allowed("blend:"+blendID, "quality") {
		return fmt.Errorf("blend quality release is missing")
	}
	if !p.book.Allowed("filter:"+filterID, "prepared") {
		return fmt.Errorf("filter preparation is incomplete")
	}
	if !p.exhaust.Available() {
		return fmt.Errorf("facility exhaust is unavailable")
	}
	return nil
}
