package database

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	natsContainer "github.com/testcontainers/testcontainers-go/modules/nats"
	tsPostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//go:embed db_schema.sql
var Schema string

func NewTestNats(t *testing.T) (string, error) {
	t.Helper()
	ctx := context.Background()

	ctr, err := natsContainer.Run(
		ctx,
		"nats:2.10-alpine",
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := ctr.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// Get nats://<host>:<port>
	uri, err := ctr.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	return uri, nil
}

// nolint:usetesting
func NewTestDB(t *testing.T) (*gorm.DB, error) {
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

func newTestPostgres(dsn string) (*gorm.DB, error) {
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

	return db, nil
}

//nolint:usetesting
func NewTestDBWithNamespace(t *testing.T, namespace string) (*gorm.DB, error) {
	t.Helper()

	db, err := NewTestDB(t)
	if err != nil {
		return nil, err
	}
	res := db.Exec("INSERT INTO namespaces(name) VALUES ($1)", namespace)
	if res.Error != nil {
		return nil, fmt.Errorf("create namespaces, err: %w", res.Error)
	}

	return db, nil
}
