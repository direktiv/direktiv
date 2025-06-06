package datasql

import (
	"github.com/direktiv/direktiv/pkg/datastore"
	"gorm.io/gorm"
)

type store struct {
	// database connection.
	db *gorm.DB
}

var _ datastore.Store = &store{}

// NewStore builds direktiv data store. Param `db` should be an opened active connection to the database. Param
// `mirrorConfigEncryptionKey` is a symmetric encryption key string used to encrypt and decrypt mirror data.
// Database transactions management should be handled by the user of this datastore.MirrorStore implementation. The caller
// can start a transaction and pass it as Param `db`. After calling different operations on the store, the caller can
// either commit or rollback the connection.

func NewStore(db *gorm.DB) datastore.Store {
	return &store{db: db}
}

func (s *store) Mirror() datastore.MirrorStore {
	return &sqlMirrorStore{db: s.db}
}

func (s *store) Secrets() datastore.SecretsStore {
	return &sqlSecretsStore{db: s.db}
}

func (s *store) HeartBeats() datastore.HeartBeatsStore {
	return &sqlHeartBeatsStore{db: s.db}
}

func (s *store) RuntimeVariables() datastore.RuntimeVariablesStore {
	return &sqlRuntimeVariablesStore{db: s.db}
}

func (s *store) StagingEvents() datastore.StagingEventStore {
	return &sqlStagingEventStore{db: s.db}
}

func (s *store) EventHistory() datastore.EventHistoryStore {
	return &sqlEventHistoryStore{db: s.db}
}

func (s *store) EventListener() datastore.EventListenerStore {
	return &sqlEventListenerStore{db: s.db}
}

func (s *store) EventListenerTopics() datastore.EventTopicsStore {
	return &sqlEventTopicsStore{db: s.db}
}

func (s *store) Namespaces() datastore.NamespacesStore {
	return &sqlNamespacesStore{db: s.db}
}

func (s *store) Traces() datastore.TracesStore {
	return &sqlTracesStore{db: s.db}
}
