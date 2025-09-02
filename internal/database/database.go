package database

import (
	"context"
	"database/sql"
	_ "embed"

	"github.com/direktiv/direktiv/internal/datastore"
	"github.com/direktiv/direktiv/internal/datastore/datasql"
	"github.com/direktiv/direktiv/internal/filestore"
	"github.com/direktiv/direktiv/internal/filestore/filesql"
	"gorm.io/gorm"
)

type DB struct {
	db *gorm.DB
}

func NewDB(db *gorm.DB) *DB {
	return &DB{
		db: db,
	}
}

func (tx *DB) Conn() *gorm.DB {
	return tx.db
}

func (tx *DB) FileStore() filestore.FileStore {
	return filesql.NewStore(tx.db)
}

func (tx *DB) DataStore() datastore.Store {
	return datasql.NewStore(tx.db)
}

func (tx *DB) Commit(ctx context.Context) error {
	return tx.db.WithContext(ctx).Commit().Error
}

func (tx *DB) Rollback() error {
	return tx.db.Rollback().Error
}

func (tx *DB) BeginTx(ctx context.Context, opts ...*sql.TxOptions) (*DB, error) {
	res := tx.db.WithContext(ctx).Begin(opts...)
	if res.Error != nil {
		return nil, res.Error
	}

	return &DB{
		db: res,
	}, nil
}
