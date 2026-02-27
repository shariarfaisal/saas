package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/pagination"
)

// Service implements user business logic.
type Service struct {
	repo *Repository
}

// NewService creates a new user service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetProfile returns the user profile by ID.
func (s *Service) GetProfile(ctx context.Context, userID uuid.UUID) (*sqlc.User, error) {
	u, err := s.repo.GetByID(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("user")
	}
	return u, err
}

// UpdateProfileRequest holds updateable profile fields.
type UpdateProfileRequest struct {
	Name            *string
	Email           *string
	AvatarURL       *string
	DevicePushToken *string
	DevicePlatform  *string
}

// UpdateProfile updates the user's profile.
func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*sqlc.User, error) {
	u, err := s.repo.Update(ctx, sqlc.UpdateUserParams{
		ID:              userID,
		Name:            nullString(req.Name),
		Email:           nullString(req.Email),
		AvatarUrl:       nullString(req.AvatarURL),
		DevicePushToken: nullString(req.DevicePushToken),
		DevicePlatform:  nullString(req.DevicePlatform),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("user")
	}
	return u, err
}

// CreateAddressRequest holds fields for creating an address.
type CreateAddressRequest struct {
	TenantID       uuid.UUID
	Label          string
	RecipientName  *string
	RecipientPhone *string
	AddressLine1   string
	AddressLine2   *string
	Area           string
	City           string
	IsDefault      bool
}

// CreateAddress creates a new delivery address for the user.
func (s *Service) CreateAddress(ctx context.Context, userID uuid.UUID, req CreateAddressRequest) (*sqlc.UserAddress, error) {
	if req.IsDefault {
		_ = s.repo.ClearDefaultAddresses(ctx, userID)
	}
	return s.repo.CreateAddress(ctx, sqlc.CreateAddressParams{
		UserID:         userID,
		TenantID:       req.TenantID,
		Label:          req.Label,
		RecipientName:  nullString(req.RecipientName),
		RecipientPhone: nullString(req.RecipientPhone),
		AddressLine1:   req.AddressLine1,
		AddressLine2:   nullString(req.AddressLine2),
		Area:           req.Area,
		City:           req.City,
		IsDefault:      req.IsDefault,
	})
}

// ListAddresses returns all delivery addresses for a user.
func (s *Service) ListAddresses(ctx context.Context, userID uuid.UUID) ([]sqlc.UserAddress, error) {
	return s.repo.ListAddresses(ctx, userID)
}

// DeleteAddress deletes an address owned by the user.
func (s *Service) DeleteAddress(ctx context.Context, userID, addressID uuid.UUID) error {
	_, err := s.repo.GetAddressByID(ctx, addressID, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.NotFound("address")
	}
	if err != nil {
		return err
	}
	return s.repo.DeleteAddress(ctx, addressID, userID)
}

// ListWallet returns paginated wallet transactions.
func (s *Service) ListWallet(ctx context.Context, userID uuid.UUID, page, perPage int) ([]sqlc.WalletTransaction, pagination.Meta, error) {
	limit, offset := pagination.FormatLimitOffset(page, perPage)
	total, err := s.repo.CountWalletTransactions(ctx, userID)
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("count wallet transactions", err)
	}
	items, err := s.repo.ListWalletTransactions(ctx, userID, int32(limit), int32(offset))
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("list wallet transactions", err)
	}
	meta := pagination.NewMeta(total, limit, "")
	return items, meta, nil
}

// ListNotifications returns paginated notifications for the user.
func (s *Service) ListNotifications(ctx context.Context, userID uuid.UUID, page, perPage int) ([]sqlc.Notification, pagination.Meta, error) {
	limit, offset := pagination.FormatLimitOffset(page, perPage)
	total, err := s.repo.CountNotifications(ctx, userID)
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("count notifications", err)
	}
	items, err := s.repo.ListNotifications(ctx, userID, int32(limit), int32(offset))
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("list notifications", err)
	}
	meta := pagination.NewMeta(total, limit, "")
	return items, meta, nil
}

// MarkNotificationRead marks a notification as read.
func (s *Service) MarkNotificationRead(ctx context.Context, userID, notifID uuid.UUID) (*sqlc.Notification, error) {
	n, err := s.repo.MarkNotificationRead(ctx, notifID, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("notification")
	}
	return n, err
}
