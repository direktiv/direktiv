package events

import (
	"context"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type EventWorker struct {
	store  StagingEventStore
	ticker *time.Ticker
	signal chan struct{}
	// eventQueue  chan []*events.StagingEvent
	handleEvent func(ctx context.Context, ns uuid.UUID, nsName string, ce *ce.Event) error
	logger      zap.SugaredLogger
}

func NewEventWorker(store StagingEventStore, interval time.Duration, logger *zap.SugaredLogger, handleEvent func(ctx context.Context, ns uuid.UUID, nsName string, ce *ce.Event) error) *EventWorker {
	return &EventWorker{
		store:  store,
		ticker: time.NewTicker(interval),
		signal: make(chan struct{}),
		// eventQueue:  events,
		handleEvent: handleEvent,
		logger:      *logger,
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
		// panic(err)
		w.logger.Errorf("Error fetching delayed events: %v\n", err)

		return
	}

	if len(receivedEvents) == 0 {
		// w.logger.Debugf("No delayed events to process.")

		return
	}

	w.logger.Debugf("Processing %d delayed events...\n", len(receivedEvents))

	// TODO:process in bulk
	for _, se := range receivedEvents {
		err := w.handleEvent(ctx, se.Namespace, se.NamespaceName, se.Event.Event)
		if err != nil {
			// panic(err)
			w.logger.Errorf("got an error handling a event: %v", err)
		}
	}

	// Delete processed events by database IDs
	databaseIDs := []uuid.UUID{}
	for _, event := range receivedEvents {
		databaseIDs = append(databaseIDs, event.DatabaseID)
	}
	if err := w.store.DeleteByDatabaseIDs(ctx, databaseIDs...); err != nil {
		w.logger.Errorf("Error deleting processed events: %v\n", err)

		return
	}

	w.logger.Debugf("Processed and deleted %d delayed events.\n", len(receivedEvents))
}

// func (w *EventWorker) EnqueueEvent(event *events.StagingEvent) {
// 	select {
// 	case w.eventQueue <- []*events.StagingEvent{event}:
// 	default:
// 		// TODO: Queue is full or blocking, log here
// 	}
// }
