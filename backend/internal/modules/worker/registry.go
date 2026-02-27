package worker

import (
	"context"
	"time"

	"github.com/munchies/platform/backend/internal/db/sqlc"
	redisclient "github.com/munchies/platform/backend/internal/platform/redis"
	"github.com/rs/zerolog/log"
)

// Worker manages background job processing.
type Worker struct {
	q     *sqlc.Queries
	redis *redisclient.Client
	stop  chan struct{}
}

// NewWorker creates a new background worker.
func NewWorker(q *sqlc.Queries, redis *redisclient.Client) *Worker {
	return &Worker{
		q:     q,
		redis: redis,
		stop:  make(chan struct{}),
	}
}

// Start starts all background job processing.
func (w *Worker) Start(ctx context.Context) {
	log.Info().Msg("starting background workers")

	// Scheduled jobs
	go w.runPeriodic(ctx, "order:auto_confirm", 1*time.Minute, w.AutoConfirmOrders)
	go w.runPeriodic(ctx, "order:auto_cancel", 5*time.Minute, w.AutoCancelOrders)
	go w.runPeriodic(ctx, "notifications:cleanup", 24*time.Hour, w.CleanupNotifications)
	go w.runPeriodic(ctx, "outbox:process", 10*time.Second, w.ProcessOutboxEvents)

	log.Info().Msg("all background workers started")
}

// Stop signals all workers to stop.
func (w *Worker) Stop() {
	close(w.stop)
}

// runPeriodic runs a job function at a fixed interval.
func (w *Worker) runPeriodic(ctx context.Context, name string, interval time.Duration, fn func(context.Context) error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Info().Str("job", name).Dur("interval", interval).Msg("periodic job started")

	for {
		select {
		case <-ctx.Done():
			log.Info().Str("job", name).Msg("periodic job stopped (context cancelled)")
			return
		case <-w.stop:
			log.Info().Str("job", name).Msg("periodic job stopped")
			return
		case <-ticker.C:
			if err := fn(ctx); err != nil {
				log.Error().Err(err).Str("job", name).Msg("periodic job failed")
			}
		}
	}
}
