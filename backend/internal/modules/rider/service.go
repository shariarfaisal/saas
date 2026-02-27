package rider

import (
	"context"
	"database/sql"
	"encoding/json"
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

// ListActiveOrders returns orders actively assigned to a rider.
func (s *Service) ListActiveOrders(ctx context.Context, riderID, tenantID uuid.UUID) ([]sqlc.Order, error) {
	orders, err := s.q.ListActiveOrdersByRider(ctx, sqlc.ListActiveOrdersByRiderParams{
		RiderID:  pgtype.UUID{Bytes: riderID, Valid: true},
		TenantID: tenantID,
	})
	if err != nil {
		return nil, apperror.Internal("list active orders", err)
	}
	return orders, nil
}

// AcceptOrder accepts an assignment offer and transitions the order.
func (s *Service) AcceptOrder(ctx context.Context, orderID, riderID, tenantID uuid.UUID) (sqlc.Order, error) {
	order, err := s.q.GetOrderByID(ctx, sqlc.GetOrderByIDParams{ID: orderID, TenantID: tenantID})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Order{}, apperror.NotFound("order")
	}
	if err != nil {
		return sqlc.Order{}, apperror.Internal("get order", err)
	}

	// Assign rider
	_, err = s.q.AssignRiderToOrder(ctx, sqlc.AssignRiderToOrderParams{
		ID:       orderID,
		TenantID: tenantID,
		RiderID:  pgtype.UUID{Bytes: riderID, Valid: true},
	})
	if err != nil {
		return sqlc.Order{}, apperror.Internal("assign rider", err)
	}

	prevStatus := order.Status
	updated, err := s.q.UpdateOrderStatus(ctx, sqlc.UpdateOrderStatusParams{
		ID: orderID, TenantID: tenantID, Status: sqlc.OrderStatusConfirmed,
	})
	if err != nil {
		return sqlc.Order{}, apperror.Internal("update order status", err)
	}

	s.q.CreateTimelineEvent(ctx, sqlc.CreateTimelineEventParams{
		OrderID:        orderID,
		TenantID:       tenantID,
		EventType:      "rider_accepted",
		PreviousStatus: sqlc.NullOrderStatus{OrderStatus: prevStatus, Valid: true},
		NewStatus:      sqlc.NullOrderStatus{OrderStatus: sqlc.OrderStatusConfirmed, Valid: true},
		Description:    "Rider accepted assignment",
		ActorID:        pgtype.UUID{Bytes: riderID, Valid: true},
		ActorType:      sqlc.ActorTypeRider,
		Metadata:       json.RawMessage(`{}`),
	})

	return updated, nil
}

