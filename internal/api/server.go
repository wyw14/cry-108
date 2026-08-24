package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	runtime *Runtime
	logger  *slog.Logger
	router  chi.Router
}

func NewServer(runtime *Runtime, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	server := &Server{runtime: runtime, logger: logger}
	server.router = server.routes()
	return server
}

func (s *Server) Handler() http.Handler { return s.router }

func (s *Server) routes() chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(20 * time.Second))
	router.Get("/healthz", s.health)
	router.Route("/api", func(api chi.Router) {
		api.Get("/blends", s.listBlends)
		api.Post("/blends", s.createBlend)
		api.Post("/blends/{id}/start", s.startBlend)
		api.Get("/manifolds", s.listRoutes)
		api.Post("/manifolds", s.createRoute)
		api.Post("/manifolds/pair", s.createRoutePair)
		api.Get("/dispenses", s.listDispenses)
		api.Post("/dispenses", s.createDispense)
		api.Post("/dispenses/{id}/close", s.closeDispense)
		api.Get("/incidents", s.listIncidents)
		api.Post("/incidents", s.createIncident)
	})
	return router
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func readJSON(request *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(nil, request.Body, 1024*1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if decoder.Decode(&struct{}{}) == nil {
		return errors.New("request body must contain one JSON value")
	}
	return nil
}

func (s *Server) problem(writer http.ResponseWriter, status int, err error) {
	s.logger.Warn("request rejected", "status", status, "error", err)
	writeJSON(writer, status, map[string]any{"error": err.Error(), "status": status})
}

func (s *Server) health(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{"status": "ok", "time": time.Now().UTC()})
}
