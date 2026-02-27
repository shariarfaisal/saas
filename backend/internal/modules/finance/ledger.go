package finance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/shopspring/decimal"
)

// Predefined ledger account codes.
const (
	AccountCustomerWallet     = "CUSTOMER_WALLET"
	AccountPlatformCommission = "PLATFORM_COMMISSION"
	AccountVendorPayable      = "VENDOR_PAYABLE"
	AccountRefundLiability    = "REFUND_LIABILITY"
	AccountDeliveryFee        = "DELIVERY_FEE"
)

// LedgerService provides append-only ledger operations.
type LedgerService struct {
	q *sqlc.Queries
}

// NewLedgerService creates a new ledger service.
func NewLedgerService(q *sqlc.Queries) *LedgerService {
	return &LedgerService{q: q}
}

// Record creates a new ledger entry. Entries are append-only (never updated).
func (s *LedgerService) Record(ctx context.Context, tenantID *uuid.UUID, accountCode string, entryType sqlc.LedgerEntryType, refType string, refID uuid.UUID, debit, credit decimal.Decimal, description string, metadata map[string]interface{}) (*sqlc.LedgerEntry, error) {
	// Look up account by code
	account, err := s.q.GetLedgerAccountByCode(ctx, accountCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("ledger account not found: %s", accountCode)
		}
		return nil, fmt.Errorf("get ledger account: %w", err)
	}

	// Get current balance
	lastBalance := decimal.Zero
	balance, err := s.q.GetLastLedgerEntryBalance(ctx, account.ID)
	if err == nil {
		lastBalance = balance
	}

	newBalance := lastBalance.Add(debit).Sub(credit)

	metadataJSON, _ := json.Marshal(metadata)

	entry, err := s.q.CreateLedgerEntry(ctx, sqlc.CreateLedgerEntryParams{
		TenantID:      tenantID,
		AccountID:     account.ID,
		EntryType:     entryType,
		ReferenceType: refType,
		ReferenceID:   refID,
		Debit:         debit,
		Credit:        credit,
		BalanceAfter:  newBalance,
		Description:   description,
		Metadata:      metadataJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("create ledger entry: %w", err)
	}
	return &entry, nil
}

// SeedPlatformAccounts creates the standard platform ledger accounts if they don't exist.
func (s *LedgerService) SeedPlatformAccounts(ctx context.Context) error {
	accounts := []struct {
		Code        string
		Name        string
		AccountType sqlc.LedgerAccountType
		Description string
	}{
		{AccountCustomerWallet, "Customer Wallet", sqlc.LedgerAccountTypeLiability, "Customer wallet balances (owed to customers)"},
		{AccountPlatformCommission, "Platform Commission", sqlc.LedgerAccountTypeRevenue, "Commission earned from restaurant orders"},
		{AccountVendorPayable, "Vendor Payable", sqlc.LedgerAccountTypeLiability, "Amounts owed to restaurant vendors"},
		{AccountRefundLiability, "Refund Liability", sqlc.LedgerAccountTypeLiability, "Pending refund obligations"},
		{AccountDeliveryFee, "Delivery Fee", sqlc.LedgerAccountTypeRevenue, "Delivery fee revenue"},
	}

	for _, a := range accounts {
		_, err := s.q.GetLedgerAccountByCode(ctx, a.Code)
		if errors.Is(err, pgx.ErrNoRows) {
			_, err = s.q.CreateLedgerAccount(ctx, sqlc.CreateLedgerAccountParams{
				Code:        a.Code,
				Name:        a.Name,
				AccountType: a.AccountType,
				Description: &a.Description,
				IsSystem:    true,
			})
			if err != nil {
				return fmt.Errorf("seed account %s: %w", a.Code, err)
			}
		} else if err != nil {
			return fmt.Errorf("check account %s: %w", a.Code, err)
		}
	}
	return nil
}

// CreateAuditLog is a helper to create audit log entries for finance operations.
func (s *Service) CreateAuditLog(ctx context.Context, tenantID, actorID uuid.UUID, action, resourceType string, resourceID uuid.UUID, reason string) {
	changes, _ := json.Marshal(map[string]interface{}{})
	s.q.CreateAuditLog(ctx, sqlc.CreateAuditLogParams{
		TenantID:     &tenantID,
		ActorID:      &actorID,
		ActorType:    sqlc.ActorTypePlatformAdmin,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   &resourceID,
		Changes:      changes,
		Reason:       &reason,
	})
}
