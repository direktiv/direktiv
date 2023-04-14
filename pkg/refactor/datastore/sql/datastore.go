package sql

import (
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"gorm.io/gorm"
)

type sqlStore struct {
	db                        *gorm.DB
	mirrorConfigEncryptionKey string
}

var _ datastore.Store = &sqlStore{}

func NewSQLStore(db *gorm.DB, mirrorConfigEncryptionKey string) datastore.Store {
	return &sqlStore{
		db:                        db,
		mirrorConfigEncryptionKey: mirrorConfigEncryptionKey,
	}
}

func (s *sqlStore) Mirror() mirror.Store {
	return &sqlMirrorStore{
		db:                  s.db,
		configEncryptionKey: s.mirrorConfigEncryptionKey,
	}
}

func (s *sqlStore) FileAnnotations() core.FileAnnotationsStore {
	return &sqlFileAnnotationsStore{
		db: s.db,
	}
}
