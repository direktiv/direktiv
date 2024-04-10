package database

import (
	"context"
	"database/sql"
	_ "embed"
	"log"
	"os"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/datastore/datastoresql"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore/filestoresql"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore/instancestoresql"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//go:embed db_schema.sql
var Schema string

func sqlLiteSchema() string {
	convertTypes := map[string]string{
		"uuid":        "text",
		"timestamptz": "datetime",
		"bytea":       "blob",
		"boolean":     "numeric",
		"serial":      "integer",
	}

	liteSchema := Schema

	for k, v := range convertTypes {
		liteSchema = strings.ReplaceAll(liteSchema, " "+k+",", " "+v+",")
		liteSchema = strings.ReplaceAll(liteSchema, " "+k+" ", " "+v+" ")
	}
	liteSchema = strings.ReplaceAll(liteSchema, "CREATE UNIQUE INDEX", "--")
	liteSchema = strings.ReplaceAll(liteSchema, "CREATE INDEX", "--")
	liteSchema = strings.ReplaceAll(liteSchema, "ALTER TABLE", "--")
	liteSchema = strings.ReplaceAll(liteSchema, "DROP TABLE", "--")
	liteSchema = strings.ReplaceAll(liteSchema, "DROP INDEX", "--")
	liteSchema = strings.ReplaceAll(liteSchema, "UNIQUE NULLS NOT DISTINCT", "UNIQUE")

	return liteSchema
}

func NewMockGorm() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold: time.Second,   // Slow SQL threshold
				LogLevel:      logger.Silent, // Log level
			},
		),
	})
	if err != nil {
		return nil, err
	}

	res := db.Exec(sqlLiteSchema())

	if res.Error != nil {
		return nil, res.Error
	}

	return db, nil
}

type SQLStore struct {
	db        *gorm.DB
	secretKey string
}

func NewSQLStore(gormDB *gorm.DB, secretKey string) *SQLStore {
	return &SQLStore{
		db:        gormDB,
		secretKey: secretKey,
	}
}

func (tx *SQLStore) FileStore() filestore.FileStore {
	return filestoresql.NewSQLFileStore(tx.db)
}

func (tx *SQLStore) DataStore() datastore.Store {
	return datastoresql.NewSQLStore(tx.db, tx.secretKey)
}

func (tx *SQLStore) InstanceStore() instancestore.Store {
	return instancestoresql.NewSQLInstanceStore(tx.db)
}

func (tx *SQLStore) Commit(ctx context.Context) error {
	return tx.db.WithContext(ctx).Commit().Error
}

func (tx *SQLStore) Rollback() error {
	return tx.db.Rollback().Error
}

func (tx *SQLStore) BeginTx(ctx context.Context, opts ...*sql.TxOptions) (*SQLStore, error) {
	res := tx.db.WithContext(ctx).Begin(opts...)
	if res.Error != nil {
		return nil, res.Error
	}

	return &SQLStore{
		db:        res,
		secretKey: tx.secretKey,
	}, nil
}
