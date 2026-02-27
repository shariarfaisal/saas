package sse

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/munchies/platform/backend/internal/modules/auth"
	"github.com/munchies/platform/backend/internal/modules/tenant"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/respond"
	redisclient "github.com/munchies/platform/backend/internal/platform/redis"
	"github.com/rs/zerolog/log"
)

// Handler handles SSE connections.
type Handler struct {
	redis *redisclient.Client
}

// NewHandler creates a new SSE handler.
func NewHandler(redis *redisclient.Client) *Handler {
	return &Handler{redis: redis}
}

// Subscribe handles GET /api/v1/events/subscribe
func (h *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		respond.Error(w, apperror.Unauthorized("authentication required"))
		return
	}
	t := tenant.FromContext(r.Context())

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		respond.Error(w, apperror.Internal("streaming unsupported", nil))
		return
	}

	// Build channel names
	channels := []string{
		fmt.Sprintf("user:%s", u.ID.String()),
	}
	if t != nil {
		channels = append(channels, fmt.Sprintf("tenant:%s", t.ID.String()))
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Subscribe to Redis pub/sub if available
	if h.redis != nil {
		pubsub := h.redis.Subscribe(ctx, channels...)
		defer pubsub.Close()

		ch := pubsub.Channel()

		// Send initial connection event
		fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\"}\n\n")
		flusher.Flush()

		// Heartbeat ticker
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				fmt.Fprintf(w, "data: %s\n\n", msg.Payload)
				flusher.Flush()
			case <-ticker.C:
				fmt.Fprintf(w, ": heartbeat\n\n")
				flusher.Flush()
			}
		}
	}

	// Fallback: heartbeat only (no Redis)
	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\",\"mode\":\"heartbeat\"}\n\n")
	flusher.Flush()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}

var _ = log.Logger // ensure zerolog import
