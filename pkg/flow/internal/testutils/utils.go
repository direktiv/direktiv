package testutils

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect"
	"github.com/direktiv/direktiv/pkg/flow/database/entwrapper"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// databaseMock is meant to mock a entwrapper.Database for testing.
type databaseMock struct {
	Postgres *embeddedpostgres.EmbeddedPostgres
	Entw     entwrapper.Database
}

// starts a in memory postgres database and passes it to ent.
func DatabaseWrapper() (databaseMock, error) {
	dbm := databaseMock{}
	dbm.Postgres = embeddedpostgres.NewDatabase()
	err := dbm.Postgres.Start()
	if err != nil {
		return dbm, err
	}
	client, err := ent.Open(dialect.Postgres, "host=localhost port=5432 user=postgres dbname=postgres password=postgres sslmode=disable ")
	if err != nil {
		_ = dbm.Postgres.Stop()
		return dbm, err
	}
	ctx := context.Background()

	if err := client.Schema.Create(ctx); err != nil {
		_ = dbm.Postgres.Stop()
		return dbm, err
	}
	sugar := zap.S()
	dbm.Entw = entwrapper.Database{
		Client: client,
		Sugar:  sugar,
	}
	return dbm, nil
}

// stopping the database also deletes the stored data.
// defer to this method to free the used port.
func (dbm databaseMock) StopDB() {
	dbm.Entw.Close()
	defer func() {
		err := dbm.Postgres.Stop()
		if err != nil {
			fmt.Sprintln(err)
		}
	}()
}

func ObservedLogger() (*zap.SugaredLogger, *observer.ObservedLogs) {
	observed, telemetrylogs := observer.New(zapcore.DebugLevel)
	sugar := zap.New(observed).Sugar()
	return sugar, telemetrylogs
}
