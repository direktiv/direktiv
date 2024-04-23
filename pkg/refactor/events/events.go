// Package events provides the core components for an event-driven system, including:
//
// Key Concepts:
// * **Event Sourcing:** The `EventHistoryStore` facilitates event persistence and retrieval for historical analysis and auditing.
// * **Event Listeners:** Represents a subscriber for one or multiple events with specific types and filtering conditions.
// * **Triggers:** Event listeners can be configured for different trigger types (StartAnd, WaitAnd, StartSimple, etc.) determining how events initiate workflows or continue existing instances.
// * **Delayed Events:** The `StagingEventStore` allows scheduling events for future delivery. Use the EventWorker` for fetching and processing delayed events. Delaying events can be used to reduce peak-load spikes
// * **Initializing the EventEngine:** To start processing events, follow these steps:
// * **Extensibility:** The `EventEngine` relies on interfaces (`WorkflowStart`, `WakeInstance`, etc.), allowing for integration with your specific data stores and workflow implementations.
// * **EventWorker Management** The `EventWorker` is responsible for fetching and processing delayed events from the `StagingEventStore`.
//
//  1. **Create an EventEngine instance:**
//     ```go
//     engine := events.EventEngine{
//     WorkflowStart:    yourWorkflowStartFunction,
//     WakeInstance:     yourWakeInstanceFunction,
//     GetListenersByTopic:   yourGetListenersByTopicFunction,
//     UpdateListeners:      yourUpdateListenersFunction,
//     }
//     ```
//  2. **Call the `ProcessEvents` method:**
//     ```go
//     engine.ProcessEvents(ctx, namespace, cloudevents, logErrors)
//     ```
//
// **Important Considerations:**
// * **Callback Functions:**  The `WorkflowStart`, `WakeInstance`, `GetListenersByTopic`, and `UpdateListeners`  fields of the `EventEngine`  struct must be assigned  implementations that integrate with your specific workflow logic and data stores.
// * **Error Handling:** The `logErrors` function is a placeholder. Ensure you have robust error logging and handling mechanisms in place.
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
	// Append adds at least one and optionally multiple events to the storage.
	// It returns the events that were successfully appended along with any errors encountered.
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
// EventListener represents a subscription to events within a specific namespace.
// It defines trigger conditions, workflow actions, and filtering criteria.
type EventListener struct {
	ID                          uuid.UUID         // Unique identifier of the listener.
	CreatedAt                   time.Time         // Timestamp when the listener was created.
	UpdatedAt                   time.Time         // Timestamp when the listener was last updated.
	Deleted                     bool              // Flag to mark a listener for deletion.
	Namespace                   string            // The Namespace the listener belongs to.
	NamespaceID                 uuid.UUID         // The namespace to which this listener belongs.
	ListeningForEventTypes      []string          // List of event types this listener subscribes to.
	ReceivedEventsForAndTrigger []*Event          // Stores events received for "And" type triggers (where all event types must be received).
	LifespanOfReceivedEvents    int               // The duration (in milliseconds) for which received events should be retained for "And" triggers.  Use 0 to disable the lifespan.
	TriggerType                 TriggerType       // Specifies the type of trigger (StartAnd, WaitAnd, etc.).
	TriggerWorkflow             string            // The ID of the workflow to initiate or continue.
	TriggerInstance             string            // Optional; The ID of a specific workflow instance to resume (for instance-waiting triggers).
	TriggerInstanceStep         int               // Optional; The step within a workflow instance to resume execution (for instance-waiting triggers).
	GlobGatekeepers             map[string]string // Map of glob-patterns for filtering events based on extensions. Key have to be prefixed with the event-type they apply to.
	Metadata                    string            // Field for storing arbitrary metadata associated with the listener.
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

type StagingEventStore interface {
	Append(ctx context.Context, events ...*StagingEvent) ([]*StagingEvent, []error)
	GetDelayedEvents(ctx context.Context, currentTime time.Time, limit int, offset int) ([]*StagingEvent, int, error)
	DeleteByDatabaseIDs(ctx context.Context, databaseIDs ...uuid.UUID) error
}
