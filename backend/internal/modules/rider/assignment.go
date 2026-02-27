package rider

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/geo"
	redisclient "github.com/munchies/platform/backend/internal/platform/redis"
	"github.com/rs/zerolog/log"
)

const (
	assignmentBatchSize = 3
	maxBatches          = 3
	batchTimeoutSec     = 60
)

// AssignmentService handles auto-assignment of riders to orders.
type AssignmentService struct {
	q     *sqlc.Queries
	redis *redisclient.Client
}

// NewAssignmentService creates a new assignment service.
func NewAssignmentService(q *sqlc.Queries, redis *redisclient.Client) *AssignmentService {
	return &AssignmentService{q: q, redis: redis}
}

type riderDistance struct {
	rider    sqlc.Rider
	distance float64
}

// AutoAssign finds and assigns the nearest available rider to an order.
func (s *AssignmentService) AutoAssign(ctx context.Context, orderID, tenantID uuid.UUID) error {
	order, err := s.q.GetOrderByID(ctx, sqlc.GetOrderByIDParams{
		ID: orderID, TenantID: tenantID,
	})
	if err != nil {
		log.Error().Err(err).Str("order_id", orderID.String()).Msg("auto-assign: order not found")
		return err
	}

	if !order.HubID.Valid {
		log.Warn().Str("order_id", orderID.String()).Msg("auto-assign: order has no hub_id")
		return nil
	}

	riders, err := s.q.ListAvailableRidersByHub(ctx, sqlc.ListAvailableRidersByHubParams{
		HubID:    order.HubID,
		TenantID: tenantID,
	})
	if err != nil {
		log.Error().Err(err).Msg("auto-assign: list available riders failed")
		return err
	}
	if len(riders) == 0 {
		log.Warn().Str("order_id", orderID.String()).Msg("auto-assign: no available riders in hub")
		return nil
	}

	deliveryLat, _ := numericToFloat64(order.DeliveryGeoLat)
	deliveryLng, _ := numericToFloat64(order.DeliveryGeoLng)

	// Calculate distance for each rider
	ranked := make([]riderDistance, 0, len(riders))
	for _, r := range riders {
		loc, err := s.q.GetRiderLocation(ctx, r.ID)
		if err != nil {
			continue
		}
		rLat, _ := numericToFloat64(loc.GeoLat)
		rLng, _ := numericToFloat64(loc.GeoLng)
		dist := geo.DistanceKm(rLat, rLng, deliveryLat, deliveryLng)
		ranked = append(ranked, riderDistance{rider: r, distance: dist})
	}

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].distance < ranked[j].distance
	})

	// Try up to maxBatches batches of assignmentBatchSize riders
	for batch := 0; batch < maxBatches; batch++ {
		start := batch * assignmentBatchSize
		if start >= len(ranked) {
			break
		}
		end := start + assignmentBatchSize
		if end > len(ranked) {
			end = len(ranked)
		}

		batchRiders := ranked[start:end]
		for _, rd := range batchRiders {
			s.sendAssignmentOffer(ctx, rd.rider, order)
		}

		log.Info().
			Int("batch", batch+1).
			Int("riders_count", len(batchRiders)).
			Str("order_id", orderID.String()).
			Msg("auto-assign: sent batch offers")

		// For simplicity, auto-accept the first rider in the first batch
		if batch == 0 && len(batchRiders) > 0 {
			chosen := batchRiders[0].rider
			_, err := s.q.AssignRiderToOrder(ctx, sqlc.AssignRiderToOrderParams{
				ID:       orderID,
				TenantID: tenantID,
				RiderID:  pgtype.UUID{Bytes: chosen.ID, Valid: true},
			})
			if err != nil {
				log.Error().Err(err).Msg("auto-assign: assign rider failed")
				return err
			}

			s.q.CreateTimelineEvent(ctx, sqlc.CreateTimelineEventParams{
				OrderID:   orderID,
				TenantID:  tenantID,
				EventType: "rider_assigned",
				NewStatus: sqlc.NullOrderStatus{OrderStatus: order.Status, Valid: true},
				Description: "Rider auto-assigned",
				ActorID:   pgtype.UUID{Bytes: chosen.ID, Valid: true},
				ActorType: sqlc.ActorTypeSystem,
				Metadata:  json.RawMessage(`{}`),
			})

			log.Info().
				Str("rider_id", chosen.ID.String()).
				Str("order_id", orderID.String()).
				Float64("distance_km", batchRiders[0].distance).
				Msg("auto-assign: rider assigned")
			return nil
		}

		// In production, wait for acceptance; for now, simulate timeout
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(batchTimeoutSec) * time.Second):
		}
	}

	log.Warn().Str("order_id", orderID.String()).Msg("auto-assign: no rider accepted after all batches")
	return nil
}

func (s *AssignmentService) sendAssignmentOffer(ctx context.Context, rider sqlc.Rider, order sqlc.Order) {
	offerData, _ := json.Marshal(map[string]interface{}{
		"type":         "assignment_offer",
		"order_id":     order.ID,
		"order_number": order.OrderNumber,
		"delivery_lat": order.DeliveryGeoLat,
		"delivery_lng": order.DeliveryGeoLng,
	})

	log.Info().
		Str("rider_id", rider.ID.String()).
		Str("order_id", order.ID.String()).
		Msg("auto-assign: sending assignment offer")

	if s.redis != nil {
		channel := "rider:" + rider.ID.String() + ":assignment"
		if err := s.redis.Publish(ctx, channel, string(offerData)); err != nil {
			log.Error().Err(err).Str("rider_id", rider.ID.String()).Msg("auto-assign: redis publish failed")
		}
	}
}
