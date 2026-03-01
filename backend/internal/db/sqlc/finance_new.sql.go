package sqlc

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ---- Invoice Adjustments ----

const createInvoiceAdjustment = `-- name: CreateInvoiceAdjustment :one
INSERT INTO invoice_adjustments (invoice_id, amount, direction, reason, created_by_admin_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, invoice_id, amount, direction, reason, created_by_admin_id, created_at
`

type CreateInvoiceAdjustmentParams struct {
	InvoiceID        uuid.UUID      `json:"invoice_id"`
	Amount           pgtype.Numeric `json:"amount"`
	Direction        string         `json:"direction"`
	Reason           string         `json:"reason"`
	CreatedByAdminID pgtype.UUID    `json:"created_by_admin_id"`
}

func (q *Queries) CreateInvoiceAdjustment(ctx context.Context, arg CreateInvoiceAdjustmentParams) (InvoiceAdjustment, error) {
	row := q.db.QueryRow(ctx, createInvoiceAdjustment,
		arg.InvoiceID,
		arg.Amount,
		arg.Direction,
		arg.Reason,
		arg.CreatedByAdminID,
	)
	var i InvoiceAdjustment
	err := row.Scan(
		&i.ID,
		&i.InvoiceID,
		&i.Amount,
		&i.Direction,
		&i.Reason,
		&i.CreatedByAdminID,
		&i.CreatedAt,
	)
	return i, err
}

const listInvoiceAdjustments = `-- name: ListInvoiceAdjustments :many
SELECT id, invoice_id, amount, direction, reason, created_by_admin_id, created_at FROM invoice_adjustments WHERE invoice_id = $1 ORDER BY created_at DESC
`

