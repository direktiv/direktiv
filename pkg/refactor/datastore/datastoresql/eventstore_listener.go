package datastoresql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _ datastore.EventHistoryStore = &sqlEventHistoryStore{}

var _ datastore.EventListenerStore = &sqlEventListenerStore{}

type sqlEventListenerStore struct {
	db *gorm.DB
}

func (s *sqlEventListenerStore) Append(ctx context.Context, listener *datastore.EventListener) error {
	q := `INSERT INTO event_listeners
	 (id, namespace_id, namespace, created_at, updated_at, deleted, received_events, trigger_type, events_lifespan, event_types, trigger_info, metadata, event_context_filters) 
	  VALUES ( $1 , $2 , $3 , $4 , $5 , $6 , $7 , $8 , $9 , $10 , $11, $12, $13);`

	trigger := triggerInfo{
		WorkflowID: listener.TriggerWorkflow,
		InstanceID: listener.TriggerInstance,
	}
	b, err := json.Marshal(trigger)
	if err != nil {
		return err
	}
	ceB, err := json.Marshal(listener.ReceivedEventsForAndTrigger)
	if err != nil {
		return err
	}

	filters, err := json.Marshal(listener.EventContextFilters)
	if err != nil {
		return err
	}

	tx := s.db.WithContext(ctx).Exec(
		q,
		listener.ID,
		listener.NamespaceID,
		listener.Namespace,
		listener.CreatedAt,
		listener.UpdatedAt,
		listener.Deleted,
		ceB,
		listener.TriggerType,
		listener.LifespanOfReceivedEvents,
		encodeStrings(listener.ListeningForEventTypes),
		string(b),
		listener.Metadata,
		string(filters))
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

func (s *sqlEventListenerStore) GetOld(ctx context.Context, namespace string, t time.Time) ([]*datastore.EventListener, error) {
	q := `SELECT 
	id, namespace_id, namespace, created_at, updated_at, deleted, received_events, trigger_type, events_lifespan, event_types, trigger_info, metadata, event_context_filters
	FROM event_listeners WHERE namespace = $1 AND created_at < $2`
	q += " ORDER BY created_at DESC LIMIT $3"

	res := make([]*gormEventListener, 0)
	tx := s.db.WithContext(ctx).Raw(q, namespace, t, pageSize).Scan(&res)
	if tx.Error != nil {
		return nil, tx.Error
	}
	conv, err := convertListerners(res)
	if err != nil {
		return nil, err
	}

	return conv, nil
}

func (s *sqlEventListenerStore) Get(ctx context.Context, namespace uuid.UUID, limit int, offset int) ([]*datastore.EventListener, int, error) {
	q := `SELECT 
	id, namespace_id, namespace, created_at, updated_at, deleted, received_events, trigger_type, events_lifespan, event_types, trigger_info, metadata, event_context_filters
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
		return make([]*datastore.EventListener, 0), 0, nil
	}
	res := make([]*gormEventListener, 0)
	tx = s.db.WithContext(ctx).Raw(q, namespace).Scan(&res)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	conv, err := convertListerners(res)
	if err != nil {
		return nil, 0, err
	}

	return conv, count, nil
}

func (s *sqlEventListenerStore) GetAll(ctx context.Context) ([]*datastore.EventListener, error) {
	q := `SELECT 
	id, namespace_id, namespace, created_at, updated_at, deleted, received_events, trigger_type, events_lifespan, event_types, trigger_info, metadata
	FROM event_listeners Where deleted = false`
	q += " ORDER BY created_at DESC;"
	res := make([]*gormEventListener, 0)
	tx := s.db.WithContext(ctx).Raw(q).Scan(&res)
	if tx.Error != nil {
		return nil, tx.Error
	}
	conv, err := convertListerners(res)
	if err != nil {
		return nil, err
	}

	return conv, nil
}

type gormEventListener struct {
	Count               int
	ID                  uuid.UUID
	CreatedAt           time.Time
	UpdatedAt           time.Time
	NamespaceID         uuid.UUID
	Namespace           string
	Deleted             bool
	TriggerType         int
	EventTypes          string
	TriggerInfo         string
	EventsLifespan      int
	ReceivedEvents      []byte
	Metadata            string
	EventContextFilters string
}

func (s *sqlEventListenerStore) GetByID(ctx context.Context, id uuid.UUID) (*datastore.EventListener, error) {
	q := "SELECT id, namespace_id, namespace, created_at, updated_at, received_events, trigger_type, events_lifespan, event_types, trigger_info, metadata, event_context_filters FROM event_listeners WHERE id = $1 ;"
	var l gormEventListener
	tx := s.db.WithContext(ctx).Raw(q, id).First(&l)
	if tx.Error != nil {
		return nil, tx.Error
	}
	var trigger triggerInfo
	var ev []*datastore.Event
	var filters []datastore.EventContextFilter

	err := json.Unmarshal([]byte(l.TriggerInfo), &trigger)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(l.EventContextFilters), &filters)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(l.ReceivedEvents, &ev)
	if err != nil {
		return nil, err
	}
	if len(l.EventContextFilters) > 0 {
		err = json.Unmarshal([]byte(l.EventContextFilters), &filters)
		if err != nil {
			return nil, err
		}
	}

	return &datastore.EventListener{
		ID:                          l.ID,
		UpdatedAt:                   l.UpdatedAt,
		CreatedAt:                   l.CreatedAt,
		Deleted:                     l.Deleted,
		NamespaceID:                 l.NamespaceID,
		Namespace:                   l.Namespace,
		ListeningForEventTypes:      decodeString(l.EventTypes),
		LifespanOfReceivedEvents:    l.EventsLifespan,
		TriggerType:                 datastore.TriggerType(l.TriggerType),
		TriggerWorkflow:             trigger.WorkflowID,
		TriggerInstance:             trigger.InstanceID,
		ReceivedEventsForAndTrigger: ev,
		Metadata:                    l.Metadata,
		EventContextFilters:         filters,
	}, nil
}

func (s *sqlEventListenerStore) UpdateOrDelete(ctx context.Context, listeners []*datastore.EventListener) []error {
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

// encodeStrings uses a custom, non-standard encoding to maintain compatibility
// with existing database entries that used spaces as literal data. This was
// due to a historical specification oversight. Ideally, a structured format
// like JSON should be used for storing arrays or lists.
func encodeStrings(s []string) string {
	encodedStrings := make([]string, len(s))
	for i, str := range s {
		encodedStrings[i] = strings.ReplaceAll(str, " ", "\u00A0")
	}

	return strings.Join(encodedStrings, " ")
}

// decodeString reverses the custom encoding applied by encodeStrings.
func decodeString(s string) []string {
	parts := strings.Split(s, " ")
	for i, part := range parts {
		parts[i] = strings.ReplaceAll(part, "\u00A0", " ")
	}

	return parts
}

func convertListerners(from []*gormEventListener) ([]*datastore.EventListener, error) {
	into := make([]*datastore.EventListener, 0, len(from))
	for _, l := range from {
		var trigger triggerInfo
		var ev []*datastore.Event
		var filters []datastore.EventContextFilter

		err := json.Unmarshal([]byte(l.TriggerInfo), &trigger)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(l.ReceivedEvents, &ev)
		if err != nil {
			return nil, err
		}
		if len(l.EventContextFilters) > 0 {
			err = json.Unmarshal([]byte(l.EventContextFilters), &filters)
			if err != nil {
				return nil, err
			}
		}
		into = append(into, &datastore.EventListener{
			ID:                          l.ID,
			UpdatedAt:                   l.UpdatedAt,
			CreatedAt:                   l.CreatedAt,
			Deleted:                     l.Deleted,
			NamespaceID:                 l.NamespaceID,
			Namespace:                   l.Namespace,
			ListeningForEventTypes:      decodeString(l.EventTypes),
			LifespanOfReceivedEvents:    l.EventsLifespan,
			TriggerType:                 datastore.TriggerType(l.TriggerType),
			TriggerWorkflow:             trigger.WorkflowID,
			TriggerInstance:             trigger.InstanceID,
			ReceivedEventsForAndTrigger: ev,
			Metadata:                    l.Metadata,
			EventContextFilters:         filters,
		})
	}

	return into, nil
}
