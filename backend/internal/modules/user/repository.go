package user

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
)

// Repository provides user data access.
type Repository struct {
	q *sqlc.Queries
}

// NewRepository creates a new user repository.
func NewRepository(q *sqlc.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*sqlc.User, error) {
	u, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) Update(ctx context.Context, arg sqlc.UpdateUserParams) (*sqlc.User, error) {
	u, err := r.q.UpdateUser(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.q.SoftDeleteUser(ctx, id)
}

func (r *Repository) CreateAddress(ctx context.Context, arg sqlc.CreateAddressParams) (*sqlc.UserAddress, error) {
	a, err := r.q.CreateAddress(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repository) GetAddressByID(ctx context.Context, id, userID uuid.UUID) (*sqlc.UserAddress, error) {
	a, err := r.q.GetAddressByID(ctx, sqlc.GetAddressByIDParams{ID: id, UserID: userID})
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repository) ListAddresses(ctx context.Context, userID uuid.UUID) ([]sqlc.UserAddress, error) {
	return r.q.ListAddresses(ctx, userID)
}

func (r *Repository) UpdateAddress(ctx context.Context, arg sqlc.UpdateAddressParams) (*sqlc.UserAddress, error) {
	a, err := r.q.UpdateAddress(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repository) DeleteAddress(ctx context.Context, id, userID uuid.UUID) error {
	return r.q.DeleteAddress(ctx, sqlc.DeleteAddressParams{ID: id, UserID: userID})
}

func (r *Repository) ClearDefaultAddresses(ctx context.Context, userID uuid.UUID) error {
	return r.q.ClearDefaultAddresses(ctx, userID)
}

func (r *Repository) ListWalletTransactions(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]sqlc.WalletTransaction, error) {
	return r.q.ListWalletTransactions(ctx, sqlc.ListWalletTransactionsParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
}

func (r *Repository) CountWalletTransactions(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.q.CountWalletTransactions(ctx, userID)
}

func (r *Repository) ListNotifications(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]sqlc.Notification, error) {
	return r.q.ListNotifications(ctx, sqlc.ListNotificationsParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
}

func (r *Repository) CountNotifications(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.q.CountNotifications(ctx, userID)
}

func (r *Repository) MarkNotificationRead(ctx context.Context, id, userID uuid.UUID) (*sqlc.Notification, error) {
	n, err := r.q.MarkNotificationRead(ctx, sqlc.MarkNotificationReadParams{ID: id, UserID: userID})
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// uuidToPgtype converts a *uuid.UUID to pgtype.UUID.
func uuidToPgtype(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

// nullString returns a sql.NullString for a non-empty string pointer.
func nullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}
