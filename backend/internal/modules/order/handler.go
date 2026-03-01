package order

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
	redisclient "github.com/munchies/platform/backend/internal/platform/redis"
	"github.com/shopspring/decimal"
)

// Handler handles order HTTP requests.
type Handler struct {
	svc   *Service
	redis *redisclient.Client
}

// NewHandler creates a new order handler.
func NewHandler(svc *Service, redis *redisclient.Client) *Handler {
	return &Handler{svc: svc, redis: redis}
}

// CalculateCharges handles POST /api/v1/orders/charges/calculate
func (h *Handler) CalculateCharges(w http.ResponseWriter, r *http.Request) {
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
		Items []struct {
			ProductID           string `json:"product_id"`
			RestaurantID        string `json:"restaurant_id"`
			CategoryID          string `json:"category_id"`
			Quantity            int32  `json:"quantity"`
			UnitPrice           string `json:"unit_price"`
			ModifierPrice       string `json:"modifier_price"`
			ItemDiscount        string `json:"item_discount"`
			ItemVat             string `json:"item_vat"`
			ProductName         string `json:"product_name"`
		} `json:"items"`
		PromoCode    string `json:"promo_code"`
		DeliveryArea string `json:"delivery_area"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	items, err := parseCartItems(req.Items)
	if err != nil {
		respond.Error(w, err.(*apperror.AppError))
		return
	}

	result, err := h.svc.CalculateCharges(r.Context(), CalculateChargesRequest{
		TenantID:     t.ID,
		UserID:       u.ID,
		Items:        items,
		PromoCode:    req.PromoCode,
		DeliveryArea: req.DeliveryArea,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, result)
}

// CreateOrder handles POST /api/v1/orders
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
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
		Items []struct {
			ProductID           string          `json:"product_id"`
			RestaurantID        string          `json:"restaurant_id"`
			CategoryID          string          `json:"category_id"`
			Quantity            int32           `json:"quantity"`
			UnitPrice           string          `json:"unit_price"`
			ModifierPrice       string          `json:"modifier_price"`
			ProductName         string          `json:"product_name"`
			ProductSnapshot     json.RawMessage `json:"product_snapshot"`
			SelectedModifiers   json.RawMessage `json:"selected_modifiers"`
			SpecialInstructions string          `json:"special_instructions"`
			ItemDiscount        string          `json:"item_discount"`
			ItemVat             string          `json:"item_vat"`
		} `json:"items"`
		PromoCode              string          `json:"promo_code"`
		PaymentMethod          string          `json:"payment_method"`
		Platform               string          `json:"platform"`
		DeliveryAddressID      *string         `json:"delivery_address_id"`
		DeliveryAddress        json.RawMessage `json:"delivery_address"`
		DeliveryRecipientName  string          `json:"delivery_recipient_name"`
		DeliveryRecipientPhone string          `json:"delivery_recipient_phone"`
		DeliveryArea           string          `json:"delivery_area"`
		DeliveryGeoLat         *string         `json:"delivery_geo_lat"`
		DeliveryGeoLng         *string         `json:"delivery_geo_lng"`
		CustomerNote           string          `json:"customer_note"`
		IsPriority             bool            `json:"is_priority"`
		IsReorder              bool            `json:"is_reorder"`
		AutoConfirmMinutes     *int            `json:"auto_confirm_minutes"`
		EstimatedDeliveryMins  *int32          `json:"estimated_delivery_minutes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	if req.DeliveryRecipientName == "" || req.DeliveryRecipientPhone == "" || req.DeliveryArea == "" {
		respond.Error(w, apperror.BadRequest("delivery_recipient_name, delivery_recipient_phone, and delivery_area are required"))
		return
	}

	if len(req.DeliveryAddress) == 0 {
		req.DeliveryAddress = json.RawMessage("{}")
	}

	paymentMethod := sqlc.PaymentMethod(req.PaymentMethod)
	validPayments := map[sqlc.PaymentMethod]bool{
		sqlc.PaymentMethodCod:        true,
		sqlc.PaymentMethodBkash:      true,
		sqlc.PaymentMethodAamarpay:   true,
		sqlc.PaymentMethodWallet:     true,
	}
	if !validPayments[paymentMethod] {
		respond.Error(w, apperror.BadRequest("invalid payment_method"))
		return
	}

	platform := sqlc.PlatformSource(req.Platform)
	if platform == "" {
		platform = sqlc.PlatformSourceWeb
	}

	// Parse cart items
	cartItems := make([]CartItemRequest, 0, len(req.Items))
	for _, item := range req.Items {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			respond.Error(w, apperror.BadRequest("invalid product_id"))
			return
		}
		restaurantID, err := uuid.Parse(item.RestaurantID)
		if err != nil {
			respond.Error(w, apperror.BadRequest("invalid restaurant_id"))
			return
		}
		var categoryID uuid.UUID
		if item.CategoryID != "" {
			categoryID, _ = uuid.Parse(item.CategoryID)
		}

		unitPrice := decimal.Zero
		if item.UnitPrice != "" {
			unitPrice, err = decimal.NewFromString(item.UnitPrice)
			if err != nil {
				respond.Error(w, apperror.BadRequest("invalid unit_price"))
				return
			}
		}
		modPrice := decimal.Zero
		if item.ModifierPrice != "" {
			modPrice, err = decimal.NewFromString(item.ModifierPrice)
			if err != nil {
				respond.Error(w, apperror.BadRequest("invalid modifier_price"))
				return
			}
		}
		itemDisc := decimal.Zero
		if item.ItemDiscount != "" {
			itemDisc, err = decimal.NewFromString(item.ItemDiscount)
			if err != nil {
				respond.Error(w, apperror.BadRequest("invalid item_discount"))
				return
			}
		}
		itemVat := decimal.Zero
		if item.ItemVat != "" {
			itemVat, err = decimal.NewFromString(item.ItemVat)
			if err != nil {
				respond.Error(w, apperror.BadRequest("invalid item_vat"))
				return
			}
		}

		cartItems = append(cartItems, CartItemRequest{
			ProductID:           productID,
			RestaurantID:        restaurantID,
			CategoryID:          categoryID,
			Quantity:            item.Quantity,
			UnitPrice:           unitPrice,
			ModifierPrice:       modPrice,
			ProductName:         item.ProductName,
			ProductSnapshot:     item.ProductSnapshot,
			SelectedModifiers:   item.SelectedModifiers,
			SpecialInstructions: item.SpecialInstructions,
			ItemDiscount:        itemDisc,
			ItemVat:             itemVat,
		})
	}

	var deliveryAddrID *uuid.UUID
	if req.DeliveryAddressID != nil {
		id, err := uuid.Parse(*req.DeliveryAddressID)
		if err != nil {
			respond.Error(w, apperror.BadRequest("invalid delivery_address_id"))
			return
		}
		deliveryAddrID = &id
	}

	var geoLat, geoLng *decimal.Decimal
	if req.DeliveryGeoLat != nil {
		v, err := decimal.NewFromString(*req.DeliveryGeoLat)
		if err != nil {
			respond.Error(w, apperror.BadRequest("invalid delivery_geo_lat"))
			return
		}
		geoLat = &v
	}
	if req.DeliveryGeoLng != nil {
		v, err := decimal.NewFromString(*req.DeliveryGeoLng)
		if err != nil {
			respond.Error(w, apperror.BadRequest("invalid delivery_geo_lng"))
			return
		}
		geoLng = &v
	}

	result, err := h.svc.CreateOrder(r.Context(), CreateOrderRequest{
		TenantID:               t.ID,
		CustomerID:             u.ID,
		Items:                  cartItems,
		PromoCode:              req.PromoCode,
		PaymentMethod:          paymentMethod,
		Platform:               platform,
		DeliveryAddressID:      deliveryAddrID,
		DeliveryAddress:        req.DeliveryAddress,
		DeliveryRecipientName:  req.DeliveryRecipientName,
		DeliveryRecipientPhone: req.DeliveryRecipientPhone,
		DeliveryArea:           req.DeliveryArea,
		DeliveryGeoLat:         geoLat,
		DeliveryGeoLng:         geoLng,
		CustomerNote:           req.CustomerNote,
		IsPriority:             req.IsPriority,
		IsReorder:              req.IsReorder,
		AutoConfirmMinutes:     req.AutoConfirmMinutes,
		EstimatedDeliveryMins:  req.EstimatedDeliveryMins,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	// For online payment methods, return payment URL
	if paymentMethod == sqlc.PaymentMethodBkash || paymentMethod == sqlc.PaymentMethodAamarpay {
		respond.JSON(w, http.StatusCreated, map[string]interface{}{
			"order":       result,
			"payment_url": "/payment/redirect/" + result.Order.ID.String(),
			"message":     "Complete payment to confirm order",
		})
		return
	}

	respond.JSON(w, http.StatusCreated, result)
}

// GetOrder handles GET /api/v1/orders/{id}
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.svc.GetOrder(r.Context(), t.ID, orderID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	// Ensure customer can only see their own orders
	if result.Order.CustomerID != u.ID && u.Role == sqlc.UserRoleCustomer {
		respond.Error(w, apperror.Forbidden("access denied"))
		return
	}

	respond.JSON(w, http.StatusOK, result)
}

// TrackOrder handles GET /api/v1/orders/{id}/tracking (SSE stream)
func (h *Handler) TrackOrder(w http.ResponseWriter, r *http.Request) {
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

	// Get initial order state
	result, err := h.svc.GetOrder(r.Context(), t.ID, orderID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	if result.Order.CustomerID != u.ID && u.Role == sqlc.UserRoleCustomer {
		respond.Error(w, apperror.Forbidden("access denied"))
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		respond.Error(w, apperror.Internal("streaming not supported", nil))
		return
	}

	// Send initial state
	data, _ := json.Marshal(result)
	fmt.Fprintf(w, "data: %s\n\n", string(data))
	flusher.Flush()

	// Subscribe to Redis for live updates if available, otherwise heartbeat-only
	if h.redis != nil {
		channel := fmt.Sprintf("order:%s", orderID.String())
		pubsub := h.redis.Subscribe(r.Context(), channel)
		defer pubsub.Close()

		ch := pubsub.Channel()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				fmt.Fprintf(w, "data: %s\n\n", msg.Payload)
				flusher.Flush()
			case <-ticker.C:
				fmt.Fprintf(w, ": heartbeat\n\n")
				flusher.Flush()
			}
		}
	}

	// Fallback: heartbeat-only when Redis is not available
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}

