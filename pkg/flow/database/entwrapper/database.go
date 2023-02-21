package entwrapper

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	"github.com/google/uuid"
	"go.uber.org/zap"

	entnote "github.com/direktiv/direktiv/pkg/flow/ent/annotation"
	entino "github.com/direktiv/direktiv/pkg/flow/ent/inode"
	entinst "github.com/direktiv/direktiv/pkg/flow/ent/instance"
	entrt "github.com/direktiv/direktiv/pkg/flow/ent/instanceruntime"
	entmir "github.com/direktiv/direktiv/pkg/flow/ent/mirror"
	entmiract "github.com/direktiv/direktiv/pkg/flow/ent/mirroractivity"
	entns "github.com/direktiv/direktiv/pkg/flow/ent/namespace"
	entrev "github.com/direktiv/direktiv/pkg/flow/ent/revision"
	entroute "github.com/direktiv/direktiv/pkg/flow/ent/route"
	entvardata "github.com/direktiv/direktiv/pkg/flow/ent/vardata"
	entvar "github.com/direktiv/direktiv/pkg/flow/ent/varref"
	entwf "github.com/direktiv/direktiv/pkg/flow/ent/workflow"
)

// TODO: un-export EntClients
type EntClients struct {
	Namespace         *ent.NamespaceClient
	Inode             *ent.InodeClient
	Annotation        *ent.AnnotationClient
	Events            *ent.EventsClient
	CloudEvents       *ent.CloudEventsClient
	CloudEventFilters *ent.CloudEventFiltersClient
	Route             *ent.RouteClient
	Ref               *ent.RefClient
	Revision          *ent.RevisionClient
	VarRef            *ent.VarRefClient
	VarData           *ent.VarDataClient
	Instance          *ent.InstanceClient
	Workflow          *ent.WorkflowClient
	LogMsg            *ent.LogMsgClient
	Mirror            *ent.MirrorClient
	MirrorActivity    *ent.MirrorActivityClient
	InstanceRuntime   *ent.InstanceRuntimeClient
}

// TODO: delete
func (db *Database) Clients(tx database.Transaction) *EntClients {
	return db.clients(tx)
}

func (db *Database) clients(tx database.Transaction) *EntClients {

	if tx == nil {
		return &EntClients{
			Namespace:         db.client.Namespace,
			Inode:             db.client.Inode,
			Annotation:        db.client.Annotation,
			Events:            db.client.Events,
			CloudEvents:       db.client.CloudEvents,
			CloudEventFilters: db.client.CloudEventFilters,
			Route:             db.client.Route,
			Ref:               db.client.Ref,
			Revision:          db.client.Revision,
			VarRef:            db.client.VarRef,
			VarData:           db.client.VarData,
			Instance:          db.client.Instance,
			Workflow:          db.client.Workflow,
			LogMsg:            db.client.LogMsg,
			Mirror:            db.client.Mirror,
			MirrorActivity:    db.client.MirrorActivity,
			InstanceRuntime:   db.client.InstanceRuntime,
		}
	}

	x := tx.(*ent.Tx)

	return &EntClients{
		Namespace:         x.Namespace,
		Inode:             x.Inode,
		Annotation:        x.Annotation,
		Events:            x.Events,
		CloudEvents:       x.CloudEvents,
		CloudEventFilters: x.CloudEventFilters,
		Route:             x.Route,
		Ref:               x.Ref,
		Revision:          x.Revision,
		VarRef:            x.VarRef,
		VarData:           x.VarData,
		Instance:          x.Instance,
		Workflow:          x.Workflow,
		LogMsg:            x.LogMsg,
		Mirror:            x.Mirror,
		MirrorActivity:    x.MirrorActivity,
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
			_, err = tx.Exec(fmt.Sprintf(`INSERT INTO db_generation(generation) VALUES('%s')`, "0.7.1")) // this value needs to be manually updated each time there's an important database change
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

func (db *Database) Tx(ctx context.Context) (database.Transaction, error) {
	return db.client.Tx(ctx)
}

func (db *Database) DB() *sql.DB {
	return db.client.DB()
}

func (db *Database) Namespace(ctx context.Context, tx database.Transaction, id uuid.UUID) (*database.Namespace, error) {

	clients := db.clients(tx)

	ns, err := clients.Namespace.Query().Where(entns.ID(id)).WithInodes(func(q *ent.InodeQuery) {
		q.Where(entino.NameIsNil()).Select(entino.FieldID)
	}).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve namespace: %v", parent(), err)
		return nil, err
	}

	return db.entNamespace(ns), nil

}

func (db *Database) NamespaceByName(ctx context.Context, tx database.Transaction, name string) (*database.Namespace, error) {

	clients := db.clients(tx)

	ns, err := clients.Namespace.Query().Where(entns.Name(name)).WithInodes(func(q *ent.InodeQuery) {
		q.Where(entino.NameIsNil()).Select(entino.FieldID)
	}).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve namespace: %v", parent(), err)
		return nil, err
	}

	return db.entNamespace(ns), nil

}

