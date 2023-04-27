package testutils

import (
	"context"
	"fmt"
	"log"
	"os"

	"entgo.io/ent/dialect"
	"github.com/direktiv/direktiv/pkg/flow/database/entwrapper"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// databaseMock is meant to mock a entwrapper.Database for testing.
type databaseMock struct {
	Postgres *embeddedpostgres.EmbeddedPostgres
	Entw     entwrapper.Database
}

func DatabaseGorm() (*gorm.DB, func() error, error) {
	pdb := embeddedpostgres.NewDatabase()
	err := pdb.Start()
	if err != nil {
		return nil, nil, err
	}
	ps := "postgres"
	cleanup := pdb.Stop
	host := "localhost"
	user := ps
	password := ps
	dbname := ps
	port := "5432"
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: false, // disables implicit prepared statement usage
		// Conn:                 edb.DB(),
	}), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				LogLevel: logger.Silent,
			},
		),
	})
	if err != nil {
		return nil, cleanup, err
	}
	client, err := ent.Open(dialect.Postgres, "host=localhost port=5432 user=postgres dbname=postgres password=postgres sslmode=disable ")
	if err != nil {
		_ = pdb.Stop()
		return nil, nil, err
	}
	ctx := context.Background()

	if err := client.Schema.Create(ctx); err != nil {
		_ = pdb.Stop()
		return nil, nil, err
	}

	tx := gormDB.Exec(`
	 CREATE TABLE IF NOT EXISTS "filesystem_roots"
			(
				"id" uuid,
				"created_at" timestamptz,
				"updated_at" timestamptz,
				PRIMARY KEY ("id"),
				CONSTRAINT "fk_namespaces_filesystem_roots"
				FOREIGN KEY ("id") REFERENCES "namespaces"("oid") ON DELETE CASCADE ON UPDATE CASCADE
				);
	 CREATE TABLE IF NOT EXISTS "filesystem_files"
			(
				"id" uuid,
				"path" text,
				"depth" integer,
				"typ" text,
				"root_id" uuid,
				"created_at" timestamptz,
				"updated_at" timestamptz,
				PRIMARY KEY ("id"),
				CONSTRAINT "fk_filesystem_roots_filesystem_files"
				FOREIGN KEY ("root_id") REFERENCES "filesystem_roots"("id") ON DELETE CASCADE ON UPDATE CASCADE
				);
	 CREATE TABLE IF NOT EXISTS "filesystem_revisions"
			(
				"id" uuid,
				"tags" text,
				"is_current" boolean,
				"data" bytea,
				"checksum" text,
				"file_id" uuid,
				"created_at" timestamptz,
				"updated_at" timestamptz,
				PRIMARY KEY ("id"),
				CONSTRAINT "fk_filesystem_files_filesystem_revisions"
				FOREIGN KEY ("file_id") REFERENCES "filesystem_files"("id") ON DELETE CASCADE ON UPDATE CASCADE
				);
	 CREATE TABLE IF NOT EXISTS "file_annotations"
			(
				"file_id" uuid,
				"data" text,
				"created_at" timestamptz,
				"updated_at" timestamptz,
				PRIMARY KEY ("file_id"),
				CONSTRAINT "fk_filesystem_files_file_annotations"
				FOREIGN KEY ("file_id") REFERENCES "filesystem_files"("id") ON DELETE CASCADE ON UPDATE CASCADE
				);
	 CREATE TABLE IF NOT EXISTS "mirror_configs" 
	 		(
	 		    "namespace_id" uuid,
	 		    "url" text,
	 		    "git_ref" text,
	 		    "git_commit_hash" text,
	 		    "public_key" text,
	 		    "private_key" text,
	 		    "private_key_passphrase" text,
	 		    "created_at" timestamptz,
	 		    "updated_at" timestamptz,
	 		    PRIMARY KEY ("namespace_id"),
				CONSTRAINT "fk_namespaces_mirror_configs"
				FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("oid") ON DELETE CASCADE ON UPDATE CASCADE
	     );
	 CREATE TABLE IF NOT EXISTS "mirror_processes" 
	 		(
	 		    "id" uuid,
	 		    "namespace_id" uuid,
	 		    "status" text,
				"typ" 	 text,
	 		    "ended_at" timestamptz,
	 		    "created_at" timestamptz,
	 		    "updated_at" timestamptz,
	 		    PRIMARY KEY ("id"),
	 		    CONSTRAINT "fk_namespaces_mirror_processes"
				FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("oid") ON DELETE CASCADE ON UPDATE CASCADE
	 		);
	 CREATE TABLE IF NOT EXISTS "log_msgs" 
			(
				"oid" uuid,
				"t" timestamptz,
				"msg" text,
				"level" text,
				"root_instance_id" uuid,
				"log_instance_call_path" text,
				"tags" jsonb,
				"workflow_id" uuid,
				"mirror_activity_id" uuid,
				"instance_logs" uuid,
				"namespace_logs" uuid,
				PRIMARY KEY ("oid")
			);
`)
	//				CONSTRAINT "log_msgs_workflow_id"
	// FOREIGN KEY ("workflow_id") REFERENCES "workflows"("oid"),
	// CONSTRAINT "log_msgs_namespace_logs"
	// FOREIGN KEY ("namespace_logs") REFERENCES "namespaces"("oid"),
	// CONSTRAINT "log_msgs_instances_logs"
	// FOREIGN KEY ("instance_logs") REFERENCES "instances"("oid")

	if tx.Error != nil {
		return nil, cleanup, tx.Error
	}
	return tx, cleanup, nil
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
