package review_services

import (
	"context"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

// StartRetentionJob starts a goroutine that periodically deletes analysis results older than 'days'.
// It returns immediately; cancellation is controlled by the provided ctx.
func StartRetentionJob(ctx context.Context, repo AnalysisRepositoryInterface, days int, interval time.Duration, l logger.Interface) {
	if repo == nil {
		l.Warn("Retention job: analysisRepo is nil; retention disabled")
		return
	}

	if days <= 0 {
		l.Info("Retention job: days <= 0 -> retention disabled")
		return
	}

	ticker := time.NewTicker(interval)
	go func() {
		l.Info("Retention job started", "days", days, "interval", interval.String())
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				l.Info("Retention job stopping due to context cancellation")
				return
			case <-ticker.C:
				cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
				if err := repo.DeleteOlderThan(context.Background(), cutoff); err != nil {
					l.Error("Retention job: failed to delete old analysis results", "error", err)
				} else {
					l.Info("Retention job: deleted old analysis results", "cutoff", cutoff.Format(time.RFC3339))
				}
			}
		}
	}()
}
