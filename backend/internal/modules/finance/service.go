package finance

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/pagination"
	"github.com/shopspring/decimal"
)

// Service implements invoice and finance business logic.
type Service struct {
	q *sqlc.Queries
}

// NewService creates a new finance service.
func NewService(q *sqlc.Queries) *Service {
	return &Service{q: q}
}

// GenerateForRestaurant generates an invoice for a restaurant for a given period.
// It is idempotent: if an invoice already exists for the same restaurant+period, it returns the existing one.
func (s *Service) GenerateForRestaurant(ctx context.Context, tenantID, restaurantID uuid.UUID, periodStart, periodEnd time.Time, generatedBy *uuid.UUID) (*sqlc.Invoice, error) {
	// Check for existing invoice (idempotent)
	existing, err := s.q.GetInvoiceByPeriod(ctx, sqlc.GetInvoiceByPeriodParams{
		TenantID:     tenantID,
		RestaurantID: restaurantID,
		PeriodStart:  pgDateFromTime(periodStart),
		PeriodEnd:    pgDateFromTime(periodEnd),
	})
	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("check existing invoice: %w", err)
	}

	// Get order pickups for this restaurant in the period
	pickups, err := s.q.GetOrderPickupsByRestaurantAndPeriod(ctx, sqlc.GetOrderPickupsByRestaurantAndPeriodParams{
		TenantID:     tenantID,
		RestaurantID: restaurantID,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
	})
	if err != nil {
		return nil, fmt.Errorf("get order pickups: %w", err)
	}

	// Calculate invoice amounts from pickups
	grossSales := decimal.Zero
	itemDiscounts := decimal.Zero
	vatCollected := decimal.Zero
	commissionTotal := decimal.Zero
	var commissionRate decimal.Decimal

	for _, p := range pickups {
		grossSales = grossSales.Add(pgNumericToDecimal(p.ItemsSubtotal))
		itemDiscounts = itemDiscounts.Add(pgNumericToDecimal(p.ItemsDiscount))
		vatCollected = vatCollected.Add(pgNumericToDecimal(p.ItemsVat))
		commissionTotal = commissionTotal.Add(pgNumericToDecimal(p.CommissionAmount))
		commissionRate = pgNumericToDecimal(p.CommissionRate)
	}

	vendorPromoDiscounts := decimal.Zero
	// Fetch vendor-funded promo discounts for the period
	if promoTotal, err := s.q.GetVendorPromoDiscountsForInvoice(ctx, sqlc.GetVendorPromoDiscountsForInvoiceParams{
		RestaurantID: restaurantID,
		TenantID:     tenantID,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
	}); err == nil {
		vendorPromoDiscounts = pgNumericToDecimal(promoTotal)
	}

	penaltyAmount := decimal.Zero
	// Fetch restaurant penalty amounts for the period
	if penaltyTotal, err := s.q.GetPenaltyAmountForInvoice(ctx, sqlc.GetPenaltyAmountForInvoiceParams{
		RestaurantID: restaurantID,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		TenantID:     tenantID,
	}); err == nil {
		penaltyAmount = pgNumericToDecimal(penaltyTotal)
	}

	deliveryChargeTotal := decimal.Zero
	// Fetch delivery charge total for platform-managed restaurants
	if dcTotal, err := s.q.GetDeliveryChargeTotalForInvoice(ctx, sqlc.GetDeliveryChargeTotalForInvoiceParams{
		RestaurantID: restaurantID,
		TenantID:     tenantID,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
	}); err == nil {
		deliveryChargeTotal = pgNumericToDecimal(dcTotal)
	}

	netSales := grossSales.Sub(itemDiscounts).Sub(vendorPromoDiscounts)
	adjustmentAmount := decimal.Zero

	// net_payable = gross_sales - item_discounts - vendor_promo_discounts - commission_amount - penalty_amount + adjustment_amount
	// Per design.md: VAT collected is tracked separately; not part of vendor net payable.
	// delivery_charge_total is retained by the platform when delivery_managed_by = 'platform'.
	netPayable := netSales.Sub(commissionTotal).Sub(penaltyAmount).Add(adjustmentAmount).Add(deliveryChargeTotal)

	// Get order counts for the period
	counts, err := s.q.CountOrdersByRestaurantAndPeriod(ctx, sqlc.CountOrdersByRestaurantAndPeriodParams{
		TenantID:    tenantID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	})
	if err != nil {
		return nil, fmt.Errorf("count orders: %w", err)
	}

	// Generate invoice number
	invoiceNumber := fmt.Sprintf("INV-%s-%s", periodStart.Format("20060102"), restaurantID.String()[:8])

	inv, err := s.q.CreateInvoice(ctx, sqlc.CreateInvoiceParams{
		TenantID:             tenantID,
		RestaurantID:         restaurantID,
		InvoiceNumber:        invoiceNumber,
		PeriodStart:          pgDateFromTime(periodStart),
		PeriodEnd:            pgDateFromTime(periodEnd),
		GrossSales:           toPgNumeric(grossSales),
		ItemDiscounts:        toPgNumeric(itemDiscounts),
		VendorPromoDiscounts: toPgNumeric(vendorPromoDiscounts),
		NetSales:             toPgNumeric(netSales),
		VatCollected:         toPgNumeric(vatCollected),
		CommissionRate:       toPgNumeric(commissionRate),
		CommissionAmount:     toPgNumeric(commissionTotal),
		PenaltyAmount:        toPgNumeric(penaltyAmount),
		AdjustmentAmount:     toPgNumeric(adjustmentAmount),
		NetPayable:           toPgNumeric(netPayable),
		TotalOrders:          counts.TotalOrders,
		DeliveredOrders:      counts.DeliveredOrders,
		CancelledOrders:      counts.CancelledOrders,
		RejectedOrders:       counts.RejectedOrders,
		Status:               sqlc.InvoiceStatusDraft,
		GeneratedBy:          toPgUUIDPtr(generatedBy),
		DeliveryChargeTotal:  toPgNumeric(deliveryChargeTotal),
	})
	if err != nil {
		return nil, fmt.Errorf("create invoice: %w", err)
	}
	return &inv, nil
}

