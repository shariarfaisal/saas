package order

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/modules/inventory"
	"github.com/munchies/platform/backend/internal/modules/promo"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/pagination"
	"github.com/shopspring/decimal"
)

// Service handles order business logic.
type Service struct {
	q        *sqlc.Queries
	pool     *pgxpool.Pool
	invSvc   *inventory.Service
	promoSvc *promo.Service
}

// NewService creates a new order service.
func NewService(q *sqlc.Queries, pool *pgxpool.Pool, invSvc *inventory.Service, promoSvc *promo.Service) *Service {
	return &Service{q: q, pool: pool, invSvc: invSvc, promoSvc: promoSvc}
}

// --- Request/Response Types ---

// CartItemRequest represents an item in the order request.
type CartItemRequest struct {
	ProductID           uuid.UUID       `json:"product_id"`
	RestaurantID        uuid.UUID       `json:"restaurant_id"`
	Quantity            int32           `json:"quantity"`
	UnitPrice           decimal.Decimal `json:"unit_price"`
	ModifierPrice       decimal.Decimal `json:"modifier_price"`
	ProductName         string          `json:"product_name"`
	ProductSnapshot     json.RawMessage `json:"product_snapshot"`
	SelectedModifiers   json.RawMessage `json:"selected_modifiers"`
	SpecialInstructions string          `json:"special_instructions"`
	CategoryID          uuid.UUID       `json:"category_id"`
	ItemDiscount        decimal.Decimal `json:"item_discount"`
	ItemVat             decimal.Decimal `json:"item_vat"`
}

// CalculateChargesRequest holds the request for charge pre-calculation.
type CalculateChargesRequest struct {
	TenantID      uuid.UUID
	UserID        uuid.UUID
	Items         []CartItemRequest
	PromoCode     string
	DeliveryArea  string
	PaymentMethod string
}

// ChargeBreakdown is the response for charge pre-calculation.
type ChargeBreakdown struct {
	Subtotal           decimal.Decimal          `json:"subtotal"`
	ItemDiscountTotal  decimal.Decimal          `json:"item_discount_total"`
	PromoDiscountTotal decimal.Decimal          `json:"promo_discount_total"`
	VatTotal           decimal.Decimal          `json:"vat_total"`
	DeliveryCharge     decimal.Decimal          `json:"delivery_charge"`
	ServiceFee         decimal.Decimal          `json:"service_fee"`
	TotalAmount        decimal.Decimal          `json:"total_amount"`
	PromoResult        *promo.PromoValidationResult `json:"promo_result,omitempty"`
	Items              []ItemBreakdown          `json:"items"`
}

// ItemBreakdown shows the price breakdown for a single item.
type ItemBreakdown struct {
	ProductID    uuid.UUID       `json:"product_id"`
	RestaurantID uuid.UUID       `json:"restaurant_id"`
	Quantity     int32           `json:"quantity"`
	UnitPrice    decimal.Decimal `json:"unit_price"`
	ModifierPrice decimal.Decimal `json:"modifier_price"`
	ItemSubtotal decimal.Decimal `json:"item_subtotal"`
	ItemDiscount decimal.Decimal `json:"item_discount"`
	ItemVat      decimal.Decimal `json:"item_vat"`
	PromoDiscount decimal.Decimal `json:"promo_discount"`
	ItemTotal    decimal.Decimal `json:"item_total"`
}

// CreateOrderRequest holds all data needed to create an order.
type CreateOrderRequest struct {
	TenantID               uuid.UUID
	CustomerID             uuid.UUID
	Items                  []CartItemRequest
	PromoCode              string
	PaymentMethod          sqlc.PaymentMethod
	Platform               sqlc.PlatformSource
	DeliveryAddressID      *uuid.UUID
	DeliveryAddress        json.RawMessage
	DeliveryRecipientName  string
	DeliveryRecipientPhone string
	DeliveryArea           string
	DeliveryGeoLat         *decimal.Decimal
	DeliveryGeoLng         *decimal.Decimal
	CustomerNote           string
	IsPriority             bool
	IsReorder              bool
	AutoConfirmMinutes     *int
	EstimatedDeliveryMins  *int32
}

// OrderDetail is the full order response including items and pickups.
type OrderDetail struct {
	Order    sqlc.Order              `json:"order"`
	Items    []sqlc.OrderItem        `json:"items"`
	Pickups  []sqlc.OrderPickup      `json:"pickups"`
	Timeline []sqlc.OrderTimelineEvent `json:"timeline,omitempty"`
}

// CalculateCharges pre-calculates order charges without creating an order.
func (s *Service) CalculateCharges(ctx context.Context, req CalculateChargesRequest) (*ChargeBreakdown, error) {
	if len(req.Items) == 0 {
		return nil, apperror.BadRequest("at least one item is required")
	}

	breakdown := &ChargeBreakdown{
		Items: make([]ItemBreakdown, 0, len(req.Items)),
	}

	subtotal := decimal.Zero
	itemDiscountTotal := decimal.Zero
	vatTotal := decimal.Zero

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, apperror.BadRequest("item quantity must be positive")
		}
		itemSubtotal := item.UnitPrice.Add(item.ModifierPrice).Mul(decimal.NewFromInt32(item.Quantity))
		itemTotal := itemSubtotal.Sub(item.ItemDiscount).Add(item.ItemVat)

		breakdown.Items = append(breakdown.Items, ItemBreakdown{
			ProductID:     item.ProductID,
			RestaurantID:  item.RestaurantID,
			Quantity:      item.Quantity,
			UnitPrice:     item.UnitPrice,
			ModifierPrice: item.ModifierPrice,
			ItemSubtotal:  itemSubtotal,
			ItemDiscount:  item.ItemDiscount,
			ItemVat:       item.ItemVat,
			PromoDiscount: decimal.Zero,
			ItemTotal:     itemTotal,
		})

		subtotal = subtotal.Add(itemSubtotal)
		itemDiscountTotal = itemDiscountTotal.Add(item.ItemDiscount)
		vatTotal = vatTotal.Add(item.ItemVat)
	}

	// Apply promo if provided
	promoDiscountTotal := decimal.Zero
	if req.PromoCode != "" {
		cartItems := make([]promo.CartItem, 0, len(req.Items))
		for _, item := range req.Items {
			cartItems = append(cartItems, promo.CartItem{
				ProductID:    item.ProductID,
				RestaurantID: item.RestaurantID,
				CategoryID:   item.CategoryID,
				Quantity:     item.Quantity,
				UnitPrice:    item.UnitPrice,
				ItemSubtotal: item.UnitPrice.Add(item.ModifierPrice).Mul(decimal.NewFromInt32(item.Quantity)),
			})
		}

		result, err := s.promoSvc.Validate(ctx, req.TenantID, req.UserID, req.PromoCode, subtotal, cartItems)
		if err != nil {
			return nil, err
		}
		breakdown.PromoResult = result
		if result.Valid {
			promoDiscountTotal = result.DiscountAmount
		}
	}

	// Calculate delivery charge (simplified zone-based)
	deliveryCharge := decimal.NewFromInt(60) // default delivery charge in BDT
	serviceFee := decimal.Zero

	totalAmount := subtotal.Sub(itemDiscountTotal).Sub(promoDiscountTotal).Add(vatTotal).Add(deliveryCharge).Add(serviceFee)
	if totalAmount.IsNegative() {
		totalAmount = decimal.Zero
	}

	breakdown.Subtotal = subtotal
	breakdown.ItemDiscountTotal = itemDiscountTotal
	breakdown.PromoDiscountTotal = promoDiscountTotal
	breakdown.VatTotal = vatTotal
	breakdown.DeliveryCharge = deliveryCharge
	breakdown.ServiceFee = serviceFee
	breakdown.TotalAmount = totalAmount

	return breakdown, nil
}

