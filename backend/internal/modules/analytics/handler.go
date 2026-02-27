package analytics

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/munchies/platform/backend/internal/modules/tenant"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/respond"
)

// Handler handles analytics HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new analytics handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GetDashboard handles GET /partner/dashboard
func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	dashboard, err := h.svc.GetDashboard(r.Context(), t.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, dashboard)
}

// GetSalesReport handles GET /partner/reports/sales
func (h *Handler) GetSalesReport(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	startDate, endDate := parseDateRange(r)
	report, err := h.svc.GetSalesReport(r.Context(), t.ID, startDate, endDate)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, report)
}

// GetTopProducts handles GET /partner/reports/products/top-selling
func (h *Handler) GetTopProducts(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	startDate, endDate := parseDateRange(r)
	products, err := h.svc.GetTopProducts(r.Context(), t.ID, startDate, endDate, 20)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, products)
}

// GetOrderBreakdown handles GET /partner/reports/orders/breakdown
func (h *Handler) GetOrderBreakdown(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	startDate, endDate := parseDateRange(r)
	breakdown, err := h.svc.GetOrderBreakdown(r.Context(), t.ID, startDate, endDate)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, breakdown)
}

// GetPeakHours handles GET /partner/reports/peak-hours
func (h *Handler) GetPeakHours(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	startDate, endDate := parseDateRange(r)
	hours, err := h.svc.GetPeakHours(r.Context(), t.ID, startDate, endDate)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, hours)
}

// GetRiderAnalytics handles GET /partner/reports/riders
func (h *Handler) GetRiderAnalytics(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	startDate, endDate := parseDateRange(r)
	riders, err := h.svc.GetRiderAnalytics(r.Context(), t.ID, startDate, endDate)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, riders)
}

// --- Admin Analytics ---

// AdminOverview handles GET /admin/analytics/overview
func (h *Handler) AdminOverview(w http.ResponseWriter, r *http.Request) {
	startDate, endDate := parseDateRange(r)
	overview, err := h.svc.GetAdminOverview(r.Context(), startDate, endDate)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, overview)
}

// AdminRevenue handles GET /admin/analytics/revenue
func (h *Handler) AdminRevenue(w http.ResponseWriter, r *http.Request) {
	startDate, endDate := parseDateRange(r)
	revenue, err := h.svc.GetAdminRevenue(r.Context(), startDate, endDate)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, revenue)
}

// AdminOrderVolume handles GET /admin/analytics/orders
func (h *Handler) AdminOrderVolume(w http.ResponseWriter, r *http.Request) {
	startDate, endDate := parseDateRange(r)
	volume, err := h.svc.GetAdminOrderVolume(r.Context(), startDate, endDate)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, volume)
}

// AdminTenantAnalytics handles GET /admin/analytics/tenants/:id
func (h *Handler) AdminTenantAnalytics(w http.ResponseWriter, r *http.Request) {
	tenantID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid tenant id"))
		return
	}
	startDate, endDate := parseDateRange(r)
	analytics, err := h.svc.GetTenantAnalytics(r.Context(), tenantID, startDate, endDate)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, analytics)
}

func parseDateRange(r *http.Request) (time.Time, time.Time) {
	q := r.URL.Query()
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if s := q.Get("start_date"); s != "" {
		if parsed, err := time.Parse("2006-01-02", s); err == nil {
			startDate = parsed
		}
	}
	if e := q.Get("end_date"); e != "" {
		if parsed, err := time.Parse("2006-01-02", e); err == nil {
			endDate = parsed
		}
	}
	return startDate, endDate
}

func toAppError(err error) *apperror.AppError {
	if e, ok := err.(*apperror.AppError); ok {
		return e
	}
	return apperror.Internal("unexpected error", err)
}
