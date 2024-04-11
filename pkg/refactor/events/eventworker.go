package events

import (
	"context"
	"log/slog"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

type EventWorker struct {
	store  StagingEventStore
	ticker *time.Ticker
	signal chan struct{}
	// eventQueue  chan []*events.StagingEvent
	handleEvent func(ctx context.Context, ns uuid.UUID, nsName string, ce *ce.Event) error
}

func NewEventWorker(store StagingEventStore, interval time.Duration, handleEvent func(ctx context.Context, ns uuid.UUID, nsName string, ce *ce.Event) error) *EventWorker {
	return &EventWorker{
		store:  store,
		ticker: time.NewTicker(interval),
		signal: make(chan struct{}),
		// eventQueue:  events,
		handleEvent: handleEvent,
	}
}

func (w *EventWorker) Start(ctx context.Context) {
	defer w.ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.ticker.C:
			w.getDelayedEvents(ctx)
		case <-w.signal:
			w.getDelayedEvents(ctx)
		}
	}
}

func (w *EventWorker) Signal() {
	select {
	case w.signal <- struct{}{}:
	default:
		// Signal channel is non-blocking to prevent blocking the worker
	}
}

func (w *EventWorker) getDelayedEvents(ctx context.Context) {
	currentTime := time.Now().UTC()
	limit := 100
	offset := 0
	receivedEvents, _, err := w.store.GetDelayedEvents(ctx, currentTime, limit, offset)
	if err != nil {
		slog.Error("fetching delayed events", "err", err)

		return
	}

	if len(receivedEvents) == 0 {
		// w.logger.Debugf("No delayed events to process.")

		return
	}

	slog.Debug("starting processing delayed events")

	// TODO: possible process events in bulk
	for _, se := range receivedEvents {
		err := w.handleEvent(ctx, se.Namespace, se.NamespaceName, se.Event.Event)
		if err != nil {
			slog.Error("handle a event", "err", err)
		}
	}

	// Delete processed events by database IDs
	databaseIDs := []uuid.UUID{}
	for _, event := range receivedEvents {
		databaseIDs = append(databaseIDs, event.DatabaseID)
	}
	if err := w.store.DeleteByDatabaseIDs(ctx, databaseIDs...); err != nil {
		slog.Error("failed deleting processed events", "err", err)

		return
	}

	slog.Debug("processed and deleted delayed events")
}
