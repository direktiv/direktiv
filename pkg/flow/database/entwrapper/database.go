package entwrapper

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entnote "github.com/direktiv/direktiv/pkg/flow/ent/annotation"
	entns "github.com/direktiv/direktiv/pkg/flow/ent/namespace"
	database2 "github.com/direktiv/direktiv/pkg/refactor/database"
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
	LogMsg            *ent.LogMsgClient
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
			LogMsg:            db.Client.LogMsg,
		}
	}

	x := a.(*ent.Tx)

	return &EntClients{
		Namespace:         x.Namespace,
		Annotation:        x.Annotation,
		Events:            x.Events,
		CloudEvents:       x.CloudEvents,
		CloudEventFilters: x.CloudEventFilters,
		LogMsg:            x.LogMsg,
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
	_, err = db.DB().Exec(database2.Schema)
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

	annotation, err := clients.Annotation.Query().Where(entnote.InstanceID(instID), entnote.Name(key)).Only(ctx)
	if err != nil {
		db.Sugar.Debugf("%s failed to resolve instance annotation: %v", parent(), err)
		return nil, err
	}

	return db.entAnnotation(annotation), nil
}