func (db *Database) Inode(ctx context.Context, tx database.Transaction, id uuid.UUID) (*database.Inode, error) {

	clients := db.clients(tx)

	ino, err := clients.Inode.Query().Where(entino.ID(id)).WithChildren(func(q *ent.InodeQuery) {
		q.Order(ent.Asc(entino.FieldName)).Select(entino.FieldID, entino.FieldName, entino.FieldType, entino.FieldExtendedType)
	}).WithNamespace(func(q *ent.NamespaceQuery) {
		q.Select(entns.FieldID)
	}).WithParent(func(q *ent.InodeQuery) {
		q.Select(entino.FieldID)
	}).WithWorkflow(func(q *ent.WorkflowQuery) {
		q.Select(entwf.FieldID)
	}).WithMirror(func(q *ent.MirrorQuery) {
		q.Select(entmir.FieldID)
	}).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve inode: %v", parent(), err)
		return nil, err
	}

	return entInode(ino), nil

}

func (db *Database) CreateInode(ctx context.Context, tx database.Transaction, args *database.CreateInodeArgs) (*database.Inode, error) {

	clients := db.clients(tx)

	ino, err := clients.Inode.Create().
		SetName(args.Name).
		SetType(args.Type).
		SetAttributes(args.Attributes).
		SetExtendedType(args.ExtendedType).
		SetReadOnly(args.ReadOnly).
		SetNamespaceID(args.Namespace).
		SetParentID(args.Parent).
		Save(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, os.ErrExist
		}
		db.sugar.Debugf("%s failed to create inode: %v", parent(), err)
		return nil, err
	}

	ino.Edges.Namespace = &ent.Namespace{
		ID: args.Namespace,
	}
	ino.Edges.Parent = &ent.Inode{
		ID: args.Parent,
	}

	return entInode(ino), nil

}

func (db *Database) UpdateInode(ctx context.Context, tx database.Transaction, args *database.UpdateInodeArgs) (*database.Inode, error) {

	clients := db.clients(tx)

	query := clients.Inode.UpdateOneID(args.Inode.ID).SetUpdatedAt(time.Now())

	if args.Name != nil {
		query = query.SetName(*args.Name)
	}

	if args.Attributes != nil {
		query = query.SetAttributes(*args.Attributes)
	}

	if args.ReadOnly != nil {
		query = query.SetReadOnly(*args.ReadOnly)
	}

	if args.Parent != nil {
		query = query.SetParentID(*args.Parent)
	}

	ino, err := query.Save(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to update inode: %v", parent(), err)
		return nil, err
	}

	ino.Edges.Namespace = &ent.Namespace{
		ID: args.Inode.Namespace,
	}
	ino.Edges.Parent = &ent.Inode{
		ID: args.Inode.Parent,
	}

	x := entInode(ino)
	x.Children = args.Inode.Children
	x.Workflow = args.Inode.Workflow

	return x, nil

}

