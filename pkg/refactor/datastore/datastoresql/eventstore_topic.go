package datastoresql

import (
	"context"

	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _ datastore.EventTopicsStore = &sqlEventTopicsStore{}

type sqlEventTopicsStore struct {
	db *gorm.DB
}

func (s *sqlEventTopicsStore) Append(ctx context.Context, namespaceID uuid.UUID, namespace string, eventListenerID uuid.UUID, topic string, filter string) error {
	q := "INSERT INTO event_topics (id, event_listener_id, namespace_id, namespace, topic, filter) VALUES ( $1 , $2 , $3 , $4 , $5, $6 );"
	tx := s.db.WithContext(ctx).Exec(q, uuid.NewString(), eventListenerID, namespaceID, namespace, topic, filter)
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

func (s *sqlEventTopicsStore) GetListeners(ctx context.Context, topic string) ([]*datastore.EventListener, error) {
	q := `SELECT 
	id, namespace_id, namespace,  created_at, updated_at, deleted, received_events, trigger_type, events_lifespan, event_types, trigger_info, metadata, glob_gates
	FROM event_listeners E WHERE E.deleted = false and E.id in 
	(SELECT T.event_listener_id FROM event_topics T WHERE topic= $1 )` //,

	res := make([]*gormEventListener, 0)
	tx := s.db.WithContext(ctx).Raw(q, topic).Scan(&res)
	if tx.Error != nil {
		return nil, tx.Error
	}
	conv, err := convertListerners(res)
	if err != nil {
		return nil, err
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
