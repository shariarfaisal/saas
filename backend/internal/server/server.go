package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/munchies/platform/backend/internal/config"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/middleware"
	authmod "github.com/munchies/platform/backend/internal/modules/auth"
	catalogmod "github.com/munchies/platform/backend/internal/modules/catalog"
	deliverymod "github.com/munchies/platform/backend/internal/modules/delivery"
	hubmod "github.com/munchies/platform/backend/internal/modules/hub"
	inventorymod "github.com/munchies/platform/backend/internal/modules/inventory"
	mediamod "github.com/munchies/platform/backend/internal/modules/media"
	ordermod "github.com/munchies/platform/backend/internal/modules/order"
	promomod "github.com/munchies/platform/backend/internal/modules/promo"
	restaurantmod "github.com/munchies/platform/backend/internal/modules/restaurant"
	storefrontmod "github.com/munchies/platform/backend/internal/modules/storefront"
	tenantmod "github.com/munchies/platform/backend/internal/modules/tenant"
	usermod "github.com/munchies/platform/backend/internal/modules/user"
	redisclient "github.com/munchies/platform/backend/internal/platform/redis"
	"github.com/munchies/platform/backend/internal/platform/sms"
	"github.com/rs/zerolog/log"
)

// Deps holds all external dependencies required by the server.
type Deps struct {
	Queries *sqlc.Queries
	Pool    *pgxpool.Pool
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

	// Hub module
	hubRepo := hubmod.NewRepository(deps.Queries)
	hubSvc := hubmod.NewService(hubRepo)
	hubHandler := hubmod.NewHandler(hubSvc)

	// Restaurant module
	restaurantRepo := restaurantmod.NewRepository(deps.Queries)
	restaurantSvc := restaurantmod.NewService(restaurantRepo)
	restaurantHandler := restaurantmod.NewHandler(restaurantSvc)

	// Catalog module
	catalogRepo := catalogmod.NewRepository(deps.Queries)
	catalogSvc := catalogmod.NewService(catalogRepo)
	catalogHandler := catalogmod.NewHandler(catalogSvc)

	// Storefront module
	storefrontSvc := storefrontmod.NewService(deps.Queries)
	storefrontHandler := storefrontmod.NewHandler(storefrontSvc)

	// Delivery module
	deliverySvc := deliverymod.NewService(deps.Queries)
	deliveryHandler := deliverymod.NewHandler(deliverySvc)

	// Media module
	mediaHandler := mediamod.NewHandler()

	// Inventory module
	inventorySvc := inventorymod.NewService(deps.Queries)
	inventoryHandler := inventorymod.NewHandler(inventorySvc)

	// Promo module
	promoSvc := promomod.NewService(deps.Queries)
	promoHandler := promomod.NewHandler(promoSvc)

	// Order module
	orderSvc := ordermod.NewService(deps.Queries, deps.Pool, inventorySvc, promoSvc)
	orderHandler := ordermod.NewHandler(orderSvc)

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

		// Public storefront routes
		r.Get("/storefront/config", storefrontHandler.GetConfig)
		r.Get("/storefront/areas", storefrontHandler.ListAreas)
		r.Get("/storefront/restaurants", storefrontHandler.ListRestaurants)
		r.Get("/restaurants/{slug}", storefrontHandler.GetRestaurant)
		r.Get("/products/{id}", storefrontHandler.GetProduct)

		// Delivery charge calculation (public)
		r.Post("/delivery/charges", deliveryHandler.CalculateCharge)

		// Media upload
		r.Post("/media/upload", mediaHandler.Upload)

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
				r.Get("/orders", orderHandler.ListMyOrders)
			})

			// Order endpoints (customer)
			r.Route("/orders", func(r chi.Router) {
				r.Post("/charges/calculate", orderHandler.CalculateCharges)
				r.Post("/", orderHandler.CreateOrder)
				r.Get("/{id}", orderHandler.GetOrder)
				r.Get("/{id}/tracking", orderHandler.TrackOrder)
				r.Patch("/{id}/cancel", orderHandler.CancelOrder)
			})

			// Partner endpoints
			r.Route("/partner", func(r chi.Router) {
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

				// Inventory management
				r.Route("/inventory", func(r chi.Router) {
					inventoryHandler.RegisterRoutes(r)
				})

				// Promo management
				r.Route("/promos", func(r chi.Router) {
					promoHandler.RegisterRoutes(r)
				})

				// Order management (partner)
				r.Route("/orders", func(r chi.Router) {
					r.Get("/", orderHandler.ListPartnerOrders)
					r.Patch("/{id}/confirm", orderHandler.ConfirmOrderPartner)
					r.Patch("/{id}/reject", orderHandler.RejectOrderPartner)
					r.Patch("/{id}/preparing", orderHandler.PreparingOrderPartner)
					r.Patch("/{id}/ready", orderHandler.ReadyOrderPartner)
				})
			})

			// Rider endpoints
			r.Route("/rider", func(r chi.Router) {
				r.Use(authmod.RequireRoles(sqlc.UserRoleRider))

				r.Route("/orders", func(r chi.Router) {
					r.Patch("/{id}/picked/{restaurantID}", orderHandler.PickedByRider)
				})
			})

			// Admin endpoints
			r.Route("/admin", func(r chi.Router) {
				r.Use(authmod.RequireRoles(
					sqlc.UserRolePlatformAdmin,
					sqlc.UserRolePlatformSupport,
				))

				r.Route("/orders", func(r chi.Router) {
					r.Patch("/{id}/force-cancel", orderHandler.ForceCancelOrder)
				})
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

