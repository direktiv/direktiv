package datastoresql

import (
	"context"
	"encoding/json"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/events"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _ events.EventHistoryStore = &SQLEventHistoryStore{}

type SQLEventHistoryStore struct {
	DB *gorm.DB
}

func (hs *SQLEventHistoryStore) Append(ctx context.Context, event *events.Event, more ...*events.Event) ([]*events.Event, error) {
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

func (*SQLEventHistoryStore) DeleteOld(ctx context.Context, sinceWhen time.Time) error {
	panic("unimplemented")
}

func (*SQLEventHistoryStore) Get(ctx context.Context, namespace uuid.UUID, limit int, offset int) ([]*events.Event, int, error) {
	panic("unimplemented")
}

func (*SQLEventHistoryStore) GetAll(ctx context.Context) ([]*events.Event, error) {
	panic("unimplemented")
}

func (*SQLEventHistoryStore) GetByID(ctx context.Context, id uuid.UUID) (*events.Event, error) {
	panic("unimplemented")
}

var _ events.EventTopicsStore = &SQLEventTopicsStore{}

type SQLEventTopicsStore struct{}

func (*SQLEventTopicsStore) Append(ctx context.Context, namespaceID uuid.UUID, eventListenerID uuid.UUID, topic string) error {
	panic("unimplemented")
}

func (*SQLEventTopicsStore) GetListeners(ctx context.Context, namespaceID uuid.UUID, eventType string) ([]*events.EventListener, error) {
	panic("unimplemented")
}

var _ events.EventListenerStore = &SQLEventListenerStore{}

type SQLEventListenerStore struct{}

func (*SQLEventListenerStore) Append(ctx context.Context, listener *events.EventListener) error {
	panic("unimplemented")
}

func (*SQLEventListenerStore) Delete(ctx context.Context) error {
	panic("unimplemented")
}

func (*SQLEventListenerStore) DeleteAllForInstance(ctx context.Context, instID uuid.UUID) error {
	panic("unimplemented")
}

func (*SQLEventListenerStore) DeleteAllForWorkflow(ctx context.Context, workflowID uuid.UUID) error {
	panic("unimplemented")
}

func (*SQLEventListenerStore) Get(ctx context.Context, namespace uuid.UUID, limit int, offet int) ([]*events.EventListener, int, error) {
	panic("unimplemented")
}

func (*SQLEventListenerStore) GetAll(ctx context.Context) ([]*events.EventListener, error) {
	panic("unimplemented")
}

func (*SQLEventListenerStore) GetByID(ctx context.Context, id uuid.UUID) (*events.EventListener, error) {
	panic("unimplemented")
}

func (*SQLEventListenerStore) Update(ctx context.Context, listener *events.EventListener, more ...*events.EventListener) (error, []error) {
	panic("unimplemented")
}

var _ events.CloudEventsFilterStore = &SQLNamespaceCloudEventFilter{}

type SQLNamespaceCloudEventFilter struct{}

func (*SQLNamespaceCloudEventFilter) Create(ctx context.Context, nsID uuid.UUID, filterName string, script string) error {
	panic("unimplemented")
}

func (*SQLNamespaceCloudEventFilter) Delete(ctx context.Context, nsID uuid.UUID, filterName string) error {
	panic("unimplemented")
}

func (*SQLNamespaceCloudEventFilter) Get(ctx context.Context, nsID uuid.UUID, filterName string, limit int, offset int) (events.NamespaceCloudEventFilter, int, error) {
	panic("unimplemented")
}

func (*SQLNamespaceCloudEventFilter) GetAll(ctx context.Context, nsID uuid.UUID) ([]*events.NamespaceCloudEventFilter, error) {
	panic("unimplemented")
}