// CancelOrder handles PATCH /api/v1/orders/{id}/cancel
func (h *Handler) CancelOrder(w http.ResponseWriter, r *http.Request) {
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
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	result, err := h.svc.CancelOrder(r.Context(), t.ID, orderID, u.ID, req.Reason)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, result)
}

// ListMyOrders handles GET /api/v1/me/orders
func (h *Handler) ListMyOrders(w http.ResponseWriter, r *http.Request) {
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

	page, perPage := parsePagination(r)
	orders, meta, err := h.svc.ListOrdersByCustomer(r.Context(), t.ID, u.ID, page, perPage)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, pagination.PagedResponse{Data: orders, Meta: meta})
}

// --- Partner Handlers ---

// ConfirmOrderPartner handles PATCH /partner/orders/{id}/confirm
func (h *Handler) ConfirmOrderPartner(w http.ResponseWriter, r *http.Request) {
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
		RestaurantID string `json:"restaurant_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	restaurantID, err := uuid.Parse(req.RestaurantID)
	if err != nil {
		respond.Error(w, apperror.BadRequest("restaurant_id is required"))
		return
	}

	result, err := h.svc.ConfirmOrder(r.Context(), t.ID, orderID, restaurantID, u.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, result)
}

// RejectOrderPartner handles PATCH /partner/orders/{id}/reject
func (h *Handler) RejectOrderPartner(w http.ResponseWriter, r *http.Request) {
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
		RestaurantID string `json:"restaurant_id"`
		Reason       string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	restaurantID, err := uuid.Parse(req.RestaurantID)
	if err != nil {
		respond.Error(w, apperror.BadRequest("restaurant_id is required"))
		return
	}
	if req.Reason == "" {
		respond.Error(w, apperror.BadRequest("reason is required for rejection"))
		return
	}

	result, err := h.svc.RejectOrder(r.Context(), t.ID, orderID, restaurantID, u.ID, req.Reason)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, result)
}

// PreparingOrderPartner handles PATCH /partner/orders/{id}/preparing
func (h *Handler) PreparingOrderPartner(w http.ResponseWriter, r *http.Request) {
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
		RestaurantID string `json:"restaurant_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	restaurantID, err := uuid.Parse(req.RestaurantID)
	if err != nil {
		respond.Error(w, apperror.BadRequest("restaurant_id is required"))
		return
	}

	result, err := h.svc.MarkPreparing(r.Context(), t.ID, orderID, restaurantID, u.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, result)
}