// CreateOrder creates a new order atomically within a database transaction.
func (s *Service) CreateOrder(ctx context.Context, req CreateOrderRequest) (*OrderDetail, error) {
	if len(req.Items) == 0 {
		return nil, apperror.BadRequest("at least one item is required")
	}

	// Determine initial status based on payment method
	initialStatus := sqlc.OrderStatusCreated
	initialPaymentStatus := sqlc.PaymentStatusUnpaid
	if req.PaymentMethod == sqlc.PaymentMethodBkash || req.PaymentMethod == sqlc.PaymentMethodAamarpay {
		initialStatus = sqlc.OrderStatusPending
	}
	if req.PaymentMethod == sqlc.PaymentMethodCod {
		initialPaymentStatus = sqlc.PaymentStatusUnpaid
	}

	var result *OrderDetail

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, apperror.Internal("begin transaction", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.q.WithTx(tx)

	// 1. Generate order number
	orderNumResult, err := qtx.GenerateOrderNumber(ctx, sqlc.GenerateOrderNumberParams{
		Prefix:   "MUN",
		TenantID: req.TenantID,
	})
	if err != nil {
		return nil, apperror.Internal("generate order number", err)
	}
	orderNumber := fmt.Sprintf("%v", orderNumResult)

	// 2. Calculate charges
	subtotal := decimal.Zero
	itemDiscountTotal := decimal.Zero
	vatTotal := decimal.Zero

	type itemCalc struct {
		ItemSubtotal  decimal.Decimal
		ItemDiscount  decimal.Decimal
		ItemVat       decimal.Decimal
		PromoDiscount decimal.Decimal
		ItemTotal     decimal.Decimal
	}
	itemCalcs := make([]itemCalc, len(req.Items))

	for i, item := range req.Items {
		is := item.UnitPrice.Add(item.ModifierPrice).Mul(decimal.NewFromInt32(item.Quantity))
		it := is.Sub(item.ItemDiscount).Add(item.ItemVat)

		itemCalcs[i] = itemCalc{
			ItemSubtotal:  is,
			ItemDiscount:  item.ItemDiscount,
			ItemVat:       item.ItemVat,
			PromoDiscount: decimal.Zero,
			ItemTotal:     it,
		}

		subtotal = subtotal.Add(is)
		itemDiscountTotal = itemDiscountTotal.Add(item.ItemDiscount)
		vatTotal = vatTotal.Add(item.ItemVat)
	}

	// 3. Reserve stock
	stockReservations := make([]inventory.StockReservation, 0, len(req.Items))
	for _, item := range req.Items {
		stockReservations = append(stockReservations, inventory.StockReservation{
			ProductID:    item.ProductID,
			RestaurantID: item.RestaurantID,
			Quantity:     item.Quantity,
		})
	}
	if err := s.invSvc.CheckAndReserveStock(ctx, qtx, req.TenantID, stockReservations); err != nil {
		return nil, err
	}

	// 4. Validate and apply promo
	promoDiscountTotal := decimal.Zero
	var promoID pgtype.UUID
	var promoCode sql.NullString
	var promoSnapshot []byte

	if req.PromoCode != "" {
		cartItems := make([]promo.CartItem, 0, len(req.Items))
		for _, item := range req.Items {
			cartItems = append(cartItems, promo.CartItem{
				ProductID:    item.ProductID,
				RestaurantID: item.RestaurantID,
				CategoryID:   item.CategoryID,
				Quantity:     item.Quantity,
				UnitPrice:    item.UnitPrice,
				ItemSubtotal: item.UnitPrice.Add(item.ModifierPrice).Mul(decimal.NewFromInt32(item.Quantity)),
			})
		}

		promoResult, err := s.promoSvc.Validate(ctx, req.TenantID, req.CustomerID, req.PromoCode, subtotal, cartItems)
		if err != nil {
			return nil, err
		}
		if !promoResult.Valid {
			return nil, apperror.BadRequest("promo code invalid: " + promoResult.ErrorMessage)
		}

		promoDiscountTotal = promoResult.DiscountAmount
		promoID = pgtype.UUID{Bytes: promoResult.PromoID, Valid: true}
		promoCode = sql.NullString{String: promoResult.Code, Valid: true}
		snapshot, _ := json.Marshal(promoResult)
		promoSnapshot = snapshot
	}

	// 5. Calculate delivery charge and totals
	deliveryCharge := decimal.NewFromInt(60)
	serviceFee := decimal.Zero
	totalAmount := subtotal.Sub(itemDiscountTotal).Sub(promoDiscountTotal).Add(vatTotal).Add(deliveryCharge).Add(serviceFee)
	if totalAmount.IsNegative() {
		totalAmount = decimal.Zero
	}

	// 6. Auto-confirm timestamp
	var autoConfirmAt pgtype.Timestamptz
	if req.AutoConfirmMinutes != nil && *req.AutoConfirmMinutes > 0 {
		autoConfirmAt = pgtype.Timestamptz{
			Time:  time.Now().Add(time.Duration(*req.AutoConfirmMinutes) * time.Minute),
			Valid: true,
		}
	} else {
		// Default: auto-confirm after 5 minutes
		autoConfirmAt = pgtype.Timestamptz{
			Time:  time.Now().Add(5 * time.Minute),
			Valid: true,
		}
	}

	// 7. Prepare delivery address
	var deliveryAddrID pgtype.UUID
	if req.DeliveryAddressID != nil {
		deliveryAddrID = pgtype.UUID{Bytes: *req.DeliveryAddressID, Valid: true}
	}

	var geoLat, geoLng pgtype.Numeric
	if req.DeliveryGeoLat != nil {
		geoLat = pgtype.Numeric{Valid: true}
		_ = geoLat.Scan(req.DeliveryGeoLat.String())
	}
	if req.DeliveryGeoLng != nil {
		geoLng = pgtype.Numeric{Valid: true}
		_ = geoLng.Scan(req.DeliveryGeoLng.String())
	}

	// 8. Create order
	subtotalPg := pgtype.Numeric{Valid: true}
	_ = subtotalPg.Scan(subtotal.String())
	itemDiscTotalPg := pgtype.Numeric{Valid: true}
	_ = itemDiscTotalPg.Scan(itemDiscountTotal.String())
	promoDiscTotalPg := pgtype.Numeric{Valid: true}
	_ = promoDiscTotalPg.Scan(promoDiscountTotal.String())
	vatTotalPg := pgtype.Numeric{Valid: true}
	_ = vatTotalPg.Scan(vatTotal.String())
	delChargePg := pgtype.Numeric{Valid: true}
	_ = delChargePg.Scan(deliveryCharge.String())
	serviceFeePg := pgtype.Numeric{Valid: true}
	_ = serviceFeePg.Scan(serviceFee.String())
	totalAmountPg := pgtype.Numeric{Valid: true}
	_ = totalAmountPg.Scan(totalAmount.String())

	order, err := qtx.CreateOrder(ctx, sqlc.CreateOrderParams{
		TenantID:               req.TenantID,
		OrderNumber:            orderNumber,
		CustomerID:             req.CustomerID,
		Status:                 initialStatus,
		PaymentStatus:          initialPaymentStatus,
		PaymentMethod:          req.PaymentMethod,
		Platform:               req.Platform,
		DeliveryAddressID:      deliveryAddrID,
		DeliveryAddress:        req.DeliveryAddress,
		DeliveryRecipientName:  req.DeliveryRecipientName,
		DeliveryRecipientPhone: req.DeliveryRecipientPhone,
		DeliveryArea:           req.DeliveryArea,
		DeliveryGeoLat:         geoLat,
		DeliveryGeoLng:         geoLng,
		Subtotal:               subtotalPg,
		ItemDiscountTotal:      itemDiscTotalPg,
		PromoDiscountTotal:     promoDiscTotalPg,
		VatTotal:               vatTotalPg,
		DeliveryCharge:         delChargePg,
		ServiceFee:             serviceFeePg,
		TotalAmount:            totalAmountPg,
		PromoID:                promoID,
		PromoCode:              promoCode,
		PromoSnapshot:          promoSnapshot,
		IsPriority:             req.IsPriority,
		IsReorder:              req.IsReorder,
		CustomerNote:           sql.NullString{String: req.CustomerNote, Valid: req.CustomerNote != ""},
		AutoConfirmAt:          autoConfirmAt,
		EstimatedDeliveryMinutes: req.EstimatedDeliveryMins,
	})
	if err != nil {
		return nil, apperror.Internal("create order", err)
	}

	// 9. Create order items
	var orderItems []sqlc.OrderItem
	for i, item := range req.Items {
		modifiers := item.SelectedModifiers
		if modifiers == nil {
			modifiers = json.RawMessage("[]")
		}
		snapshot := item.ProductSnapshot
		if snapshot == nil {
			snapshot = json.RawMessage("{}")
		}

		unitPricePg := pgtype.Numeric{Valid: true}
		_ = unitPricePg.Scan(item.UnitPrice.String())
		modPricePg := pgtype.Numeric{Valid: true}
		_ = modPricePg.Scan(item.ModifierPrice.String())
		isSubPg := pgtype.Numeric{Valid: true}
		_ = isSubPg.Scan(itemCalcs[i].ItemSubtotal.String())
		idPg := pgtype.Numeric{Valid: true}
		_ = idPg.Scan(itemCalcs[i].ItemDiscount.String())
		ivPg := pgtype.Numeric{Valid: true}
		_ = ivPg.Scan(itemCalcs[i].ItemVat.String())
		pdPg := pgtype.Numeric{Valid: true}
		_ = pdPg.Scan(itemCalcs[i].PromoDiscount.String())
		itPg := pgtype.Numeric{Valid: true}
		_ = itPg.Scan(itemCalcs[i].ItemTotal.String())

		oi, err := qtx.CreateOrderItem(ctx, sqlc.CreateOrderItemParams{
			OrderID:             order.ID,
			RestaurantID:        item.RestaurantID,
			ProductID:           item.ProductID,
			TenantID:            req.TenantID,
			ProductName:         item.ProductName,
			ProductSnapshot:     snapshot,
			Quantity:            item.Quantity,
			UnitPrice:           unitPricePg,
			ModifierPrice:       modPricePg,
			ItemSubtotal:        isSubPg,
			ItemDiscount:        idPg,
			ItemVat:             ivPg,
			PromoDiscount:       pdPg,
			ItemTotal:           itPg,
			SelectedModifiers:   modifiers,
			SpecialInstructions: sql.NullString{String: item.SpecialInstructions, Valid: item.SpecialInstructions != ""},
		})
		if err != nil {
			return nil, apperror.Internal("create order item", err)
		}
		orderItems = append(orderItems, oi)
	}

	// 10. Create order pickups (grouped by restaurant)
	restaurantItems := make(map[uuid.UUID][]int)
	for i, item := range req.Items {
		restaurantItems[item.RestaurantID] = append(restaurantItems[item.RestaurantID], i)
	}

	var orderPickups []sqlc.OrderPickup
	pickupIdx := 1
	for restID, indices := range restaurantItems {
		pickupSubtotal := decimal.Zero
		pickupDiscount := decimal.Zero
		pickupVat := decimal.Zero

		for _, idx := range indices {
			pickupSubtotal = pickupSubtotal.Add(itemCalcs[idx].ItemSubtotal)
			pickupDiscount = pickupDiscount.Add(itemCalcs[idx].ItemDiscount)
			pickupVat = pickupVat.Add(itemCalcs[idx].ItemVat)
		}
		pickupTotal := pickupSubtotal.Sub(pickupDiscount).Add(pickupVat)

		pSubPg := pgtype.Numeric{Valid: true}
		_ = pSubPg.Scan(pickupSubtotal.String())
		pDiscPg := pgtype.Numeric{Valid: true}
		_ = pDiscPg.Scan(pickupDiscount.String())
		pVatPg := pgtype.Numeric{Valid: true}
		_ = pVatPg.Scan(pickupVat.String())
		pTotalPg := pgtype.Numeric{Valid: true}
		_ = pTotalPg.Scan(pickupTotal.String())
		commRatePg := pgtype.Numeric{Valid: true}
		_ = commRatePg.Scan("0")
		commAmtPg := pgtype.Numeric{Valid: true}
		_ = commAmtPg.Scan("0")

		pickupNumber := fmt.Sprintf("%s-P%d", orderNumber, pickupIdx)
		pickup, err := qtx.CreateOrderPickup(ctx, sqlc.CreateOrderPickupParams{
			OrderID:          order.ID,
			RestaurantID:     restID,
			TenantID:         req.TenantID,
			PickupNumber:     pickupNumber,
			Status:           sqlc.PickupStatusNew,
			ItemsSubtotal:    pSubPg,
			ItemsDiscount:    pDiscPg,
			ItemsVat:         pVatPg,
			ItemsTotal:       pTotalPg,
			CommissionRate:   commRatePg,
			CommissionAmount: commAmtPg,
		})
		if err != nil {
			return nil, apperror.Internal("create order pickup", err)
		}
		orderPickups = append(orderPickups, pickup)
		pickupIdx++
	}

	// 11. Add timeline event
	timeline, err := qtx.AddTimelineEvent(ctx, sqlc.AddTimelineEventParams{
		OrderID:        order.ID,
		TenantID:       req.TenantID,
		EventType:      "order_created",
		PreviousStatus: sqlc.NullOrderStatus{},
		NewStatus:      sqlc.NullOrderStatus{OrderStatus: initialStatus, Valid: true},
		Description:    "Order created",
		ActorID:        pgtype.UUID{Bytes: req.CustomerID, Valid: true},
		ActorType:      sqlc.ActorTypeCustomer,
		Metadata:       json.RawMessage("{}"),
	})
	if err != nil {
		return nil, apperror.Internal("add timeline event", err)
	}

	// 12. Record promo usage
	if promoID.Valid {
		discAmtPg := pgtype.Numeric{Valid: true}
		_ = discAmtPg.Scan(promoDiscountTotal.String())
		cbAmtPg := pgtype.Numeric{Valid: true}
		_ = cbAmtPg.Scan("0")

		_, err = qtx.CreatePromoUsage(ctx, sqlc.CreatePromoUsageParams{
			PromoID:        promoID.Bytes,
			UserID:         req.CustomerID,
			OrderID:        order.ID,
			TenantID:       req.TenantID,
			DiscountAmount: discAmtPg,
			CashbackAmount: cbAmtPg,
		})
		if err != nil {
			return nil, apperror.Internal("create promo usage", err)
		}

		err = qtx.IncrementPromoUsage(ctx, sqlc.IncrementPromoUsageParams{
			DiscountAmount: discAmtPg,
			ID:             promoID.Bytes,
			TenantID:       req.TenantID,
		})
		if err != nil {
			return nil, apperror.Internal("increment promo usage", err)
		}
	}

	// 13. Create outbox event for async processing (rider assignment)
	payload, _ := json.Marshal(map[string]interface{}{
		"order_id":     order.ID.String(),
		"tenant_id":    req.TenantID.String(),
		"order_number": orderNumber,
	})
	_, err = qtx.CreateOutboxEvent(ctx, sqlc.CreateOutboxEventParams{
		TenantID:      pgtype.UUID{Bytes: req.TenantID, Valid: true},
		AggregateType: "order",
		AggregateID:   order.ID,
		EventType:     "order.created",
		Payload:       payload,
		MaxAttempts:   5,
	})
	if err != nil {
		return nil, apperror.Internal("create outbox event", err)
	}

	// 14. Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, apperror.Internal("commit transaction", err)
	}

	result = &OrderDetail{
		Order:    order,
		Items:    orderItems,
		Pickups:  orderPickups,
		Timeline: []sqlc.OrderTimelineEvent{timeline},
	}

	return result, nil
}

