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
	Get      func(ctx context.Context, offset int, params map[string]string) ([]core.FeatureLogEntry, error)
	Interval time.Duration
	LogCh    chan []byte
	Params   map[string]string
}

// Start starts the log polling worker.
func (lw *logStoreWorker) start(ctx context.Context) {
	go func() {
		defer close(lw.LogCh)
		ticker := time.NewTicker(lw.Interval)
		defer ticker.Stop()
		offset := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				logs, err := lw.Get(ctx, offset, lw.Params)
				offset += len(logs)
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
			}
		}
	}()
}
