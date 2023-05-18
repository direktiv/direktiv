package entwrapper

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entnote "github.com/direktiv/direktiv/pkg/flow/ent/annotation"
	entinst "github.com/direktiv/direktiv/pkg/flow/ent/instance"
	entrt "github.com/direktiv/direktiv/pkg/flow/ent/instanceruntime"
	entns "github.com/direktiv/direktiv/pkg/flow/ent/namespace"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ctxKey string

const (
	ctxKeyTx = ctxKey("entwrapperCtxTxKey")
)

// TODO: un-export EntClients.
type EntClients struct {
	Namespace         *ent.NamespaceClient
	Annotation        *ent.AnnotationClient
	Events            *ent.EventsClient
	CloudEvents       *ent.CloudEventsClient
	CloudEventFilters *ent.CloudEventFiltersClient
	Instance          *ent.InstanceClient
	LogMsg            *ent.LogMsgClient
	InstanceRuntime   *ent.InstanceRuntimeClient
}

// TODO: delete.
func (db *Database) Clients(ctx context.Context) *EntClients {
	return db.clients(ctx)
}

func (db *Database) clients(ctx context.Context) *EntClients {
	a := ctx.Value(ctxKeyTx)

	if a == nil {
		return &EntClients{
			Namespace:         db.Client.Namespace,
			Annotation:        db.Client.Annotation,
			Events:            db.Client.Events,
			CloudEvents:       db.Client.CloudEvents,
			CloudEventFilters: db.Client.CloudEventFilters,
			Instance:          db.Client.Instance,
			LogMsg:            db.Client.LogMsg,
			InstanceRuntime:   db.Client.InstanceRuntime,
		}
	}

	x := a.(*ent.Tx)

	return &EntClients{
		Namespace:         x.Namespace,
		Annotation:        x.Annotation,
		Events:            x.Events,
		CloudEvents:       x.CloudEvents,
		CloudEventFilters: x.CloudEventFilters,
		Instance:          x.Instance,
		LogMsg:            x.LogMsg,
		InstanceRuntime:   x.InstanceRuntime,
	}
}

type Database struct {
	Sugar  *zap.SugaredLogger
	Client *ent.Client
}