func (q *Queries) ListInvoiceAdjustments(ctx context.Context, invoiceID uuid.UUID) ([]InvoiceAdjustment, error) {
	rows, err := q.db.Query(ctx, listInvoiceAdjustments, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []InvoiceAdjustment{}
	for rows.Next() {
		var i InvoiceAdjustment
		if err := rows.Scan(
			&i.ID,
			&i.InvoiceID,
			&i.Amount,
			&i.Direction,
			&i.Reason,
			&i.CreatedByAdminID,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const getInvoiceAdjustmentTotal = `-- name: GetInvoiceAdjustmentTotal :one
SELECT
    COALESCE(SUM(CASE WHEN direction = 'credit' THEN amount ELSE 0 END), 0) -
    COALESCE(SUM(CASE WHEN direction = 'debit' THEN amount ELSE 0 END), 0) AS net_adjustment
FROM invoice_adjustments
WHERE invoice_id = $1
`

func (q *Queries) GetInvoiceAdjustmentTotal(ctx context.Context, invoiceID uuid.UUID) (pgtype.Numeric, error) {
	row := q.db.QueryRow(ctx, getInvoiceAdjustmentTotal, invoiceID)
	var netAdjustment pgtype.Numeric
	err := row.Scan(&netAdjustment)
	return netAdjustment, err
}

// ---- Subscription Invoices ----

const createSubscriptionInvoice = `-- name: CreateSubscriptionInvoice :one
INSERT INTO subscription_invoices (tenant_id, amount, billing_period_start, billing_period_end, due_date)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, tenant_id, amount, status, billing_period_start, billing_period_end, due_date, paid_at, created_at
`

type CreateSubscriptionInvoiceParams struct {
	TenantID           uuid.UUID      `json:"tenant_id"`
	Amount             pgtype.Numeric `json:"amount"`
	BillingPeriodStart pgtype.Date    `json:"billing_period_start"`
	BillingPeriodEnd   pgtype.Date    `json:"billing_period_end"`
	DueDate            pgtype.Date    `json:"due_date"`
}

func (q *Queries) CreateSubscriptionInvoice(ctx context.Context, arg CreateSubscriptionInvoiceParams) (SubscriptionInvoice, error) {
	row := q.db.QueryRow(ctx, createSubscriptionInvoice,
		arg.TenantID,
		arg.Amount,
		arg.BillingPeriodStart,
		arg.BillingPeriodEnd,
		arg.DueDate,
	)
	var i SubscriptionInvoice
	err := row.Scan(
		&i.ID,
		&i.TenantID,
		&i.Amount,
		&i.Status,
		&i.BillingPeriodStart,
		&i.BillingPeriodEnd,
		&i.DueDate,
		&i.PaidAt,
		&i.CreatedAt,
	)
	return i, err
}

const listSubscriptionInvoices = `-- name: ListSubscriptionInvoices :many
SELECT id, tenant_id, amount, status, billing_period_start, billing_period_end, due_date, paid_at, created_at
FROM subscription_invoices
ORDER BY created_at DESC
LIMIT $1 OFFSET $2
`

type ListSubscriptionInvoicesParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListSubscriptionInvoices(ctx context.Context, arg ListSubscriptionInvoicesParams) ([]SubscriptionInvoice, error) {
	rows, err := q.db.Query(ctx, listSubscriptionInvoices, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []SubscriptionInvoice{}
	for rows.Next() {
		var i SubscriptionInvoice
		if err := rows.Scan(
			&i.ID,
			&i.TenantID,
			&i.Amount,
			&i.Status,
			&i.BillingPeriodStart,
			&i.BillingPeriodEnd,
			&i.DueDate,
			&i.PaidAt,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const getSubscriptionInvoiceByTenantAndPeriod = `-- name: GetSubscriptionInvoiceByTenantAndPeriod :one
SELECT id, tenant_id, amount, status, billing_period_start, billing_period_end, due_date, paid_at, created_at
FROM subscription_invoices
WHERE tenant_id = $1 AND billing_period_start = $2
LIMIT 1
`

type GetSubscriptionInvoiceByTenantAndPeriodParams struct {
	TenantID           uuid.UUID   `json:"tenant_id"`
	BillingPeriodStart pgtype.Date `json:"billing_period_start"`
}

func (q *Queries) GetSubscriptionInvoiceByTenantAndPeriod(ctx context.Context, arg GetSubscriptionInvoiceByTenantAndPeriodParams) (SubscriptionInvoice, error) {
	row := q.db.QueryRow(ctx, getSubscriptionInvoiceByTenantAndPeriod, arg.TenantID, arg.BillingPeriodStart)
	var i SubscriptionInvoice
	err := row.Scan(
		&i.ID,
		&i.TenantID,
		&i.Amount,
		&i.Status,
		&i.BillingPeriodStart,
		&i.BillingPeriodEnd,
		&i.DueDate,
		&i.PaidAt,
		&i.CreatedAt,
	)
	return i, err
}

// ---- Cash Collection Records ----

const createCashCollectionRecord = `-- name: CreateCashCollectionRecord :one
INSERT INTO cash_collection_records (tenant_id, rider_id, order_id, amount)
VALUES ($1, $2, $3, $4)
RETURNING id, tenant_id, rider_id, order_id, amount, status, collected_at, remitted_at
`

type CreateCashCollectionRecordParams struct {
	TenantID uuid.UUID      `json:"tenant_id"`
	RiderID  uuid.UUID      `json:"rider_id"`
	OrderID  uuid.UUID      `json:"order_id"`
	Amount   pgtype.Numeric `json:"amount"`
}

func (q *Queries) CreateCashCollectionRecord(ctx context.Context, arg CreateCashCollectionRecordParams) (CashCollectionRecord, error) {
	row := q.db.QueryRow(ctx, createCashCollectionRecord,
		arg.TenantID,
		arg.RiderID,
		arg.OrderID,
		arg.Amount,
	)
	var i CashCollectionRecord
	err := row.Scan(
		&i.ID,
		&i.TenantID,
		&i.RiderID,
		&i.OrderID,
		&i.Amount,
		&i.Status,
		&i.CollectedAt,
		&i.RemittedAt,
	)
	return i, err
}

const listCashCollectionRecords = `-- name: ListCashCollectionRecords :many
SELECT id, tenant_id, rider_id, order_id, amount, status, collected_at, remitted_at
FROM cash_collection_records
WHERE tenant_id = $1
ORDER BY collected_at DESC
LIMIT $2 OFFSET $3
`

type ListCashCollectionRecordsParams struct {
	TenantID uuid.UUID `json:"tenant_id"`
	Limit    int32     `json:"limit"`
	Offset   int32     `json:"offset"`
}

func (q *Queries) ListCashCollectionRecords(ctx context.Context, arg ListCashCollectionRecordsParams) ([]CashCollectionRecord, error) {
	rows, err := q.db.Query(ctx, listCashCollectionRecords, arg.TenantID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []CashCollectionRecord{}
	for rows.Next() {
		var i CashCollectionRecord
		if err := rows.Scan(
			&i.ID,
			&i.TenantID,
			&i.RiderID,
			&i.OrderID,
			&i.Amount,
			&i.Status,
			&i.CollectedAt,
			&i.RemittedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const remitCashCollectionByRider = `-- name: RemitCashCollectionByRider :exec
UPDATE cash_collection_records
SET status = 'remitted', remitted_at = NOW()
WHERE rider_id = $1 AND status = 'collected'
`

func (q *Queries) RemitCashCollectionByRider(ctx context.Context, riderID uuid.UUID) error {
	_, err := q.db.Exec(ctx, remitCashCollectionByRider, riderID)
	return err
}

const markOverdueCashCollections = `-- name: MarkOverdueCashCollections :exec
UPDATE cash_collection_records
SET status = 'overdue'
WHERE status = 'collected' AND collected_at < $1
`

func (q *Queries) MarkOverdueCashCollections(ctx context.Context, before time.Time) error {
	_, err := q.db.Exec(ctx, markOverdueCashCollections, before)
	return err
}

// ---- Reconciliation Alerts ----

const createReconciliationAlert = `-- name: CreateReconciliationAlert :one
INSERT INTO reconciliation_alerts (tenant_id, payment_transaction_id, alert_type)
VALUES ($1, $2, $3)
RETURNING id, tenant_id, payment_transaction_id, alert_type, status, resolution_notes, resolved_by, resolved_at, created_at
`

type CreateReconciliationAlertParams struct {
	TenantID             pgtype.UUID `json:"tenant_id"`
	PaymentTransactionID pgtype.UUID `json:"payment_transaction_id"`
	AlertType            string      `json:"alert_type"`
}

func (q *Queries) CreateReconciliationAlert(ctx context.Context, arg CreateReconciliationAlertParams) (ReconciliationAlert, error) {
	row := q.db.QueryRow(ctx, createReconciliationAlert,
		arg.TenantID,
		arg.PaymentTransactionID,
		arg.AlertType,
	)
	var i ReconciliationAlert
	err := row.Scan(
		&i.ID,
		&i.TenantID,
		&i.PaymentTransactionID,
		&i.AlertType,
		&i.Status,
		&i.ResolutionNotes,
		&i.ResolvedBy,
		&i.ResolvedAt,
		&i.CreatedAt,
	)
	return i, err
}

const listOpenReconciliationAlerts = `-- name: ListOpenReconciliationAlerts :many
SELECT id, tenant_id, payment_transaction_id, alert_type, status, resolution_notes, resolved_by, resolved_at, created_at
FROM reconciliation_alerts
WHERE status = 'open'
ORDER BY created_at DESC
LIMIT $1 OFFSET $2
`

type ListOpenReconciliationAlertsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListOpenReconciliationAlerts(ctx context.Context, arg ListOpenReconciliationAlertsParams) ([]ReconciliationAlert, error) {
	rows, err := q.db.Query(ctx, listOpenReconciliationAlerts, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ReconciliationAlert{}
	for rows.Next() {
		var i ReconciliationAlert
		if err := rows.Scan(
			&i.ID,
			&i.TenantID,
			&i.PaymentTransactionID,
			&i.AlertType,
			&i.Status,
			&i.ResolutionNotes,
			&i.ResolvedBy,
			&i.ResolvedAt,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const resolveReconciliationAlert = `-- name: ResolveReconciliationAlert :one
UPDATE reconciliation_alerts
SET status = 'resolved', resolution_notes = $2, resolved_by = $3, resolved_at = NOW()
WHERE id = $1 AND status = 'open'
RETURNING id, tenant_id, payment_transaction_id, alert_type, status, resolution_notes, resolved_by, resolved_at, created_at
`

type ResolveReconciliationAlertParams struct {
	ID              uuid.UUID      `json:"id"`
	ResolutionNotes sql.NullString `json:"resolution_notes"`
	ResolvedBy      pgtype.UUID    `json:"resolved_by"`
}

func (q *Queries) ResolveReconciliationAlert(ctx context.Context, arg ResolveReconciliationAlertParams) (ReconciliationAlert, error) {
	row := q.db.QueryRow(ctx, resolveReconciliationAlert,
		arg.ID,
		arg.ResolutionNotes,
		arg.ResolvedBy,
	)
	var i ReconciliationAlert
	err := row.Scan(
		&i.ID,
		&i.TenantID,
		&i.PaymentTransactionID,
		&i.AlertType,
		&i.Status,
		&i.ResolutionNotes,
		&i.ResolvedBy,
		&i.ResolvedAt,
		&i.CreatedAt,
	)
	return i, err
}

// ---- Finance Calculation Queries ----

const getVendorPromoDiscountsForInvoice = `-- name: GetVendorPromoDiscountsForInvoice :one
SELECT COALESCE(SUM(pu.discount_amount), 0)::NUMERIC AS total
FROM promo_usages pu
JOIN promos p ON p.id = pu.promo_id
JOIN order_items oi ON oi.order_id = pu.order_id AND oi.restaurant_id = $1
WHERE pu.tenant_id = $2
  AND pu.created_at BETWEEN $3 AND $4
  AND p.funded_by = 'vendor'
`

type GetVendorPromoDiscountsForInvoiceParams struct {
	RestaurantID uuid.UUID `json:"restaurant_id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
}

func (q *Queries) GetVendorPromoDiscountsForInvoice(ctx context.Context, arg GetVendorPromoDiscountsForInvoiceParams) (pgtype.Numeric, error) {
	row := q.db.QueryRow(ctx, getVendorPromoDiscountsForInvoice,
		arg.RestaurantID,
		arg.TenantID,
		arg.PeriodStart,
		arg.PeriodEnd,
	)
	var total pgtype.Numeric
	err := row.Scan(&total)
	return total, err
}

const getPenaltyAmountForInvoice = `-- name: GetPenaltyAmountForInvoice :one
SELECT COALESCE(SUM(oi2.restaurant_penalty_amount), 0)::NUMERIC AS total
FROM order_issues oi2
JOIN orders o ON o.id = oi2.order_id
JOIN order_items oi ON oi.order_id = o.id AND oi.restaurant_id = $1
WHERE oi2.status = 'resolved'
  AND oi2.created_at BETWEEN $2 AND $3
  AND o.tenant_id = $4
`

type GetPenaltyAmountForInvoiceParams struct {
	RestaurantID uuid.UUID `json:"restaurant_id"`
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
	TenantID     uuid.UUID `json:"tenant_id"`
}

func (q *Queries) GetPenaltyAmountForInvoice(ctx context.Context, arg GetPenaltyAmountForInvoiceParams) (pgtype.Numeric, error) {
	row := q.db.QueryRow(ctx, getPenaltyAmountForInvoice,
		arg.RestaurantID,
		arg.PeriodStart,
		arg.PeriodEnd,
		arg.TenantID,
	)
	var total pgtype.Numeric
	err := row.Scan(&total)
	return total, err
}

const getDeliveryChargeTotalForInvoice = `-- name: GetDeliveryChargeTotalForInvoice :one
SELECT COALESCE(SUM(o.delivery_charge), 0)::NUMERIC AS total
FROM orders o
JOIN order_pickups op ON op.order_id = o.id AND op.restaurant_id = $1
JOIN restaurants r ON r.id = $1
WHERE o.tenant_id = $2
  AND o.status = 'delivered'
  AND o.created_at BETWEEN $3 AND $4
  AND r.delivery_managed_by = 'platform'
`

type GetDeliveryChargeTotalForInvoiceParams struct {
	RestaurantID uuid.UUID `json:"restaurant_id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
}

func (q *Queries) GetDeliveryChargeTotalForInvoice(ctx context.Context, arg GetDeliveryChargeTotalForInvoiceParams) (pgtype.Numeric, error) {
	row := q.db.QueryRow(ctx, getDeliveryChargeTotalForInvoice,
		arg.RestaurantID,
		arg.TenantID,
		arg.PeriodStart,
		arg.PeriodEnd,
	)
	var total pgtype.Numeric
	err := row.Scan(&total)
	return total, err
}

// ---- Rider Payout Queries ----

const createRiderPayout = `-- name: CreateRiderPayout :one
INSERT INTO rider_payouts (rider_id, tenant_id, amount, earnings_from, earnings_to, payment_method, status)
VALUES ($1, $2, $3, $4, $5, $6, 'pending')
RETURNING id, rider_id, tenant_id, amount, earnings_from, earnings_to, payment_method, payment_reference, status, processed_by, processed_at, note, created_at, updated_at
`

type CreateRiderPayoutParams struct {
	RiderID       uuid.UUID      `json:"rider_id"`
	TenantID      uuid.UUID      `json:"tenant_id"`
	Amount        pgtype.Numeric `json:"amount"`
	EarningsFrom  pgtype.Date    `json:"earnings_from"`
	EarningsTo    pgtype.Date    `json:"earnings_to"`
	PaymentMethod string         `json:"payment_method"`
}

func (q *Queries) CreateRiderPayout(ctx context.Context, arg CreateRiderPayoutParams) (RiderPayout, error) {
	row := q.db.QueryRow(ctx, createRiderPayout,
		arg.RiderID,
		arg.TenantID,
		arg.Amount,
		arg.EarningsFrom,
		arg.EarningsTo,
		arg.PaymentMethod,
	)
	var i RiderPayout
	err := row.Scan(
		&i.ID,
		&i.RiderID,
		&i.TenantID,
		&i.Amount,
		&i.EarningsFrom,
		&i.EarningsTo,
		&i.PaymentMethod,
		&i.PaymentReference,
		&i.Status,
		&i.ProcessedBy,
		&i.ProcessedAt,
		&i.Note,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listRiderPayouts = `-- name: ListRiderPayouts :many
SELECT id, rider_id, tenant_id, amount, earnings_from, earnings_to, payment_method, payment_reference, status, processed_by, processed_at, note, created_at, updated_at
FROM rider_payouts
ORDER BY created_at DESC
LIMIT $1 OFFSET $2
`

type ListRiderPayoutsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListRiderPayouts(ctx context.Context, arg ListRiderPayoutsParams) ([]RiderPayout, error) {
	rows, err := q.db.Query(ctx, listRiderPayouts, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []RiderPayout{}
	for rows.Next() {
		var i RiderPayout
		if err := rows.Scan(
			&i.ID,
			&i.RiderID,
			&i.TenantID,
			&i.Amount,
			&i.EarningsFrom,
			&i.EarningsTo,
			&i.PaymentMethod,
			&i.PaymentReference,
			&i.Status,
			&i.ProcessedBy,
			&i.ProcessedAt,
			&i.Note,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const approveRiderPayout = `-- name: ApproveRiderPayout :one
UPDATE rider_payouts
SET status = 'processing', processed_by = $2, processed_at = NOW()
WHERE id = $1 AND status = 'pending'
RETURNING id, rider_id, tenant_id, amount, earnings_from, earnings_to, payment_method, payment_reference, status, processed_by, processed_at, note, created_at, updated_at
`

type ApproveRiderPayoutParams struct {
	ID          uuid.UUID   `json:"id"`
	ProcessedBy pgtype.UUID `json:"processed_by"`
}

func (q *Queries) ApproveRiderPayout(ctx context.Context, arg ApproveRiderPayoutParams) (RiderPayout, error) {
	row := q.db.QueryRow(ctx, approveRiderPayout, arg.ID, arg.ProcessedBy)
	var i RiderPayout
	err := row.Scan(
		&i.ID,
		&i.RiderID,
		&i.TenantID,
		&i.Amount,
		&i.EarningsFrom,
		&i.EarningsTo,
		&i.PaymentMethod,
		&i.PaymentReference,
		&i.Status,
		&i.ProcessedBy,
		&i.ProcessedAt,
		&i.Note,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listUnpaidEarningsByRider = `-- name: ListUnpaidEarningsByRider :many
SELECT id, rider_id, tenant_id, order_id, base_earning, distance_bonus, peak_bonus, tip_amount, total_earning, is_paid_out, payout_id, created_at
FROM rider_earnings
WHERE rider_id = $1 AND is_paid_out = false
ORDER BY created_at ASC
`

func (q *Queries) ListUnpaidEarningsByRider(ctx context.Context, riderID uuid.UUID) ([]RiderEarning, error) {
	rows, err := q.db.Query(ctx, listUnpaidEarningsByRider, riderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []RiderEarning{}
	for rows.Next() {
		var i RiderEarning
		if err := rows.Scan(
			&i.ID,
			&i.RiderID,
			&i.TenantID,
			&i.OrderID,
			&i.BaseEarning,
			&i.DistanceBonus,
			&i.PeakBonus,
			&i.TipAmount,
			&i.TotalEarning,
			&i.IsPaidOut,
			&i.PayoutID,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const linkEarningsToPayout = `-- name: LinkEarningsToPayout :exec
UPDATE rider_earnings
SET is_paid_out = true, payout_id = $2
WHERE rider_id = $1 AND is_paid_out = false
`

type LinkEarningsToPayoutParams struct {
	RiderID  uuid.UUID   `json:"rider_id"`
	PayoutID pgtype.UUID `json:"payout_id"`
}

func (q *Queries) LinkEarningsToPayout(ctx context.Context, arg LinkEarningsToPayoutParams) error {
	_, err := q.db.Exec(ctx, linkEarningsToPayout, arg.RiderID, arg.PayoutID)
	return err
}

const listRidersWithUnpaidEarnings = `-- name: ListRidersWithUnpaidEarnings :many
SELECT DISTINCT rider_id, tenant_id
FROM rider_earnings
WHERE is_paid_out = false
`

type ListRidersWithUnpaidEarningsRow struct {
	RiderID  uuid.UUID `json:"rider_id"`
	TenantID uuid.UUID `json:"tenant_id"`
}

func (q *Queries) ListRidersWithUnpaidEarnings(ctx context.Context) ([]ListRidersWithUnpaidEarningsRow, error) {
	rows, err := q.db.Query(ctx, listRidersWithUnpaidEarnings)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListRidersWithUnpaidEarningsRow{}
	for rows.Next() {
		var i ListRidersWithUnpaidEarningsRow
		if err := rows.Scan(&i.RiderID, &i.TenantID); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

// ---- Stale Payment Queries ----

const listStaleProcessingPayments = `-- name: ListStaleProcessingPayments :many
SELECT id, tenant_id, order_id, user_id, payment_method, status, amount, currency, gateway_transaction_id, gateway_reference_id, gateway_response, gateway_fee, ip_address, user_agent, callback_received_at, created_at, updated_at
FROM payment_transactions
WHERE status = 'processing' AND created_at < $1
ORDER BY created_at ASC
LIMIT $2
`

type ListStaleProcessingPaymentsParams struct {
	Before time.Time `json:"before"`
	Limit  int32     `json:"limit"`
}

func (q *Queries) ListStaleProcessingPayments(ctx context.Context, arg ListStaleProcessingPaymentsParams) ([]PaymentTransaction, error) {
	rows, err := q.db.Query(ctx, listStaleProcessingPayments, arg.Before, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []PaymentTransaction{}
	for rows.Next() {
		var i PaymentTransaction
		if err := rows.Scan(
			&i.ID,
			&i.TenantID,
			&i.OrderID,
			&i.UserID,
			&i.PaymentMethod,
			&i.Status,
			&i.Amount,
			&i.Currency,
			&i.GatewayTransactionID,
			&i.GatewayReferenceID,
			&i.GatewayResponse,
			&i.GatewayFee,
			&i.IpAddress,
			&i.UserAgent,
			&i.CallbackReceivedAt,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

// ---- Revenue Report Queries ----

const getCommissionRevenueSummary = `-- name: GetCommissionRevenueSummary :one
SELECT
    COALESCE(SUM(CASE WHEN DATE_TRUNC('month', o.delivered_at) = DATE_TRUNC('month', NOW()) THEN op.commission_amount ELSE 0 END), 0)::NUMERIC AS current_month,
    COALESCE(SUM(CASE WHEN DATE_TRUNC('month', o.delivered_at) = DATE_TRUNC('month', NOW() - INTERVAL '1 month') THEN op.commission_amount ELSE 0 END), 0)::NUMERIC AS previous_month
FROM order_pickups op
JOIN orders o ON o.id = op.order_id
WHERE o.status = 'delivered'
  AND o.delivered_at >= DATE_TRUNC('month', NOW() - INTERVAL '1 month')
`

type GetCommissionRevenueSummaryRow struct {
	CurrentMonth  pgtype.Numeric `json:"current_month"`
	PreviousMonth pgtype.Numeric `json:"previous_month"`
}

func (q *Queries) GetCommissionRevenueSummary(ctx context.Context) (GetCommissionRevenueSummaryRow, error) {
	row := q.db.QueryRow(ctx, getCommissionRevenueSummary)
	var i GetCommissionRevenueSummaryRow
	err := row.Scan(&i.CurrentMonth, &i.PreviousMonth)
	return i, err
}

const getDeliveryFeeRevenueSummary = `-- name: GetDeliveryFeeRevenueSummary :one
SELECT
    COALESCE(SUM(CASE WHEN DATE_TRUNC('month', o.delivered_at) = DATE_TRUNC('month', NOW()) THEN o.delivery_charge ELSE 0 END), 0)::NUMERIC AS current_month,
    COALESCE(SUM(CASE WHEN DATE_TRUNC('month', o.delivered_at) = DATE_TRUNC('month', NOW() - INTERVAL '1 month') THEN o.delivery_charge ELSE 0 END), 0)::NUMERIC AS previous_month
FROM orders o
JOIN restaurants r ON r.id = (SELECT restaurant_id FROM order_pickups WHERE order_id = o.id LIMIT 1)
WHERE o.status = 'delivered'
  AND r.delivery_managed_by = 'platform'
  AND o.delivered_at >= DATE_TRUNC('month', NOW() - INTERVAL '1 month')
`

type GetDeliveryFeeRevenueSummaryRow struct {
	CurrentMonth  pgtype.Numeric `json:"current_month"`
	PreviousMonth pgtype.Numeric `json:"previous_month"`
}

func (q *Queries) GetDeliveryFeeRevenueSummary(ctx context.Context) (GetDeliveryFeeRevenueSummaryRow, error) {
	row := q.db.QueryRow(ctx, getDeliveryFeeRevenueSummary)
	var i GetDeliveryFeeRevenueSummaryRow
	err := row.Scan(&i.CurrentMonth, &i.PreviousMonth)
	return i, err
}

const listCommissionByRestaurant = `-- name: ListCommissionByRestaurant :many
SELECT op.restaurant_id, r.name AS restaurant_name,
    COALESCE(SUM(op.commission_amount), 0)::NUMERIC AS total_commission,
    COUNT(op.id)::INT AS order_count
FROM order_pickups op
JOIN restaurants r ON r.id = op.restaurant_id
JOIN orders o ON o.id = op.order_id
WHERE o.status = 'delivered'
  AND o.delivered_at >= $1 AND o.delivered_at < $2
GROUP BY op.restaurant_id, r.name
ORDER BY total_commission DESC
LIMIT $3 OFFSET $4
`

type ListCommissionByRestaurantParams struct {
	From   time.Time `json:"from"`
	To     time.Time `json:"to"`
	Limit  int32     `json:"limit"`
	Offset int32     `json:"offset"`
}

type ListCommissionByRestaurantRow struct {
	RestaurantID    uuid.UUID      `json:"restaurant_id"`
	RestaurantName  string         `json:"restaurant_name"`
	TotalCommission pgtype.Numeric `json:"total_commission"`
	OrderCount      int32          `json:"order_count"`
}

func (q *Queries) ListCommissionByRestaurant(ctx context.Context, arg ListCommissionByRestaurantParams) ([]ListCommissionByRestaurantRow, error) {
	rows, err := q.db.Query(ctx, listCommissionByRestaurant, arg.From, arg.To, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListCommissionByRestaurantRow{}
	for rows.Next() {
		var i ListCommissionByRestaurantRow
		if err := rows.Scan(
			&i.RestaurantID,
			&i.RestaurantName,
			&i.TotalCommission,
			&i.OrderCount,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const listDeliveryFeesByTenant = `-- name: ListDeliveryFeesByTenant :many
SELECT o.tenant_id, t.name AS tenant_name,
    COALESCE(SUM(o.delivery_charge), 0)::NUMERIC AS total_delivery_fees,
    COUNT(o.id)::INT AS order_count
FROM orders o
JOIN tenants t ON t.id = o.tenant_id
JOIN order_pickups op ON op.order_id = o.id
JOIN restaurants r ON r.id = op.restaurant_id
WHERE o.status = 'delivered'
  AND r.delivery_managed_by = 'platform'
  AND o.delivered_at >= $1 AND o.delivered_at < $2
GROUP BY o.tenant_id, t.name
ORDER BY total_delivery_fees DESC
LIMIT $3 OFFSET $4
`

type ListDeliveryFeesByTenantParams struct {
	From   time.Time `json:"from"`
	To     time.Time `json:"to"`
	Limit  int32     `json:"limit"`
	Offset int32     `json:"offset"`
}

type ListDeliveryFeesByTenantRow struct {
	TenantID          uuid.UUID      `json:"tenant_id"`
	TenantName        string         `json:"tenant_name"`
	TotalDeliveryFees pgtype.Numeric `json:"total_delivery_fees"`
	OrderCount        int32          `json:"order_count"`
}

func (q *Queries) ListDeliveryFeesByTenant(ctx context.Context, arg ListDeliveryFeesByTenantParams) ([]ListDeliveryFeesByTenantRow, error) {
	rows, err := q.db.Query(ctx, listDeliveryFeesByTenant, arg.From, arg.To, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListDeliveryFeesByTenantRow{}
	for rows.Next() {
		var i ListDeliveryFeesByTenantRow
		if err := rows.Scan(
			&i.TenantID,
			&i.TenantName,
			&i.TotalDeliveryFees,
			&i.OrderCount,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

// ---- Tenant Listing for Subscription Billing ----

const listAllActiveTenants = `-- name: ListAllActiveTenants :many
SELECT id, slug, name, status, plan, subscription_plan_id, commission_rate, settings, custom_domain, logo_url, favicon_url, primary_color, secondary_color, contact_email, contact_phone, address, timezone, currency, locale, created_at, updated_at, billing_day
FROM tenants
WHERE status = 'active'
ORDER BY created_at ASC
`

func (q *Queries) ListAllActiveTenants(ctx context.Context) ([]Tenant, error) {
	rows, err := q.db.Query(ctx, listAllActiveTenants)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Tenant{}
	for rows.Next() {
		var i Tenant
		if err := rows.Scan(
			&i.ID,
			&i.Slug,
			&i.Name,
			&i.Status,
			&i.Plan,
			&i.SubscriptionPlanID,
			&i.CommissionRate,
			&i.Settings,
			&i.CustomDomain,
			&i.LogoUrl,
			&i.FaviconUrl,
			&i.PrimaryColor,
			&i.SecondaryColor,
			&i.ContactEmail,
			&i.ContactPhone,
			&i.Address,
			&i.Timezone,
			&i.Currency,
			&i.Locale,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.BillingDay,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}
