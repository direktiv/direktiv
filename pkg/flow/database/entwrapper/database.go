package entwrapper

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	"github.com/google/uuid"
	"go.uber.org/zap"

	entnote "github.com/direktiv/direktiv/pkg/flow/ent/annotation"
	entinst "github.com/direktiv/direktiv/pkg/flow/ent/instance"
	entrt "github.com/direktiv/direktiv/pkg/flow/ent/instanceruntime"
	entns "github.com/direktiv/direktiv/pkg/flow/ent/namespace"
	entvardata "github.com/direktiv/direktiv/pkg/flow/ent/vardata"
	entvar "github.com/direktiv/direktiv/pkg/flow/ent/varref"
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
	VarRef            *ent.VarRefClient
	VarData           *ent.VarDataClient
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
			Namespace:         db.client.Namespace,
			Annotation:        db.client.Annotation,
			Events:            db.client.Events,
			CloudEvents:       db.client.CloudEvents,
			CloudEventFilters: db.client.CloudEventFilters,
			VarRef:            db.client.VarRef,
			VarData:           db.client.VarData,
			Instance:          db.client.Instance,
			LogMsg:            db.client.LogMsg,
			InstanceRuntime:   db.client.InstanceRuntime,
		}
	}

	x := a.(*ent.Tx)

	return &EntClients{
		Namespace:         x.Namespace,
		Annotation:        x.Annotation,
		Events:            x.Events,
		CloudEvents:       x.CloudEvents,
		CloudEventFilters: x.CloudEventFilters,
		VarRef:            x.VarRef,
		VarData:           x.VarData,
		Instance:          x.Instance,
		LogMsg:            x.LogMsg,
		InstanceRuntime:   x.InstanceRuntime,
	}
}

type Database struct {
	sugar  *zap.SugaredLogger
	client *ent.Client
}

func New(ctx context.Context, sugar *zap.SugaredLogger, addr string) (*Database, error) {
	db, err := ent.Open("postgres", addr)
	if err != nil {
		return nil, err
	}

	udb := db.DB()
	udb.SetMaxIdleConns(64)
	udb.SetMaxOpenConns(32)

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
		sugar:  sugar,
		client: db,
	}, nil
}

func (db *Database) Close() error {
	return db.client.Close()
}

func (db *Database) AddTxToCtx(ctx context.Context, tx database.Transaction) context.Context {
	return context.WithValue(ctx, ctxKeyTx, tx)
}

func (db *Database) Tx(ctx context.Context) (context.Context, database.Transaction, error) {
	tx, err := db.client.Tx(ctx)
	if err != nil {
		return ctx, nil, err
	}

	ctx = db.AddTxToCtx(ctx, tx)

	return ctx, tx, nil
}

func (db *Database) DB() *sql.DB {
	return db.client.DB()
}

func (db *Database) Namespace(ctx context.Context, id uuid.UUID) (*database.Namespace, error) {
	// TODO: yassir, need refactor.
	return nil, nil
	//clients := db.clients(ctx)
	//
	//ns, err := clients.Namespace.Query().Where(entns.ID(id)).WithInodes(func(q *ent.InodeQuery) {
	//	q.Where(entino.NameIsNil()).Select(entino.FieldID)
	//}).Only(ctx)
	//if err != nil {
	//	db.sugar.Debugf("%s failed to resolve namespace: %v", parent(), err)
	//	return nil, err
	//}
	//
	//return db.entNamespace(ns), nil
}

func (db *Database) NamespaceByName(ctx context.Context, name string) (*database.Namespace, error) {
	// TODO: yassir, need refactor.
	return nil, nil
	//clients := db.clients(ctx)
	//
	//ns, err := clients.Namespace.Query().Where(entns.Name(name)).WithInodes(func(q *ent.InodeQuery) {
	//	q.Where(entino.NameIsNil()).Select(entino.FieldID)
	//}).Only(ctx)
	//if err != nil {
	//	db.sugar.Debugf("%s failed to resolve namespace: %v", parent(), err)
	//	return nil, err
	//}
	//
	//return db.entNamespace(ns), nil
}

func (db *Database) Instance(ctx context.Context, id uuid.UUID) (*database.Instance, error) {
	// TODO: yassir, need refactor.
	return nil, nil
	//clients := db.clients(ctx)
	//
	//inst, err := clients.Instance.Query().Where(entinst.ID(id)).WithNamespace(func(q *ent.NamespaceQuery) {
	//	q.Select(entns.FieldID)
	//}).WithWorkflow(func(q *ent.WorkflowQuery) {
	//	q.Select(entwf.FieldID)
	//}).WithRevision(func(q *ent.RevisionQuery) {
	//	q.Select(entrev.FieldID)
	//}).WithRuntime(func(q *ent.InstanceRuntimeQuery) {
	//	q.Select(entrt.FieldID)
	//}).Only(ctx)
	//if err != nil {
	//	db.sugar.Debugf("%s failed to resolve instance: %v", parent(), err)
	//	return nil, err
	//}
	//
	//return entInstance(inst), nil
}

func (db *Database) InstanceRuntime(ctx context.Context, id uuid.UUID) (*database.InstanceRuntime, error) {
	clients := db.clients(ctx)

	rt, err := clients.InstanceRuntime.Query().Where(entrt.ID(id)).WithCaller(func(q *ent.InstanceQuery) {
		q.Select(entinst.FieldID)
	}).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve instance runtime data: %v", parent(), err)
		return nil, err
	}

	return entInstanceRuntime(rt), nil
}

