package rider

import (
	"context"
	"database/sql"
	"errors"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/rs/zerolog/log"
)

// Service implements rider business logic.
type Service struct {
	q *sqlc.Queries
}

// NewService creates a new rider service.
func NewService(q *sqlc.Queries) *Service {
	return &Service{q: q}
}

// CreateRiderParams holds input for rider creation.
type CreateRiderParams struct {
	UserID              uuid.UUID
	HubID               *uuid.UUID
	VehicleType         sqlc.VehicleType
	VehicleRegistration string
	LicenseNumber       string
	NidNumber           string
	NidVerified         bool
}

// CreateRider creates a new rider profile.
func (s *Service) CreateRider(ctx context.Context, tenantID uuid.UUID, p CreateRiderParams) (sqlc.Rider, error) {
	hubID := pgtype.UUID{}
	if p.HubID != nil {
		hubID = pgtype.UUID{Bytes: *p.HubID, Valid: true}
	}

	rider, err := s.q.CreateRider(ctx, sqlc.CreateRiderParams{
		TenantID:            tenantID,
		UserID:              p.UserID,
		HubID:               hubID,
		VehicleType:         p.VehicleType,
		VehicleRegistration: sql.NullString{String: p.VehicleRegistration, Valid: p.VehicleRegistration != ""},
		LicenseNumber:       sql.NullString{String: p.LicenseNumber, Valid: p.LicenseNumber != ""},
		NidNumber:           sql.NullString{String: p.NidNumber, Valid: p.NidNumber != ""},
		NidVerified:         p.NidVerified,
	})
	if err != nil {
		return sqlc.Rider{}, apperror.Internal("create rider", err)
	}
	return rider, nil
}

// GetRider returns a rider by ID and tenant.
func (s *Service) GetRider(ctx context.Context, id, tenantID uuid.UUID) (sqlc.Rider, error) {
	rider, err := s.q.GetRiderByID(ctx, sqlc.GetRiderByIDParams{ID: id, TenantID: tenantID})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Rider{}, apperror.NotFound("rider")
	}
	if err != nil {
		return sqlc.Rider{}, apperror.Internal("get rider", err)
	}
	return rider, nil
}

// GetRiderByUserID returns a rider by user ID and tenant.
func (s *Service) GetRiderByUserID(ctx context.Context, userID, tenantID uuid.UUID) (sqlc.Rider, error) {
	rider, err := s.q.GetRiderByUserID(ctx, sqlc.GetRiderByUserIDParams{UserID: userID, TenantID: tenantID})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Rider{}, apperror.NotFound("rider")
	}
	if err != nil {
		return sqlc.Rider{}, apperror.Internal("get rider by user", err)
	}
	return rider, nil
}

