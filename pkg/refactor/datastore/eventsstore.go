package datastore

import (
	"context"
	"fmt"
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
	Event       *cloudevents.Event `json:"event"`
	NamespaceID uuid.UUID          `json:"namespaceID,omitempty"`
	Namespace   string             `json:"namespace,omitempty"`
	ReceivedAt  time.Time          `json:"receivedAt"` // marks when the events received by the web-API or created via internal logic.
	SerialID    int                `json:"serialID"`
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
	GetStartingIDUntilTime(ctx context.Context, namespace string, lastID int, t time.Time, keyAndValues ...string) ([]*Event, error)
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
	ID                          uuid.UUID            `json:"id"`                          // Unique identifier of the listener.
	CreatedAt                   time.Time            `json:"createdAt"`                   // Timestamp when the listener was created.
	UpdatedAt                   time.Time            `json:"updatedAt"`                   // Timestamp when the listener was last updated.
	Deleted                     bool                 `json:"deleted"`                     // Flag to mark a listener for deletion.
	Namespace                   string               `json:"namespace"`                   // The Namespace the listener belongs to.
	NamespaceID                 uuid.UUID            `json:"namespaceID"`                 // The namespace to which this listener belongs.
	ListeningForEventTypes      []string             `json:"listeningForEventTypes"`      // List of event types this listener subscribes to.
	ReceivedEventsForAndTrigger []*Event             `json:"receivedEventsForAndTrigger"` // Stores events received for "And" type triggers (where all event types must be received).
	LifespanOfReceivedEvents    int                  `json:"lifespanOfReceivedEvents"`    // The duration (in milliseconds) for which received events should be retained for "And" triggers.  Use 0 to disable the lifespan.
	TriggerType                 TriggerType          `json:"triggerType"`                 // Specifies the type of trigger (StartAnd, WaitAnd, etc.).
	TriggerWorkflow             string               `json:"triggerWorkflow,omitempty"`   // The ID of the workflow to initiate or continue.
	TriggerInstance             string               `json:"triggerInstance,omitempty"`   // Optional; The ID of a specific workflow instance to resume (for instance-waiting triggers).
	EventContextFilter          []EventContextFilter `json:"eventContextFilters,omitempty"`
	Metadata                    string               `json:"metadata"` // Field for storing arbitrary metadata associated with the listener.
}

type EventContextFilter struct {
	Typ     string            `json:"typ"`
	Context map[string]string `json:"context"`
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

func (t TriggerType) String() string {
	switch t {
	case StartAnd:
		return "StartAnd"
	case WaitAnd:
		return "WaitAnd"
	case StartSimple:
		return "StartSimple"
	case WaitSimple:
		return "WaitSimple"
	case StartOR:
		return "StartOR"
	case WaitOR:
		return "WaitOR"
	default:
		return fmt.Sprintf("Unknown TriggerType: %d", t)
	}
}

type EventListenerStore interface {
	// adds a EventListener to the storage.
	Append(ctx context.Context, listener *EventListener) error
	// updates the EventListeners.
	UpdateOrDelete(ctx context.Context, listener []*EventListener) []error
	GetByID(ctx context.Context, id uuid.UUID) (*EventListener, error)
	GetAll(ctx context.Context) ([]*EventListener, error)
	// return EventListeners for a given namespace with the total row count for pagination.
	Get(ctx context.Context, namespace uuid.UUID, limit, offet int) ([]*EventListener, int, error)
	GetOld(ctx context.Context, namespace string, t time.Time) ([]*EventListener, error)
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
