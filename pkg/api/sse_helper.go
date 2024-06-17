package api

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

type CoursoredEvent struct {
	Event
	time.Time
}

type Event struct {
	ID   string
	Data string
	Type string
}

type sseHandle func(ctx context.Context, cursorTime time.Time) ([]CoursoredEvent, error)

// sseWorker manages the server side event polling and channel communication.
type seeWorker struct {
	Get      sseHandle
	Interval time.Duration
	Ch       chan Event
	Cursor   time.Time // Cursor instead of Offset.
}

// Start starts the sse polling worker.
func (lw *seeWorker) start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(lw.Interval)
		defer ticker.Stop()
		defer close(lw.Ch)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				sseEvents, err := lw.Get(ctx, lw.Cursor)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return
					}

					slog.Error("TODO: should we quit with an error?", "err", err)

					continue
				}
				for _, e := range sseEvents {
					lw.Ch <- e.Event
				}

				// Update cursorTime for the next iteration.
				if len(sseEvents) > 0 {
					lw.Cursor = sseEvents[len(sseEvents)-1].Time
				}
			}
		}
	}()
}
