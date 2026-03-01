package worker

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

// MarkOverdueCOD marks cash collections older than 24 hours as overdue.
func (w *Worker) MarkOverdueCOD(ctx context.Context) error {
	threshold := time.Now().Add(-24 * time.Hour)
	if err := w.q.MarkOverdueCashCollections(ctx, threshold); err != nil {
		return err
	}
	log.Debug().Msg("marked overdue COD cash collections")
	return nil
}
