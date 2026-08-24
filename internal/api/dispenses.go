package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/wyw14/fabchem/internal/dispense"
)

func (s *Server) listDispenses(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{"items": s.runtime.Dispenses.List()})
}

func (s *Server) createDispense(writer http.ResponseWriter, request *http.Request) {
	var payload dispense.Request
	if err := readJSON(request, &payload); err != nil {
		s.problem(writer, http.StatusBadRequest, err)
		return
	}
	session, err := s.runtime.Dispenses.Create(request.Context(), payload)
	if err != nil {
		s.problem(writer, http.StatusUnprocessableEntity, err)
		return
	}
	if err := s.runtime.Dispenses.Start(request.Context(), session.ID); err != nil {
		s.problem(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusCreated, session)
}

type closePayload struct {
	TailDelayMillis int     `json:"tail_delay_millis"`
	TailVolume      float64 `json:"tail_volume"`
}

func (s *Server) closeDispense(writer http.ResponseWriter, request *http.Request) {
	var payload closePayload
	if err := readJSON(request, &payload); err != nil {
		s.problem(writer, http.StatusBadRequest, err)
		return
	}
	session, err := s.runtime.Dispenses.Close(request.Context(), chi.URLParam(request, "id"), time.Duration(payload.TailDelayMillis)*time.Millisecond, payload.TailVolume)
	if err != nil {
		s.problem(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusOK, session)
}
