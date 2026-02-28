package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/munchies/platform/backend/internal/config"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/middleware"
	analyticsmod "github.com/munchies/platform/backend/internal/modules/analytics"
	authmod "github.com/munchies/platform/backend/internal/modules/auth"
	catalogmod "github.com/munchies/platform/backend/internal/modules/catalog"
	contentmod "github.com/munchies/platform/backend/internal/modules/content"
	deliverymod "github.com/munchies/platform/backend/internal/modules/delivery"
	financemod "github.com/munchies/platform/backend/internal/modules/finance"
	hubmod "github.com/munchies/platform/backend/internal/modules/hub"
	inventorymod "github.com/munchies/platform/backend/internal/modules/inventory"
	issuemod "github.com/munchies/platform/backend/internal/modules/issue"
	mediamod "github.com/munchies/platform/backend/internal/modules/media"
	ordermod "github.com/munchies/platform/backend/internal/modules/order"
	paymentmod "github.com/munchies/platform/backend/internal/modules/payment"
	promomod "github.com/munchies/platform/backend/internal/modules/promo"
	ratingmod "github.com/munchies/platform/backend/internal/modules/rating"
	restaurantmod "github.com/munchies/platform/backend/internal/modules/restaurant"
	ridermod "github.com/munchies/platform/backend/internal/modules/rider"
	searchmod "github.com/munchies/platform/backend/internal/modules/search"
	ssemod "github.com/munchies/platform/backend/internal/modules/sse"
	storefrontmod "github.com/munchies/platform/backend/internal/modules/storefront"
	tenantmod "github.com/munchies/platform/backend/internal/modules/tenant"
	usermod "github.com/munchies/platform/backend/internal/modules/user"
	workermod "github.com/munchies/platform/backend/internal/modules/worker"
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
	Pool    *pgxpool.Pool
	Redis   *redisclient.Client
	SMS     sms.Sender
}

