package database

import (
	"context"
	"database/sql"
	_ "embed"

	"github.com/direktiv/direktiv/internal/datastore"
	"github.com/direktiv/direktiv/internal/datastore/datasql"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/filestore/filesql"
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

func (d *DB) Conn() *gorm.DB {
	return d.db
}

func (d *DB) FileStore() filestore.FileStore {
	return filesql.NewStore(d.db)
}

func (d *DB) DataStore() datastore.Store {
	return datasql.NewStore(d.db)
}

func (d *DB) BeginTx(ctx context.Context, opts ...*sql.TxOptions) (*DB, error) {
	res := d.db.WithContext(ctx).Begin(opts...)
	if res.Error != nil {
		return nil, res.Error
	}

	return &DB{
		db: res,
	}, nil
}
