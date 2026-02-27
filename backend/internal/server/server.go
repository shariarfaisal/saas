package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/munchies/platform/backend/internal/config"
	"github.com/munchies/platform/backend/internal/middleware"
	"github.com/rs/zerolog/log"
)

// Server holds the HTTP router and dependencies.
type Server struct {
	router chi.Router
	cfg    *config.Config
}

// New creates a new Server with all routes and middleware configured.
func New(cfg *config.Config) *Server {
	r := chi.NewRouter()

	// Global middleware stack (order matters)
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.StructuredLogger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.ContentTypeJSON)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.Server.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-Tenant-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	s := &Server{
		router: r,
		cfg:    cfg,
	}

	s.registerRoutes()

	return s
}

// Router returns the chi.Router for use with http.Server.
func (s *Server) Router() chi.Router {
	return s.router
}

func (s *Server) registerRoutes() {
	r := s.router

	// Health check endpoints (no auth required)
	r.Get("/healthz", s.handleHealthz)
	r.Get("/readyz", s.handleReadyz)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Future module routes will be mounted here
		log.Debug().Msg("API v1 routes registered")
	})
}

// handleHealthz is a liveness probe — always returns 200 if the process is running.
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"alive"}`))
}

// handleReadyz is a readiness probe — returns 200 when the service is ready to accept traffic.
func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	// TODO: Add database and redis connectivity checks
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready"}`))
}
