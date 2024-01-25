package logcollection

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

// LogStoreWorker manages the log polling and channel communication.
type logStoreWorker struct {
	Get      func(ctx context.Context, cursorTime time.Time, params map[string]string) ([]core.FeatureLogEntry, error)
	Interval time.Duration
	LogCh    chan string
	Params   map[string]string
	Cursor   time.Time // Cursor instead of Offset
}

// Start starts the log polling worker.
func (lw *logStoreWorker) start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(lw.Interval)
		defer ticker.Stop()
		defer close(lw.LogCh)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				slog.Info("data", "message", lw.Params)
				logs, err := lw.Get(ctx, lw.Cursor, lw.Params)
				if err != nil {
					slog.Error("TODO: should we quit with an error?", "error", err)

					continue
				}
				for _, fle := range logs {
					b, err := json.Marshal(fle)
					if err != nil {
						slog.Error("TODO: should we quit with an error?", "error", err)

						continue
					}
					slog.Info("data", "message", string(b))
					lw.LogCh <- string(b)
				}

				// Update cursorTime for the next iteration
				if len(logs) > 0 {
					lw.Cursor = logs[len(logs)-1].Time
				}
			}
		}
	}()
}
