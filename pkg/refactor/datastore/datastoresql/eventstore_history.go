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
	q := "INSERT INTO events_history (id, type, source, cloudevent, namespace_id, namespace, received_at, created_at) VALUES ( $1 , $2 , $3 , $4 , $5 , $6, $7, $8 )"
	errs := make([]error, len(events))
	for i := range events {
		v := events[i]
		if v.Event == nil {
			errs[i] = fmt.Errorf("event was nil")

			continue
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
		values = append(values, v.NamespaceName)
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
	Namespace                string
	Type, Source, Cloudevent string
	CreatedAt, ReceivedAt    time.Time
}

func (hs *sqlEventHistoryStore) GetOld(ctx context.Context, namespace string, t time.Time, keyAndValues ...string) ([]*events.Event, error) {
	return hs.getEvents(ctx, namespace, t, "where (namespace= ? and received_at < ? )", keyAndValues...)
}

func (hs *sqlEventHistoryStore) GetNew(ctx context.Context, namespace string, t time.Time, keyAndValues ...string) ([]*events.Event, error) {
	return hs.getEvents(ctx, namespace, t, "where (namespace= ? and received_at >= ? )", keyAndValues...)
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
	q := `SELECT id, type, source, cloudevent, namespace_id, namespace, received_at, created_at FROM events_history
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
		conv = append(conv, &events.Event{Namespace: v.NamespaceID, NamespaceName: v.Namespace, ReceivedAt: v.ReceivedAt, Event: &finalCE})
	}

	return conv, count, nil
}

func (hs *sqlEventHistoryStore) GetAll(ctx context.Context) ([]*events.Event, error) {
	q := "SELECT id, type, source, cloudevent, namespace_id, namespace, received_at, created_at FROM events_history;"
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
	q := "SELECT id, type, source, cloudevent, namespace_id, namespace, received_at, created_at FROM events_history WHERE id = $1 ;"

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

	return &events.Event{Namespace: e.NamespaceID, NamespaceName: e.Namespace, ReceivedAt: e.ReceivedAt, Event: &finalCE}, nil
}

func unzipAndAppendToQueryParams(qs []string, qv []interface{}, keyAndValues []string) ([]string, []interface{}) {
	for i := 0; i < len(keyAndValues); i += 2 {
		v := keyAndValues[i+1]
		if keyAndValues[i] == "created_before" {
			qs = append(qs, " and created_at < ?")
			qv = append(qv, v)
		}
		if keyAndValues[i] == "created_after" {
			qs = append(qs, " and created_at >= ?")
			qv = append(qv, v)
		}
		if keyAndValues[i] == "received_before" {
			qs = append(qs, " and received_at < ?")
			qv = append(qv, v)
		}
		if keyAndValues[i] == "received_after" {
			qs = append(qs, " and received_at >= ?")
			qv = append(qv, v)
		}
		if keyAndValues[i] == "event_contains" {
			qs = append(qs, " and cloudevent like ?")
			qv = append(qv, fmt.Sprintf("%%%v%%", v))
		}
		if keyAndValues[i] == "type_contains" {
			qs = append(qs, " and type like ?")
			qv = append(qv, fmt.Sprintf("%%%v%%", v))
		}
	}

	return qs, qv
}

func (hs *sqlEventHistoryStore) getEvents(ctx context.Context, namespace string, t time.Time, whereClause string, keyAndValues ...string) ([]*events.Event, error) {
	if len(keyAndValues)%2 != 0 {
		return nil, fmt.Errorf("keyAndValues have to be a pair of keys and values")
	}

	qs := []string{whereClause}
	qv := []interface{}{namespace, t}
	qs, qv = unzipAndAppendToQueryParams(qs, qv, keyAndValues)
	qv = append(qv, pageSize)

	q := fmt.Sprintf(`SELECT id, type, source, cloudevent, namespace_id, namespace, received_at, created_at FROM events_history
	%v ORDER BY created_at DESC LIMIT ?`, strings.Join(qs, ""))

	res := make([]gormEventHistoryEntry, 0, pageSize)
	tx := hs.db.WithContext(ctx).Raw(q, qv...).Scan(&res)
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
		conv = append(conv, &events.Event{Namespace: v.NamespaceID, NamespaceName: v.Namespace, ReceivedAt: v.ReceivedAt, Event: &finalCE})
	}

	return conv, nil
}
