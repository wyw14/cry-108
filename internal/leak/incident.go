package leak

import "time"

type State string

const (
	Detected  State = "detected"
	Isolating State = "isolating"
	Draining  State = "draining"
	Contained State = "contained"
	Unsecured State = "unsecured"
)

type Incident struct {
	ID                 string    `json:"id"`
	Cabinet            string    `json:"cabinet"`
	Chemical           string    `json:"chemical"`
	Rate               float64   `json:"rate"`
	State              State     `json:"state"`
	IsolationConfirmed bool      `json:"isolation_confirmed"`
	DrainedVolume      float64   `json:"drained_volume"`
	Alarm              string    `json:"alarm,omitempty"`
	StartedAt          time.Time `json:"started_at"`
	ContainedAt        time.Time `json:"contained_at,omitempty"`
}

func (i Incident) Safe() bool {
	return i.State == Contained && i.IsolationConfirmed && i.Alarm == ""
}

func (i Incident) Open() bool {
	return i.State != Contained && i.State != Unsecured
}

func (i Incident) Summary() map[string]any {
	return map[string]any{
		"id": i.ID, "cabinet": i.Cabinet, "chemical": i.Chemical,
		"state": i.State, "isolation_confirmed": i.IsolationConfirmed,
		"drained_volume": i.DrainedVolume, "alarm": i.Alarm,
	}
}