// MarkPickupPicked marks a per-restaurant pickup as picked and transitions parent order if all picked.
func (s *Service) MarkPickupPicked(ctx context.Context, orderID, restaurantID, riderID, tenantID uuid.UUID) (sqlc.OrderPickup, error) {
	pickup, err := s.q.GetPickupByOrderAndRestaurant(ctx, sqlc.GetPickupByOrderAndRestaurantParams{
		OrderID: orderID, RestaurantID: restaurantID, TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.OrderPickup{}, apperror.NotFound("pickup")
	}
	if err != nil {
		return sqlc.OrderPickup{}, apperror.Internal("get pickup", err)
	}

	updated, err := s.q.UpdatePickupStatus(ctx, sqlc.UpdatePickupStatusParams{
		ID: pickup.ID, TenantID: tenantID, Status: sqlc.PickupStatusPicked,
	})
	if err != nil {
		return sqlc.OrderPickup{}, apperror.Internal("update pickup status", err)
	}

	s.q.CreateTimelineEvent(ctx, sqlc.CreateTimelineEventParams{
		OrderID:     orderID,
		TenantID:    tenantID,
		EventType:   "pickup_picked",
		Description: "Pickup picked from restaurant",
		ActorID:     pgtype.UUID{Bytes: riderID, Valid: true},
		ActorType:   sqlc.ActorTypeRider,
		Metadata:    json.RawMessage(`{"restaurant_id":"` + restaurantID.String() + `"}`),
	})

	// Check if all pickups are picked
	allPickups, err := s.q.ListPickupsByOrder(ctx, sqlc.ListPickupsByOrderParams{
		OrderID: orderID, TenantID: tenantID,
	})
	if err == nil {
		allPicked := true
		for _, p := range allPickups {
			if p.Status != sqlc.PickupStatusPicked {
				allPicked = false
				break
			}
		}
		if allPicked {
			s.q.UpdateOrderStatus(ctx, sqlc.UpdateOrderStatusParams{
				ID: orderID, TenantID: tenantID, Status: sqlc.OrderStatusPicked,
			})
			s.q.CreateTimelineEvent(ctx, sqlc.CreateTimelineEventParams{
				OrderID:   orderID,
				TenantID:  tenantID,
				EventType: "status_changed",
				NewStatus: sqlc.NullOrderStatus{OrderStatus: sqlc.OrderStatusPicked, Valid: true},
				Description: "All pickups collected",
				ActorID:   pgtype.UUID{Bytes: riderID, Valid: true},
				ActorType: sqlc.ActorTypeRider,
				Metadata:  json.RawMessage(`{}`),
			})
		}
	}

	return updated, nil
}

// MarkDelivered marks an order as delivered and updates rider stats.
func (s *Service) MarkDelivered(ctx context.Context, orderID, riderID, tenantID uuid.UUID) (sqlc.Order, error) {
	order, err := s.q.GetOrderByID(ctx, sqlc.GetOrderByIDParams{ID: orderID, TenantID: tenantID})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Order{}, apperror.NotFound("order")
	}
	if err != nil {
		return sqlc.Order{}, apperror.Internal("get order", err)
	}

	prevStatus := order.Status
	updated, err := s.q.UpdateOrderStatus(ctx, sqlc.UpdateOrderStatusParams{
		ID: orderID, TenantID: tenantID, Status: sqlc.OrderStatusDelivered,
	})
	if err != nil {
		return sqlc.Order{}, apperror.Internal("update order status", err)
	}

	s.q.CreateTimelineEvent(ctx, sqlc.CreateTimelineEventParams{
		OrderID:        orderID,
		TenantID:       tenantID,
		EventType:      "status_changed",
		PreviousStatus: sqlc.NullOrderStatus{OrderStatus: prevStatus, Valid: true},
		NewStatus:      sqlc.NullOrderStatus{OrderStatus: sqlc.OrderStatusDelivered, Valid: true},
		Description:    "Order delivered",
		ActorID:        pgtype.UUID{Bytes: riderID, Valid: true},
		ActorType:      sqlc.ActorTypeRider,
		Metadata:       json.RawMessage(`{}`),
	})

	// Calculate and record earnings
	if err := s.CalculateAndRecordEarning(ctx, riderID, tenantID, orderID); err != nil {
		log.Error().Err(err).Str("order_id", orderID.String()).Msg("failed to record earnings")
	}

	return updated, nil
}

// ReportIssue creates an issue record for an order.
func (s *Service) ReportIssue(ctx context.Context, orderID, riderID, tenantID uuid.UUID, issueType, details string) (sqlc.OrderIssue, error) {
	issue, err := s.q.CreateOrderIssue(ctx, sqlc.CreateOrderIssueParams{
		OrderID:          orderID,
		TenantID:         tenantID,
		IssueType:        sqlc.IssueType(issueType),
		ReportedByID:     riderID,
		Details:          details,
		AccountableParty: sqlc.AccountablePlatform,
	})
	if err != nil {
		return sqlc.OrderIssue{}, apperror.Internal("create issue", err)
	}

	s.q.CreateTimelineEvent(ctx, sqlc.CreateTimelineEventParams{
		OrderID:     orderID,
		TenantID:    tenantID,
		EventType:   "issue_reported",
		Description: "Rider reported issue: " + issueType,
		ActorID:     pgtype.UUID{Bytes: riderID, Valid: true},
		ActorType:   sqlc.ActorTypeRider,
		Metadata:    json.RawMessage(`{"issue_type":"` + issueType + `"}`),
	})

	return issue, nil
}

