package events

import (
	"context"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

type StagingEvent struct {
	*Event
	DatabaseID   uuid.UUID
	DelayedUntil time.Time
}

// wraps the cloud-event and adds contextual information.
type Event struct {
	Event         *cloudevents.Event
	Namespace     uuid.UUID
	NamespaceName string
	ReceivedAt    time.Time // marks when the events received by the web-API or created via internal logic.
}

// Persists and gets events.
type EventHistoryStore interface {
	// adds at least one and optionally multiple events to the storage.
	// returns the events that where successfully appended
	Append(ctx context.Context, event []*Event) ([]*Event, []error)
	GetByID(ctx context.Context, id string) (*Event, error)
	// the result will be sorted by the AcceptedAt value.
	// pass 0 for limit or offset to get all events.
	// The total row count is also returned for pagination.
	// keyValues MUST be passed as key value pairs.
	// passed keyValues will be used filter results.
	// supported keys are created_before, created_after,
	// received_before, received_after, event_contains, type_contains.
	Get(ctx context.Context, limit, offset int, namespace uuid.UUID, keyValues ...string) ([]*Event, int, error)
	GetOld(ctx context.Context, namespace string, t time.Time, keyAndValues ...string) ([]*Event, error)
	GetNew(ctx context.Context, namespace string, t time.Time, keyAndValues ...string) ([]*Event, error)
	GetAll(ctx context.Context) ([]*Event, error)
	// deletes events that are older then the given timestamp.
	DeleteOld(ctx context.Context, sinceWhen time.Time) error
}

// Helps query the proper event-listeners for a namespace and event-type.
type EventTopicsStore interface {
	// topic SHOULD be a compound of namespaceID and the eventType like this: "uuid-eventType"
	Append(ctx context.Context, namespaceID uuid.UUID, namespace string, eventListenerID uuid.UUID, topic string, filter string) error
	GetListeners(ctx context.Context, topic string) ([]*EventListener, error)
	Delete(ctx context.Context, eventListenerID uuid.UUID) error
}

// represents a listener for one or multiple events with specific types.
type EventListener struct {
	ID                          uuid.UUID
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
	Deleted                     bool      // set true to remove the subscription.
	NamespaceID                 uuid.UUID // the namespace to which the listener belongs.
	Namespace                   string
	ListeningForEventTypes      []string    // the types of the event the listener is waiting for to be triggered.
	ReceivedEventsForAndTrigger []*Event    // events already received for the EventsAnd trigger.
	LifespanOfReceivedEvents    int         // set 0 to omit the value.
	TriggerType                 TriggerType // set true for EventsAnd.
	TriggerWorkflow             string      // the id of the workflow.
	TriggerInstance             string      // optional fill for instance-waiting trigger.
	TriggerInstanceStep         int         // optional fill for instance-waiting trigger.
	GlobGatekeepers             map[string]string
	Metadata                    string
}

type TriggerType int

const (
	StartAnd    TriggerType = iota
	WaitAnd     TriggerType = iota
	StartSimple TriggerType = iota
	WaitSimple  TriggerType = iota
	StartOR     TriggerType = iota
	WaitOR      TriggerType = iota
)

type EventListenerStore interface {
	// adds a EventListener to the storage.
	Append(ctx context.Context, listener *EventListener) error
	// updates the EventListeners.
	UpdateOrDelete(ctx context.Context, listener []*EventListener) []error
	GetByID(ctx context.Context, id uuid.UUID) (*EventListener, error)
	GetAll(ctx context.Context) ([]*EventListener, error)
	// return EventListeners for a given namespace with the total row count for pagination.
	Get(ctx context.Context, namespace uuid.UUID, limit, offet int) ([]*EventListener, int, error)
	GetNew(ctx context.Context, namespace string, t time.Time) ([]*EventListener, error)
	// deletes EventListeners that have the deleted flag set.
	Delete(ctx context.Context) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
	// deletes the entries associated with the given instance ID.
	DeleteAllForInstance(ctx context.Context, instID uuid.UUID) ([]*uuid.UUID, error)
	// deletes the entries associated with the given workflow ID.
	DeleteAllForWorkflow(ctx context.Context, workflowID uuid.UUID) ([]*uuid.UUID, error)
}

// Currently only in use for delayed events.
type StagingEventStore interface {
	Append(ctx context.Context, events ...*StagingEvent) ([]*StagingEvent, []error)
	GetDelayedEvents(ctx context.Context, currentTime time.Time, limit int, offset int) ([]*StagingEvent, int, error)
	DeleteByDatabaseIDs(ctx context.Context, databaseIDs ...uuid.UUID) error
}