func (db *Database) Workflow(ctx context.Context, tx database.Transaction, id uuid.UUID) (*database.Workflow, error) {

	clients := db.clients(tx)

	wf, err := clients.Workflow.Query().Where(entwf.ID(id)).WithInode(func(q *ent.InodeQuery) {
		q.Select(entino.FieldID)
	}).WithNamespace(func(q *ent.NamespaceQuery) {
		q.Select(entns.FieldID)
	}).WithRefs(func(q *ent.RefQuery) {
		q.WithRevision(func(q *ent.RevisionQuery) {
			q.Select(entrev.FieldID)
		})
	}).WithRevisions(func(q *ent.RevisionQuery) {
		q.Select(entrev.FieldID, entrev.FieldHash)
	}).WithRoutes(func(q *ent.RouteQuery) {
		q.WithRef(func(q *ent.RefQuery) {
			q.WithRevision(func(q *ent.RevisionQuery) {
				q.Select(entrev.FieldID)
			})
		})
	}).Order(ent.Desc(entroute.FieldID)).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve workflow: %v", parent(), err)
		return nil, err
	}

	return entWorkflow(wf), nil

}

func (db *Database) CreateWorkflow(ctx context.Context, tx database.Transaction, args *database.CreateWorkflowArgs) (*database.Workflow, error) {

	clients := db.clients(tx)

	wf, err := clients.Workflow.Create().
		SetInodeID(args.Inode.ID).
		SetNamespaceID(args.Inode.Namespace).
		SetReadOnly(args.Inode.ReadOnly).
		Save(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to create workflow: %v", parent(), err)
		return nil, err
	}

	wf.Edges.Namespace = &ent.Namespace{
		ID: args.Inode.Namespace,
	}

	wf.Edges.Inode = &ent.Inode{
		ID: args.Inode.ID,
	}

	return entWorkflow(wf), nil

}

func (db *Database) UpdateWorkflow(ctx context.Context, tx database.Transaction, args *database.UpdateWorkflowArgs) (*database.Workflow, error) {

	clients := db.clients(tx)

	query := clients.Workflow.UpdateOneID(args.ID).SetUpdatedAt(time.Now())

	if args.ReadOnly != nil {
		query = query.SetReadOnly(*args.ReadOnly)
	}

	wf, err := query.Save(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to update workflow: %v", parent(), err)
		return nil, err
	}

	return entWorkflow(wf), nil

}

func (db *Database) CreateRef(ctx context.Context, tx database.Transaction, args *database.CreateRefArgs) (*database.Ref, error) {

	clients := db.clients(tx)

	ref, err := clients.Ref.Create().
		SetImmutable(args.Immutable).
		SetName(args.Name).
		SetWorkflowID(args.Workflow).
		SetRevisionID(args.Revision).
		Save(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to create ref: %v", parent(), err)
		return nil, err
	}

	ref.Edges.Revision = &ent.Revision{
		ID: args.Revision,
	}

	ref.Edges.Workflow = &ent.Workflow{
		ID: args.Workflow,
	}

	return entRef(ref), nil

}

func (db *Database) Revision(ctx context.Context, tx database.Transaction, id uuid.UUID) (*database.Revision, error) {

	clients := db.clients(tx)

	rev, err := clients.Revision.Query().Where(entrev.ID(id)).WithWorkflow(func(q *ent.WorkflowQuery) {
		q.Select(entwf.FieldID)
	}).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve revision '%s': %v", parent(), id, err)
		return nil, err
	}

	return entRevision(rev), nil

}

func (db *Database) CreateRevision(ctx context.Context, tx database.Transaction, args *database.CreateRevisionArgs) (*database.Revision, error) {

	clients := db.clients(tx)

	rev, err := clients.Revision.Create().
		SetHash(args.Hash).
		SetSource(args.Source).
		SetWorkflowID(args.Workflow).
		SetMetadata(args.Metadata).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	rev.Edges.Workflow = &ent.Workflow{
		ID: args.Workflow,
	}

	return entRevision(rev), nil

}