// Server holds the HTTP router and dependencies.
type Server struct {
	router            chi.Router
	cfg               *config.Config
	reconciliationJob *paymentmod.ReconciliationJob
	worker            *workermod.Worker
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

	// Inventory module
	inventorySvc := inventorymod.NewService(deps.Queries)
	inventoryHandler := inventorymod.NewHandler(inventorySvc)

	// Promo module
	promoSvc := promomod.NewService(deps.Queries)
	promoHandler := promomod.NewHandler(promoSvc)

	// Order module
	orderSvc := ordermod.NewService(deps.Queries, deps.Pool, inventorySvc, promoSvc)
	orderHandler := ordermod.NewHandler(orderSvc)

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

	// Finance module
	financeSvc := financemod.NewService(deps.Queries)
	financeHandler := financemod.NewHandler(financeSvc)

	// Issue module
	issueSvc := issuemod.NewService(deps.Queries)
	issueHandler := issuemod.NewHandler(issueSvc)

	// Rating module
	ratingSvc := ratingmod.NewService(deps.Queries)
	ratingHandler := ratingmod.NewHandler(ratingSvc)

	// Search module
	searchSvc := searchmod.NewService(deps.Queries)
	searchHandler := searchmod.NewHandler(searchSvc)

	// Content module
	contentSvc := contentmod.NewService(deps.Queries)
	contentHandler := contentmod.NewHandler(contentSvc)

	// Analytics module
	analyticsSvc := analyticsmod.NewService(deps.Queries)
	analyticsHandler := analyticsmod.NewHandler(analyticsSvc)

	// SSE module
	sseHandler := ssemod.NewHandler(deps.Redis)

	// Background worker
	s.worker = workermod.NewWorker(deps.Queries, deps.Redis)

	// Ledger service (seed platform accounts)
	ledgerSvc := financemod.NewLedgerService(deps.Queries)
	_ = ledgerSvc

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
				r.Get("/orders", orderHandler.ListMyOrders)
			})

			// Order issues (customer create)
			r.Post("/orders/{id}/issue", issueHandler.CreateIssue)

			// Ratings (customer rate delivered order)
			r.Post("/orders/{id}/rate", ratingHandler.RateOrder)

			// Customer order endpoints
			r.Route("/orders", func(r chi.Router) {
				r.Post("/charges/calculate", orderHandler.CalculateCharges)
				r.Post("/", orderHandler.CreateOrder)
				r.Get("/{id}", orderHandler.GetOrder)
				r.Get("/{id}/tracking", orderHandler.TrackOrder)
				r.Patch("/{id}/cancel", orderHandler.CancelOrder)
			})

			// SSE events
			r.Get("/events/subscribe", sseHandler.Subscribe)
		})

		// Public storefront routes
		r.Get("/storefront/config", storefrontHandler.GetConfig)
		r.Get("/storefront/areas", storefrontHandler.ListAreas)
		r.Get("/storefront/restaurants", storefrontHandler.ListRestaurants)
		r.Get("/storefront/banners", contentHandler.StorefrontBanners)
		r.Get("/storefront/stories", contentHandler.StorefrontStories)
		r.Get("/storefront/sections", contentHandler.StorefrontSections)
		r.Get("/restaurants/{slug}", storefrontHandler.GetRestaurant)
		r.Get("/restaurants/{id}/ratings", ratingHandler.ListReviews)
		r.Get("/products/{id}", storefrontHandler.GetProduct)

		// Search routes (public)
		r.Get("/search", searchHandler.Search)
		r.Get("/search/autocomplete", searchHandler.Autocomplete)

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

			// Order flow
			r.Get("/orders/active", riderHandler.ListActiveOrders)
			r.Patch("/orders/{id}/accept", riderHandler.AcceptOrder)
			r.Patch("/orders/{id}/picked/{restaurant_id}", riderHandler.MarkPickupPicked)
			r.Patch("/orders/{id}/delivered", riderHandler.MarkDelivered)
			r.Patch("/orders/{id}/issue", riderHandler.ReportIssue)

			// Earnings & history
			r.Get("/earnings", riderHandler.ListEarnings)
			r.Get("/history", riderHandler.ListDeliveryHistory)

			// Order module rider routes
			r.Route("/orders", func(r chi.Router) {
				r.Patch("/{id}/picked/{restaurantID}", orderHandler.PickedByRider)
			})
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

		// Order rider assignment
		r.Post("/orders/{id}/assign-rider", riderHandler.ManualAssignRider)

		// Rider management
		r.Get("/riders", riderHandler.ListRiders)
		r.Post("/riders", riderHandler.CreateRider)
		r.Get("/riders/attendance", riderHandler.ListAttendance)
		r.Get("/riders/tracking", riderHandler.ListRiderTracking)
		r.Get("/riders/{id}", riderHandler.GetRider)
		r.Put("/riders/{id}", riderHandler.UpdateRider)
		r.Delete("/riders/{id}", riderHandler.DeleteRider)
		r.Get("/riders/{id}/travel-log", riderHandler.GetTravelLog)
		r.Get("/riders/{id}/penalties", riderHandler.ListPenalties)
		r.Post("/riders/{id}/penalties", riderHandler.CreatePenalty)
		r.Patch("/riders/{id}/penalties/{penalty_id}", riderHandler.UpdatePenalty)

		// Finance management (partner)
		r.Get("/finance/summary", financeHandler.GetSummary)
		r.Get("/finance/invoices", financeHandler.ListInvoices)
		r.Get("/finance/invoices/{id}", financeHandler.GetInvoice)
		r.Get("/finance/invoices/{id}/pdf", financeHandler.GetInvoicePDF)

		// Issue management (partner)
		r.Get("/issues", issueHandler.ListIssues)
		r.Get("/issues/{id}", issueHandler.GetIssue)
		r.Post("/issues/{id}/message", issueHandler.AddMessage)
		r.Get("/issues/{id}/messages", issueHandler.ListMessages)

		// Review management (partner respond)
		r.Post("/reviews/{id}/respond", ratingHandler.RespondToReview)

		// Content management (partner)
		r.Get("/content/banners", contentHandler.ListBanners)
		r.Post("/content/banners", contentHandler.CreateBanner)
		r.Put("/content/banners/{id}", contentHandler.UpdateBanner)
		r.Delete("/content/banners/{id}", contentHandler.DeleteBanner)
		r.Get("/content/stories", contentHandler.ListStories)
		r.Post("/content/stories", contentHandler.CreateStory)
		r.Delete("/content/stories/{id}", contentHandler.DeleteStory)
		r.Get("/content/sections", contentHandler.ListSections)
		r.Put("/content/sections/{id}", contentHandler.UpdateSection)

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

		// Dashboard & analytics (partner)
		r.Get("/dashboard", analyticsHandler.GetDashboard)
		r.Get("/reports/sales", analyticsHandler.GetSalesReport)
		r.Get("/reports/products/top-selling", analyticsHandler.GetTopProducts)
		r.Get("/reports/orders/breakdown", analyticsHandler.GetOrderBreakdown)
		r.Get("/reports/peak-hours", analyticsHandler.GetPeakHours)
		r.Get("/reports/riders", analyticsHandler.GetRiderAnalytics)
		r.Get("/reports/searches", searchHandler.TopSearchTerms)
	})

	// Admin routes (authenticated, admin role required)
	r.Route("/admin", func(r chi.Router) {
		r.Use(tenantmod.NewResolver(tenantRepo, deps.Redis).Middleware)
		r.Use(authMiddleware.Authenticate)
		r.Use(authmod.RequireRoles(sqlc.UserRolePlatformAdmin, sqlc.UserRolePlatformFinance))

		// Invoice management (admin)
		r.Post("/finance/invoices/generate", financeHandler.GenerateInvoice)
		r.Patch("/finance/invoices/{id}/finalize", financeHandler.FinalizeInvoice)
		r.Patch("/finance/invoices/{id}/mark-paid", financeHandler.MarkInvoicePaid)

		// Issue resolution (admin)
		r.Patch("/issues/{id}/resolve", issueHandler.ResolveIssue)
		r.Patch("/issues/{id}/refund/approve", issueHandler.ApproveRefund)
		r.Patch("/issues/{id}/refund/reject", issueHandler.RejectRefund)

		// Order management (admin)
		r.Route("/orders", func(r chi.Router) {
			r.Patch("/{id}/force-cancel", orderHandler.ForceCancelOrder)
		})

		// Analytics (admin — cross-tenant)
		r.Get("/analytics/overview", analyticsHandler.AdminOverview)
		r.Get("/analytics/revenue", analyticsHandler.AdminRevenue)
		r.Get("/analytics/orders", analyticsHandler.AdminOrderVolume)
		r.Get("/analytics/tenants/{id}", analyticsHandler.AdminTenantAnalytics)
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
	if s.worker != nil {
		go s.worker.Start(ctx)
	}
}

