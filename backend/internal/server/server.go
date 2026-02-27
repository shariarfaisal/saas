package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/munchies/platform/backend/internal/config"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/middleware"
	authmod "github.com/munchies/platform/backend/internal/modules/auth"
	catalogmod "github.com/munchies/platform/backend/internal/modules/catalog"
	deliverymod "github.com/munchies/platform/backend/internal/modules/delivery"
	hubmod "github.com/munchies/platform/backend/internal/modules/hub"
	mediamod "github.com/munchies/platform/backend/internal/modules/media"
	paymentmod "github.com/munchies/platform/backend/internal/modules/payment"
	restaurantmod "github.com/munchies/platform/backend/internal/modules/restaurant"
	ridermod "github.com/munchies/platform/backend/internal/modules/rider"
	storefrontmod "github.com/munchies/platform/backend/internal/modules/storefront"
	tenantmod "github.com/munchies/platform/backend/internal/modules/tenant"
	usermod "github.com/munchies/platform/backend/internal/modules/user"
	gatewaypkg "github.com/munchies/platform/backend/internal/platform/payment"
	"github.com/munchies/platform/backend/internal/platform/payment/aamarpay"
	"github.com/munchies/platform/backend/internal/platform/payment/bkash"
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
	router           chi.Router
	cfg              *config.Config
	reconciliationJob *paymentmod.ReconciliationJob
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

	hubRepo := hubmod.NewRepository(deps.Queries)
	hubSvc := hubmod.NewService(hubRepo)
	hubHandler := hubmod.NewHandler(hubSvc)

	restaurantRepo := restaurantmod.NewRepository(deps.Queries)
	restaurantSvc := restaurantmod.NewService(restaurantRepo)
	restaurantHandler := restaurantmod.NewHandler(restaurantSvc)

	catalogRepo := catalogmod.NewRepository(deps.Queries)
	catalogSvc := catalogmod.NewService(catalogRepo)
	catalogHandler := catalogmod.NewHandler(catalogSvc)

	storefrontSvc := storefrontmod.NewService(deps.Queries)
	storefrontHandler := storefrontmod.NewHandler(storefrontSvc)

	deliverySvc := deliverymod.NewService(deps.Queries)
	deliveryHandler := deliverymod.NewHandler(deliverySvc)

	mediaHandler := mediamod.NewHandler()

	// Payment gateways
	paymentGateways := map[sqlc.PaymentMethod]gatewaypkg.Gateway{
		sqlc.PaymentMethodBkash: bkash.New(bkash.Config{
			AppKey:    s.cfg.Services.BkashAppKey,
			AppSecret: s.cfg.Services.BkashAppSecret,
			BaseURL:   s.cfg.Services.BkashBaseURL,
		}),
		sqlc.PaymentMethodAamarpay: aamarpay.New(aamarpay.Config{
			StoreID:      s.cfg.Services.AamarPayStoreID,
			SignatureKey: s.cfg.Services.AamarPayAPIKey,
			BaseURL:      s.cfg.Services.AamarPayBaseURL,
		}),
	}
	paymentSvc := paymentmod.NewService(deps.Queries, paymentGateways)
	callbackBaseURL := s.cfg.Server.PublicBaseURL
	if callbackBaseURL == "" {
		callbackBaseURL = fmt.Sprintf("http://localhost:%d", s.cfg.Server.Port)
	}
	paymentHandler := paymentmod.NewHandler(paymentSvc, callbackBaseURL)

	// Rider module
	riderSvc := ridermod.NewService(deps.Queries)
	riderHandler := ridermod.NewHandler(riderSvc)
	riderWSHandler := ridermod.NewWSHandler(deps.Queries, tokenCfg, deps.Redis)

	// Reconciliation job
	s.reconciliationJob = paymentmod.NewReconciliationJob(deps.Queries, paymentGateways)

	partnerRoles := authmod.RequireRoles(
		sqlc.UserRoleTenantOwner,
		sqlc.UserRoleTenantAdmin,
		sqlc.UserRoleRestaurantManager,
		sqlc.UserRoleRestaurantStaff,
	)

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

		// Public storefront routes
		r.Get("/storefront/config", storefrontHandler.GetConfig)
		r.Get("/storefront/areas", storefrontHandler.ListAreas)
		r.Get("/storefront/restaurants", storefrontHandler.ListRestaurants)
		r.Get("/restaurants/{slug}", storefrontHandler.GetRestaurant)
		r.Get("/products/{id}", storefrontHandler.GetProduct)

		// Delivery charge calculation (public)
		r.Post("/orders/charges/calculate", deliveryHandler.CalculateCharge)

		// Payment callbacks (no auth required — called by gateways)
		r.Get("/payments/bkash/callback", paymentHandler.BkashCallback)
		r.Post("/payments/aamarpay/success", paymentHandler.AamarpaySuccess)
		r.Post("/payments/aamarpay/fail", paymentHandler.AamarpayFail)
		r.Post("/payments/aamarpay/cancel", paymentHandler.AamarpayCancel)

		// Payment initiation (authenticated)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)

			r.Post("/payments/bkash/initiate", paymentHandler.InitiateBkash)
			r.Post("/payments/aamarpay/initiate", paymentHandler.InitiateAamarpay)
		})

		// Media upload
		r.Post("/media/upload", mediaHandler.Upload)

		// Rider-facing routes (authenticated, rider role)
		r.Route("/rider", func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)

			r.Post("/attendance/checkin", riderHandler.CheckIn)
			r.Post("/attendance/checkout", riderHandler.CheckOut)
			r.Patch("/availability", riderHandler.UpdateAvailability)
		})

		// Rider WebSocket (custom auth via query param)
		r.Get("/rider/ws", riderWSHandler.HandleWS)
	})

	// Partner routes (authenticated, role-restricted)
	r.Route("/partner", func(r chi.Router) {
		r.Use(tenantmod.NewResolver(tenantRepo, deps.Redis).Middleware)
		r.Use(authMiddleware.Authenticate)
		r.Use(partnerRoles)

		// Hub management
		r.Get("/hubs", hubHandler.ListHubs)
		r.Post("/hubs", hubHandler.CreateHub)
		r.Get("/hubs/{id}", hubHandler.GetHub)
		r.Put("/hubs/{id}", hubHandler.UpdateHub)
		r.Delete("/hubs/{id}", hubHandler.DeleteHub)
		r.Get("/hubs/{id}/areas", hubHandler.ListHubAreas)
		r.Post("/hubs/{id}/areas", hubHandler.CreateHubArea)
		r.Put("/hubs/{id}/areas/{area_id}", hubHandler.UpdateHubArea)
		r.Delete("/hubs/{id}/areas/{area_id}", hubHandler.DeleteHubArea)
		r.Get("/delivery/config", hubHandler.GetDeliveryZoneConfig)
		r.Put("/delivery/config", hubHandler.UpsertDeliveryZoneConfig)

		// Restaurant management
		r.Get("/restaurants", restaurantHandler.ListRestaurants)
		r.Post("/restaurants", restaurantHandler.CreateRestaurant)
		r.Get("/restaurants/{id}", restaurantHandler.GetRestaurant)
		r.Put("/restaurants/{id}", restaurantHandler.UpdateRestaurant)
		r.Delete("/restaurants/{id}", restaurantHandler.DeleteRestaurant)
		r.Patch("/restaurants/{id}/availability", restaurantHandler.UpdateAvailability)
		r.Get("/restaurants/{id}/hours", restaurantHandler.GetOperatingHours)
		r.Put("/restaurants/{id}/hours", restaurantHandler.UpsertOperatingHours)

		// Category management
		r.Get("/restaurants/{id}/categories", catalogHandler.ListCategories)
		r.Post("/restaurants/{id}/categories", catalogHandler.CreateCategory)
		r.Put("/restaurants/{id}/categories/{cat_id}", catalogHandler.UpdateCategory)
		r.Delete("/restaurants/{id}/categories/{cat_id}", catalogHandler.DeleteCategory)
		r.Patch("/restaurants/{id}/categories/reorder", catalogHandler.ReorderCategories)

		// Product management
		r.Get("/restaurants/{id}/products", catalogHandler.ListProducts)
		r.Post("/restaurants/{id}/products", catalogHandler.CreateProduct)
		r.Get("/products/{id}", catalogHandler.GetProduct)
		r.Put("/products/{id}", catalogHandler.UpdateProduct)
		r.Delete("/products/{id}", catalogHandler.DeleteProduct)
		r.Patch("/products/{id}/availability", catalogHandler.UpdateProductAvailability)
		r.Post("/products/{id}/discount", catalogHandler.UpsertDiscount)
		r.Delete("/products/{id}/discount", catalogHandler.DeactivateDiscount)
		r.Post("/products/bulk-upload", catalogHandler.BulkUpload)

		// Menu duplication
		r.Post("/restaurants/{id}/menu/duplicate", catalogHandler.DuplicateMenu)

		// Order refund
		r.Post("/orders/{id}/refund", paymentHandler.ProcessRefund)

		// Rider management
		r.Get("/riders", riderHandler.ListRiders)
		r.Post("/riders", riderHandler.CreateRider)
		r.Get("/riders/attendance", riderHandler.ListAttendance)
		r.Get("/riders/{id}", riderHandler.GetRider)
		r.Put("/riders/{id}", riderHandler.UpdateRider)
		r.Delete("/riders/{id}", riderHandler.DeleteRider)
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

// StartBackgroundJobs launches background goroutines such as payment reconciliation.
// The provided context controls the lifecycle of all background jobs.
func (s *Server) StartBackgroundJobs(ctx context.Context) {
	if s.reconciliationJob != nil {
		go s.reconciliationJob.StartReconciliation(ctx, 5*time.Minute)
	}
}