func (db *Database) Instance(ctx context.Context, tx database.Transaction, id uuid.UUID) (*database.Instance, error) {

	clients := db.clients(tx)

	inst, err := clients.Instance.Query().Where(entinst.ID(id)).WithNamespace(func(q *ent.NamespaceQuery) {
		q.Select(entns.FieldID)
	}).WithWorkflow(func(q *ent.WorkflowQuery) {
		q.Select(entwf.FieldID)
	}).WithRevision(func(q *ent.RevisionQuery) {
		q.Select(entrev.FieldID)
	}).WithRuntime(func(q *ent.InstanceRuntimeQuery) {
		q.Select(entrt.FieldID)
	}).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve instance: %v", parent(), err)
		return nil, err
	}

	return entInstance(inst), nil

}

func (db *Database) InstanceRuntime(ctx context.Context, tx database.Transaction, id uuid.UUID) (*database.InstanceRuntime, error) {

	clients := db.clients(tx)

	rt, err := clients.InstanceRuntime.Query().Where(entrt.ID(id)).WithCaller(func(q *ent.InstanceQuery) {
		q.Select(entinst.FieldID)
	}).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve instance runtime data: %v", parent(), err)
		return nil, err
	}

	return entInstanceRuntime(rt), nil

}

func (db *Database) NamespaceAnnotation(ctx context.Context, tx database.Transaction, nsID uuid.UUID, key string) (*database.Annotation, error) {

	clients := db.clients(tx)

	annotation, err := clients.Annotation.Query().Where(entnote.HasNamespaceWith(entns.ID(nsID)), entnote.Name(key)).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve namespace annotation: %v", parent(), err)
		return nil, err
	}

	return db.entAnnotation(annotation), nil

}

func (db *Database) InodeAnnotation(ctx context.Context, tx database.Transaction, inodeID uuid.UUID, key string) (*database.Annotation, error) {

	clients := db.clients(tx)

	annotation, err := clients.Annotation.Query().Where(entnote.HasInodeWith(entino.ID(inodeID)), entnote.Name(key)).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve inode annotation: %v", parent(), err)
		return nil, err
	}

	return db.entAnnotation(annotation), nil

}

func (db *Database) WorkflowAnnotation(ctx context.Context, tx database.Transaction, wfID uuid.UUID, key string) (*database.Annotation, error) {

	clients := db.clients(tx)

	annotation, err := clients.Annotation.Query().Where(entnote.HasWorkflowWith(entwf.ID(wfID)), entnote.Name(key)).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve workflow annotation: %v", parent(), err)
		return nil, err
	}

	return db.entAnnotation(annotation), nil

}

func (db *Database) InstanceAnnotation(ctx context.Context, tx database.Transaction, instID uuid.UUID, key string) (*database.Annotation, error) {

	clients := db.clients(tx)

	annotation, err := clients.Annotation.Query().Where(entnote.HasInstanceWith(entinst.ID(instID)), entnote.Name(key)).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve instance annotation: %v", parent(), err)
		return nil, err
	}

	return db.entAnnotation(annotation), nil

}

func (db *Database) ThreadVariables(ctx context.Context, tx database.Transaction, instID uuid.UUID) ([]*database.VarRef, error) {

	clients := db.clients(tx)

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

func (db *Database) NamespaceVariableRef(ctx context.Context, tx database.Transaction, nsID uuid.UUID, key string) (*database.VarRef, error) {

	clients := db.clients(tx)

	varref, err := clients.VarRef.Query().Where(entvar.HasNamespaceWith(entns.ID(nsID)), entvar.NameEQ(key)).WithVardata(func(q *ent.VarDataQuery) {
		q.Select(entvardata.FieldID)
	}).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve namespace variable: %v", parent(), err)
		return nil, err
	}

	return db.entVarRef(varref), nil

}

