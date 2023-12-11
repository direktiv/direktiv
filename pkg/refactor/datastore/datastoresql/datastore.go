package datastoresql

import (
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/events"
	"github.com/direktiv/direktiv/pkg/refactor/logengine"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
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
// Database transactions management should be handled by the user of this datastore.Store implementation. The caller
// can start a transaction and pass it as Param `db`. After calling different operations on the store, the caller can
// either commit or rollback the connection.

func NewSQLStore(db *gorm.DB, mirrorConfigEncryptionKey string) datastore.Store {
	return &sqlStore{
		db:                        db,
		mirrorConfigEncryptionKey: mirrorConfigEncryptionKey,
	}
}

// Mirror returns mirror store.
func (s *sqlStore) Mirror() mirror.Store {
	return &sqlMirrorStore{
		db:                  s.db,
		configEncryptionKey: s.mirrorConfigEncryptionKey,
	}
}

// FileAnnotations returns file annotations store.
func (s *sqlStore) FileAnnotations() core.FileAnnotationsStore {
	return &sqlFileAnnotationsStore{
		db: s.db,
	}
}

// Logs returns a log store.
func (s *sqlStore) Logs() logengine.LogStore {
	return &sqlLogStore{
		db: s.db,
	}
}

// Secrets returns secrets store.
func (s *sqlStore) Secrets() core.SecretsStore {
	return &sqlSecretsStore{
		db:        s.db,
		secretKey: s.mirrorConfigEncryptionKey,
	}
}

func (s *sqlStore) RuntimeVariables() core.RuntimeVariablesStore {
	return &sqlRuntimeVariablesStore{
		db: s.db,
	}
}

func (s *sqlStore) EventFilter() events.CloudEventsFilterStore {
	return &sqlNamespaceCloudEventFilter{db: s.db}
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

func (s *sqlStore) NamespaceCloudEventFilter() events.CloudEventsFilterStore {
	return &sqlNamespaceCloudEventFilter{db: s.db}
}

func (s *sqlStore) Namespaces() core.NamespacesStore {
	return &sqlNamespacesStore{db: s.db}
}
