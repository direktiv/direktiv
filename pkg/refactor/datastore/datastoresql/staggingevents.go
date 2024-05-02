package datastoresql

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _ datastore.StagingEventStore = &sqlStagingEventStore{}

type sqlStagingEventStore struct {
	db *gorm.DB
}

func (ss *sqlStagingEventStore) Append(ctx context.Context, events ...*datastore.StagingEvent) ([]*datastore.StagingEvent, []error) {
	q := "INSERT INTO staging_events (id, event_id, type, source, cloudevent, namespace_id, namespace_name, received_at, created_at, delayed_until) VALUES (  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10 )"
	errs := make([]error, len(events))
	for i := range events {
		v := events[i]
		if v.Event.Event == nil {
			panic("event was nil") // TODO handle by logging
		}
		eventByte, err := json.Marshal(v.Event.Event)
		if err != nil {
			errs[i] = err

			continue
		}
		if v.DatabaseID == uuid.Nil {
			v.DatabaseID = uuid.New()
		}
		values := make([]interface{}, 0)
		values = append(values, v.DatabaseID)
		values = append(values, v.Event.Event.ID())
		values = append(values, v.Event.Event.Type())
		values = append(values, v.Event.Event.Source())
		values = append(values, string(eventByte))
		values = append(values, v.Namespace)
		values = append(values, v.NamespaceName)
		values = append(values, v.ReceivedAt)
		values = append(values, time.Now().UTC())
		values = append(values, v.DelayedUntil)
		tx := ss.db.WithContext(ctx).Exec(q, values...)
		if tx.Error != nil {
			errs[i] = tx.Error

			continue
		}
	}

	return events, nil
}

type gormStagingEvent struct {
	ID            uuid.UUID
	EventID       string
	Source        string
	Type          string
	Cloudevent    string
	NamespaceID   uuid.UUID
	NamespaceName string
	ReceivedAt    time.Time
	CreatedAt     time.Time
	DelayedUntil  time.Time
}

func (ss *sqlStagingEventStore) GetDelayedEvents(ctx context.Context, currentTime time.Time, limit int, offset int) ([]*datastore.StagingEvent, int, error) {
	q := `SELECT id, source, type, cloudevent, namespace_id, namespace_name, created_at, delayed_until FROM staging_events WHERE delayed_until < $1`

	var count int
	tx := ss.db.WithContext(ctx).Raw(`SELECT COUNT(id) FROM staging_events WHERE delayed_until < $1`, currentTime).Scan(&count)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}

	var res []*gormStagingEvent
	tx = ss.db.WithContext(ctx).Raw(q, currentTime).Offset(offset).Limit(limit).Scan(&res)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}

	ev := make([]*datastore.StagingEvent, 0, len(res))

	for _, gse := range res {
		var finalCE ce.Event
		err := json.Unmarshal([]byte(gse.Cloudevent), &finalCE)
		if err != nil {
			return nil, 0, fmt.Errorf("res len(): %v, event: %v, err: %w ", len(res), gse.Cloudevent, err)
		}
		ev = append(ev, &datastore.StagingEvent{Event: &datastore.Event{Namespace: gse.NamespaceID, ReceivedAt: gse.ReceivedAt, Event: &finalCE, NamespaceName: gse.NamespaceName}, DatabaseID: gse.ID, DelayedUntil: gse.DelayedUntil})
	}

	return ev, count, nil
}

func (ss *sqlStagingEventStore) DeleteByDatabaseIDs(ctx context.Context, databaseIDs ...uuid.UUID) error {
	if len(databaseIDs) == 0 {
		return nil
	}

	q := "DELETE FROM staging_events WHERE id IN (?)"
	tx := ss.db.WithContext(ctx).Exec(q, databaseIDs)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}
