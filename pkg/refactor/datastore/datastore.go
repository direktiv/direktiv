package datastore

import (
	"context"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"gorm.io/gorm"
)

type DataStore interface {
	Mirror() mirror.Store
	Namespace() core.NamespaceStore

	Begin(ctx context.Context) (DataStore, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type sqlDataStore struct {
	db *gorm.DB
}

func (s sqlDataStore) Begin(ctx context.Context) (DataStore, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return sqlDataStore{
		db: tx,
	}, nil
}

func (s sqlDataStore) Commit(ctx context.Context) error {
	return s.db.WithContext(ctx).Commit().Error
}

func (s sqlDataStore) Rollback(ctx context.Context) error {
	return s.db.WithContext(ctx).Rollback().Error
}

func (s sqlDataStore) Mirror() mirror.Store {
	//nolint:gosimple
	return sqlMirrorStore{
		db: s.db,
	}
}

func (s sqlDataStore) Namespace() core.NamespaceStore {
	//nolint:gosimple
	return sqlNamespaceStore{
		db: s.db,
	}
}

var _ DataStore = sqlDataStore{}

func NewSQLDataStore(db *gorm.DB) DataStore {
	return sqlDataStore{
		db: db,
	}
}
