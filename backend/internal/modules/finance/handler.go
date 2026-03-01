package finance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/modules/auth"
	"github.com/munchies/platform/backend/internal/modules/tenant"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/pagination"
	"github.com/munchies/platform/backend/internal/pkg/respond"
	"github.com/shopspring/decimal"
)

// Handler handles finance HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new finance handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GetSummary handles GET /partner/finance/summary
func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	summary, err := h.svc.GetFinanceSummary(r.Context(), t.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, summary)
}

// ListInvoices handles GET /partner/finance/invoices
func (h *Handler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	page, perPage := parsePagination(r)
	items, meta, err := h.svc.ListByTenant(r.Context(), t.ID, page, perPage)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, pagination.PagedResponse{Data: items, Meta: meta})
}

// GetInvoice handles GET /partner/finance/invoices/:id
func (h *Handler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	invoiceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid invoice id"))
		return
	}
	inv, err := h.svc.GetByID(r.Context(), t.ID, invoiceID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, inv)
}

// GenerateInvoice handles POST /admin/finance/invoices/generate
func (h *Handler) GenerateInvoice(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		respond.Error(w, apperror.Unauthorized("authentication required"))
		return
	}
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}

	var req struct {
		RestaurantID string `json:"restaurant_id"`
		PeriodStart  string `json:"period_start"`
		PeriodEnd    string `json:"period_end"`
		Reason       string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	restaurantID, err := uuid.Parse(req.RestaurantID)
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid restaurant_id"))
		return
	}

	periodStart, err := time.Parse("2006-01-02", req.PeriodStart)
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid period_start format (YYYY-MM-DD)"))
		return
	}
	periodEnd, err := time.Parse("2006-01-02", req.PeriodEnd)
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid period_end format (YYYY-MM-DD)"))
		return
	}

	inv, err := h.svc.GenerateForRestaurant(r.Context(), t.ID, restaurantID, periodStart, periodEnd, &u.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	// Create audit log
	h.svc.CreateAuditLog(r.Context(), t.ID, u.ID, "invoice.generated", "invoice", inv.ID, req.Reason)

	respond.JSON(w, http.StatusCreated, inv)
}

// FinalizeInvoice handles PATCH /admin/finance/invoices/:id/finalize
func (h *Handler) FinalizeInvoice(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		respond.Error(w, apperror.Unauthorized("authentication required"))
		return
	}
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	invoiceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid invoice id"))
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	inv, err := h.svc.FinalizeInvoice(r.Context(), t.ID, invoiceID, u.ID, &req.Reason)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	h.svc.CreateAuditLog(r.Context(), t.ID, u.ID, "invoice.finalized", "invoice", inv.ID, req.Reason)

	respond.JSON(w, http.StatusOK, inv)
}

// MarkInvoicePaid handles PATCH /admin/finance/invoices/:id/mark-paid
func (h *Handler) MarkInvoicePaid(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		respond.Error(w, apperror.Unauthorized("authentication required"))
		return
	}
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	invoiceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid invoice id"))
		return
	}

	var req struct {
		PaymentReference string `json:"payment_reference"`
		Reason           string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	inv, err := h.svc.MarkPaid(r.Context(), t.ID, invoiceID, u.ID, req.PaymentReference, &req.Reason)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	h.svc.CreateAuditLog(r.Context(), t.ID, u.ID, "invoice.paid", "invoice", inv.ID, req.Reason)

	respond.JSON(w, http.StatusOK, inv)
}

// GetInvoicePDF handles GET /partner/finance/invoices/:id/pdf
func (h *Handler) GetInvoicePDF(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	invoiceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid invoice id"))
		return
	}

	inv, err := h.svc.GetByID(r.Context(), t.ID, invoiceID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	pdfBytes, err := GenerateInvoicePDF(inv)
	if err != nil {
		respond.Error(w, apperror.Internal("failed to generate PDF", err))
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+inv.InvoiceNumber+".pdf\"")
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes)
}

func parsePagination(r *http.Request) (page, perPage int) {
	q := r.URL.Query()
	page, _ = strconv.Atoi(q.Get("page"))
	perPage, _ = strconv.Atoi(q.Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = pagination.DefaultPageSize
	}
	return page, perPage
}

// CreateInvoiceAdjustment handles POST /admin/invoices/:id/adjustments
func (h *Handler) CreateInvoiceAdjustment(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		respond.Error(w, apperror.Unauthorized("authentication required"))
		return
	}
	invoiceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid invoice id"))
		return
	}
	var req struct {
		Amount    string `json:"amount"`
		Direction string `json:"direction"`
		Reason    string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	amt, err := decimalFromString(req.Amount)
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid amount"))
		return
	}
	adj, err := h.svc.CreateAdjustment(r.Context(), invoiceID, u.ID, amt, req.Direction, req.Reason)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusCreated, adj)
}

// ListInvoiceAdjustments handles GET /admin/invoices/:id/adjustments
func (h *Handler) ListInvoiceAdjustments(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid invoice id"))
		return
	}
	items, err := h.svc.ListAdjustments(r.Context(), invoiceID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, items)
}

