package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/munchies/platform/backend/internal/config"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/middleware"
	authmod "github.com/munchies/platform/backend/internal/modules/auth"
	tenantmod "github.com/munchies/platform/backend/internal/modules/tenant"
	usermod "github.com/munchies/platform/backend/internal/modules/user"
	redisclient "github.com/munchies/platform/backend/internal/platform/redis"
	"github.com/munchies/platform/backend/internal/platform/sms"
	"github.com/rs/zerolog/log"
)

// Deps holds all external dependencies required by the server.
type Deps struct {
	Queries *sqlc.Queries
	Redis   *redisclient.Client
	SMS     sms.Sender
}

// Server holds the HTTP router and dependencies.
type Server struct {
	router chi.Router
	cfg    *config.Config
}

// New creates a new Server with all routes and middleware configured.
func New(cfg *config.Config, deps Deps) *Server {
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
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-Tenant-ID", "Idempotency-Key"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	s := &Server{
		router: r,
		cfg:    cfg,
	}

	s.registerRoutes(deps)

	return s
}

// Router returns the chi.Router for use with http.Server.
func (s *Server) Router() chi.Router {
	return s.router
}

func (s *Server) registerRoutes(deps Deps) {
	r := s.router

	// Health check endpoints (no auth required)
	r.Get("/healthz", s.handleHealthz)
	r.Get("/readyz", s.handleReadyz)

	// Build module dependencies
	tokenCfg := authmod.TokenConfig{
		AccessSecret:  s.cfg.JWT.AccessTokenSecret,
		RefreshSecret: s.cfg.JWT.RefreshTokenSecret,
		AccessExpiry:  s.cfg.JWT.AccessTokenExpiry,
		RefreshExpiry: s.cfg.JWT.RefreshTokenExpiry,
	}

	tenantRepo := tenantmod.NewRepository(deps.Queries)
	tenantResolver := tenantmod.NewResolver(tenantRepo, deps.Redis)

	authSvc := authmod.NewService(deps.Queries, deps.Redis, deps.SMS, tokenCfg)
	authHandler := authmod.NewHandler(authSvc, tokenCfg)
	authMiddleware := authmod.NewAuthMiddleware(deps.Queries, tokenCfg)

	userRepo := usermod.NewRepository(deps.Queries)
	userSvc := usermod.NewService(userRepo)
	userHandler := usermod.NewHandler(userSvc)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		log.Debug().Msg("API v1 routes registered")

		// Tenant resolution (optional — proceeds without tenant if not found)
		r.Use(tenantResolver.Middleware)

		// Auth routes (no authentication required)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/otp/send", authHandler.SendOTP)
			r.Post("/otp/verify", authHandler.VerifyOTP)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/logout", authHandler.Logout)
			r.Post("/password/reset-request", authHandler.RequestPasswordReset)
			r.Post("/password/reset", authHandler.ResetPassword)
		})

		// Authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Use(middleware.Idempotency(deps.Queries))

			// /me endpoints
			r.Route("/me", func(r chi.Router) {
				r.Get("/", userHandler.GetMe)
				r.Patch("/", userHandler.UpdateMe)
				r.Get("/addresses", userHandler.ListAddresses)
				r.Post("/addresses", userHandler.CreateAddress)
				r.Delete("/addresses/{id}", userHandler.DeleteAddress)
				r.Get("/wallet", userHandler.ListWallet)
				r.Get("/notifications", userHandler.ListNotifications)
				r.Patch("/notifications/{id}/read", userHandler.MarkNotificationRead)
			})
		})
	})
}

// handleHealthz is a liveness probe — always returns 200 if the process is running.
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"alive"}`))
}

// handleReadyz is a readiness probe — returns 200 when the service is ready to accept traffic.
func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ready"}`))
}

