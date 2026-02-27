package worker

import (
	"context"

	"github.com/rs/zerolog/log"
)

// ProcessOutboxEvents processes pending outbox events.
func (w *Worker) ProcessOutboxEvents(ctx context.Context) error {
	events, err := w.q.ListPendingOutboxEvents(ctx, 50)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		return nil
	}

	processed := 0
	failed := 0

	for _, event := range events {
		// Publish to Redis pub/sub for SSE routing
		if w.redis != nil {
			channel := event.AggregateType + ":" + event.AggregateID.String()
			if err := w.redis.Publish(ctx, channel, string(event.Payload)); err != nil {
				log.Error().Err(err).
					Str("event_id", event.ID.String()).
					Str("event_type", event.EventType).
					Msg("failed to publish outbox event to Redis")

				if err := w.q.MarkOutboxEventFailed(ctx, event.ID, err.Error()); err != nil {
					log.Error().Err(err).Str("event_id", event.ID.String()).Msg("failed to mark outbox event as failed")
				}
				failed++
				continue
			}
		}

		// Mark as processed
		if err := w.q.MarkOutboxEventProcessed(ctx, event.ID); err != nil {
			log.Error().Err(err).Str("event_id", event.ID.String()).Msg("failed to mark outbox event as processed")
			failed++
			continue
		}
		processed++
	}

	if processed > 0 || failed > 0 {
		log.Info().
			Int("processed", processed).
			Int("failed", failed).
			Msg("outbox events processed")
	}

	return nil
}