// ManualAssignRider assigns a specific rider to an order (partner action).
func (s *Service) ManualAssignRider(ctx context.Context, orderID, riderID, tenantID, actorID uuid.UUID) (sqlc.Order, error) {
	order, err := s.q.AssignRiderToOrder(ctx, sqlc.AssignRiderToOrderParams{
		ID:       orderID,
		TenantID: tenantID,
		RiderID:  pgtype.UUID{Bytes: riderID, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Order{}, apperror.NotFound("order")
	}
	if err != nil {
		return sqlc.Order{}, apperror.Internal("assign rider", err)
	}

	s.q.CreateTimelineEvent(ctx, sqlc.CreateTimelineEventParams{
		OrderID:     orderID,
		TenantID:    tenantID,
		EventType:   "rider_assigned",
		Description: "Rider manually assigned",
		ActorID:     pgtype.UUID{Bytes: actorID, Valid: true},
		ActorType:   sqlc.ActorTypePlatformAdmin,
		Metadata:    json.RawMessage(`{"rider_id":"` + riderID.String() + `"}`),
	})

	return order, nil
}

// CalculateAndRecordEarning computes and saves a rider earning for an order.
func (s *Service) CalculateAndRecordEarning(ctx context.Context, riderID, tenantID, orderID uuid.UUID) error {
	baseEarning := int64(5000) // 50.00 BDT in cents

	// Distance bonus from location history
	var distanceBonus int64
	history, err := s.q.ListLocationHistoryByRider(ctx, sqlc.ListLocationHistoryByRiderParams{
		RiderID: riderID,
		Limit:   1000,
		Since:   time.Now().Add(-4 * time.Hour),
	})
	if err == nil && len(history) > 0 {
		var totalKm float64
		for _, h := range history {
			d, ok := numericToFloat64(h.DistanceFromPrevKm)
			if ok {
				totalKm += d
			}
		}
		// 5 BDT per km
		distanceBonus = int64(totalKm * 500)
	}

	// Peak hour bonus
	var peakBonus int64
	hour := time.Now().Hour()
	if (hour >= 12 && hour < 14) || (hour >= 18 && hour < 21) {
		peakBonus = 2000 // 20.00 BDT
	}

	totalEarning := baseEarning + distanceBonus + peakBonus

	earning, err := s.q.CreateRiderEarning(ctx, sqlc.CreateRiderEarningParams{
		RiderID:       riderID,
		TenantID:      tenantID,
		OrderID:       orderID,
		BaseEarning:   pgtype.Numeric{Int: big.NewInt(baseEarning), Exp: -2, Valid: true},
		DistanceBonus: pgtype.Numeric{Int: big.NewInt(distanceBonus), Exp: -2, Valid: true},
		PeakBonus:     pgtype.Numeric{Int: big.NewInt(peakBonus), Exp: -2, Valid: true},
		TipAmount:     pgtype.Numeric{Int: big.NewInt(0), Exp: -2, Valid: true},
		TotalEarning:  pgtype.Numeric{Int: big.NewInt(totalEarning), Exp: -2, Valid: true},
	})
	if err != nil {
		return apperror.Internal("create earning", err)
	}

	// Update rider stats
	err = s.q.UpdateRiderStats(ctx, sqlc.UpdateRiderStatsParams{
		ID:             riderID,
		TenantID:       tenantID,
		TotalEarnings:  earning.TotalEarning,
		PendingBalance: earning.TotalEarning,
	})
	if err != nil {
		log.Error().Err(err).Msg("update rider stats failed")
	}

	return nil
}

// ListEarnings returns paginated rider earnings.
func (s *Service) ListEarnings(ctx context.Context, riderID uuid.UUID, limit, offset int32) ([]sqlc.RiderEarning, error) {
	earnings, err := s.q.ListEarningsByRider(ctx, sqlc.ListEarningsByRiderParams{
		RiderID: riderID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, apperror.Internal("list earnings", err)
	}
	return earnings, nil
}

// ListDeliveryHistory returns paginated completed orders for a rider.
func (s *Service) ListDeliveryHistory(ctx context.Context, riderID, tenantID uuid.UUID, limit, offset int32) ([]sqlc.Order, error) {
	orders, err := s.q.ListDeliveredOrdersByRider(ctx, sqlc.ListDeliveredOrdersByRiderParams{
		RiderID:  pgtype.UUID{Bytes: riderID, Valid: true},
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, apperror.Internal("list delivery history", err)
	}
	return orders, nil
}

// ListLocationHistory returns location history for a rider since a given time.
func (s *Service) ListLocationHistory(ctx context.Context, riderID uuid.UUID, since time.Time, limit int32) ([]sqlc.RiderLocationHistory, error) {
	history, err := s.q.ListLocationHistoryByRider(ctx, sqlc.ListLocationHistoryByRiderParams{
		RiderID: riderID, Limit: limit, Since: since,
	})
	if err != nil {
		return nil, apperror.Internal("list location history", err)
	}
	return history, nil
}

// ListRiderLocations returns live locations of all riders for a tenant.
func (s *Service) ListRiderLocations(ctx context.Context, tenantID uuid.UUID) ([]sqlc.RiderLocation, error) {
	locs, err := s.q.ListRiderLocationsByTenant(ctx, tenantID)
	if err != nil {
		return nil, apperror.Internal("list rider locations", err)
	}
	return locs, nil
}

// ListPenalties returns paginated penalties for a rider.
func (s *Service) ListPenalties(ctx context.Context, riderID, tenantID uuid.UUID, limit, offset int32) ([]sqlc.RiderPenalty, error) {
	penalties, err := s.q.ListPenaltiesByRider(ctx, sqlc.ListPenaltiesByRiderParams{
		RiderID: riderID, TenantID: tenantID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, apperror.Internal("list penalties", err)
	}
	return penalties, nil
}

// CreatePenalty creates a new rider penalty.
func (s *Service) CreatePenalty(ctx context.Context, riderID, tenantID uuid.UUID, reason string, amount int64, orderID, issueID *uuid.UUID) (sqlc.RiderPenalty, error) {
	oid := pgtype.UUID{}
	if orderID != nil {
		oid = pgtype.UUID{Bytes: *orderID, Valid: true}
	}
	iid := pgtype.UUID{}
	if issueID != nil {
		iid = pgtype.UUID{Bytes: *issueID, Valid: true}
	}

	penalty, err := s.q.CreateRiderPenalty(ctx, sqlc.CreateRiderPenaltyParams{
		RiderID:  riderID,
		TenantID: tenantID,
		OrderID:  oid,
		IssueID:  iid,
		Reason:   reason,
		Amount:   pgtype.Numeric{Int: big.NewInt(amount), Exp: -2, Valid: true},
	})
	if err != nil {
		return sqlc.RiderPenalty{}, apperror.Internal("create penalty", err)
	}
	return penalty, nil
}

// UpdatePenalty updates a penalty status.
func (s *Service) UpdatePenalty(ctx context.Context, penaltyID, tenantID uuid.UUID, status string, clearedBy *uuid.UUID) (sqlc.RiderPenalty, error) {
	clearedAt := pgtype.Timestamptz{}
	clearedByUUID := pgtype.UUID{}
	if status == string(sqlc.PenaltyStatusCleared) {
		now := time.Now()
		clearedAt = pgtype.Timestamptz{Time: now, InfinityModifier: pgtype.Finite, Valid: true}
		if clearedBy != nil {
			clearedByUUID = pgtype.UUID{Bytes: *clearedBy, Valid: true}
		}
	}

	penalty, err := s.q.UpdatePenaltyStatus(ctx, sqlc.UpdatePenaltyStatusParams{
		ID:        penaltyID,
		TenantID:  tenantID,
		Status:    sqlc.PenaltyStatus(status),
		ClearedAt: clearedAt,
		ClearedBy: clearedByUUID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.RiderPenalty{}, apperror.NotFound("penalty")
	}
	if err != nil {
		return sqlc.RiderPenalty{}, apperror.Internal("update penalty", err)
	}
	return penalty, nil
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