// GetOrder returns full order details.
func (s *Service) GetOrder(ctx context.Context, tenantID, orderID uuid.UUID) (*OrderDetail, error) {
	order, err := s.q.GetOrderByID(ctx, sqlc.GetOrderByIDParams{
		ID:       orderID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("order")
	}
	if err != nil {
		return nil, apperror.Internal("get order", err)
	}

	items, err := s.q.GetOrderItemsByOrder(ctx, orderID)
	if err != nil {
		return nil, apperror.Internal("get order items", err)
	}

	pickups, err := s.q.GetOrderPickupsByOrder(ctx, orderID)
	if err != nil {
		return nil, apperror.Internal("get order pickups", err)
	}

	timeline, err := s.q.ListTimelineEvents(ctx, sqlc.ListTimelineEventsParams{
		OrderID:  orderID,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, apperror.Internal("list timeline events", err)
	}

	return &OrderDetail{
		Order:    order,
		Items:    items,
		Pickups:  pickups,
		Timeline: timeline,
	}, nil
}

// ListOrdersByCustomer returns paginated orders for a customer.
func (s *Service) ListOrdersByCustomer(ctx context.Context, tenantID, customerID uuid.UUID, page, perPage int) ([]sqlc.Order, pagination.Meta, error) {
	limit, offset := pagination.FormatLimitOffset(page, perPage)

	total, err := s.q.CountOrdersByCustomer(ctx, sqlc.CountOrdersByCustomerParams{
		CustomerID: customerID,
		TenantID:   tenantID,
	})
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("count orders", err)
	}

	orders, err := s.q.ListOrdersByCustomer(ctx, sqlc.ListOrdersByCustomerParams{
		CustomerID: customerID,
		TenantID:   tenantID,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("list orders", err)
	}

	meta := pagination.NewMeta(total, limit, "")
	return orders, meta, nil
}

// ListOrdersByRestaurant returns paginated orders for a restaurant.
func (s *Service) ListOrdersByRestaurant(ctx context.Context, tenantID, restaurantID uuid.UUID, page, perPage int) ([]sqlc.Order, pagination.Meta, error) {
	limit, offset := pagination.FormatLimitOffset(page, perPage)

	total, err := s.q.CountOrdersByRestaurant(ctx, sqlc.CountOrdersByRestaurantParams{
		RestaurantID: restaurantID,
		TenantID:     tenantID,
	})
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("count orders", err)
	}

	orders, err := s.q.ListOrdersByRestaurant(ctx, sqlc.ListOrdersByRestaurantParams{
		RestaurantID: restaurantID,
		TenantID:     tenantID,
		Limit:        int32(limit),
		Offset:       int32(offset),
	})
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("list orders", err)
	}

	meta := pagination.NewMeta(total, limit, "")
	return orders, meta, nil
}

