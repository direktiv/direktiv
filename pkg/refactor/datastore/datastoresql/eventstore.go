package datastoresql

import (
	"context"
	"encoding/json"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/events"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _ events.EventHistoryStore = &SqlEventHistoryStore{}

type SqlEventHistoryStore struct {
	DB *gorm.DB
}

// Append implements events.EventHistoryStore
func (hs *SqlEventHistoryStore) Append(ctx context.Context, event *events.Event, more ...*events.Event) ([]*events.Event, error) {
	values := make([]interface{}, 0)
	q := "INSERT INTO events_history (id, type, source, cloudevent, namespace_id, received_at, created_at) VALUES ( $1 , $2 , $3 , $4 , $5 , $6, $7 )"
	eventByte, err := json.Marshal(event.Event)
	if err != nil {
		return nil, err
	}
	values = append(values, event.Event.ID())
	values = append(values, event.Event.Type())
	values = append(values, event.Event.Source())
	values = append(values, eventByte)
	values = append(values, event.Namespace)
	values = append(values, event.ReceivedAt)
	values = append(values, time.Now())
	tx := hs.DB.WithContext(ctx).Exec(q, values...)
	if tx.Error != nil {
		return nil, tx.Error
	}
	for _, v := range more {
		eventByte, err := json.Marshal(v.Event)
		if err != nil {
			return nil, err
		}
		values := make([]interface{}, 0)
		values = append(values, v.Event.ID())
		values = append(values, v.Event.Type())
		values = append(values, v.Event.Source())
		values = append(values, eventByte)
		values = append(values, v.Namespace)
		values = append(values, v.ReceivedAt)
		values = append(values, time.Now())
		tx := hs.DB.WithContext(ctx).Exec(q, values...)
		if tx.Error != nil {
			return nil, tx.Error
		}
	}

	return append([]*events.Event{event}, more...), nil
}

// DeleteOld implements events.EventHistoryStore
func (*SqlEventHistoryStore) DeleteOld(ctx context.Context, sinceWhen time.Time) error {
	panic("unimplemented")
}

// Get implements events.EventHistoryStore
func (*SqlEventHistoryStore) Get(ctx context.Context, namespace uuid.UUID, limit int, offset int) ([]*events.Event, int, error) {
	panic("unimplemented")
}

// GetAll implements events.EventHistoryStore
func (*SqlEventHistoryStore) GetAll(ctx context.Context) ([]*events.Event, error) {
	panic("unimplemented")
}

// GetByID implements events.EventHistoryStore
func (*SqlEventHistoryStore) GetByID(ctx context.Context, id uuid.UUID) (*events.Event, error) {
	panic("unimplemented")
}

var _ events.EventTopicsStore = &SqlEventTopicsStore{}

type SqlEventTopicsStore struct{}

// Append implements events.EventTopicsStore
func (*SqlEventTopicsStore) Append(ctx context.Context, namespaceID uuid.UUID, eventListenerID uuid.UUID, topic string) error {
	panic("unimplemented")
}

// GetListeners implements events.EventTopicsStore
func (*SqlEventTopicsStore) GetListeners(ctx context.Context, namespaceID uuid.UUID, eventType string) ([]*events.EventListener, error) {
	panic("unimplemented")
}

var _ events.EventListenerStore = &SqlEventListenerStore{}

type SqlEventListenerStore struct{}

// Append implements events.EventListenerStore
func (*SqlEventListenerStore) Append(ctx context.Context, listener *events.EventListener) error {
	panic("unimplemented")
}

// Delete implements events.EventListenerStore
func (*SqlEventListenerStore) Delete(ctx context.Context) error {
	panic("unimplemented")
}

// DeleteAllForInstance implements events.EventListenerStore
func (*SqlEventListenerStore) DeleteAllForInstance(ctx context.Context, instID uuid.UUID) error {
	panic("unimplemented")
}

// DeleteAllForWorkflow implements events.EventListenerStore
func (*SqlEventListenerStore) DeleteAllForWorkflow(ctx context.Context, workflowID uuid.UUID) error {
	panic("unimplemented")
}

// Get implements events.EventListenerStore
func (*SqlEventListenerStore) Get(ctx context.Context, namespace uuid.UUID, limit int, offet int) ([]*events.EventListener, int, error) {
	panic("unimplemented")
}

// GetAll implements events.EventListenerStore
func (*SqlEventListenerStore) GetAll(ctx context.Context) ([]*events.EventListener, error) {
	panic("unimplemented")
}

// GetByID implements events.EventListenerStore
func (*SqlEventListenerStore) GetByID(ctx context.Context, id uuid.UUID) (*events.EventListener, error) {
	panic("unimplemented")
}

// Update implements events.EventListenerStore
func (*SqlEventListenerStore) Update(ctx context.Context, listener *events.EventListener, more ...*events.EventListener) (error, []error) {
	panic("unimplemented")
}

var _ events.CloudEventsFilterStore = &SqlNamespaceCloudEventFilter{}

type SqlNamespaceCloudEventFilter struct{}

// Create implements events.CloudEventsFilterStore
func (*SqlNamespaceCloudEventFilter) Create(ctx context.Context, nsID uuid.UUID, filterName string, script string) error {
	panic("unimplemented")
}

// Delete implements events.CloudEventsFilterStore
func (*SqlNamespaceCloudEventFilter) Delete(ctx context.Context, nsID uuid.UUID, filterName string) error {
	panic("unimplemented")
}

// Get implements events.CloudEventsFilterStore
func (*SqlNamespaceCloudEventFilter) Get(ctx context.Context, nsID uuid.UUID, filterName string, limit int, offset int) (events.NamespaceCloudEventFilter, int, error) {
	panic("unimplemented")
}

// GetAll implements events.CloudEventsFilterStore
func (*SqlNamespaceCloudEventFilter) GetAll(ctx context.Context, nsID uuid.UUID) ([]*events.NamespaceCloudEventFilter, error) {
	panic("unimplemented")
}
