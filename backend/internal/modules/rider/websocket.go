package rider

import (
	"encoding/json"
	"math/big"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/modules/auth"
	"github.com/munchies/platform/backend/internal/modules/tenant"
	"github.com/munchies/platform/backend/internal/pkg/geo"
	redisclient "github.com/munchies/platform/backend/internal/platform/redis"
	"github.com/rs/zerolog/log"
)

// upgrader allows all origins because rider apps connect from mobile clients.
// Authentication is enforced via JWT before upgrading.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WSHandler handles WebSocket connections for rider location tracking.
type WSHandler struct {
	q      *sqlc.Queries
	tokens auth.TokenConfig
	redis  *redisclient.Client
}

// NewWSHandler creates a new WebSocket handler.
func NewWSHandler(q *sqlc.Queries, tokens auth.TokenConfig, redis *redisclient.Client) *WSHandler {
	return &WSHandler{q: q, tokens: tokens, redis: redis}
}

type wsMessage struct {
	Type        string  `json:"type"`
	Lat         float64 `json:"lat,omitempty"`
	Lng         float64 `json:"lng,omitempty"`
	Heading     float64 `json:"heading,omitempty"`
	Speed       float64 `json:"speed,omitempty"`
	Accuracy    float64 `json:"accuracy,omitempty"`
	IsAvailable *bool   `json:"is_available,omitempty"`
}

// HandleWS handles WS /api/v1/rider/ws
func (h *WSHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	// Authenticate via query param or Authorization header
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		hdr := r.Header.Get("Authorization")
		if strings.HasPrefix(hdr, "Bearer ") {
			tokenStr = strings.TrimPrefix(hdr, "Bearer ")
		}
	}
	if tokenStr == "" {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ParseAccessToken(h.tokens, tokenStr)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	user, err := h.q.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}
	if user.Role != sqlc.UserRoleRider {
		http.Error(w, "rider role required", http.StatusForbidden)
		return
	}

	t := tenant.FromContext(r.Context())
	if t == nil {
		http.Error(w, "tenant not found", http.StatusBadRequest)
		return
	}

	rider, err := h.q.GetRiderByUserID(r.Context(), sqlc.GetRiderByUserIDParams{
		UserID:   user.ID,
		TenantID: t.ID,
	})
	if err != nil {
		http.Error(w, "rider profile not found", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("websocket upgrade failed")
		return
	}
	defer conn.Close()

	log.Info().Str("rider_id", rider.ID.String()).Msg("rider websocket connected")

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Error().Err(err).Str("rider_id", rider.ID.String()).Msg("websocket read error")
			}
			break
		}

		var msg wsMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Warn().Err(err).Msg("invalid websocket message")
			continue
		}

		switch msg.Type {
		case "location":
			h.handleLocation(r, conn, rider, t.ID, msg)
		case "status":
			h.handleStatus(r, rider, t.ID, msg)
		default:
			log.Warn().Str("type", msg.Type).Msg("unknown websocket message type")
		}
	}

	log.Info().Str("rider_id", rider.ID.String()).Msg("rider websocket disconnected")
}

func (h *WSHandler) handleLocation(r *http.Request, conn *websocket.Conn, rider sqlc.Rider, tenantID uuid.UUID, msg wsMessage) {
	ctx := r.Context()
	latNum := numericFromFloat64(msg.Lat)
	lngNum := numericFromFloat64(msg.Lng)

	// Get previous location for distance calculation
	var distKm float64
	prevLoc, err := h.q.GetRiderLocation(ctx, rider.ID)
	if err == nil {
		prevLat, _ := numericToFloat64(prevLoc.GeoLat)
		prevLng, _ := numericToFloat64(prevLoc.GeoLng)
		distKm = geo.DistanceKm(prevLat, prevLng, msg.Lat, msg.Lng)
	}

	// Upsert current location
	_, err = h.q.UpsertRiderLocation(ctx, sqlc.UpsertRiderLocationParams{
		RiderID:        rider.ID,
		TenantID:       tenantID,
		GeoLat:         latNum,
		GeoLng:         lngNum,
		Heading:        numericFromFloat64(msg.Heading),
		SpeedKmh:       numericFromFloat64(msg.Speed),
		AccuracyMeters: numericFromFloat64(msg.Accuracy),
	})
	if err != nil {
		log.Error().Err(err).Msg("upsert rider location failed")
		return
	}

	// Append to location history
	_, err = h.q.AppendLocationHistory(ctx, sqlc.AppendLocationHistoryParams{
		RiderID:            rider.ID,
		TenantID:           tenantID,
		GeoLat:             latNum,
		GeoLng:             lngNum,
		EventType:          sqlc.RiderSubjectLocationUpdate,
		DistanceFromPrevKm: numericFromFloat64(distKm),
	})
	if err != nil {
		log.Error().Err(err).Msg("append location history failed")
	}

	// Publish to Redis
	locData, _ := json.Marshal(map[string]interface{}{
		"rider_id": rider.ID,
		"lat":      msg.Lat,
		"lng":      msg.Lng,
		"heading":  msg.Heading,
		"speed":    msg.Speed,
	})
	if h.redis != nil {
		channel := "rider:" + rider.ID.String() + ":location"
		if err := h.redis.Publish(ctx, channel, string(locData)); err != nil {
			log.Error().Err(err).Msg("redis publish failed")
		}
	}
}

func (h *WSHandler) handleStatus(r *http.Request, rider sqlc.Rider, tenantID uuid.UUID, msg wsMessage) {
	if msg.IsAvailable == nil {
		return
	}
	_, err := h.q.UpdateRiderAvailability(r.Context(), sqlc.UpdateRiderAvailabilityParams{
		ID: rider.ID, TenantID: tenantID, IsAvailable: *msg.IsAvailable,
	})
	if err != nil {
		log.Error().Err(err).Msg("update rider availability failed")
	}
}

func numericFromFloat64(f float64) pgtype.Numeric {
	// Represent as integer * 10^-6 for 6 decimal places of precision
	scaled := int64(f * 1_000_000)
	return pgtype.Numeric{
		Int:   big.NewInt(scaled),
		Exp:   -6,
		Valid: true,
	}
}

func numericToFloat64(n pgtype.Numeric) (float64, bool) {
	if !n.Valid || n.Int == nil {
		return 0, false
	}
	f := new(big.Float).SetInt(n.Int)
	if n.Exp != 0 {
		exp := new(big.Float).SetFloat64(1)
		base := new(big.Float).SetFloat64(10)
		e := int(n.Exp)
		if e < 0 {
			e = -e
			for i := 0; i < e; i++ {
				exp.Mul(exp, base)
			}
			f.Quo(f, exp)
		} else {
			for i := 0; i < e; i++ {
				exp.Mul(exp, base)
			}
			f.Mul(f, exp)
		}
	}
	result, _ := f.Float64()
	return result, true
}
