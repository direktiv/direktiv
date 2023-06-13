package datastoresql

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/direktiv/direktiv/pkg/refactor/events"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _ events.EventHistoryStore = &sqlEventHistoryStore{}

type sqlEventHistoryStore struct {
	db *gorm.DB
}

func (hs *sqlEventHistoryStore) Append(ctx context.Context, event *events.Event, more ...*events.Event) ([]*events.Event, error) {
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
	tx := hs.db.WithContext(ctx).Exec(q, values...)
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
		tx := hs.db.WithContext(ctx).Exec(q, values...)
		if tx.Error != nil {
			return nil, tx.Error
		}
	}

	return append([]*events.Event{event}, more...), nil
}

func (*sqlEventHistoryStore) DeleteOld(ctx context.Context, sinceWhen time.Time) error {
	panic("unimplemented")
}

func (hs *sqlEventHistoryStore) Get(ctx context.Context, limit int, offset int, namespace uuid.UUID, keyAndValues ...string) ([]*events.Event, int, error) {
	q := "SELECT id, type, source, cloudevent, namespace_id, received_at, created_at FROM events_history WHERE namespace_id = $1;"
	res := make([]*events.Event, 0)

	rows, err := hs.db.WithContext(ctx).Raw(q).Rows()
	if err != nil {
		return nil, 0, err
	}
	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}
	defer rows.Close()
	if rows.Next() {
		var id, ns uuid.UUID
		var t, source, ce string
		var created, received time.Time
		err := rows.Scan(&id, &t, &source, &ce, &ns, &received, &created)
		if err != nil {
			return nil, 0, err
		}
		var finalCE event.Event
		err = json.Unmarshal([]byte(ce), &finalCE)
		if err != nil {
			return nil, 0, err
		}
		res = append(res, &events.Event{Namespace: ns, ReceivedAt: received, Event: &finalCE})
	}

	return res, 0, nil
}

func (hs *sqlEventHistoryStore) GetAll(ctx context.Context) ([]*events.Event, error) {
	q := "SELECT id, type, source, cloudevent, namespace_id, received_at, created_at FROM events_history;"
	res := make([]*events.Event, 0)

	rows, err := hs.db.WithContext(ctx).Raw(q).Rows()
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	defer rows.Close()
	if rows.Next() {
		var id, ns uuid.UUID
		var t, source, ce string
		var created, received time.Time
		err := rows.Scan(&id, &t, &source, &ce, &ns, &received, &created)
		if err != nil {
			return nil, err
		}
		var finalCE event.Event
		err = json.Unmarshal([]byte(ce), &finalCE)
		if err != nil {
			return nil, err
		}
		res = append(res, &events.Event{Namespace: ns, ReceivedAt: received, Event: &finalCE})
	}

	return res, nil
}

func (*sqlEventHistoryStore) GetByID(ctx context.Context, id uuid.UUID) (*events.Event, error) {
	panic("unimplemented")
}

var _ events.EventTopicsStore = &sqlEventTopicsStore{}

type sqlEventTopicsStore struct{}

func (*sqlEventTopicsStore) Append(ctx context.Context, namespaceID uuid.UUID, eventListenerID uuid.UUID, topic string) error {
	panic("unimplemented")
}

func (*sqlEventTopicsStore) GetListeners(ctx context.Context, namespaceID uuid.UUID, eventType string) ([]*events.EventListener, error) {
	panic("unimplemented")
}

var _ events.EventListenerStore = &sqlEventListenerStore{}

type sqlEventListenerStore struct{}

func (*sqlEventListenerStore) Append(ctx context.Context, listener *events.EventListener) error {
	panic("unimplemented")
}

func (*sqlEventListenerStore) Delete(ctx context.Context) error {
	panic("unimplemented")
}

func (*sqlEventListenerStore) DeleteAllForInstance(ctx context.Context, instID uuid.UUID) error {
	panic("unimplemented")
}

func (*sqlEventListenerStore) DeleteAllForWorkflow(ctx context.Context, workflowID uuid.UUID) error {
	panic("unimplemented")
}

func (*sqlEventListenerStore) Get(ctx context.Context, namespace uuid.UUID, limit int, offet int) ([]*events.EventListener, int, error) {
	panic("unimplemented")
}

func (*sqlEventListenerStore) GetAll(ctx context.Context) ([]*events.EventListener, error) {
	panic("unimplemented")
}

func (*sqlEventListenerStore) GetByID(ctx context.Context, id uuid.UUID) (*events.EventListener, error) {
	panic("unimplemented")
}

func (*sqlEventListenerStore) Update(ctx context.Context, listener *events.EventListener, more ...*events.EventListener) (error, []error) {
	panic("unimplemented")
}

var _ events.CloudEventsFilterStore = &sqlNamespaceCloudEventFilter{}

type sqlNamespaceCloudEventFilter struct{}

func (*sqlNamespaceCloudEventFilter) Create(ctx context.Context, nsID uuid.UUID, filterName string, script string) error {
	panic("unimplemented")
}

func (*sqlNamespaceCloudEventFilter) Delete(ctx context.Context, nsID uuid.UUID, filterName string) error {
	panic("unimplemented")
}

func (*sqlNamespaceCloudEventFilter) Get(ctx context.Context, nsID uuid.UUID, filterName string, limit int, offset int) (events.NamespaceCloudEventFilter, int, error) {
	panic("unimplemented")
}

func (*sqlNamespaceCloudEventFilter) GetAll(ctx context.Context, nsID uuid.UUID) ([]*events.NamespaceCloudEventFilter, error) {
	panic("unimplemented")
}
