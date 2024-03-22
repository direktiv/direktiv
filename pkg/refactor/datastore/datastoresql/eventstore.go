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

func (hs *sqlEventHistoryStore) Append(ctx context.Context, events []*events.Event) ([]*events.Event, []error) {
	q := "INSERT INTO events_history (id, type, source, cloudevent, namespace_id, received_at, created_at) VALUES ( $1 , $2 , $3 , $4 , $5 , $6, $7 )"
	errs := make([]error, len(events))
	for i := range events {
		v := events[i]
		if v.Event == nil {
			panic("event was nil") // TODO hadle by logging
		}
		eventByte, err := json.Marshal(v.Event)
		if err != nil {
			errs[i] = err

			continue
		}
		values := make([]interface{}, 0)
		values = append(values, v.Event.ID())
		values = append(values, v.Event.Type())
		values = append(values, v.Event.Source())
		values = append(values, string(eventByte))
		values = append(values, v.Namespace)
		values = append(values, v.ReceivedAt)
		values = append(values, time.Now().UTC())
		tx := hs.db.WithContext(ctx).Exec(q, values...)
		if tx.Error != nil {
			errs[i] = tx.Error

			continue
		}
	}

	return events, nil
}

func (hs *sqlEventHistoryStore) DeleteOld(ctx context.Context, sinceWhen time.Time) error {
	q := "DELETE FROM events_history WHERE received_at < $1;"
	tx := hs.db.WithContext(ctx).Exec(q, sinceWhen)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

type gormEventHistoryEntry struct {
	ID                       string
	NamespaceID              uuid.UUID
	Type, Source, Cloudevent string
	CreatedAt, ReceivedAt    time.Time
}

func (hs *sqlEventHistoryStore) Get(ctx context.Context, limit int, offset int, namespace uuid.UUID, keyAndValues ...string) ([]*events.Event, int, error) {
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
			qs = append(qs, " and created_at < $%v")
			qv = append(qv, v)
		}
		if keyAndValues[i] == "created_after" {
			qs = append(qs, " and created_at >= $%v")
			qv = append(qv, v)
		}
		if keyAndValues[i] == "received_before" {
			qs = append(qs, " and received_at < $%v")
			qv = append(qv, v)
		}
		if keyAndValues[i] == "received_after" {
			qs = append(qs, " and received_at >= $%v")
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
	o := 0
	if limit > 0 {
		i++
		q += fmt.Sprintf(" LIMIT $%v ", i)
		qv = append(qv, limit)
		o++
	}
	if offset > 0 {
		i++
		q += fmt.Sprintf(" OFFSET $%v ", i)
		qv = append(qv, offset)
		o++
	}
	q = fmt.Sprintf(q, tail)

	qCount := `select count(id) from events_history `
	qCount += tail + ";"
	count := 0
	tx := hs.db.Raw(qCount, qv[:len(qv)-o]...).Scan(&count)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	res := make([]*gormEventHistoryEntry, 0)

	tx = hs.db.WithContext(ctx).Raw(q, qv...).Scan(&res)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}

	conv := make([]*events.Event, 0, len(res))

	for _, v := range res {
		var finalCE event.Event
		err := json.Unmarshal([]byte(v.Cloudevent), &finalCE)
		if err != nil {
			return nil, 0, err
		}
		conv = append(conv, &events.Event{Namespace: v.NamespaceID, ReceivedAt: v.ReceivedAt, Event: &finalCE})
	}

	return conv, count, nil
}

func (hs *sqlEventHistoryStore) GetAll(ctx context.Context) ([]*events.Event, error) {
	q := "SELECT id, type, source, cloudevent, namespace_id, received_at, created_at FROM events_history;"
	res := make([]*gormEventHistoryEntry, 0)

	tx := hs.db.WithContext(ctx).Raw(q).Scan(&res)
	if tx.Error != nil {
		return nil, tx.Error
	}
	conv := make([]*events.Event, 0, len(res))

	for _, v := range res {
		var finalCE event.Event
		err := json.Unmarshal([]byte(v.Cloudevent), &finalCE)
		if err != nil {
			return nil, err
		}
		conv = append(conv, &events.Event{Namespace: v.NamespaceID, ReceivedAt: v.ReceivedAt, Event: &finalCE})
	}

	return conv, nil
}

