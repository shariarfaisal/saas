package payment

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/modules/auth"
	"github.com/munchies/platform/backend/internal/modules/tenant"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/respond"
)

// Handler handles payment HTTP requests.
type Handler struct {
	svc         *Service
	callbackURL string
}

// NewHandler creates a new payment handler.
func NewHandler(svc *Service, callbackURL string) *Handler {
	return &Handler{svc: svc, callbackURL: callbackURL}
}

// InitiateBkash handles POST /api/v1/payments/bkash/initiate
func (h *Handler) InitiateBkash(w http.ResponseWriter, r *http.Request) {
	h.initiatePayment(w, r, sqlc.PaymentMethodBkash)
}

// InitiateAamarpay handles POST /api/v1/payments/aamarpay/initiate
func (h *Handler) InitiateAamarpay(w http.ResponseWriter, r *http.Request) {
	h.initiatePayment(w, r, sqlc.PaymentMethodAamarpay)
}

func (h *Handler) initiatePayment(w http.ResponseWriter, r *http.Request, method sqlc.PaymentMethod) {
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
		OrderID uuid.UUID `json:"order_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if req.OrderID == uuid.Nil {
		respond.Error(w, apperror.BadRequest("order_id is required"))
		return
	}

	callbackURL := h.callbackURL + "/api/v1/payments/" + string(method)

	gwResp, err := h.svc.InitiatePayment(r.Context(), InitiatePaymentRequest{
		OrderID:   req.OrderID,
		TenantID:  t.ID,
		UserID:    u.ID,
		Method:    method,
		IPAddr:    r.RemoteAddr,
		UserAgent: r.UserAgent(),
	}, callbackURL, u.Name, u.Phone.String)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"payment_id":   gwResp.GatewayPaymentID,
		"redirect_url": gwResp.RedirectURL,
		"status":       gwResp.Status,
	})
}

// BkashCallback handles GET /api/v1/payments/bkash/callback
func (h *Handler) BkashCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	paymentID := q.Get("paymentID")
	status := q.Get("status")

	if paymentID == "" {
		respond.Error(w, apperror.BadRequest("paymentID is required"))
		return
	}

	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}

	if status == "cancel" || status == "failure" {
		gwResponse, _ := json.Marshal(map[string]string{
			"status":    status,
			"paymentID": paymentID,
		})
		_, err := h.svc.ProcessCallback(r.Context(), paymentID, t.ID, sqlc.PaymentMethodBkash, gwResponse)
		if err != nil {
			respond.Error(w, toAppError(err))
			return
		}
		respond.JSON(w, http.StatusOK, map[string]string{
			"status":  "failed",
			"message": "payment " + status,
		})
		return
	}

	// Success flow: execute payment
	txn, err := h.svc.ProcessCallback(r.Context(), paymentID, t.ID, sqlc.PaymentMethodBkash, nil)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"status":         string(txn.Status),
		"transaction_id": txn.ID,
		"order_id":       txn.OrderID,
	})
}

// AamarpaySuccess handles POST /api/v1/payments/aamarpay/success
func (h *Handler) AamarpaySuccess(w http.ResponseWriter, r *http.Request) {
	h.handleAamarpayCallback(w, r, sqlc.TxnStatusSuccess)
}

// AamarpayFail handles POST /api/v1/payments/aamarpay/fail
func (h *Handler) AamarpayFail(w http.ResponseWriter, r *http.Request) {
	h.handleAamarpayCallback(w, r, sqlc.TxnStatusFailed)
}

// AamarpayCancel handles POST /api/v1/payments/aamarpay/cancel
func (h *Handler) AamarpayCancel(w http.ResponseWriter, r *http.Request) {
	h.handleAamarpayCallback(w, r, sqlc.TxnStatusCancelled)
}

func (h *Handler) handleAamarpayCallback(w http.ResponseWriter, r *http.Request, expectedStatus sqlc.TxnStatus) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}

	// AamarPay sends form data or JSON in the callback
	var callbackData map[string]interface{}
	if err := r.ParseForm(); err == nil && r.Form.Get("mer_txnid") != "" {
		callbackData = make(map[string]interface{})
		for k, v := range r.Form {
			if len(v) > 0 {
				callbackData[k] = v[0]
			}
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&callbackData); err != nil {
			callbackData = make(map[string]interface{})
		}
	}

	gwResponse, _ := json.Marshal(callbackData)

	merTxnID := ""
	if v, ok := callbackData["mer_txnid"]; ok {
		if s, ok := v.(string); ok {
			merTxnID = s
		}
	}
	if merTxnID == "" {
		respond.Error(w, apperror.BadRequest("mer_txnid is required"))
		return
	}

	if expectedStatus != sqlc.TxnStatusSuccess {
		// For fail/cancel, directly mark as failed
		_, _ = h.svc.ProcessCallback(r.Context(), merTxnID, t.ID, sqlc.PaymentMethodAamarpay, gwResponse)
		respond.JSON(w, http.StatusOK, map[string]string{
			"status":  string(expectedStatus),
			"message": "payment " + string(expectedStatus),
		})
		return
	}

	txn, err := h.svc.ProcessCallback(r.Context(), merTxnID, t.ID, sqlc.PaymentMethodAamarpay, gwResponse)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"status":         string(txn.Status),
		"transaction_id": txn.ID,
		"order_id":       txn.OrderID,
	})
}

func toAppError(err error) *apperror.AppError {
	if e, ok := err.(*apperror.AppError); ok {
		return e
	}
	return apperror.Internal("unexpected error", err)
}

// ProcessRefund handles POST /partner/orders/{id}/refund
func (h *Handler) ProcessRefund(w http.ResponseWriter, r *http.Request) {
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
		respond.Error(w, apperror.BadRequest("invalid order ID"))
		return
	}

	var req struct {
		Amount float64 `json:"amount"`
		Reason string  `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if req.Amount <= 0 {
		respond.Error(w, apperror.BadRequest("amount must be positive"))
		return
	}
	if req.Reason == "" {
		respond.Error(w, apperror.BadRequest("reason is required"))
		return
	}

	refund, err := h.svc.ProcessRefund(r.Context(), orderID, t.ID, req.Amount, req.Reason, u.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"refund_id": refund.ID,
		"order_id":  refund.OrderID,
		"amount":    req.Amount,
		"status":    string(refund.Status),
		"reason":    refund.Reason,
	})
}
