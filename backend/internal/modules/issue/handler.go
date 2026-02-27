package issue

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/modules/auth"
	"github.com/munchies/platform/backend/internal/modules/tenant"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/pagination"
	"github.com/munchies/platform/backend/internal/pkg/respond"
)

// Handler handles order issue HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new issue handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// CreateIssue handles POST /api/v1/orders/:id/issue
func (h *Handler) CreateIssue(w http.ResponseWriter, r *http.Request) {
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
	orderID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid order id"))
		return
	}

	var req struct {
		IssueType string `json:"issue_type"`
		Details   string `json:"details"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if req.Details == "" {
		respond.Error(w, apperror.BadRequest("details is required"))
		return
	}

	issue, err := h.svc.CreateIssue(r.Context(), t.ID, CreateIssueRequest{
		OrderID:          orderID,
		IssueType:        sqlc.IssueType(req.IssueType),
		ReportedByID:     u.ID,
		Details:          req.Details,
		AccountableParty: sqlc.AccountablePlatform,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusCreated, issue)
}

// GetIssue handles GET /partner/issues/:id
func (h *Handler) GetIssue(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	issueID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid issue id"))
		return
	}
	issue, err := h.svc.GetByID(r.Context(), t.ID, issueID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, issue)
}

// ListIssues handles GET /partner/issues
func (h *Handler) ListIssues(w http.ResponseWriter, r *http.Request) {
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

// AddMessage handles POST /partner/issues/:id/message
func (h *Handler) AddMessage(w http.ResponseWriter, r *http.Request) {
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
	issueID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid issue id"))
		return
	}

	var req struct {
		Message     string   `json:"message"`
		Attachments []string `json:"attachments"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if req.Message == "" {
		respond.Error(w, apperror.BadRequest("message is required"))
		return
	}

	attachments := req.Attachments
	if attachments == nil {
		attachments = []string{}
	}

	msg, err := h.svc.AddMessage(r.Context(), t.ID, issueID, u.ID, req.Message, attachments)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusCreated, msg)
}

// ListMessages handles GET /partner/issues/:id/messages
func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	issueID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid issue id"))
		return
	}
	messages, err := h.svc.ListMessages(r.Context(), t.ID, issueID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, messages)
}

// ResolveIssue handles PATCH /admin/issues/:id/resolve
func (h *Handler) ResolveIssue(w http.ResponseWriter, r *http.Request) {
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
	issueID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid issue id"))
		return
	}

	var req struct {
		Note   string `json:"note"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	issue, err := h.svc.Resolve(r.Context(), t.ID, issueID, u.ID, req.Note)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	h.svc.CreateAuditLog(r.Context(), t.ID, u.ID, "issue.resolved", "order_issue", issueID, req.Reason)

	respond.JSON(w, http.StatusOK, issue)
}

// ApproveRefund handles PATCH /admin/issues/:id/refund/approve
func (h *Handler) ApproveRefund(w http.ResponseWriter, r *http.Request) {
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
	issueID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid issue id"))
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	issue, err := h.svc.ApproveRefund(r.Context(), t.ID, issueID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	h.svc.CreateAuditLog(r.Context(), t.ID, u.ID, "refund.approved", "order_issue", issueID, req.Reason)

	respond.JSON(w, http.StatusOK, issue)
}

// RejectRefund handles PATCH /admin/issues/:id/refund/reject
func (h *Handler) RejectRefund(w http.ResponseWriter, r *http.Request) {
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
	issueID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid issue id"))
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	issue, err := h.svc.RejectRefund(r.Context(), t.ID, issueID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	h.svc.CreateAuditLog(r.Context(), t.ID, u.ID, "refund.rejected", "order_issue", issueID, req.Reason)

	respond.JSON(w, http.StatusOK, issue)
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

func toAppError(err error) *apperror.AppError {
	if e, ok := err.(*apperror.AppError); ok {
		return e
	}
	return apperror.Internal("unexpected error", err)
}