func (hs *sqlEventHistoryStore) GetByID(ctx context.Context, id string) (*events.Event, error) {
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

func (s *sqlEventTopicsStore) Append(ctx context.Context, namespaceID uuid.UUID, eventListenerID uuid.UUID, topic string, filter string) error {
	q := "INSERT INTO event_topics (id, event_listener_id, namespace_id, topic, filter) VALUES ( $1 , $2 , $3 , $4 , $5 );"
	tx := s.db.WithContext(ctx).Exec(q, uuid.NewString(), eventListenerID, namespaceID, topic, filter)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

type triggerInfo struct {
	WorkflowID string
	InstanceID string
	Step       int
}

func (s *sqlEventTopicsStore) GetListeners(ctx context.Context, topic string) ([]*events.EventListener, error) {
	q := `SELECT 
	id, namespace_id, created_at, updated_at, deleted, received_events, trigger_type, events_lifespan, event_types, trigger_info, metadata, glob_gates
	FROM event_listeners E WHERE E.deleted = false and E.id in 
	(SELECT T.event_listener_id FROM event_topics T WHERE topic= $1 )` //,

	res := make([]*gormEventListener, 0)
	tx := s.db.WithContext(ctx).Raw(q, topic).Scan(&res)
	if tx.Error != nil {
		return nil, tx.Error
	}
	conv := make([]*events.EventListener, 0)

	conv, err := convertListeners(res, conv)
	if err != nil {
		return nil, err
	}

	return conv, nil
}

func convertListeners(res []*gormEventListener, conv []*events.EventListener) ([]*events.EventListener, error) {
	for _, l := range res {
		var trigger triggerInfo
		var ev []*events.Event

		err := json.Unmarshal([]byte(l.TriggerInfo), &trigger)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(l.ReceivedEvents, &ev)
		if err != nil {
			return nil, err
		}
		var glob map[string]string
		err = json.Unmarshal([]byte(l.GlobGates), &glob)
		if err != nil {
			return nil, err
		}

		conv = append(conv, &events.EventListener{
			ID:                          l.ID,
			UpdatedAt:                   l.UpdatedAt,
			CreatedAt:                   l.CreatedAt,
			Deleted:                     l.Deleted,
			NamespaceID:                 l.NamespaceID,
			ListeningForEventTypes:      strings.Split(l.EventTypes, " "),
			LifespanOfReceivedEvents:    l.EventsLifespan,
			TriggerType:                 events.TriggerType(l.TriggerType),
			TriggerWorkflow:             trigger.WorkflowID,
			TriggerInstance:             trigger.InstanceID,
			TriggerInstanceStep:         trigger.Step,
			ReceivedEventsForAndTrigger: ev,
			Metadata:                    l.Metadata,
			GlobGatekeepers:             glob,
		})
	}

	return conv, nil
}

func (s *sqlEventTopicsStore) Delete(ctx context.Context, eventListenerID uuid.UUID) error {
	q := "DELETE FROM event_topics WHERE event_listener_id = $1;"
	tx := s.db.WithContext(ctx).Exec(q, eventListenerID)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

var _ events.EventListenerStore = &sqlEventListenerStore{}

type sqlEventListenerStore struct {
	db *gorm.DB
}

func (s *sqlEventListenerStore) Append(ctx context.Context, listener *events.EventListener) error {
	q := `INSERT INTO event_listeners
	 (id, namespace_id, created_at, updated_at, deleted, received_events, trigger_type, events_lifespan, event_types, trigger_info, metadata, glob_gates) 
	  VALUES ( $1 , $2 , $3 , $4 , $5 , $6 , $7 , $8 , $9 , $10 , $11, $12);`

	trigger := triggerInfo{
		WorkflowID: listener.TriggerWorkflow,
		InstanceID: listener.TriggerInstance,
		Step:       listener.TriggerInstanceStep,
	}
	b, err := json.Marshal(trigger)
	if err != nil {
		return err
	}
	ceB, err := json.Marshal(listener.ReceivedEventsForAndTrigger)
	if err != nil {
		return err
	}

	glob, err := json.Marshal(listener.GlobGatekeepers)
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
		string(b),
		listener.Metadata,
		string(glob))
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

func (s *sqlEventListenerStore) DeleteAllForInstance(ctx context.Context, instID uuid.UUID) ([]*uuid.UUID, error) {
	res := []*uuid.UUID{}

	q := "SELECT id FROM event_listeners WHERE trigger_info like $1;"
	val := fmt.Sprintf("%%%v%%", instID)
	tx := s.db.WithContext(ctx).Exec(q, val).Scan(&res)
	if tx.Error != nil {
		return nil, tx.Error
	}

	for _, u := range res {
		err := s.DeleteByID(ctx, *u)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (s *sqlEventListenerStore) DeleteByID(ctx context.Context, id uuid.UUID) error {
	q := "DELETE FROM event_listeners WHERE id = $1;"
	tx := s.db.WithContext(ctx).Exec(q, id)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *sqlEventListenerStore) DeleteAllForWorkflow(ctx context.Context, workflowID uuid.UUID) ([]*uuid.UUID, error) {
	res := []*uuid.UUID{}

	q := "SELECT id FROM event_listeners WHERE trigger_info like $1;"
	val := fmt.Sprintf("%%%v%%", workflowID)
	tx := s.db.WithContext(ctx).Raw(q, val).Scan(&res)
	if tx.Error != nil {
		return nil, tx.Error
	}

	for _, u := range res {
		err := s.DeleteByID(ctx, *u)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (s *sqlEventListenerStore) Get(ctx context.Context, namespace uuid.UUID, limit int, offset int) ([]*events.EventListener, int, error) {
	q := `SELECT 
	id, namespace_id, created_at, updated_at, deleted, received_events, trigger_type, events_lifespan, event_types, trigger_info, metadata, glob_gates
	FROM event_listeners WHERE namespace_id = $1 `
	q += " ORDER BY created_at DESC "
	if limit > 0 {
		q += fmt.Sprintf(" LIMIT %v", limit)
	}
	if offset > 0 {
		q += fmt.Sprintf(" OFFSET %v", offset)
	}
	qCount := `SELECT count(id) FROM event_listeners WHERE namespace_id = $1 and deleted = false;`
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
		var trigger triggerInfo
		var ev []*events.Event
		err := json.Unmarshal([]byte(l.TriggerInfo), &trigger)
		if err != nil {
			return nil, 0, err
		}
		err = json.Unmarshal(l.ReceivedEvents, &ev)
		if err != nil {
			return nil, 0, err
		}
		var glob map[string]string
		err = json.Unmarshal([]byte(l.GlobGates), &glob)
		if err != nil {
			return nil, 0, err
		}
		conv = append(conv, &events.EventListener{
			ID:                          l.ID,
			UpdatedAt:                   l.UpdatedAt,
			CreatedAt:                   l.CreatedAt,
			Deleted:                     l.Deleted,
			NamespaceID:                 l.NamespaceID,
			ListeningForEventTypes:      strings.Split(l.EventTypes, " "),
			LifespanOfReceivedEvents:    l.EventsLifespan,
			TriggerType:                 events.TriggerType(l.TriggerType),
			TriggerWorkflow:             trigger.WorkflowID,
			TriggerInstance:             trigger.InstanceID,
			TriggerInstanceStep:         trigger.Step,
			ReceivedEventsForAndTrigger: ev,
			Metadata:                    l.Metadata,
			GlobGatekeepers:             glob,
		})
	}

	return conv, count, nil
}

func (s *sqlEventListenerStore) GetAll(ctx context.Context) ([]*events.EventListener, error) {
	q := `SELECT 
	id, namespace_id, created_at, updated_at, deleted, received_events, trigger_type, events_lifespan, event_types, trigger_info, metadata
	FROM event_listeners Where deleted = false`
	q += " ORDER BY created_at DESC;"
	res := make([]*gormEventListener, 0)
	tx := s.db.WithContext(ctx).Raw(q).Scan(&res)
	if tx.Error != nil {
		return nil, tx.Error
	}
	conv := make([]*events.EventListener, 0)
	conv, err := convertListeners(res, conv)
	if err != nil {
		return nil, err
	}

	return conv, nil
}

type gormEventListener struct {
	Count          int
	ID             uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	NamespaceID    uuid.UUID
	Deleted        bool
	TriggerType    int
	EventTypes     string
	TriggerInfo    string
	EventsLifespan int
	ReceivedEvents []byte
	Metadata       string
	GlobGates      string
}

func (s *sqlEventListenerStore) GetByID(ctx context.Context, id uuid.UUID) (*events.EventListener, error) {
	q := "SELECT count(id), id, namespace_id, created_at, updated_at, received_events, trigger_type, events_lifespan, event_types, trigger_info, metadata, glob_gates FROM event_listeners WHERE id = $1 ;"
	var l gormEventListener
	tx := s.db.WithContext(ctx).Raw(q, id).First(&l)
	if tx.Error != nil {
		return nil, tx.Error
	}
	var trigger triggerInfo
	var ev []*events.Event
	var glob map[string]string

	err := json.Unmarshal([]byte(l.TriggerInfo), &trigger)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(l.GlobGates), &glob)
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
		ListeningForEventTypes:      strings.Split(l.EventTypes, " "),
		LifespanOfReceivedEvents:    l.EventsLifespan,
		TriggerType:                 events.TriggerType(l.TriggerType),
		TriggerWorkflow:             trigger.WorkflowID,
		TriggerInstance:             trigger.InstanceID,
		TriggerInstanceStep:         trigger.Step,
		ReceivedEventsForAndTrigger: ev,
		Metadata:                    l.Metadata,
		GlobGatekeepers:             glob,
	}, nil
}

func (s *sqlEventListenerStore) UpdateOrDelete(ctx context.Context, listeners []*events.EventListener) []error {
	q := `UPDATE event_listeners SET
	 updated_at = $1 , deleted = $2, received_events = $3 WHERE id = $4;`

	errs := make([]error, len(listeners))
	for i := range listeners {
		e := listeners[i]
		if e.Deleted {
			err := s.DeleteByID(ctx, e.ID)
			if err != nil {
				errs[i] = err
			}

			continue
		}
		b, err := json.Marshal(e.ReceivedEventsForAndTrigger)
		if err != nil {
			errs[i] = err

			continue
		}
		tx := s.db.WithContext(ctx).Exec(
			q,
			e.UpdatedAt,
			e.Deleted,
			string(b),
			e.ID)
		if tx.Error != nil {
			errs[i] = tx.Error

			continue
		}
	}

	return errs
}
