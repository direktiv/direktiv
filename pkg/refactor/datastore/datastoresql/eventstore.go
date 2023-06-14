package datastoresql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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

type gormEventHistoryEntry struct {
	ID, NamespaceID          uuid.UUID
	Type, Source, Cloudevent string
	CreatedAt, ReceivedAt    time.Time
}

func (hs *sqlEventHistoryStore) GetByID(ctx context.Context, id uuid.UUID) (*events.Event, error) {
	q := "SELECT id, type, source, cloudevent, namespace_id, received_at, created_at FROM events_history WHERE id = $1 ;"

	e := gormEventHistoryEntry{}
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

type sqlEventTopicsStore struct {
	db *gorm.DB
}

func (s *sqlEventTopicsStore) Append(ctx context.Context, namespaceID uuid.UUID, eventListenerID uuid.UUID, topic string) error {
	q := "INSERT INTO event_topics (id, event_listener_id, namespace_id, topic) VALUES ( $1 , $2 , $3 , $4 );"
	tx := s.db.WithContext(ctx).Exec(q, uuid.NewString(), eventListenerID, namespaceID, topic)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *sqlEventTopicsStore) GetListeners(ctx context.Context, topic string) ([]*events.EventListener, error) {
	q := `SELECT 
	id, namespace_id, created_at, updated_at, deleted, received_events, trigger_type, events_lifespan, event_types, trigger_info
	FROM event_listeners E WHERE E.deleted = 0 and E.id in 
	(SELECT T.event_listener_id FROM event_topics T WHERE topic= $1 )` //,

	res := make([]*gormEventListener, 0)
	tx := s.db.WithContext(ctx).Raw(q, topic).Scan(&res)
	if tx.Error != nil {
		return nil, tx.Error
	}
	conv := make([]*events.EventListener, 0)

	for _, l := range res {
		var trigger events.TriggerInfo
		var ev []*events.Event

		err := json.Unmarshal(l.TriggerInfo, &trigger)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(l.ReceivedEvents, &ev)
		if err != nil {
			return nil, err
		}
		conv = append(conv, &events.EventListener{
			ID:                          l.ID,
			UpdatedAt:                   l.UpdatedAt,
			CreatedAt:                   l.CreatedAt,
			Deleted:                     l.Deleted,
			NamespaceID:                 l.NamespaceID,
			ListeningForEventTypes:      strings.Split(l.EventType, " "),
			LifespanOfReceivedEvents:    l.EventsLifespan,
			TriggerType:                 events.TriggerType(l.TriggerType),
			Trigger:                     trigger,
			ReceivedEventsForAndTrigger: ev,
		})
	}

	return conv, nil
}

var _ events.EventListenerStore = &sqlEventListenerStore{}

type sqlEventListenerStore struct {
	db *gorm.DB
}

func (s *sqlEventListenerStore) Append(ctx context.Context, listener *events.EventListener) error {
	q := `INSERT INTO event_listeners
	 (id, namespace_id, created_at, updated_at, deleted, received_events, trigger_type, events_lifespan, event_types, trigger_info) 
	  VALUES ( $1 , $2 , $3 , $4 , $5 , $6 , $7 , $8 , $9 , $10 );`
	b, err := json.Marshal(listener.Trigger)
	if err != nil {
		return err
	}
	ceB, err := json.Marshal(listener.ReceivedEventsForAndTrigger)
	if err != nil {
		return err
	}

	tx := s.db.WithContext(ctx).Exec(
		q,
		listener.ID,
		listener.NamespaceID,
		listener.CreatedAt,
		listener.UpdatedAt,
		listener.Deleted,
		ceB,
		listener.TriggerType,
		listener.LifespanOfReceivedEvents,
		strings.Join(listener.ListeningForEventTypes, " "),
		b)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *sqlEventListenerStore) Delete(ctx context.Context) error {
	q := "DELETE FROM event_listeners WHERE deleted = TRUE;"
	tx := s.db.WithContext(ctx).Exec(q)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (*sqlEventListenerStore) DeleteAllForInstance(ctx context.Context, instID uuid.UUID) error {
	panic("unimplemented")
}

func (*sqlEventListenerStore) DeleteAllForWorkflow(ctx context.Context, workflowID uuid.UUID) error {
	panic("unimplemented")
}

func (s *sqlEventListenerStore) Get(ctx context.Context, namespace uuid.UUID, limit int, offet int) ([]*events.EventListener, int, error) {
	q := `SELECT 
	id, namespace_id, created_at, updated_at, deleted, received_events, trigger_type, events_lifespan, event_types, trigger_info
	FROM event_listeners WHERE namespace_id = $1 `
	if limit > 0 {
		q += fmt.Sprintf("LIMIT %v", limit)
	}
	if offet > 0 {
		q += fmt.Sprintf("OFFSET %v", offet)
	}
	q += " ORDER BY updated_at DESC;"
	qCount := `SELECT count(id) FROM event_listeners WHERE namespace_id = $1 ;`
	var count int
	tx := s.db.WithContext(ctx).Raw(qCount, namespace).Scan(&count)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	if count == 0 {
		return make([]*events.EventListener, 0), 0, nil
	}
	res := make([]*gormEventListener, 0)
	tx = s.db.WithContext(ctx).Raw(q, namespace).Scan(&res)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	conv := make([]*events.EventListener, 0)

	for _, l := range res {
		var trigger events.TriggerInfo
		var ev []*events.Event

		err := json.Unmarshal(l.TriggerInfo, &trigger)
		if err != nil {
			return nil, 0, err
		}
		err = json.Unmarshal(l.ReceivedEvents, &ev)
		if err != nil {
			return nil, 0, err
		}
		conv = append(conv, &events.EventListener{
			ID:                          l.ID,
			UpdatedAt:                   l.UpdatedAt,
			CreatedAt:                   l.CreatedAt,
			Deleted:                     l.Deleted,
			NamespaceID:                 l.NamespaceID,
			ListeningForEventTypes:      strings.Split(l.EventType, " "),
			LifespanOfReceivedEvents:    l.EventsLifespan,
			TriggerType:                 events.TriggerType(l.TriggerType),
			Trigger:                     trigger,
			ReceivedEventsForAndTrigger: ev,
		})
	}

	return conv, count, nil
}

func (*sqlEventListenerStore) GetAll(ctx context.Context) ([]*events.EventListener, error) {
	panic("unimplemented")
}

type gormEventListener struct {
	Count          int
	ID             uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	NamespaceID    uuid.UUID
	Deleted        bool
	TriggerType    int
	EventType      string
	TriggerInfo    []byte
	EventsLifespan int
	ReceivedEvents []byte
}

func (s *sqlEventListenerStore) GetByID(ctx context.Context, id uuid.UUID) (*events.EventListener, error) {
	q := "SELECT count(id), id, namespace_id, created_at, updated_at ,received_events, trigger_type, events_lifespan, event_types, trigger_info FROM event_listeners WHERE id = $1 ;"
	var l gormEventListener
	tx := s.db.WithContext(ctx).Raw(q, id).First(&l)
	if tx.Error != nil {
		return nil, tx.Error
	}
	var trigger events.TriggerInfo
	var ev []*events.Event

	err := json.Unmarshal(l.TriggerInfo, &trigger)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(l.ReceivedEvents, &ev)
	if err != nil {
		return nil, err
	}

	return &events.EventListener{
		ID:                          l.ID,
		UpdatedAt:                   l.UpdatedAt,
		CreatedAt:                   l.CreatedAt,
		Deleted:                     l.Deleted,
		NamespaceID:                 l.NamespaceID,
		ListeningForEventTypes:      strings.Split(l.EventType, " "),
		LifespanOfReceivedEvents:    l.EventsLifespan,
		TriggerType:                 events.TriggerType(l.TriggerType),
		Trigger:                     trigger,
		ReceivedEventsForAndTrigger: ev,
	}, nil
}

func (s *sqlEventListenerStore) Update(ctx context.Context, listener *events.EventListener, more ...*events.EventListener) (error, []error) {
	q := `UPDATE event_listeners SET
	 updated_at = $1 , deleted = $2, received_events = $3 WHERE id = $4;`
	b, err := json.Marshal(listener.ReceivedEventsForAndTrigger)
	if err != nil {
		return err, nil
	}
	tx := s.db.WithContext(ctx).Exec(
		q,
		listener.UpdatedAt,
		listener.Deleted,
		b,
		listener.ID)
	if tx.Error != nil {
		return tx.Error, nil
	}
	errs := make([]error, len(more))
	for i := range more {
		e := more[i]
		b, err := json.Marshal(e.ReceivedEventsForAndTrigger)
		if err != nil {
			return err, nil
		}
		tx := s.db.WithContext(ctx).Exec(
			q,
			e.UpdatedAt,
			e.Deleted,
			b,
			e.ID)
		if tx.Error != nil {
			errs[i] = tx.Error
		}
	}
	return nil, errs
}

var _ events.CloudEventsFilterStore = &sqlNamespaceCloudEventFilter{}

type sqlNamespaceCloudEventFilter struct {
	db *gorm.DB
}

func (sf *sqlNamespaceCloudEventFilter) Create(ctx context.Context, nsID uuid.UUID, filterName string, script string) error {
	q := `INSERT INTO events_filters (id, namespace_id, name, jscode) VALUES ( $1 , $2 , $3 , $4 )`
	id := uuid.New()

	tx := sf.db.WithContext(ctx).Raw(
		q, id, nsID, filterName, script)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (sf *sqlNamespaceCloudEventFilter) Delete(ctx context.Context, nsID uuid.UUID, filterName string) error {
	q := `DELETE FROM events_filters WHERE namespace_id = $1 AND name = $2`
	tx := sf.db.WithContext(ctx).Raw(
		q, nsID, filterName)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (sf *sqlNamespaceCloudEventFilter) Get(ctx context.Context, nsID uuid.UUID, filterName string, limit int, offset int) ([]*events.NamespaceCloudEventFilter, int, error) {
	q := `SELECT namespace_id, name, jscode FROM events_filters WHERE namespace_id = $1 `
	qCount := `SELECT count(id) FROM events_filters WHERE namespace_id = $1 `
	var count int
	tx := sf.db.WithContext(ctx).Exec(
		qCount, nsID).Scan(&count)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	if count == 0 {
		return make([]*events.NamespaceCloudEventFilter, 0), 0, nil
	}
	res := make([]*events.NamespaceCloudEventFilter, 0)
	tx = sf.db.WithContext(ctx).Exec(
		q, nsID).Scan(&res)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	return res, count, nil
}

func (sf *sqlNamespaceCloudEventFilter) GetAll(ctx context.Context, nsID uuid.UUID) ([]*events.NamespaceCloudEventFilter, error) {
	q := `SELECT namespace_id, name, jscode FROM events_filters WHERE namespace_id = $1 `
	res := make([]*events.NamespaceCloudEventFilter, 0)
	tx := sf.db.WithContext(ctx).Exec(
		q, nsID).Scan(&res)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return res, nil
}
