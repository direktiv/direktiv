package events

import (
	"context"
	"log/slog"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/google/uuid"
)

type EventWorker struct {
	store  datastore.StagingEventStore
	ticker *time.Ticker
	signal chan struct{}
	// eventQueue  chan []*events.StagingEvent
	handleEvent func(ctx context.Context, ns uuid.UUID, nsName string, ce *ce.Event) error
}

// NewEventWorker creates a new EventWorker instance.
// * `store`: The StagingEventStore used for retrieving and deleting delayed events.
// * `interval`:  The interval at which the worker checks for delayed events.
// * `handleEvent`: The function invoked to process each delayed event.
func NewEventWorker(store datastore.StagingEventStore, interval time.Duration, handleEvent func(ctx context.Context, ns uuid.UUID, nsName string, ce *ce.Event) error) *EventWorker {
	return &EventWorker{
		store:  store,
		ticker: time.NewTicker(interval),
		signal: make(chan struct{}),
		// eventQueue:  events,
		handleEvent: handleEvent,
	}
}

// Start initializes the EventWorker's processing loop. The worker will periodically fetch and handle delayed events until the provided context is canceled.
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

// Signal triggers an immediate check for delayed events. This can be used to process events outside the regular interval.
func (w *EventWorker) Signal() {
	select {
	case w.signal <- struct{}{}:
	default:
		// Signal channel is non-blocking to prevent blocking the worker
	}
}

// getDelayedEvents retrieves delayed events from the store, processes them using the configured handler, and removes them upon successful processing.
func (w *EventWorker) getDelayedEvents(ctx context.Context) {
	currentTime := time.Now().UTC()
	limit := 100
	offset := 0
	receivedEvents, _, err := w.store.GetDelayedEvents(ctx, currentTime, limit, offset)
	//  TODO: myMetrics.events_delayed_processing_duration.Observe(processDuration.Seconds())
	if err != nil {
		slog.Error("fetching delayed events", "err", err)
		// TODO: myMetrics.events_delayed_fetch_errors.Inc()

		return
	}

	if len(receivedEvents) == 0 {
		// w.logger.Debugf("No delayed events to process.")

		return
	}

	slog.Debug("starting processing delayed events")

	// TODO: possible process events in bulk
	for _, se := range receivedEvents {
		err := w.handleEvent(ctx, se.NamespaceID, se.Namespace, se.Event.Event)
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