// ListRiders returns a paginated list of riders for a tenant.
func (s *Service) ListRiders(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]sqlc.Rider, int64, error) {
	riders, err := s.q.ListRidersByTenant(ctx, sqlc.ListRidersByTenantParams{
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, 0, apperror.Internal("list riders", err)
	}
	total, err := s.q.CountRidersByTenant(ctx, tenantID)
	if err != nil {
		return nil, 0, apperror.Internal("count riders", err)
	}
	return riders, total, nil
}

// UpdateRiderParams holds input for rider update.
type UpdateRiderParams struct {
	HubID               *uuid.UUID
	VehicleType         string
	VehicleRegistration string
	LicenseNumber       string
	NidNumber           string
	NidVerified         *bool
}

// UpdateRider updates a rider profile.
func (s *Service) UpdateRider(ctx context.Context, id, tenantID uuid.UUID, p UpdateRiderParams) (sqlc.Rider, error) {
	hubID := pgtype.UUID{}
	if p.HubID != nil {
		hubID = pgtype.UUID{Bytes: *p.HubID, Valid: true}
	}

	vt := sqlc.NullVehicleType{}
	if p.VehicleType != "" {
		vt = sqlc.NullVehicleType{VehicleType: sqlc.VehicleType(p.VehicleType), Valid: true}
	}

	rider, err := s.q.UpdateRider(ctx, sqlc.UpdateRiderParams{
		ID:                  id,
		TenantID:            tenantID,
		HubID:               hubID,
		VehicleType:         vt,
		VehicleRegistration: sql.NullString{String: p.VehicleRegistration, Valid: p.VehicleRegistration != ""},
		LicenseNumber:       sql.NullString{String: p.LicenseNumber, Valid: p.LicenseNumber != ""},
		NidNumber:           sql.NullString{String: p.NidNumber, Valid: p.NidNumber != ""},
		NidVerified:         p.NidVerified,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Rider{}, apperror.NotFound("rider")
	}
	if err != nil {
		return sqlc.Rider{}, apperror.Internal("update rider", err)
	}
	return rider, nil
}

// DeleteRider removes a rider profile.
func (s *Service) DeleteRider(ctx context.Context, id, tenantID uuid.UUID) error {
	if err := s.q.DeleteRider(ctx, sqlc.DeleteRiderParams{ID: id, TenantID: tenantID}); err != nil {
		return apperror.Internal("delete rider", err)
	}
	return nil
}

// CheckIn creates an attendance record and sets the rider on duty.
func (s *Service) CheckIn(ctx context.Context, riderID, tenantID uuid.UUID, hubID uuid.UUID) (sqlc.RiderAttendance, error) {
	now := time.Now()

	// Update rider hub and set on duty
	_, err := s.q.UpdateRider(ctx, sqlc.UpdateRiderParams{
		ID:       riderID,
		TenantID: tenantID,
		HubID:    pgtype.UUID{Bytes: hubID, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.RiderAttendance{}, apperror.NotFound("rider")
	}
	if err != nil {
		return sqlc.RiderAttendance{}, apperror.Internal("update rider hub", err)
	}

	_, err = s.q.UpdateRiderDutyStatus(ctx, sqlc.UpdateRiderDutyStatusParams{
		ID: riderID, TenantID: tenantID, IsOnDuty: true,
	})
	if err != nil {
		return sqlc.RiderAttendance{}, apperror.Internal("set on duty", err)
	}

	att, err := s.q.CreateAttendance(ctx, sqlc.CreateAttendanceParams{
		RiderID:  riderID,
		TenantID: tenantID,
		WorkDate: pgtype.Date{
			Time:             now,
			InfinityModifier: pgtype.Finite,
			Valid:            true,
		},
		CheckedInAt: pgtype.Timestamptz{
			Time:             now,
			InfinityModifier: pgtype.Finite,
			Valid:            true,
		},
	})
	if err != nil {
		return sqlc.RiderAttendance{}, apperror.Internal("create attendance", err)
	}
	return att, nil
}

// CheckOut closes the active attendance record and sets the rider off duty.
func (s *Service) CheckOut(ctx context.Context, riderID, tenantID uuid.UUID) (sqlc.RiderAttendance, error) {
	att, err := s.q.GetActiveAttendance(ctx, riderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.RiderAttendance{}, apperror.BadRequest("no active attendance record")
	}
	if err != nil {
		return sqlc.RiderAttendance{}, apperror.Internal("get active attendance", err)
	}

	now := time.Now()
	hours := now.Sub(att.CheckedInAt.Time).Hours()

	updated, err := s.q.UpdateAttendanceCheckout(ctx, sqlc.UpdateAttendanceCheckoutParams{
		ID: att.ID,
		CheckedOutAt: pgtype.Timestamptz{
			Time:             now,
			InfinityModifier: pgtype.Finite,
			Valid:            true,
		},
		TotalHours:      numericFromFloat(hours),
		TotalDistanceKm: pgtype.Numeric{Int: big.NewInt(0), Exp: 0, Valid: true},
	})
	if err != nil {
		return sqlc.RiderAttendance{}, apperror.Internal("checkout attendance", err)
	}

	// Set off duty and unavailable
	if _, err := s.q.UpdateRiderDutyStatus(ctx, sqlc.UpdateRiderDutyStatusParams{
		ID: riderID, TenantID: tenantID, IsOnDuty: false,
	}); err != nil {
		log.Error().Err(err).Str("rider_id", riderID.String()).Msg("failed to set rider off duty")
	}
	if _, err := s.q.UpdateRiderAvailability(ctx, sqlc.UpdateRiderAvailabilityParams{
		ID: riderID, TenantID: tenantID, IsAvailable: false,
	}); err != nil {
		log.Error().Err(err).Str("rider_id", riderID.String()).Msg("failed to set rider unavailable")
	}

	return updated, nil
}

// UpdateAvailability toggles rider availability without affecting duty status.
func (s *Service) UpdateAvailability(ctx context.Context, riderID, tenantID uuid.UUID, available bool) (sqlc.Rider, error) {
	rider, err := s.q.UpdateRiderAvailability(ctx, sqlc.UpdateRiderAvailabilityParams{
		ID: riderID, TenantID: tenantID, IsAvailable: available,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Rider{}, apperror.NotFound("rider")
	}
	if err != nil {
		return sqlc.Rider{}, apperror.Internal("update availability", err)
	}
	return rider, nil
}

// ListAttendanceByDate returns attendance records for a tenant on a given date.
func (s *Service) ListAttendanceByDate(ctx context.Context, tenantID uuid.UUID, date time.Time, limit, offset int32) ([]sqlc.RiderAttendance, error) {
	records, err := s.q.ListAttendanceByTenant(ctx, sqlc.ListAttendanceByTenantParams{
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
		WorkDate: pgtype.Date{
			Time:             date,
			InfinityModifier: pgtype.Finite,
			Valid:            true,
		},
	})
	if err != nil {
		return nil, apperror.Internal("list attendance", err)
	}
	return records, nil
}

func numericFromFloat(f float64) pgtype.Numeric {
	// Convert to centesimal integer to preserve two decimal places
	cents := int64(f * 100)
	return pgtype.Numeric{
		Int:   big.NewInt(cents),
		Exp:   -2,
		Valid: true,
	}
}
