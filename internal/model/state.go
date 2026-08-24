package model

import "fmt"

type BatchState string

const (
	BatchPrepared     BatchState = "prepared"
	BatchRouting      BatchState = "routing"
	BatchConditioning BatchState = "conditioning"
	BatchAvailable    BatchState = "available"
	BatchDispensing   BatchState = "dispensing"
	BatchReturning    BatchState = "returning"
	BatchClosed       BatchState = "closed"
	BatchIsolated     BatchState = "isolated"
)

var transitions = map[BatchState]map[BatchState]bool{
	BatchPrepared:     {BatchRouting: true, BatchIsolated: true},
	BatchRouting:      {BatchConditioning: true, BatchIsolated: true},
	BatchConditioning: {BatchAvailable: true, BatchIsolated: true},
	BatchAvailable:    {BatchDispensing: true, BatchIsolated: true},
	BatchDispensing:   {BatchReturning: true, BatchAvailable: true, BatchIsolated: true},
	BatchReturning:    {BatchClosed: true, BatchAvailable: true, BatchIsolated: true},
}

func CanTransition(from, to BatchState) bool {
	return transitions[from][to]
}

func RequireTransition(from, to BatchState) error {
	if !CanTransition(from, to) {
		return fmt.Errorf("invalid batch transition %s -> %s", from, to)
	}
	return nil
}

type Proof struct {
	SessionID string  `json:"session_id"`
	Kind      string  `json:"kind"`
	Value     float64 `json:"value"`
	Valid     bool    `json:"valid"`
}

func SameSession(proofs ...Proof) bool {
	if len(proofs) == 0 || proofs[0].SessionID == "" {
		return false
	}
	for _, proof := range proofs {
		if !proof.Valid || proof.SessionID != proofs[0].SessionID {
			return false
		}
	}
	return true
}
