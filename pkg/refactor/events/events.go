package events

import (
	"context"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

// wraps the cloud-event and adds contextual information.
type Event struct {
	Event      *cloudevents.Event
	Namespace  uuid.UUID
	ReceivedAt time.Time // marks when the events received by the web-API or created via internal logic.
	Round      int       // this value MUST be increased if the event is passed back into the queue.
}

// Persists events.
type EventHistoryStore interface {
	// adds at least one and optionally multiple events to the storage.
	// returns the events that where successfully appended
	Append(ctx context.Context, event *Event, more ...*Event) ([]*Event, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Event, error)
	// the result will be sorted by the AcceptedAt value.
	// pass 0 for limit or offset to get all events.
	Get(ctx context.Context, namespace uuid.UUID, limit, offset int) ([]*Event, error)
	GetAll(ctx context.Context) ([]*Event, error)
	// deletes events that are older then the given timestamp.
	DeleteOld(ctx context.Context, sinceWhen time.Time) error
}

// represents a listener for one or multiple events with specific types.
type EventListener struct {
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
	Deleted                     bool        // set true to remove the subscription.
	NamespaceID                 uuid.UUID   // the namespace to which the listener belongs.
	ListeningForEventTypes      []string    // the types of the event the listener is waiting for to be triggered.
	ReceivedEventsForAndTrigger []*Event    // events already received for the EventsAnd trigger.
	LifespanOfReceivedEvents    int         // set 0 to omit the value.
	TriggerType                 TriggerType // set true for EventsAnd.
	Trigger                     TriggerInfo // hold the information to decide what to do if the listener has satisfied.
}

type TriggerInfo struct {
	WorkflowID uuid.UUID // the id of the workflow.
	InstanceID uuid.UUID // optional fill for instance-waiting trigger.
	Step       int       // optional fill for instance-waiting trigger.
}

type TriggerType int

const (
	StartAnd    TriggerType = iota
	WaitAnd     TriggerType = iota
	StartSimple TriggerType = iota
	WaitSimple  TriggerType = iota
	StartOR     TriggerType = iota
	WaitOR      TriggerType = iota
	StartXOR    TriggerType = iota
	WaitXOR     TriggerType = iota
)

type EventListenerStore interface {
	// adds a EventListener to the storage.
	Append(ctx context.Context, listener *EventListener) error
	// updates the Eventlisteners.
	Update(ctx context.Context, listener *EventListener, more ...*EventListener) (error, []error)
	GetByID(ctx context.Context, id uuid.UUID) (*EventListener, error)
	GetAll(ctx context.Context) ([]*EventListener, error)
	// return all Eventlisteners for a given namespace.
	Get(ctx context.Context, namespace uuid.UUID) ([]*EventListener, error)
	// returns all Eventlisteners for a given namespace that have a subscription for the given eventtype.
	GetByTopic(ctx context.Context, namespace uuid.UUID, eventType string) ([]*EventListener, error)
	// deletes Eventlisteners that have the deleted flag set.
	Delete(ctx context.Context) error
	// deletes the entries associated with the given instance ID.
	DeleteAllForInstance(ctx context.Context, instID uuid.UUID) error
	// deletes the entries associated with the given workflow ID.
	DeleteAllForWorkflow(ctx context.Context, workflowID uuid.UUID) error
}

type NamespaceCloudEventFilter struct {
	Name        string
	JSCode      string
	NamespaceID uuid.UUID
}

type CloudEventsFilterStore interface {
	Delete(ctx context.Context, nsID uuid.UUID, filterName string) error
	Create(ctx context.Context, nsID uuid.UUID, filterName string, script string) error
	Get(ctx context.Context, nsID uuid.UUID, filterName string) (NamespaceCloudEventFilter, error)
	GetAll(ctx context.Context, nsID uuid.UUID) ([]*NamespaceCloudEventFilter, error)
}
