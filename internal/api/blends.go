package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/wyw14/fabchem/internal/blend"
	"github.com/wyw14/fabchem/internal/model"
)

type blendPayload struct {
	Chemical     string  `json:"chemical"`
	TargetLiters float64 `json:"target_liters"`
}

func (s *Server) listBlends(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{"items": s.runtime.Blends.List()})
}

func (s *Server) createBlend(writer http.ResponseWriter, request *http.Request) {
	var payload blendPayload
	if err := readJSON(request, &payload); err != nil {
		s.problem(writer, http.StatusBadRequest, err)
		return
	}
	chemical, ok := model.Catalog()[payload.Chemical]
	if !ok {
		s.problem(writer, http.StatusBadRequest, &unknownChemical{name: payload.Chemical})
		return
	}
	batch, err := s.runtime.Blends.Create(request.Context(), blend.CreateRequest{Chemical: chemical, TargetLiters: payload.TargetLiters})
	if err != nil {
		s.problem(writer, http.StatusUnprocessableEntity, err)
		return
	}
	writeJSON(writer, http.StatusCreated, batch)
}

type unknownChemical struct{ name string }

func (e *unknownChemical) Error() string { return "unknown chemical " + e.name }

type blendStartPayload struct {
	WaterFeedbackMillis int     `json:"water_feedback_millis"`
	MinimumWaterFlow    float64 `json:"minimum_water_flow"`
}

func (s *Server) startBlend(writer http.ResponseWriter, request *http.Request) {
	var payload blendStartPayload
	if err := readJSON(request, &payload); err != nil {
		s.problem(writer, http.StatusBadRequest, err)
		return
	}
	batch, err := s.runtime.Blends.Start(request.Context(), chi.URLParam(request, "id"), blend.StartOptions{
		WaterFeedbackDelay: time.Duration(payload.WaterFeedbackMillis) * time.Millisecond,
		MinimumWaterFlow:   payload.MinimumWaterFlow,
	})
	if err != nil {
		s.problem(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusOK, batch)
}
