package datastoresql

import (
	"context"
	"encoding/json"
	"fmt"
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
	res := make([]*events.Event, 0)
	if len(keyAndValues)%2 != 0 {
		return nil, 0, fmt.Errorf("keyAnValues have to be a pair of keys and values")
	}
	qs := make([]string, 0)
	qv := make([]interface{}, 0)
	qs = append(qs, "where namespace_id= $%v ")
	qv = append(qv, namespace)

	for i := 0; i < len(keyAndValues); i += 2 {
		v := keyAndValues[i+1]
		if keyAndValues[i] == "created_before" {
			qs = append(qs, " and created_at > $%v")
			qv = append(qv, v)
		}
		if keyAndValues[i] == "created_after" {
			qs = append(qs, " and created_at <= $%v")
			qv = append(qv, v)
		}
		if keyAndValues[i] == "received_before" {
			qs = append(qs, " and received_at > $%v")
			qv = append(qv, v)
		}
		if keyAndValues[i] == "received_after" {
			qs = append(qs, " and received_at <= $%v")
			qv = append(qv, v)
		}
		if keyAndValues[i] == "event_contains" {
			qs = append(qs, " and cloudevent like $%v")
			qv = append(qv, fmt.Sprintf("%%%v%%", v))
		}
		if keyAndValues[i] == "type_contains" {
			qs = append(qs, " and type like $%v")
			qv = append(qv, fmt.Sprintf("%%%v%%", v))
		}
	}

	tail := ""
	i := 0
	var x string
	q := `SELECT id, type, source, cloudevent, namespace_id, received_at, created_at FROM events_history
	%v ORDER BY created_at DESC`

	for i, x = range qs {
		tail += fmt.Sprintf(x, i+1)
	}
	i++
	if limit > 0 {
		i++
		q += fmt.Sprintf(" LIMIT $%v ", i)
	}
	if offset > 0 {
		i++
		q += fmt.Sprintf(" OFFSET $%v ", i)
	}
	qv = append(qv, limit, offset)
	q = fmt.Sprintf(q, tail)

	qCount := `select count(id) as count from events_history `
	qCount += tail + ";"
	count := 0
	tx := hs.db.Raw(qCount, qv[:len(qv)-2]...).Scan(&count)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	rows, err := hs.db.WithContext(ctx).Raw(q, qv...).Rows()
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

	return res, count, nil
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

type gormEventHitoryEntry struct {
	ID, NamespaceID          uuid.UUID
	Type, Source, Cloudevent string
	CreatedAt, ReceivedAt    time.Time
}

func (hs *sqlEventHistoryStore) GetByID(ctx context.Context, id uuid.UUID) (*events.Event, error) {
	q := "SELECT id, type, source, cloudevent, namespace_id, received_at, created_at FROM events_history WHERE id = $1 ;"

	e := gormEventHitoryEntry{}
	tx := hs.db.WithContext(ctx).Raw(q, id).Scan(&e)
	if tx.Error != nil {
		return nil, tx.Error
	}

	var finalCE event.Event
	err := json.Unmarshal([]byte(e.Cloudevent), &finalCE)
	if err != nil {
		return nil, err
	}
	
	return &events.Event{Namespace: e.NamespaceID, ReceivedAt: e.ReceivedAt, Event: &finalCE}, nil
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