func (db *Database) WorkflowVariableRef(ctx context.Context, tx database.Transaction, wfID uuid.UUID, key string) (*database.VarRef, error) {

	clients := db.clients(tx)

	varref, err := clients.VarRef.Query().Where(entvar.HasWorkflowWith(entwf.ID(wfID)), entvar.NameEQ(key)).WithVardata(func(q *ent.VarDataQuery) {
		q.Select(entvardata.FieldID)
	}).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve workflow variable: %v", parent(), err)
		return nil, err
	}

	return db.entVarRef(varref), nil

}

func (db *Database) InstanceVariableRef(ctx context.Context, tx database.Transaction, instID uuid.UUID, key string) (*database.VarRef, error) {

	clients := db.clients(tx)

	varref, err := clients.VarRef.Query().Where(entvar.HasInstanceWith(entinst.ID(instID)), entvar.BehaviourNEQ("thread"), entvar.NameEQ(key)).WithVardata(func(q *ent.VarDataQuery) {
		q.Select(entvardata.FieldID)
	}).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve instance variable: %v", parent(), err)
		return nil, err
	}

	return db.entVarRef(varref), nil

}

func (db *Database) ThreadVariableRef(ctx context.Context, tx database.Transaction, instID uuid.UUID, key string) (*database.VarRef, error) {

	clients := db.clients(tx)

	varref, err := clients.VarRef.Query().Where(entvar.HasInstanceWith(entinst.ID(instID)), entvar.BehaviourEQ("thread"), entvar.NameEQ(key)).WithVardata(func(q *ent.VarDataQuery) {
		q.Select(entvardata.FieldID)
	}).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve thread variable: %v", parent(), err)
		return nil, err
	}

	return db.entVarRef(varref), nil

}

func (db *Database) VariableData(ctx context.Context, tx database.Transaction, id uuid.UUID, load bool) (*database.VarData, error) {

	var err error
	var vardata *ent.VarData

	clients := db.clients(tx)

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

func (db *Database) Mirror(ctx context.Context, tx database.Transaction, id uuid.UUID) (*database.Mirror, error) {

	clients := db.clients(tx)

	mir, err := clients.Mirror.Query().Where(entmir.ID(id)).WithInode().Only(ctx)
	if err != nil {
		return nil, err
	}

	return entMirror(mir), nil

}

func (db *Database) Mirrors(ctx context.Context, tx database.Transaction) ([]uuid.UUID, error) {

	clients := db.clients(tx)

	var ids = make([]uuid.UUID, 0)

	rows, err := clients.Mirror.Query().Select(entmir.FieldID).All(ctx)
	if err != nil {
		return nil, err
	}

	for idx := range rows {
		ids = append(ids, rows[idx].ID)
	}

	return ids, nil

}

func (db *Database) MirrorActivity(ctx context.Context, tx database.Transaction, id uuid.UUID) (*database.MirrorActivity, error) {

	clients := db.clients(tx)

	act, err := clients.MirrorActivity.Query().Where(entmiract.ID(id)).WithNamespace().WithMirror().Only(ctx)
	if err != nil {
		return nil, err
	}

	return entMirrorActivity(act), nil

}

func (db *Database) CreateMirrorActivity(ctx context.Context, tx database.Transaction, args *database.CreateMirrorActivityArgs) (*database.MirrorActivity, error) {

	clients := db.clients(tx)

	act, err := clients.MirrorActivity.Create().
		SetType(args.Type).
		SetStatus(args.Status).
		SetEndAt(args.EndAt).
		SetMirrorID(args.Mirror).
		SetNamespaceID(args.Namespace).
		SetController(args.Controller).
		SetDeadline(args.Deadline).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	act.Edges.Namespace = &ent.Namespace{
		ID: args.Namespace,
	}

	act.Edges.Mirror = &ent.Mirror{
		ID: args.Mirror,
	}

	return entMirrorActivity(act), nil

}