func (db *Database) NamespaceAnnotation(ctx context.Context, nsID uuid.UUID, key string) (*database.Annotation, error) {
	clients := db.clients(ctx)

	annotation, err := clients.Annotation.Query().Where(entnote.HasNamespaceWith(entns.ID(nsID)), entnote.Name(key)).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve namespace annotation: %v", parent(), err)
		return nil, err
	}

	return db.entAnnotation(annotation), nil
}

func (db *Database) WorkflowAnnotation(ctx context.Context, wfID uuid.UUID, key string) (*database.Annotation, error) {
	// TODO: yassir, need refactor.
	return nil, nil
	//clients := db.clients(ctx)
	//
	//annotation, err := clients.Annotation.Query().Where(entnote.HasWorkflowWith(entwf.ID(wfID)), entnote.Name(key)).Only(ctx)
	//if err != nil {
	//	db.sugar.Debugf("%s failed to resolve workflow annotation: %v", parent(), err)
	//	return nil, err
	//}
	//
	//return db.entAnnotation(annotation), nil
}

func (db *Database) ThreadVariables(ctx context.Context, instID uuid.UUID) ([]*database.VarRef, error) {
	clients := db.clients(ctx)

	varrefs, err := clients.VarRef.Query().Where(entvar.HasInstanceWith(entinst.ID(instID)), entvar.BehaviourEQ("thread")).WithVardata(func(q *ent.VarDataQuery) {
		q.Select(entvardata.FieldID)
	}).All(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve instance thread variables: %v", parent(), err)
		return nil, err
	}

	x := make([]*database.VarRef, 0)

	for _, y := range varrefs {
		x = append(x, db.entVarRef(y))
	}

	return x, nil
}

func (db *Database) NamespaceVariableRef(ctx context.Context, nsID uuid.UUID, key string) (*database.VarRef, error) {
	clients := db.clients(ctx)

	varref, err := clients.VarRef.Query().Where(entvar.HasNamespaceWith(entns.ID(nsID)), entvar.NameEQ(key)).WithVardata(func(q *ent.VarDataQuery) {
		q.Select(entvardata.FieldID)
	}).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve namespace variable: %v", parent(), err)
		return nil, err
	}

	return db.entVarRef(varref), nil
}

func (db *Database) WorkflowVariableRef(ctx context.Context, wfID uuid.UUID, key string) (*database.VarRef, error) {
	// TODO: yassir, need refactor.
	return nil, nil
	//clients := db.clients(ctx)
	//
	//varref, err := clients.VarRef.Query().Where(entvar.HasWorkflowWith(entwf.ID(wfID)), entvar.NameEQ(key)).WithVardata(func(q *ent.VarDataQuery) {
	//	q.Select(entvardata.FieldID)
	//}).Only(ctx)
	//if err != nil {
	//	db.sugar.Debugf("%s failed to resolve workflow variable: %v", parent(), err)
	//	return nil, err
	//}
	//
	//return db.entVarRef(varref), nil
}

func (db *Database) InstanceVariableRef(ctx context.Context, instID uuid.UUID, key string) (*database.VarRef, error) {
	clients := db.clients(ctx)

	varref, err := clients.VarRef.Query().Where(entvar.HasInstanceWith(entinst.ID(instID)), entvar.BehaviourNEQ("thread"), entvar.NameEQ(key)).WithVardata(func(q *ent.VarDataQuery) {
		q.Select(entvardata.FieldID)
	}).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve instance variable: %v", parent(), err)
		return nil, err
	}

	return db.entVarRef(varref), nil
}

func (db *Database) ThreadVariableRef(ctx context.Context, instID uuid.UUID, key string) (*database.VarRef, error) {
	clients := db.clients(ctx)

	varref, err := clients.VarRef.Query().Where(entvar.HasInstanceWith(entinst.ID(instID)), entvar.BehaviourEQ("thread"), entvar.NameEQ(key)).WithVardata(func(q *ent.VarDataQuery) {
		q.Select(entvardata.FieldID)
	}).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve thread variable: %v", parent(), err)
		return nil, err
	}

	return db.entVarRef(varref), nil
}

func (db *Database) VariableData(ctx context.Context, id uuid.UUID, load bool) (*database.VarData, error) {
	var err error
	var vardata *ent.VarData

	clients := db.clients(ctx)

	if load {
		vardata, err = clients.VarData.Get(ctx, id)
		if err != nil {
			db.sugar.Debugf("%s failed to resolve variable data: %v", parent(), err)
			return nil, err
		}
	} else {
		vardata, err = clients.VarData.Query().Where(entvardata.ID(id)).Select(entvardata.FieldID, entvardata.FieldCreatedAt, entvardata.FieldUpdatedAt, entvardata.FieldSize, entvardata.FieldHash, entvardata.FieldMimeType).Only(ctx)
		if err != nil {
			db.sugar.Debugf("%s failed to resolve variable data: %v", parent(), err)
			return nil, err
		}
	}

	x := db.entVarData(vardata)

	k, err := clients.VarRef.Query().Where(entvar.HasVardataWith(entvardata.ID(vardata.ID))).Count(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to count variable references: %v", parent(), err)
		return nil, err
	}

	x.RefCount = k

	return x, err
}
