package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/datastore/datasql"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/filestore/filesql"
	"github.com/direktiv/direktiv/pkg/instancestore"
	"github.com/direktiv/direktiv/pkg/instancestore/instancestoresql"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tsPostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//go:embed db_schema.sql
var Schema string

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

func (tx *DB) InstanceStore() instancestore.Store {
	return instancestoresql.NewSQLInstanceStore(tx.db)
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

func NewTestDB(t *testing.T) (*DB, error) {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := tsPostgres.Run(ctx, "postgres:15.3-alpine",
		tsPostgres.WithDatabase("mydb"),
		tsPostgres.WithUsername("myadmin"),
		tsPostgres.WithPassword("mypassword"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate pgContainer: %s", err)
		}
	})
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	return newTestPostgres(connStr)
}

func newTestPostgres(dsn string) (*DB, error) {
	gormConf := &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				LogLevel:                  logger.Silent,
				IgnoreRecordNotFoundError: true,
			},
		),
	}
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: false, // disables implicit prepared statement usage
		// Conn:                 edb.DB(),
	}), gormConf)
	if err != nil {
		return nil, fmt.Errorf("connecting to db, err: %w", err)
	}

	res := db.Exec(Schema)
	if res.Error != nil {
		return nil, fmt.Errorf("creating schema, err: %w", res.Error)
	}
	res = db.Exec("DELETE FROM namespaces;")
	if res.Error != nil {
		return nil, fmt.Errorf("delete namespaces, err: %w", res.Error)
	}

	return NewDB(db), nil
}

func NewTestDBWithNamespace(t *testing.T, namespace string) (*DB, *datastore.Namespace, error) {
	t.Helper()

	db, err := NewTestDB(t)
	if err != nil {
		return nil, nil, err
	}
	ns, err := db.DataStore().Namespaces().Create(context.Background(), &datastore.Namespace{
		Name: namespace,
	})
	if err != nil {
		return nil, nil, err
	}

	return db, ns, nil
}