// --- Order Status Transitions ---

// ConfirmOrder transitions order from CREATED to CONFIRMED.
func (s *Service) ConfirmOrder(ctx context.Context, tenantID, orderID, restaurantID, actorID uuid.UUID) (*sqlc.Order, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, apperror.Internal("begin transaction", err)
	}
	defer tx.Rollback(ctx)
	qtx := s.q.WithTx(tx)

	order, err := qtx.GetOrderForUpdate(ctx, sqlc.GetOrderForUpdateParams{
		ID:       orderID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("order")
	}
	if err != nil {
		return nil, apperror.Internal("get order", err)
	}

	if order.Status != sqlc.OrderStatusCreated {
		return nil, apperror.BadRequest("order can only be confirmed from CREATED status")
	}

	// Transition pickup status
	_, err = qtx.TransitionPickupStatus(ctx, sqlc.TransitionPickupStatusParams{
		NewStatus:       sqlc.PickupStatusConfirmed,
		OrderID:         orderID,
		RestaurantID:    restaurantID,
		RejectionReason: sql.NullString{},
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.Internal("transition pickup status", err)
	}

	// Check if all pickups are confirmed
	allConfirmed, err := qtx.CheckAllPickupsInStatus(ctx, sqlc.CheckAllPickupsInStatusParams{
		OrderID:        orderID,
		ExpectedStatus: sqlc.PickupStatusConfirmed,
	})
	if err != nil {
		return nil, apperror.Internal("check pickups status", err)
	}

	var updated sqlc.Order
	if allConfirmed {
		updated, err = qtx.TransitionOrderStatus(ctx, sqlc.TransitionOrderStatusParams{
			NewStatus:           sqlc.OrderStatusConfirmed,
			CancellationReason:  sql.NullString{},
			CancelledBy:         sqlc.NullActorType{},
			RejectionReason:     sql.NullString{},
			RejectedBy:          sqlc.NullActorType{},
			ID:                  orderID,
			TenantID:            tenantID,
		})
		if err != nil {
			return nil, apperror.Internal("transition order status", err)
		}

		_, err = qtx.AddTimelineEvent(ctx, sqlc.AddTimelineEventParams{
			OrderID:        orderID,
			TenantID:       tenantID,
			EventType:      "status_changed",
			PreviousStatus: sqlc.NullOrderStatus{OrderStatus: sqlc.OrderStatusCreated, Valid: true},
			NewStatus:      sqlc.NullOrderStatus{OrderStatus: sqlc.OrderStatusConfirmed, Valid: true},
			Description:    "Order confirmed by restaurant",
			ActorID:        pgtype.UUID{Bytes: actorID, Valid: true},
			ActorType:      sqlc.ActorTypeRestaurant,
			Metadata:       json.RawMessage("{}"),
		})
		if err != nil {
			return nil, apperror.Internal("add timeline event", err)
		}
	} else {
		updated = order
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperror.Internal("commit transaction", err)
	}

	return &updated, nil
}

// RejectOrder transitions order from CREATED to REJECTED.
func (s *Service) RejectOrder(ctx context.Context, tenantID, orderID, restaurantID, actorID uuid.UUID, reason string) (*sqlc.Order, error) {
	if reason == "" {
		return nil, apperror.BadRequest("rejection reason is required")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, apperror.Internal("begin transaction", err)
	}
	defer tx.Rollback(ctx)
	qtx := s.q.WithTx(tx)

	order, err := qtx.GetOrderForUpdate(ctx, sqlc.GetOrderForUpdateParams{
		ID:       orderID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("order")
	}
	if err != nil {
		return nil, apperror.Internal("get order", err)
	}

	if order.Status != sqlc.OrderStatusCreated {
		return nil, apperror.BadRequest("order can only be rejected from CREATED status")
	}

	// Reject the pickup for this restaurant
	_, err = qtx.TransitionPickupStatus(ctx, sqlc.TransitionPickupStatusParams{
		NewStatus:       sqlc.PickupStatusRejected,
		OrderID:         orderID,
		RestaurantID:    restaurantID,
		RejectionReason: sql.NullString{String: reason, Valid: true},
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.Internal("transition pickup status", err)
	}

	// For multi-restaurant: check if all pickups are rejected
	pickupCount, err := qtx.GetPickupCountByOrder(ctx, orderID)
	if err != nil {
		return nil, apperror.Internal("get pickup count", err)
	}

	allRejected, err := qtx.CheckAllPickupsInStatus(ctx, sqlc.CheckAllPickupsInStatusParams{
		OrderID:        orderID,
		ExpectedStatus: sqlc.PickupStatusRejected,
	})
	if err != nil {
		return nil, apperror.Internal("check all pickups rejected", err)
	}

	// If all pickups rejected, or single-restaurant order, reject the whole order
	var updated sqlc.Order
	if allRejected || pickupCount == 1 {
		updated, err = qtx.TransitionOrderStatus(ctx, sqlc.TransitionOrderStatusParams{
			NewStatus:          sqlc.OrderStatusRejected,
			CancellationReason: sql.NullString{},
			CancelledBy:        sqlc.NullActorType{},
			RejectionReason:    sql.NullString{String: reason, Valid: true},
			RejectedBy:         sqlc.NullActorType{ActorType: sqlc.ActorTypeRestaurant, Valid: true},
			ID:                 orderID,
			TenantID:           tenantID,
		})
		if err != nil {
			return nil, apperror.Internal("transition order status", err)
		}

		// Release stock for rejected items
		items, err := qtx.GetOrderItemsByRestaurant(ctx, sqlc.GetOrderItemsByRestaurantParams{
			OrderID:      orderID,
			RestaurantID: restaurantID,
		})
		if err != nil {
			return nil, apperror.Internal("get order items for stock release", err)
		}
		stockReleases := make([]inventory.StockReservation, 0, len(items))
		for _, item := range items {
			stockReleases = append(stockReleases, inventory.StockReservation{
				ProductID:    item.ProductID,
				RestaurantID: item.RestaurantID,
				Quantity:     item.Quantity,
			})
		}
		if err := s.invSvc.ReleaseStockForOrder(ctx, qtx, tenantID, stockReleases); err != nil {
			return nil, err
		}
	} else {
		updated = order
	}

	_, err = qtx.AddTimelineEvent(ctx, sqlc.AddTimelineEventParams{
		OrderID:        orderID,
		TenantID:       tenantID,
		EventType:      "status_changed",
		PreviousStatus: sqlc.NullOrderStatus{OrderStatus: sqlc.OrderStatusCreated, Valid: true},
		NewStatus:      sqlc.NullOrderStatus{OrderStatus: sqlc.OrderStatusRejected, Valid: true},
		Description:    "Order rejected by restaurant: " + reason,
		ActorID:        pgtype.UUID{Bytes: actorID, Valid: true},
		ActorType:      sqlc.ActorTypeRestaurant,
		Metadata:       json.RawMessage("{}"),
	})
	if err != nil {
		return nil, apperror.Internal("add timeline event", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperror.Internal("commit transaction", err)
	}

	return &updated, nil
}

// MarkPreparing transitions pickup status to PREPARING.
func (s *Service) MarkPreparing(ctx context.Context, tenantID, orderID, restaurantID, actorID uuid.UUID) (*sqlc.Order, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, apperror.Internal("begin transaction", err)
	}
	defer tx.Rollback(ctx)
	qtx := s.q.WithTx(tx)

	order, err := qtx.GetOrderForUpdate(ctx, sqlc.GetOrderForUpdateParams{
		ID:       orderID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("order")
	}
	if err != nil {
		return nil, apperror.Internal("get order", err)
	}

	if order.Status != sqlc.OrderStatusConfirmed && order.Status != sqlc.OrderStatusCreated {
		return nil, apperror.BadRequest("order must be in CONFIRMED or CREATED status to mark as preparing")
	}

	_, err = qtx.TransitionPickupStatus(ctx, sqlc.TransitionPickupStatusParams{
		NewStatus:       sqlc.PickupStatusPreparing,
		OrderID:         orderID,
		RestaurantID:    restaurantID,
		RejectionReason: sql.NullString{},
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.Internal("transition pickup status", err)
	}

	// Update parent order to PREPARING if any pickup is preparing
	updated := order
	if order.Status == sqlc.OrderStatusConfirmed {
		updated, err = qtx.TransitionOrderStatus(ctx, sqlc.TransitionOrderStatusParams{
			NewStatus:          sqlc.OrderStatusPreparing,
			CancellationReason: sql.NullString{},
			CancelledBy:        sqlc.NullActorType{},
			RejectionReason:    sql.NullString{},
			RejectedBy:         sqlc.NullActorType{},
			ID:                 orderID,
			TenantID:           tenantID,
		})
		if err != nil {
			return nil, apperror.Internal("transition order status", err)
		}
	}

	_, err = qtx.AddTimelineEvent(ctx, sqlc.AddTimelineEventParams{
		OrderID:        orderID,
		TenantID:       tenantID,
		EventType:      "status_changed",
		PreviousStatus: sqlc.NullOrderStatus{OrderStatus: order.Status, Valid: true},
		NewStatus:      sqlc.NullOrderStatus{OrderStatus: sqlc.OrderStatusPreparing, Valid: true},
		Description:    "Order is being prepared",
		ActorID:        pgtype.UUID{Bytes: actorID, Valid: true},
		ActorType:      sqlc.ActorTypeRestaurant,
		Metadata:       json.RawMessage("{}"),
	})
	if err != nil {
		return nil, apperror.Internal("add timeline event", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperror.Internal("commit transaction", err)
	}

	return &updated, nil
}

// MarkReady transitions pickup status to READY.
func (s *Service) MarkReady(ctx context.Context, tenantID, orderID, restaurantID, actorID uuid.UUID) (*sqlc.Order, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, apperror.Internal("begin transaction", err)
	}
	defer tx.Rollback(ctx)
	qtx := s.q.WithTx(tx)

	order, err := qtx.GetOrderForUpdate(ctx, sqlc.GetOrderForUpdateParams{
		ID:       orderID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("order")
	}
	if err != nil {
		return nil, apperror.Internal("get order", err)
	}

	if order.Status != sqlc.OrderStatusPreparing && order.Status != sqlc.OrderStatusConfirmed {
		return nil, apperror.BadRequest("order must be in PREPARING or CONFIRMED status to mark as ready")
	}

	_, err = qtx.TransitionPickupStatus(ctx, sqlc.TransitionPickupStatusParams{
		NewStatus:       sqlc.PickupStatusReady,
		OrderID:         orderID,
		RestaurantID:    restaurantID,
		RejectionReason: sql.NullString{},
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.Internal("transition pickup status", err)
	}

	// Check if all pickups are ready → parent order READY
	allReady, err := qtx.CheckAllPickupsInStatus(ctx, sqlc.CheckAllPickupsInStatusParams{
		OrderID:        orderID,
		ExpectedStatus: sqlc.PickupStatusReady,
	})
	if err != nil {
		return nil, apperror.Internal("check all pickups ready", err)
	}

	updated := order
	if allReady {
		updated, err = qtx.TransitionOrderStatus(ctx, sqlc.TransitionOrderStatusParams{
			NewStatus:          sqlc.OrderStatusReady,
			CancellationReason: sql.NullString{},
			CancelledBy:        sqlc.NullActorType{},
			RejectionReason:    sql.NullString{},
			RejectedBy:         sqlc.NullActorType{},
			ID:                 orderID,
			TenantID:           tenantID,
		})
		if err != nil {
			return nil, apperror.Internal("transition order status", err)
		}
	}

	_, err = qtx.AddTimelineEvent(ctx, sqlc.AddTimelineEventParams{
		OrderID:        orderID,
		TenantID:       tenantID,
		EventType:      "status_changed",
		PreviousStatus: sqlc.NullOrderStatus{OrderStatus: order.Status, Valid: true},
		NewStatus:      sqlc.NullOrderStatus{OrderStatus: sqlc.OrderStatusReady, Valid: true},
		Description:    "Food is ready for pickup",
		ActorID:        pgtype.UUID{Bytes: actorID, Valid: true},
		ActorType:      sqlc.ActorTypeRestaurant,
		Metadata:       json.RawMessage("{}"),
	})
	if err != nil {
		return nil, apperror.Internal("add timeline event", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperror.Internal("commit transaction", err)
	}

	return &updated, nil
}

// MarkPickedByRider marks a specific restaurant pickup as PICKED by rider.
func (s *Service) MarkPickedByRider(ctx context.Context, tenantID, orderID, restaurantID, riderID uuid.UUID) (*sqlc.Order, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, apperror.Internal("begin transaction", err)
	}
	defer tx.Rollback(ctx)
	qtx := s.q.WithTx(tx)

	order, err := qtx.GetOrderForUpdate(ctx, sqlc.GetOrderForUpdateParams{
		ID:       orderID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("order")
	}
	if err != nil {
		return nil, apperror.Internal("get order", err)
	}

	if order.Status != sqlc.OrderStatusReady && order.Status != sqlc.OrderStatusPreparing {
		return nil, apperror.BadRequest("order must be in READY or PREPARING status for pickup")
	}

	// Mark this restaurant's pickup as PICKED
	_, err = qtx.TransitionPickupStatus(ctx, sqlc.TransitionPickupStatusParams{
		NewStatus:       sqlc.PickupStatusPicked,
		OrderID:         orderID,
		RestaurantID:    restaurantID,
		RejectionReason: sql.NullString{},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("pickup for this restaurant")
		}
		return nil, apperror.Internal("transition pickup status", err)
	}

	// Check if all pickups are PICKED → parent order PICKED
	allPicked, err := qtx.CheckAllPickupsInStatus(ctx, sqlc.CheckAllPickupsInStatusParams{
		OrderID:        orderID,
		ExpectedStatus: sqlc.PickupStatusPicked,
	})
	if err != nil {
		return nil, apperror.Internal("check all pickups picked", err)
	}

	updated := order
	if allPicked {
		updated, err = qtx.TransitionOrderStatus(ctx, sqlc.TransitionOrderStatusParams{
			NewStatus:          sqlc.OrderStatusPicked,
			CancellationReason: sql.NullString{},
			CancelledBy:        sqlc.NullActorType{},
			RejectionReason:    sql.NullString{},
			RejectedBy:         sqlc.NullActorType{},
			ID:                 orderID,
			TenantID:           tenantID,
		})
		if err != nil {
			return nil, apperror.Internal("transition order status", err)
		}
	}

	_, err = qtx.AddTimelineEvent(ctx, sqlc.AddTimelineEventParams{
		OrderID:        orderID,
		TenantID:       tenantID,
		EventType:      "picked_up",
		PreviousStatus: sqlc.NullOrderStatus{OrderStatus: order.Status, Valid: true},
		NewStatus:      sqlc.NullOrderStatus{OrderStatus: sqlc.OrderStatusPicked, Valid: true},
		Description:    "Rider picked up order from restaurant",
		ActorID:        pgtype.UUID{Bytes: riderID, Valid: true},
		ActorType:      sqlc.ActorTypeRider,
		Metadata:       json.RawMessage(fmt.Sprintf(`{"restaurant_id":"%s"}`, restaurantID.String())),
	})
	if err != nil {
		return nil, apperror.Internal("add timeline event", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperror.Internal("commit transaction", err)
	}

	return &updated, nil
}

// CancelOrder cancels an order (customer can cancel PENDING or CREATED only).
func (s *Service) CancelOrder(ctx context.Context, tenantID, orderID, customerID uuid.UUID, reason string) (*sqlc.Order, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, apperror.Internal("begin transaction", err)
	}
	defer tx.Rollback(ctx)
	qtx := s.q.WithTx(tx)

	order, err := qtx.GetOrderForUpdate(ctx, sqlc.GetOrderForUpdateParams{
		ID:       orderID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("order")
	}
	if err != nil {
		return nil, apperror.Internal("get order", err)
	}

	if order.CustomerID != customerID {
		return nil, apperror.Forbidden("you can only cancel your own orders")
	}

	if order.Status != sqlc.OrderStatusPending && order.Status != sqlc.OrderStatusCreated {
		return nil, apperror.BadRequest("order can only be cancelled in PENDING or CREATED status")
	}

	// Release stock
	items, err := qtx.GetOrderItemsByOrder(ctx, orderID)
	if err != nil {
		return nil, apperror.Internal("get order items", err)
	}
	stockReleases := make([]inventory.StockReservation, 0, len(items))
	for _, item := range items {
		stockReleases = append(stockReleases, inventory.StockReservation{
			ProductID:    item.ProductID,
			RestaurantID: item.RestaurantID,
			Quantity:     item.Quantity,
		})
	}
	if err := s.invSvc.ReleaseStockForOrder(ctx, qtx, tenantID, stockReleases); err != nil {
		return nil, err
	}

	cancelReason := reason
	if cancelReason == "" {
		cancelReason = "cancelled by customer"
	}

	updated, err := qtx.TransitionOrderStatus(ctx, sqlc.TransitionOrderStatusParams{
		NewStatus:          sqlc.OrderStatusCancelled,
		CancellationReason: sql.NullString{String: cancelReason, Valid: true},
		CancelledBy:        sqlc.NullActorType{ActorType: sqlc.ActorTypeCustomer, Valid: true},
		RejectionReason:    sql.NullString{},
		RejectedBy:         sqlc.NullActorType{},
		ID:                 orderID,
		TenantID:           tenantID,
	})
	if err != nil {
		return nil, apperror.Internal("cancel order", err)
	}

	_, err = qtx.AddTimelineEvent(ctx, sqlc.AddTimelineEventParams{
		OrderID:        orderID,
		TenantID:       tenantID,
		EventType:      "status_changed",
		PreviousStatus: sqlc.NullOrderStatus{OrderStatus: order.Status, Valid: true},
		NewStatus:      sqlc.NullOrderStatus{OrderStatus: sqlc.OrderStatusCancelled, Valid: true},
		Description:    "Order cancelled by customer: " + cancelReason,
		ActorID:        pgtype.UUID{Bytes: customerID, Valid: true},
		ActorType:      sqlc.ActorTypeCustomer,
		Metadata:       json.RawMessage("{}"),
	})
	if err != nil {
		return nil, apperror.Internal("add timeline event", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperror.Internal("commit transaction", err)
	}

	return &updated, nil
}

// ForceCancelOrder admin force-cancels an order with mandatory reason.
func (s *Service) ForceCancelOrder(ctx context.Context, tenantID, orderID, adminID uuid.UUID, reason string) (*sqlc.Order, error) {
	if reason == "" {
		return nil, apperror.BadRequest("cancellation reason is required for force cancel")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, apperror.Internal("begin transaction", err)
	}
	defer tx.Rollback(ctx)
	qtx := s.q.WithTx(tx)

	order, err := qtx.GetOrderForUpdate(ctx, sqlc.GetOrderForUpdateParams{
		ID:       orderID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("order")
	}
	if err != nil {
		return nil, apperror.Internal("get order", err)
	}

	if order.Status == sqlc.OrderStatusDelivered || order.Status == sqlc.OrderStatusCancelled {
		return nil, apperror.BadRequest("cannot cancel a delivered or already cancelled order")
	}

	// Release stock
	items, err := qtx.GetOrderItemsByOrder(ctx, orderID)
	if err != nil {
		return nil, apperror.Internal("get order items", err)
	}
	stockReleases := make([]inventory.StockReservation, 0, len(items))
	for _, item := range items {
		stockReleases = append(stockReleases, inventory.StockReservation{
			ProductID:    item.ProductID,
			RestaurantID: item.RestaurantID,
			Quantity:     item.Quantity,
		})
	}
	if err := s.invSvc.ReleaseStockForOrder(ctx, qtx, tenantID, stockReleases); err != nil {
		return nil, err
	}

	updated, err := qtx.TransitionOrderStatus(ctx, sqlc.TransitionOrderStatusParams{
		NewStatus:          sqlc.OrderStatusCancelled,
		CancellationReason: sql.NullString{String: reason, Valid: true},
		CancelledBy:        sqlc.NullActorType{ActorType: sqlc.ActorTypePlatformAdmin, Valid: true},
		RejectionReason:    sql.NullString{},
		RejectedBy:         sqlc.NullActorType{},
		ID:                 orderID,
		TenantID:           tenantID,
	})
	if err != nil {
		return nil, apperror.Internal("force cancel order", err)
	}

	_, err = qtx.AddTimelineEvent(ctx, sqlc.AddTimelineEventParams{
		OrderID:        orderID,
		TenantID:       tenantID,
		EventType:      "status_changed",
		PreviousStatus: sqlc.NullOrderStatus{OrderStatus: order.Status, Valid: true},
		NewStatus:      sqlc.NullOrderStatus{OrderStatus: sqlc.OrderStatusCancelled, Valid: true},
		Description:    "Order force-cancelled by admin: " + reason,
		ActorID:        pgtype.UUID{Bytes: adminID, Valid: true},
		ActorType:      sqlc.ActorTypePlatformAdmin,
		Metadata:       json.RawMessage("{}"),
	})
	if err != nil {
		return nil, apperror.Internal("add timeline event", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperror.Internal("commit transaction", err)
	}

	return &updated, nil
}

// AutoConfirmOrders scans for CREATED orders past auto-confirm timeout and transitions them.
func (s *Service) AutoConfirmOrders(ctx context.Context, batchSize int32) (int, error) {
	orders, err := s.q.ListPendingAutoConfirmOrders(ctx, batchSize)
	if err != nil {
		return 0, apperror.Internal("list pending auto-confirm orders", err)
	}

	confirmed := 0
	for _, order := range orders {
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			continue
		}
		qtx := s.q.WithTx(tx)

		_, err = qtx.TransitionOrderStatus(ctx, sqlc.TransitionOrderStatusParams{
			NewStatus:          sqlc.OrderStatusConfirmed,
			CancellationReason: sql.NullString{},
			CancelledBy:        sqlc.NullActorType{},
			RejectionReason:    sql.NullString{},
			RejectedBy:         sqlc.NullActorType{},
			ID:                 order.ID,
			TenantID:           order.TenantID,
		})
		if err != nil {
			tx.Rollback(ctx)
			continue
		}

		// Transition all pickups to confirmed
		pickups, err := qtx.GetOrderPickupsByOrder(ctx, order.ID)
		if err != nil {
			tx.Rollback(ctx)
			continue
		}
		for _, p := range pickups {
			if p.Status == sqlc.PickupStatusNew {
				_, _ = qtx.TransitionPickupStatus(ctx, sqlc.TransitionPickupStatusParams{
					NewStatus:       sqlc.PickupStatusConfirmed,
					OrderID:         order.ID,
					RestaurantID:    p.RestaurantID,
					RejectionReason: sql.NullString{},
				})
			}
		}

		_, err = qtx.AddTimelineEvent(ctx, sqlc.AddTimelineEventParams{
			OrderID:        order.ID,
			TenantID:       order.TenantID,
			EventType:      "status_changed",
			PreviousStatus: sqlc.NullOrderStatus{OrderStatus: sqlc.OrderStatusCreated, Valid: true},
			NewStatus:      sqlc.NullOrderStatus{OrderStatus: sqlc.OrderStatusConfirmed, Valid: true},
			Description:    "Order auto-confirmed by system",
			ActorID:        pgtype.UUID{},
			ActorType:      sqlc.ActorTypeSystem,
			Metadata:       json.RawMessage("{}"),
		})
		if err != nil {
			tx.Rollback(ctx)
			continue
		}

		if err := tx.Commit(ctx); err != nil {
			continue
		}
		confirmed++
	}

	return confirmed, nil
}

// ConfirmPayment transitions a PENDING order to CREATED after payment is confirmed (online payments).
func (s *Service) ConfirmPayment(ctx context.Context, tenantID, orderID uuid.UUID) (*sqlc.Order, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, apperror.Internal("begin transaction", err)
	}
	defer tx.Rollback(ctx)
	qtx := s.q.WithTx(tx)

	order, err := qtx.GetOrderForUpdate(ctx, sqlc.GetOrderForUpdateParams{
		ID:       orderID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("order")
	}
	if err != nil {
		return nil, apperror.Internal("get order", err)
	}

	if order.Status != sqlc.OrderStatusPending {
		return nil, apperror.BadRequest("order must be in PENDING status to confirm payment")
	}

	updated, err := qtx.TransitionOrderStatus(ctx, sqlc.TransitionOrderStatusParams{
		NewStatus:          sqlc.OrderStatusCreated,
		CancellationReason: sql.NullString{},
		CancelledBy:        sqlc.NullActorType{},
		RejectionReason:    sql.NullString{},
		RejectedBy:         sqlc.NullActorType{},
		ID:                 orderID,
		TenantID:           tenantID,
	})
	if err != nil {
		return nil, apperror.Internal("transition to created", err)
	}

	_, err = qtx.UpdateOrderPaymentStatus(ctx, sqlc.UpdateOrderPaymentStatusParams{
		PaymentStatus: sqlc.PaymentStatusPaid,
		ID:            orderID,
		TenantID:      tenantID,
	})
	if err != nil {
		return nil, apperror.Internal("update payment status", err)
	}

	_, err = qtx.AddTimelineEvent(ctx, sqlc.AddTimelineEventParams{
		OrderID:        orderID,
		TenantID:       tenantID,
		EventType:      "payment_confirmed",
		PreviousStatus: sqlc.NullOrderStatus{OrderStatus: sqlc.OrderStatusPending, Valid: true},
		NewStatus:      sqlc.NullOrderStatus{OrderStatus: sqlc.OrderStatusCreated, Valid: true},
		Description:    "Payment confirmed, order created",
		ActorID:        pgtype.UUID{},
		ActorType:      sqlc.ActorTypeSystem,
		Metadata:       json.RawMessage("{}"),
	})
	if err != nil {
		return nil, apperror.Internal("add timeline event", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperror.Internal("commit transaction", err)
	}

	return &updated, nil
}

// FailPayment handles failed payment: release stock, soft-delete order.
func (s *Service) FailPayment(ctx context.Context, tenantID, orderID uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return apperror.Internal("begin transaction", err)
	}
	defer tx.Rollback(ctx)
	qtx := s.q.WithTx(tx)

	order, err := qtx.GetOrderForUpdate(ctx, sqlc.GetOrderForUpdateParams{
		ID:       orderID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.NotFound("order")
	}
	if err != nil {
		return apperror.Internal("get order", err)
	}

	if order.Status != sqlc.OrderStatusPending {
		return apperror.BadRequest("only PENDING orders can have payment failure")
	}

	// Release stock
	items, err := qtx.GetOrderItemsByOrder(ctx, orderID)
	if err != nil {
		return apperror.Internal("get order items", err)
	}
	stockReleases := make([]inventory.StockReservation, 0, len(items))
	for _, item := range items {
		stockReleases = append(stockReleases, inventory.StockReservation{
			ProductID:    item.ProductID,
			RestaurantID: item.RestaurantID,
			Quantity:     item.Quantity,
		})
	}
	if err := s.invSvc.ReleaseStockForOrder(ctx, qtx, tenantID, stockReleases); err != nil {
		return err
	}

	// Soft-delete the order
	if err := qtx.SoftDeleteOrder(ctx, sqlc.SoftDeleteOrderParams{
		ID:       orderID,
		TenantID: tenantID,
	}); err != nil {
		return apperror.Internal("soft delete order", err)
	}

	_, err = qtx.AddTimelineEvent(ctx, sqlc.AddTimelineEventParams{
		OrderID:        orderID,
		TenantID:       tenantID,
		EventType:      "payment_failed",
		PreviousStatus: sqlc.NullOrderStatus{OrderStatus: sqlc.OrderStatusPending, Valid: true},
		NewStatus:      sqlc.NullOrderStatus{OrderStatus: sqlc.OrderStatusCancelled, Valid: true},
		Description:    "Payment failed, order cancelled",
		ActorID:        pgtype.UUID{},
		ActorType:      sqlc.ActorTypeSystem,
		Metadata:       json.RawMessage("{}"),
	})
	if err != nil {
		return apperror.Internal("add timeline event", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return apperror.Internal("commit transaction", err)
	}

	return nil
}

// GetOrderItemsByRestaurant returns items for a specific restaurant in an order.
func (s *Service) GetOrderItemsByRestaurant(ctx context.Context, orderID, restaurantID uuid.UUID) ([]sqlc.OrderItem, error) {
	return s.q.GetOrderItemsByRestaurant(ctx, sqlc.GetOrderItemsByRestaurantParams{
		OrderID:      orderID,
		RestaurantID: restaurantID,
	})
}
