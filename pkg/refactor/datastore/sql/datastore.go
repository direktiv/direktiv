package sql

import (
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"gorm.io/gorm"
)

type sqlStore struct {
	db *gorm.DB
}

var _ datastore.Store = &sqlStore{}

func NewSQLStore(db *gorm.DB) datastore.Store {
	return &sqlStore{
		db: db,
	}
}

func (s *sqlStore) Mirror() mirror.Store {
	return &sqlMirrorStore{
		db: s.db,
	}
}

func (s *sqlStore) FileAttributes() core.FileAttributesStore {
	return &sqlFileAttributesStore{
		db: s.db,
	}
}
