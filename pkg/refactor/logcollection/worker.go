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
	LogCh    chan []byte
	Params   map[string]string
	Cursor   time.Time // Cursor instead of Offset
}

// Start starts the log polling worker.
func (lw *logStoreWorker) start(ctx context.Context) {
	go func() {
		defer close(lw.LogCh)
		ticker := time.NewTicker(lw.Interval)
		defer ticker.Stop()
		cursorTime := time.Time{} // Initial cursor is the zero time
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				logs, err := lw.Get(ctx, cursorTime, lw.Params)
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
					lw.LogCh <- b
				}

				// Update cursorTime for the next iteration
				if len(logs) > 0 {
					cursorTime = logs[len(logs)-1].Time
				}
			}
		}
	}()
}