// ReadyOrderPartner handles PATCH /partner/orders/{id}/ready
func (h *Handler) ReadyOrderPartner(w http.ResponseWriter, r *http.Request) {
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
		RestaurantID string `json:"restaurant_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	restaurantID, err := uuid.Parse(req.RestaurantID)
	if err != nil {
		respond.Error(w, apperror.BadRequest("restaurant_id is required"))
		return
	}

	result, err := h.svc.MarkReady(r.Context(), t.ID, orderID, restaurantID, u.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, result)
}

// PickedByRider handles PATCH /rider/orders/{id}/picked/{restaurantID}
func (h *Handler) PickedByRider(w http.ResponseWriter, r *http.Request) {
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

	restaurantID, err := uuid.Parse(chi.URLParam(r, "restaurantID"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid restaurant id"))
		return
	}

	result, err := h.svc.MarkPickedByRider(r.Context(), t.ID, orderID, restaurantID, u.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, result)
}

// MarkDelivered handles PATCH /rider/orders/{id}/deliver
func (h *Handler) MarkDelivered(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.svc.MarkDelivered(r.Context(), t.ID, orderID, u.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, result)
}

// ForceCancelOrder handles PATCH /admin/orders/{id}/force-cancel
func (h *Handler) ForceCancelOrder(w http.ResponseWriter, r *http.Request) {
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
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if req.Reason == "" {
		respond.Error(w, apperror.BadRequest("reason is required for force cancel"))
		return
	}

	result, err := h.svc.ForceCancelOrder(r.Context(), t.ID, orderID, u.ID, req.Reason)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, result)
}

// ListPartnerOrders handles GET /partner/orders
func (h *Handler) ListPartnerOrders(w http.ResponseWriter, r *http.Request) {
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

	restaurantID, err := uuid.Parse(r.URL.Query().Get("restaurant_id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("restaurant_id is required"))
		return
	}

	page, perPage := parsePagination(r)
	orders, meta, err := h.svc.ListOrdersByRestaurant(r.Context(), t.ID, restaurantID, page, perPage)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, pagination.PagedResponse{Data: orders, Meta: meta})
}

// --- Helper Functions ---

func parseCartItems(items []struct {
	ProductID    string `json:"product_id"`
	RestaurantID string `json:"restaurant_id"`
	CategoryID   string `json:"category_id"`
	Quantity     int32  `json:"quantity"`
	UnitPrice    string `json:"unit_price"`
	ModifierPrice string `json:"modifier_price"`
	ItemDiscount string `json:"item_discount"`
	ItemVat      string `json:"item_vat"`
	ProductName  string `json:"product_name"`
}) ([]CartItemRequest, error) {
	cartItems := make([]CartItemRequest, 0, len(items))
	for _, item := range items {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			return nil, apperror.BadRequest("invalid product_id")
		}
		restaurantID, err := uuid.Parse(item.RestaurantID)
		if err != nil {
			return nil, apperror.BadRequest("invalid restaurant_id")
		}
		var categoryID uuid.UUID
		if item.CategoryID != "" {
			categoryID, _ = uuid.Parse(item.CategoryID)
		}

		unitPrice := decimal.Zero
		if item.UnitPrice != "" {
			unitPrice, err = decimal.NewFromString(item.UnitPrice)
			if err != nil {
				return nil, apperror.BadRequest("invalid unit_price")
			}
		}
		modPrice := decimal.Zero
		if item.ModifierPrice != "" {
			modPrice, err = decimal.NewFromString(item.ModifierPrice)
			if err != nil {
				return nil, apperror.BadRequest("invalid modifier_price")
			}
		}
		itemDisc := decimal.Zero
		if item.ItemDiscount != "" {
			itemDisc, err = decimal.NewFromString(item.ItemDiscount)
			if err != nil {
				return nil, apperror.BadRequest("invalid item_discount")
			}
		}
		itemVat := decimal.Zero
		if item.ItemVat != "" {
			itemVat, err = decimal.NewFromString(item.ItemVat)
			if err != nil {
				return nil, apperror.BadRequest("invalid item_vat")
			}
		}

		cartItems = append(cartItems, CartItemRequest{
			ProductID:    productID,
			RestaurantID: restaurantID,
			CategoryID:   categoryID,
			Quantity:     item.Quantity,
			UnitPrice:    unitPrice,
			ModifierPrice: modPrice,
			ProductName:  item.ProductName,
			ItemDiscount: itemDisc,
			ItemVat:      itemVat,
		})
	}
	return cartItems, nil
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
