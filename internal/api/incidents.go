package api

import "net/http"

type incidentPayload struct {
	Cabinet  string  `json:"cabinet"`
	Chemical string  `json:"chemical"`
	Rate     float64 `json:"rate"`
}

func (s *Server) listIncidents(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{
		"items":        s.runtime.Leaks.List(),
		"active_drain": s.runtime.Scrubber.DrainSnapshot(),
	})
}

func (s *Server) createIncident(writer http.ResponseWriter, request *http.Request) {
	var payload incidentPayload
	if err := readJSON(request, &payload); err != nil {
		s.problem(writer, http.StatusBadRequest, err)
		return
	}
	incident, err := s.runtime.Leaks.Detect(request.Context(), payload.Cabinet, payload.Chemical, payload.Rate)
	if err != nil {
		s.problem(writer, http.StatusUnprocessableEntity, err)
		return
	}
	writeJSON(writer, http.StatusCreated, incident)
}