// GetByID returns an invoice by ID.
func (s *Service) GetByID(ctx context.Context, tenantID, invoiceID uuid.UUID) (*sqlc.Invoice, error) {
	inv, err := s.q.GetInvoiceByID(ctx, sqlc.GetInvoiceByIDParams{
		ID:       invoiceID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("invoice")
	}
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// ListByTenant returns paginated invoices for a tenant.
func (s *Service) ListByTenant(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]sqlc.Invoice, pagination.Meta, error) {
	limit, offset := pagination.FormatLimitOffset(page, perPage)
	total, err := s.q.CountInvoicesByTenant(ctx, tenantID)
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("count invoices", err)
	}
	items, err := s.q.ListInvoicesByTenant(ctx, sqlc.ListInvoicesByTenantParams{
		TenantID: tenantID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("list invoices", err)
	}
	return items, pagination.NewMeta(total, limit, ""), nil
}

// GetFinanceSummary returns a finance summary for a tenant.
func (s *Service) GetFinanceSummary(ctx context.Context, tenantID uuid.UUID) (*sqlc.GetFinanceSummaryRow, error) {
	summary, err := s.q.GetFinanceSummary(ctx, tenantID)
	if err != nil {
		return nil, apperror.Internal("get finance summary", err)
	}
	return &summary, nil
}

// FinalizeInvoice marks an invoice as finalized.
func (s *Service) FinalizeInvoice(ctx context.Context, tenantID, invoiceID, actorID uuid.UUID, reason *string) (*sqlc.Invoice, error) {
	var notes *string
	if reason != nil {
		notes = reason
	}
	inv, err := s.q.FinalizeInvoice(ctx, sqlc.FinalizeInvoiceParams{
		ID:          invoiceID,
		TenantID:    tenantID,
		FinalizedBy: toPgUUID(actorID),
		Notes:       toNullString(notes),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("invoice not found or not in draft status")
	}
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// MarkPaid marks an invoice as paid.
func (s *Service) MarkPaid(ctx context.Context, tenantID, invoiceID, actorID uuid.UUID, paymentReference string, reason *string) (*sqlc.Invoice, error) {
	inv, err := s.q.MarkInvoicePaid(ctx, sqlc.MarkInvoicePaidParams{
		ID:               invoiceID,
		TenantID:         tenantID,
		PaidBy:           toPgUUID(actorID),
		PaymentReference: toNullStringVal(paymentReference),
		Notes:            toNullString(reason),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("invoice not found or not in finalized status")
	}
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// RevenueReport holds aggregated revenue summary data.
type RevenueReport struct {
	CommissionCurrentMonth  string `json:"commission_current_month"`
	CommissionPreviousMonth string `json:"commission_previous_month"`
	DeliveryCurrentMonth    string `json:"delivery_current_month"`
	DeliveryPreviousMonth   string `json:"delivery_previous_month"`
}

// CreateAdjustment adds an adjustment (credit or debit) to an invoice.
func (s *Service) CreateAdjustment(ctx context.Context, invoiceID, adminID uuid.UUID, amount decimal.Decimal, direction, reason string) (*sqlc.InvoiceAdjustment, error) {
	adj, err := s.q.CreateInvoiceAdjustment(ctx, sqlc.CreateInvoiceAdjustmentParams{
		InvoiceID:        invoiceID,
		Amount:           toPgNumeric(amount),
		Direction:        direction,
		Reason:           reason,
		CreatedByAdminID: toPgUUID(adminID),
	})
	if err != nil {
		return nil, fmt.Errorf("create invoice adjustment: %w", err)
	}
	return &adj, nil
}

// ListAdjustments returns all adjustments for an invoice.
func (s *Service) ListAdjustments(ctx context.Context, invoiceID uuid.UUID) ([]sqlc.InvoiceAdjustment, error) {
	items, err := s.q.ListInvoiceAdjustments(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("list invoice adjustments: %w", err)
	}
	return items, nil
}

// GetRevenueSummary returns aggregated commission and delivery fee revenue.
func (s *Service) GetRevenueSummary(ctx context.Context) (*RevenueReport, error) {
	comm, err := s.q.GetCommissionRevenueSummary(ctx)
	if err != nil {
		return nil, apperror.Internal("get commission summary", err)
	}
	del, err := s.q.GetDeliveryFeeRevenueSummary(ctx)
	if err != nil {
		return nil, apperror.Internal("get delivery fee summary", err)
	}
	return &RevenueReport{
		CommissionCurrentMonth:  pgNumericToDecimal(comm.CurrentMonth).String(),
		CommissionPreviousMonth: pgNumericToDecimal(comm.PreviousMonth).String(),
		DeliveryCurrentMonth:    pgNumericToDecimal(del.CurrentMonth).String(),
		DeliveryPreviousMonth:   pgNumericToDecimal(del.PreviousMonth).String(),
	}, nil
}

// ListRiderPayouts returns paginated rider payouts.
func (s *Service) ListRiderPayouts(ctx context.Context, page, perPage int) ([]sqlc.RiderPayout, pagination.Meta, error) {
	limit, offset := pagination.FormatLimitOffset(page, perPage)
	items, err := s.q.ListRiderPayouts(ctx, sqlc.ListRiderPayoutsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("list rider payouts", err)
	}
	return items, pagination.NewMeta(int64(len(items)), limit, ""), nil
}

// ApproveRiderPayout marks a rider payout as processing.
func (s *Service) ApproveRiderPayout(ctx context.Context, payoutID, adminID uuid.UUID) (*sqlc.RiderPayout, error) {
	p, err := s.q.ApproveRiderPayout(ctx, sqlc.ApproveRiderPayoutParams{
		ID:          payoutID,
		ProcessedBy: toPgUUID(adminID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("payout not found or not in pending status")
	}
	if err != nil {
		return nil, fmt.Errorf("approve rider payout: %w", err)
	}
	return &p, nil
}

// ListOpenAlerts returns paginated open reconciliation alerts.
func (s *Service) ListOpenAlerts(ctx context.Context, page, perPage int) ([]sqlc.ReconciliationAlert, pagination.Meta, error) {
	limit, offset := pagination.FormatLimitOffset(page, perPage)
	items, err := s.q.ListOpenReconciliationAlerts(ctx, sqlc.ListOpenReconciliationAlertsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("list reconciliation alerts", err)
	}
	return items, pagination.NewMeta(int64(len(items)), limit, ""), nil
}

// ResolveAlert resolves a reconciliation alert.
func (s *Service) ResolveAlert(ctx context.Context, alertID, adminID uuid.UUID, notes string) (*sqlc.ReconciliationAlert, error) {
	alert, err := s.q.ResolveReconciliationAlert(ctx, sqlc.ResolveReconciliationAlertParams{
		ID:              alertID,
		ResolutionNotes: toNullStringVal(notes),
		ResolvedBy:      toPgUUID(adminID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("alert not found or already resolved")
	}
	if err != nil {
		return nil, fmt.Errorf("resolve reconciliation alert: %w", err)
	}
	return &alert, nil
}

// ListCashCollections returns paginated cash collection records for a tenant.
func (s *Service) ListCashCollections(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]sqlc.CashCollectionRecord, pagination.Meta, error) {
	limit, offset := pagination.FormatLimitOffset(page, perPage)
	items, err := s.q.ListCashCollectionRecords(ctx, sqlc.ListCashCollectionRecordsParams{
		TenantID: tenantID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("list cash collections", err)
	}
	return items, pagination.NewMeta(int64(len(items)), limit, ""), nil
}

// ListSubscriptionInvoices returns paginated subscription invoices.
func (s *Service) ListSubscriptionInvoices(ctx context.Context, page, perPage int) ([]sqlc.SubscriptionInvoice, pagination.Meta, error) {
	limit, offset := pagination.FormatLimitOffset(page, perPage)
	items, err := s.q.ListSubscriptionInvoices(ctx, sqlc.ListSubscriptionInvoicesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("list subscription invoices", err)
	}
	return items, pagination.NewMeta(int64(len(items)), limit, ""), nil
}

// pgDateFromTime converts a time.Time to pgtype.Date for SQLC.
func pgDateFromTime(t time.Time) pgtype.Date {
	return pgtype.Date{Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), Valid: true}
}

func pgNumericToDecimal(n pgtype.Numeric) decimal.Decimal {
	if !n.Valid {
		return decimal.Zero
	}
	f, err := n.Float64Value()
	if err != nil {
		return decimal.Zero
	}
	return decimal.NewFromFloat(f.Float64)
}

func toPgNumeric(d decimal.Decimal) pgtype.Numeric {
	n := pgtype.Numeric{}
	_ = n.Scan(d.String())
	return n
}

func toPgUUIDPtr(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func toNullStringVal(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
