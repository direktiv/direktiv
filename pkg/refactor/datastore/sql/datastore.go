package sql

import (
	"context"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"gorm.io/gorm"
)

type sqlStore struct {
	db *gorm.DB
}

func (s *sqlStore) Begin(ctx context.Context) (datastore.StoreTx, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &sqlStore{
		db: tx,
	}, nil
}

func (s *sqlStore) Commit(ctx context.Context) error {
	return s.db.WithContext(ctx).Commit().Error
}

func (s *sqlStore) Rollback(ctx context.Context) error {
	return s.db.WithContext(ctx).Rollback().Error
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
