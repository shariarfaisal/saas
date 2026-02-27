package finance

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/shopspring/decimal"
)

// WalletService manages wallet operations.
type WalletService struct {
	q *sqlc.Queries
}

// NewWalletService creates a new wallet service.
func NewWalletService(q *sqlc.Queries) *WalletService {
	return &WalletService{q: q}
}

// Credit adds funds to a user's wallet.
func (s *WalletService) Credit(ctx context.Context, userID, tenantID uuid.UUID, orderID *uuid.UUID, source sqlc.WalletSource, amount decimal.Decimal, description string) (*sqlc.WalletTransaction, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, apperror.BadRequest("amount must be positive")
	}

	// Get current balance
	currentBalancePg, err := s.q.GetUserWalletBalance(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("user")
	}
	if err != nil {
		return nil, fmt.Errorf("get wallet balance: %w", err)
	}

	currentBalance := pgNumericToDecimal(currentBalancePg)
	newBalance := currentBalance.Add(amount)

	// Credit the user's wallet balance
	if err := s.q.CreditUserWallet(ctx, sqlc.CreditUserWalletParams{
		Amount: toPgNumeric(amount),
		ID:     userID,
	}); err != nil {
		return nil, fmt.Errorf("credit wallet: %w", err)
	}

	// Create wallet transaction record
	txn, err := s.q.CreateWalletTransaction(ctx, sqlc.CreateWalletTransactionParams{
		UserID:       userID,
		TenantID:     tenantID,
		OrderID:      toPgUUIDPtr(orderID),
		Type:         sqlc.WalletTypeCredit,
		Source:       source,
		Amount:       toPgNumeric(amount),
		BalanceAfter: toPgNumeric(newBalance),
		Description:  toNullStringVal(description),
	})
	if err != nil {
		return nil, fmt.Errorf("create wallet transaction: %w", err)
	}
	return &txn, nil
}

// Debit removes funds from a user's wallet.
func (s *WalletService) Debit(ctx context.Context, userID, tenantID uuid.UUID, orderID *uuid.UUID, source sqlc.WalletSource, amount decimal.Decimal, description string) (*sqlc.WalletTransaction, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, apperror.BadRequest("amount must be positive")
	}

	// Get current balance
	currentBalancePg, err := s.q.GetUserWalletBalance(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("user")
	}
	if err != nil {
		return nil, fmt.Errorf("get wallet balance: %w", err)
	}

	currentBalance := pgNumericToDecimal(currentBalancePg)
	if currentBalance.LessThan(amount) {
		return nil, apperror.BadRequest("insufficient wallet balance")
	}

	newBalance := currentBalance.Sub(amount)

	// Debit the user's wallet balance
	if err := s.q.DebitUserWallet(ctx, sqlc.DebitUserWalletParams{
		Amount: toPgNumeric(amount),
		ID:     userID,
	}); err != nil {
		return nil, fmt.Errorf("debit wallet: %w", err)
	}

	// Create wallet transaction record
	txn, err := s.q.CreateWalletTransaction(ctx, sqlc.CreateWalletTransactionParams{
		UserID:       userID,
		TenantID:     tenantID,
		OrderID:      toPgUUIDPtr(orderID),
		Type:         sqlc.WalletTypeDebit,
		Source:       source,
		Amount:       toPgNumeric(amount),
		BalanceAfter: toPgNumeric(newBalance),
		Description:  toNullStringVal(description),
	})
	if err != nil {
		return nil, fmt.Errorf("create wallet transaction: %w", err)
	}
	return &txn, nil
}

// GetBalance returns the current wallet balance for a user.
func (s *WalletService) GetBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	balancePg, err := s.q.GetUserWalletBalance(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return decimal.Zero, apperror.NotFound("user")
	}
	if err != nil {
		return decimal.Zero, fmt.Errorf("get wallet balance: %w", err)
	}
	return pgNumericToDecimal(balancePg), nil
}