func New(ctx context.Context, sugar *zap.SugaredLogger, addr string) (*Database, error) {
	db, err := ent.Open("postgres", addr)
	if err != nil {
		return nil, err
	}

	udb := db.DB()
	udb.SetMaxIdleConns(32)
	udb.SetMaxOpenConns(16)

	// Run the auto migration tool.
	if err = db.Schema.Create(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	//
	// initialize generation table if not exists
	qstr := `CREATE TABLE IF NOT EXISTS db_generation (
		generation VARCHAR
	)`

	_, err = db.DB().Exec(qstr)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	// create the new filesystem tables.
	_, err = db.DB().Exec(`
	 CREATE TABLE IF NOT EXISTS "filesystem_roots"
			(
				"id" uuid,
				"created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
				"updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY ("id"),
				CONSTRAINT "fk_namespaces_filesystem_roots"
				FOREIGN KEY ("id") REFERENCES "namespaces"("oid") ON DELETE CASCADE ON UPDATE CASCADE
				);
	 CREATE TABLE IF NOT EXISTS "filesystem_files"
			(
				"id" uuid,
				"path" text NOT NULL,
				"depth" integer NOT NULL,
				"typ" text NOT NULL,
				"root_id" uuid NOT NULL,
				"created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
				"updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY ("id"),
				CONSTRAINT "fk_filesystem_roots_filesystem_files"
				FOREIGN KEY ("root_id") REFERENCES "filesystem_roots"("id") ON DELETE CASCADE ON UPDATE CASCADE
				);
	 CREATE TABLE IF NOT EXISTS "filesystem_revisions"
			(
				"id" uuid,
				"tags" text,
				"is_current" boolean NOT NULL,
				"data" bytea NOT NULL,
				"checksum" text NOT NULL,
				"file_id" uuid NOT NULL,
				"created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
				"updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY ("id"),
				CONSTRAINT "fk_filesystem_files_filesystem_revisions"
				FOREIGN KEY ("file_id") REFERENCES "filesystem_files"("id") ON DELETE CASCADE ON UPDATE CASCADE
				);
	 CREATE TABLE IF NOT EXISTS "file_annotations"
			(
				"file_id" uuid,
				"data" text,
				"created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
				"updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY ("file_id"),
				CONSTRAINT "fk_filesystem_files_file_annotations"
				FOREIGN KEY ("file_id") REFERENCES "filesystem_files"("id") ON DELETE CASCADE ON UPDATE CASCADE
				);
	 CREATE TABLE IF NOT EXISTS "mirror_configs" 
	 		(
	 		    "namespace_id" uuid,
	 		    "url" text NOT NULL,
	 		    "git_ref" text NOT NULL,
	 		    "git_commit_hash" text,
	 		    "public_key" text,
	 		    "private_key" text,
	 		    "private_key_passphrase" text,
	 		    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
	 		    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
	 		    PRIMARY KEY ("namespace_id"),
				CONSTRAINT "fk_namespaces_mirror_configs"
				FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("oid") ON DELETE CASCADE ON UPDATE CASCADE
	     );
	 CREATE TABLE IF NOT EXISTS "mirror_processes" 
	 		(
	 		    "id" uuid,
	 		    "namespace_id" uuid NOT NULL,
	 		    "status" text NOT NULL,
				"typ" 	 text NOT NULL,
	 		    "ended_at" timestamptz,
	 		    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
	 		    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
	 		    PRIMARY KEY ("id"),
	 		    CONSTRAINT "fk_namespaces_mirror_processes"
				FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("oid") ON DELETE CASCADE ON UPDATE CASCADE
	 		);
	 CREATE TABLE IF NOT EXISTS "secrets" 
	 		(
	 		    "id" uuid,
	 		    "namespace_id" uuid NOT NULL,
	 		    "name" text NOT NULL,
				"data" 	 text NOT NULL,
	 		    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
	 		    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
	 		    PRIMARY KEY ("id"),
	 		    CONSTRAINT "fk_namespaces_secrets"
				FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("oid") ON DELETE CASCADE ON UPDATE CASCADE
	 		);
`)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize filesystem tables: %w\n", err)
	}

	tx, err := db.DB().Begin()
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	row := tx.QueryRow(`SELECT generation FROM db_generation`)
	var gen string
	err = row.Scan(&gen)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, err = tx.Exec(fmt.Sprintf(`INSERT INTO db_generation(generation) VALUES('%s')`, "0.7.3")) // this value needs to be manually updated each time there's an important database change
			if err != nil {
				_ = db.Close()
				return nil, err
			}
			err = tx.Commit()
			if err != nil {
				_ = db.Close()
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &Database{
		Sugar:  sugar,
		Client: db,
	}, nil
}

func (db *Database) Close() error {
	return db.Client.Close()
}

func (db *Database) AddTxToCtx(ctx context.Context, tx database.Transaction) context.Context {
	return context.WithValue(ctx, ctxKeyTx, tx)
}

func (db *Database) Tx(ctx context.Context) (context.Context, database.Transaction, error) {
	tx, err := db.Client.Tx(ctx)
	if err != nil {
		return ctx, nil, err
	}

	ctx = db.AddTxToCtx(ctx, tx)

	return ctx, tx, nil
}

func (db *Database) DB() *sql.DB {
	return db.Client.DB()
}

func (db *Database) Namespace(ctx context.Context, id uuid.UUID) (*database.Namespace, error) {
	clients := db.clients(ctx)

	ns, err := clients.Namespace.Query().Where(entns.ID(id)).Only(ctx)
	if err != nil {
		db.Sugar.Debugf("%s failed to resolve namespace: %v", parent(), err)
		return nil, err
	}

	return db.entNamespace(ns), nil
}

func (db *Database) NamespaceByName(ctx context.Context, name string) (*database.Namespace, error) {
	clients := db.clients(ctx)

	ns, err := clients.Namespace.Query().Where(entns.Name(name)).Only(ctx)
	if err != nil {
		db.Sugar.Debugf("%s failed to resolve namespace: %v", parent(), err)
		return nil, err
	}

	return db.entNamespace(ns), nil
}

func (db *Database) Instance(ctx context.Context, id uuid.UUID) (*database.Instance, error) {
	clients := db.clients(ctx)

	inst, err := clients.Instance.Query().Where(entinst.ID(id)).WithNamespace(func(q *ent.NamespaceQuery) {
		q.Select(entns.FieldID)
	}).WithRuntime(func(q *ent.InstanceRuntimeQuery) {
		q.Select(entrt.FieldID)
	}).Only(ctx)
	if err != nil {
		db.Sugar.Debugf("%s failed to resolve instance: %v", parent(), err)
		return nil, err
	}

	return entInstance(inst), nil
}

func (db *Database) InstanceRuntime(ctx context.Context, id uuid.UUID) (*database.InstanceRuntime, error) {
	clients := db.clients(ctx)

	rt, err := clients.InstanceRuntime.Query().Where(entrt.ID(id)).WithCaller(func(q *ent.InstanceQuery) {
		q.Select(entinst.FieldID)
	}).Only(ctx)
	if err != nil {
		db.Sugar.Debugf("%s failed to resolve instance runtime data: %v", parent(), err)
		return nil, err
	}

	return entInstanceRuntime(rt), nil
}

func (db *Database) NamespaceAnnotation(ctx context.Context, nsID uuid.UUID, key string) (*database.Annotation, error) {
	clients := db.clients(ctx)

	annotation, err := clients.Annotation.Query().Where(entnote.HasNamespaceWith(entns.ID(nsID)), entnote.Name(key)).Only(ctx)
	if err != nil {
		db.Sugar.Debugf("%s failed to resolve namespace annotation: %v", parent(), err)
		return nil, err
	}

	return db.entAnnotation(annotation), nil
}

func (db *Database) InstanceAnnotation(ctx context.Context, instID uuid.UUID, key string) (*database.Annotation, error) {
	clients := db.clients(ctx)

	annotation, err := clients.Annotation.Query().Where(entnote.HasInstanceWith(entinst.ID(instID)), entnote.Name(key)).Only(ctx)
	if err != nil {
		db.Sugar.Debugf("%s failed to resolve instance annotation: %v", parent(), err)
		return nil, err
	}

	return db.entAnnotation(annotation), nil
}

func (db *Database) ThreadVariables(ctx context.Context, instID uuid.UUID) ([]*database.VarRef, error) {
	return nil, nil
}

func (db *Database) NamespaceVariableRef(ctx context.Context, nsID uuid.UUID, key string) (*database.VarRef, error) {
	return nil, nil
}

func (db *Database) WorkflowVariableRef(ctx context.Context, wfID uuid.UUID, key string) (*database.VarRef, error) {
	return nil, nil
}

func (db *Database) InstanceVariableRef(ctx context.Context, instID uuid.UUID, key string) (*database.VarRef, error) {
	return nil, nil
}

func (db *Database) ThreadVariableRef(ctx context.Context, instID uuid.UUID, key string) (*database.VarRef, error) {
	return nil, nil
}

func (db *Database) VariableData(ctx context.Context, id uuid.UUID, load bool) (*database.VarData, error) {
	return nil, nil
}
