package eventstore

import (
	"context"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

type CloudEventsStore interface {
	Create(ctx context.Context,
		nsID uuid.UUID,
		fireTime time.Time,
		event *cloudevents.Event,
		eventID uuid.UUID,
		processed bool) error
	UpdateAsPrecessed(eventID uuid.UUID, processed bool) error
	Get(ctx context.Context, nsID, eventID uuid.UUID)
	GetAll(ctx context.Context, nsID uuid.UUID)
	GetFirstUnprocessed(ctx context.Context) (NamespaceCloudEvent, error)
}

type EventsStore interface {
	CreateEventForWorkflow(
		ctx context.Context,
		nsID, fileID uuid.UUID,
		eventin *cloudevents.Event,
		correlations []string,
		signature []byte,
		count int) error
	CreateEventForInstance(
		ctx context.Context,
		nsID, fileID, instID uuid.UUID,
		eventin *cloudevents.Event,
		correlations []string,
		signature []byte,
		count int) error
	Delete(ctx context.Context, eventID uuid.UUID) error
	DeleteAllForInstance(ctx context.Context, InstID uuid.UUID) error
	DeleteAllForWorkflow(ctx context.Context, fileID uuid.UUID) error
	UpdateAsPrecessed(eventID uuid.UUID, processed bool) error
	Update(id uuid.UUID, events NamespaceEvents) error
}

type CloudEventsFilterStore interface {
	Delete(ctx context.Context, nsID uuid.UUID, filterName string) error
	Create(ctx context.Context, nsID uuid.UUID, filterName string, script string) error
	Get(ctx context.Context, nsID uuid.UUID, filterName string) (NamespaceCloudEventFilter, error)
	GetAll(ctx context.Context, nsID uuid.UUID) ([]*NamespaceCloudEventFilter, error)
}

type (
	NamespaceEvents struct {
		ID                    uuid.UUID
		Events                []map[string]interface{}
		Correlations          []string
		Signature             []byte
		Count                 int
		Created               time.Time
		Updated               time.Time
		WorkflowID            uuid.UUID
		EventInstanceListener uuid.UUID
		NamespaceListener     uuid.UUID
		WFEventsWait          map[string]interface{}
	}
	NamespaceCloudEvent struct {
		ID          uuid.UUID
		EventID     uuid.UUID
		NamespaceID uuid.UUID
		Event       cloudevents.Event
		Fire        time.Time
		Created     time.Time
		Processed   bool
	}
	NamespaceCloudEventFilter struct {
		Name        string
		JSCode      string
		NamespaceID uuid.UUID
	}
)