// GetRevenueSummary handles GET /admin/revenue/summary
func (h *Handler) GetRevenueSummary(w http.ResponseWriter, r *http.Request) {
	report, err := h.svc.GetRevenueSummary(r.Context())
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, report)
}

// GetCommissionReport handles GET /admin/revenue/commission
func (h *Handler) GetCommissionReport(w http.ResponseWriter, r *http.Request) {
	from, to, err := parseDateRange(r)
	if err != nil {
		respond.Error(w, apperror.BadRequest(err.Error()))
		return
	}
	page, perPage := parsePagination(r)
	limit, offset := pagination.FormatLimitOffset(page, perPage)
	items, err := h.svc.q.ListCommissionByRestaurant(r.Context(), sqlc.ListCommissionByRestaurantParams{
		From:   from,
		To:     to,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, items)
}

// GetDeliveryFeeReport handles GET /admin/revenue/delivery-fees
func (h *Handler) GetDeliveryFeeReport(w http.ResponseWriter, r *http.Request) {
	from, to, err := parseDateRange(r)
	if err != nil {
		respond.Error(w, apperror.BadRequest(err.Error()))
		return
	}
	page, perPage := parsePagination(r)
	limit, offset := pagination.FormatLimitOffset(page, perPage)
	items, err := h.svc.q.ListDeliveryFeesByTenant(r.Context(), sqlc.ListDeliveryFeesByTenantParams{
		From:   from,
		To:     to,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, items)
}

// ListRiderPayouts handles GET /admin/rider-payouts
func (h *Handler) ListRiderPayouts(w http.ResponseWriter, r *http.Request) {
	page, perPage := parsePagination(r)
	items, meta, err := h.svc.ListRiderPayouts(r.Context(), page, perPage)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, pagination.PagedResponse{Data: items, Meta: meta})
}

// ApproveRiderPayout handles POST /admin/rider-payouts/:id/approve
func (h *Handler) ApproveRiderPayout(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		respond.Error(w, apperror.Unauthorized("authentication required"))
		return
	}
	payoutID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid payout id"))
		return
	}
	p, err := h.svc.ApproveRiderPayout(r.Context(), payoutID, u.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, p)
}

// ListReconciliationAlerts handles GET /admin/reconciliation-alerts
func (h *Handler) ListReconciliationAlerts(w http.ResponseWriter, r *http.Request) {
	page, perPage := parsePagination(r)
	items, meta, err := h.svc.ListOpenAlerts(r.Context(), page, perPage)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, pagination.PagedResponse{Data: items, Meta: meta})
}

// ResolveReconciliationAlert handles PATCH /admin/reconciliation-alerts/:id/resolve
func (h *Handler) ResolveReconciliationAlert(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		respond.Error(w, apperror.Unauthorized("authentication required"))
		return
	}
	alertID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid alert id"))
		return
	}
	var req struct {
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	alert, err := h.svc.ResolveAlert(r.Context(), alertID, u.ID, req.Notes)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, alert)
}

// ListCashCollections handles GET /admin/cash-collections
func (h *Handler) ListCashCollections(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	page, perPage := parsePagination(r)
	items, meta, err := h.svc.ListCashCollections(r.Context(), t.ID, page, perPage)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, pagination.PagedResponse{Data: items, Meta: meta})
}

// RemitCashByRider handles POST /rider/cash-remittance
func (h *Handler) RemitCashByRider(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		respond.Error(w, apperror.Unauthorized("authentication required"))
		return
	}
	if err := h.svc.q.RemitCashCollectionByRider(r.Context(), u.ID); err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"status": "remitted"})
}

// ListSubscriptionInvoices handles GET /admin/subscription-invoices
func (h *Handler) ListSubscriptionInvoices(w http.ResponseWriter, r *http.Request) {
	page, perPage := parsePagination(r)
	items, meta, err := h.svc.ListSubscriptionInvoices(r.Context(), page, perPage)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, pagination.PagedResponse{Data: items, Meta: meta})
}

func parseDateRange(r *http.Request) (from, to time.Time, err error) {
	q := r.URL.Query()
	fromStr := q.Get("from")
	toStr := q.Get("to")
	if fromStr == "" || toStr == "" {
		// Default to current month
		now := time.Now().UTC()
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		to = from.AddDate(0, 1, 0)
		return from, to, nil
	}
	from, err = time.Parse("2006-01-02", fromStr)
	if err != nil {
		return from, to, fmt.Errorf("invalid from date")
	}
	to, err = time.Parse("2006-01-02", toStr)
	if err != nil {
		return from, to, fmt.Errorf("invalid to date")
	}
	return from, to, nil
}

func decimalFromString(s string) (decimal.Decimal, error) {
	return decimal.NewFromString(s)
}

func toAppError(err error) *apperror.AppError {
	if e, ok := err.(*apperror.AppError); ok {
		return e
	}
	return apperror.Internal("unexpected error", err)
}
