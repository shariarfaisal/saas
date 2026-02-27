package finance

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
		grossSales = grossSales.Add(p.ItemsSubtotal)
		itemDiscounts = itemDiscounts.Add(p.ItemsDiscount)
		vatCollected = vatCollected.Add(p.ItemsVat)
		commissionTotal = commissionTotal.Add(p.CommissionAmount)
		commissionRate = p.CommissionRate // Use last known rate
	}

	vendorPromoDiscounts := decimal.Zero
	netSales := grossSales.Sub(itemDiscounts).Sub(vendorPromoDiscounts)
	penaltyAmount := decimal.Zero
	adjustmentAmount := decimal.Zero

	// net_payable = net_sales + vat_collected - commission_amount - penalty - adjustment
	netPayable := netSales.Add(vatCollected).Sub(commissionTotal).Sub(penaltyAmount).Sub(adjustmentAmount)

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
		GrossSales:           grossSales,
		ItemDiscounts:        itemDiscounts,
		VendorPromoDiscounts: vendorPromoDiscounts,
		NetSales:             netSales,
		VatCollected:         vatCollected,
		CommissionRate:       commissionRate,
		CommissionAmount:     commissionTotal,
		PenaltyAmount:        penaltyAmount,
		AdjustmentAmount:     adjustmentAmount,
		NetPayable:           netPayable,
		TotalOrders:          counts.TotalOrders,
		DeliveredOrders:      counts.DeliveredOrders,
		CancelledOrders:      counts.CancelledOrders,
		RejectedOrders:       counts.RejectedOrders,
		Status:               sqlc.InvoiceStatusDraft,
		GeneratedBy:          generatedBy,
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
		FinalizedBy: &actorID,
		Notes:       notes,
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
		PaidBy:           &actorID,
		PaymentReference: &paymentReference,
		Notes:            reason,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("invoice not found or not in finalized status")
	}
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// pgDateFromTime converts a time.Time to pgtype.Date for SQLC.
func pgDateFromTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
