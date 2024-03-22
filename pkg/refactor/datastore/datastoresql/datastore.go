package datastoresql

import (
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/events"
	"gorm.io/gorm"
)

type sqlStore struct {
	// database connection.
	db *gorm.DB
	// symmetric encryption key to encrypt and decrypt mirror data.
	mirrorConfigEncryptionKey string
}

var _ datastore.Store = &sqlStore{}

// NewSQLStore builds direktiv data store. Param `db` should be an opened active connection to the database. Param
// `mirrorConfigEncryptionKey` is a symmetric encryption key string used to encrypt and decrypt mirror data.
// Database transactions management should be handled by the user of this datastore.MirrorStore implementation. The caller
// can start a transaction and pass it as Param `db`. After calling different operations on the store, the caller can
// either commit or rollback the connection.

func NewSQLStore(db *gorm.DB, mirrorConfigEncryptionKey string) datastore.Store {
	return &sqlStore{
		db:                        db,
		mirrorConfigEncryptionKey: mirrorConfigEncryptionKey,
	}
}

// Mirror returns mirror store.
func (s *sqlStore) Mirror() datastore.MirrorStore {
	return &sqlMirrorStore{
		db:                  s.db,
		configEncryptionKey: s.mirrorConfigEncryptionKey,
	}
}

// Logs returns a log store.
func (s *sqlStore) NewLogs() datastore.LogStore {
	return &sqlLogNewStore{
		db: s.db,
	}
}

// Secrets returns secrets store.
func (s *sqlStore) Secrets() datastore.SecretsStore {
	return &sqlSecretsStore{
		db:        s.db,
		secretKey: s.mirrorConfigEncryptionKey,
	}
}

func (s *sqlStore) RuntimeVariables() datastore.RuntimeVariablesStore {
	return &sqlRuntimeVariablesStore{
		db: s.db,
	}
}

func (s *sqlStore) StagingEvents() events.StagingEventStore {
	return &sqlStagingEventStore{db: s.db}
}

func (s *sqlStore) EventHistory() events.EventHistoryStore {
	return &sqlEventHistoryStore{db: s.db}
}

func (s *sqlStore) EventListener() events.EventListenerStore {
	return &sqlEventListenerStore{db: s.db}
}

func (s *sqlStore) EventListenerTopics() events.EventTopicsStore {
	return &sqlEventTopicsStore{db: s.db}
}

func (s *sqlStore) Namespaces() datastore.NamespacesStore {
	return &sqlNamespacesStore{db: s.db}
}
