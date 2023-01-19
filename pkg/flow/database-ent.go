package flow

import (
	"context"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	"github.com/google/uuid"
	"go.uber.org/zap"

	entnote "github.com/direktiv/direktiv/pkg/flow/ent/annotation"
	entino "github.com/direktiv/direktiv/pkg/flow/ent/inode"
	entinst "github.com/direktiv/direktiv/pkg/flow/ent/instance"
	entrt "github.com/direktiv/direktiv/pkg/flow/ent/instanceruntime"
	entmir "github.com/direktiv/direktiv/pkg/flow/ent/mirror"
	entns "github.com/direktiv/direktiv/pkg/flow/ent/namespace"
	entref "github.com/direktiv/direktiv/pkg/flow/ent/ref"
	entrev "github.com/direktiv/direktiv/pkg/flow/ent/revision"
	entroute "github.com/direktiv/direktiv/pkg/flow/ent/route"
	entvardata "github.com/direktiv/direktiv/pkg/flow/ent/vardata"
	entvar "github.com/direktiv/direktiv/pkg/flow/ent/varref"
	entwf "github.com/direktiv/direktiv/pkg/flow/ent/workflow"
)

type entClients struct {
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

func (db *EntDatabase) clients(tx Transaction) *entClients {

	if tx == nil {
		return &entClients{
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

	return &entClients{
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

func (srv *server) entClients(tx Transaction) *entClients {

	if tx == nil {
		return &entClients{
			Namespace:         srv.db.Namespace,
			Inode:             srv.db.Inode,
			Annotation:        srv.db.Annotation,
			Events:            srv.db.Events,
			CloudEvents:       srv.db.CloudEvents,
			CloudEventFilters: srv.db.CloudEventFilters,
			Route:             srv.db.Route,
			Ref:               srv.db.Ref,
			Revision:          srv.db.Revision,
			VarRef:            srv.db.VarRef,
			VarData:           srv.db.VarData,
			Instance:          srv.db.Instance,
			Workflow:          srv.db.Workflow,
			LogMsg:            srv.db.LogMsg,
			Mirror:            srv.db.Mirror,
			MirrorActivity:    srv.db.MirrorActivity,
			InstanceRuntime:   srv.db.InstanceRuntime,
		}
	}

	x := tx.(*ent.Tx)

	return &entClients{
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

type EntDatabase struct {
	client *ent.Client
	sugar  *zap.SugaredLogger
}

func (db *EntDatabase) entNamespace(ns *ent.Namespace) *Namespace {

	if ns == nil {
		return nil
	}

	return &Namespace{
		ID:        ns.ID,
		CreatedAt: ns.CreatedAt,
		UpdatedAt: ns.UpdatedAt,
		Config:    ns.Config,
		Name:      ns.Name,
		Root:      ns.Edges.Inodes[0].ID,
	}

}

func entInode(ino *ent.Inode) *Inode {

	if ino == nil {
		return nil
	}

	var children []*Inode
	for _, x := range ino.Edges.Children {
		children = append(children, &Inode{
			ID:   x.ID,
			Name: x.Name,
		})
	}

	x := &Inode{
		ID:           ino.ID,
		CreatedAt:    ino.CreatedAt,
		UpdatedAt:    ino.UpdatedAt,
		Name:         ino.Name,
		Type:         ino.Type,
		Attributes:   ino.Attributes,
		ExtendedType: ino.ExtendedType,
		ReadOnly:     ino.ReadOnly,
		Children:     children,
		Parent:       ino.Edges.Parent.ID,
		Namespace:    ino.Edges.Namespace.ID,
	}

	if ino.Edges.Workflow != nil {
		x.Workflow = ino.Edges.Workflow.ID
	}

	return x

}

func entWorkflow(wf *ent.Workflow) *Workflow {

	if wf == nil {
		return nil
	}

	var refs []*Ref
	for _, x := range wf.Edges.Refs {
		refs = append(refs, entRef(x))
	}

	var revisions []*Revision
	for _, x := range wf.Edges.Revisions {
		revisions = append(revisions, &Revision{
			ID:   x.ID,
			Hash: x.Hash,
		})
	}

	var routes []*Route
	for _, x := range wf.Edges.Routes {
		routes = append(routes, &Route{
			ID:     x.ID,
			Weight: x.Weight,
			Ref:    entRef(x.Edges.Ref),
		})
	}

	return &Workflow{
		ID:          wf.ID,
		Live:        wf.Live,
		LogToEvents: wf.LogToEvents,
		ReadOnly:    wf.ReadOnly,
		UpdatedAt:   wf.UpdatedAt,
		Namespace:   wf.Edges.Namespace.ID,
		Inode:       wf.Edges.Inode.ID,
		Refs:        refs,
		Revisions:   revisions,
		Routes:      routes,
	}

}

func entRef(ref *ent.Ref) *Ref {

	if ref == nil {
		return nil
	}

	x := &Ref{
		ID:        ref.ID,
		Name:      ref.Name,
		Immutable: ref.Immutable,
		CreatedAt: ref.CreatedAt,
	}

	if ref.Edges.Revision != nil {
		x.Revision = ref.Edges.Revision.ID
	}

	return x

}

func entRevision(rev *ent.Revision) *Revision {

	if rev == nil {
		return nil
	}

	return &Revision{
		ID:        rev.ID,
		CreatedAt: rev.CreatedAt,
		Hash:      rev.Hash,
		Source:    rev.Source,
		Metadata:  rev.Metadata,
		Workflow:  rev.Edges.Workflow.ID,
	}

}

func entInstance(inst *ent.Instance) *Instance {

	if inst == nil {
		return nil
	}

	return &Instance{
		ID:           inst.ID,
		CreatedAt:    inst.CreatedAt,
		UpdatedAt:    inst.UpdatedAt,
		EndAt:        inst.EndAt,
		Status:       inst.Status,
		As:           inst.As,
		ErrorCode:    inst.ErrorCode,
		ErrorMessage: inst.ErrorMessage,
		Invoker:      inst.Invoker,
		Namespace:    inst.Edges.Namespace.ID,
		Workflow:     inst.Edges.Workflow.ID,
		Revision:     inst.Edges.Revision.ID,
		Runtime:      inst.Edges.Runtime.ID,
	}

}

func entInstanceRuntime(rt *ent.InstanceRuntime) *InstanceRuntime {

	if rt == nil {
		return nil
	}

	x := &InstanceRuntime{
		ID:              rt.ID,
		Input:           rt.Input,
		Data:            rt.Data,
		Controller:      rt.Controller,
		Memory:          rt.Memory,
		Flow:            rt.Flow,
		Output:          rt.Output,
		StateBeginTime:  rt.StateBeginTime,
		Deadline:        rt.Deadline,
		Attempts:        rt.Attempts,
		CallerData:      rt.CallerData,
		InstanceContext: rt.InstanceContext,
		StateContext:    rt.StateContext,
		Metadata:        rt.Metadata,
	}

	if rt.Edges.Caller != nil {
		x.Caller = rt.Edges.Caller.ID
	}

	return x

}

func (db *EntDatabase) entAnnotation(annotation *ent.Annotation) *Annotation {

	if annotation == nil {
		return nil
	}

	return &Annotation{
		ID:        annotation.ID,
		Name:      annotation.Name,
		CreatedAt: annotation.CreatedAt,
		UpdatedAt: annotation.UpdatedAt,
		Size:      annotation.Size,
		Hash:      annotation.Hash,
		Data:      annotation.Data,
		MimeType:  annotation.MimeType,
	}

}

func (db *EntDatabase) entVarRef(vref *ent.VarRef) *VarRef {

	if vref == nil {
		return nil
	}

	return &VarRef{
		ID:        vref.ID,
		Name:      vref.Name,
		Behaviour: vref.Behaviour,
		VarData:   vref.Edges.Vardata.ID,
	}

}

func (db *EntDatabase) entVarData(v *ent.VarData) *VarData {

	if v == nil {
		return nil
	}

	return &VarData{
		ID:        v.ID,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
		Size:      v.Size,
		Hash:      v.Hash,
		Data:      v.Data,
		MimeType:  v.MimeType,
	}

}

func entMirror(v *ent.Mirror) *Mirror {

	if v == nil {
		return nil
	}

	return &Mirror{
		ID:         v.ID,
		URL:        v.URL,
		Ref:        v.Ref,
		Cron:       v.Cron,
		PublicKey:  v.PublicKey,
		PrivateKey: v.PrivateKey,
		Passphrase: v.Passphrase,
		Commit:     v.Commit,
		LastSync:   v.LastSync,
		UpdatedAt:  v.UpdatedAt,
	}

}

func entMirrorActivity(v *ent.MirrorActivity) *MirrorActivity {

	if v == nil {
		return nil
	}

	return &MirrorActivity{
		ID:         v.ID,
		Type:       v.Type,
		Status:     v.Status,
		CreatedAt:  v.CreatedAt,
		UpdatedAt:  v.UpdatedAt,
		EndAt:      v.EndAt,
		Controller: v.Controller,
		Deadline:   v.Deadline,
	}

}

func (db *EntDatabase) Namespace(ctx context.Context, tx Transaction, id uuid.UUID) (*Namespace, error) {

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

func (db *EntDatabase) NamespaceByName(ctx context.Context, tx Transaction, name string) (*Namespace, error) {

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

func (db *EntDatabase) Inode(ctx context.Context, tx Transaction, id uuid.UUID) (*Inode, error) {

	clients := db.clients(tx)

	ino, err := clients.Inode.Query().Where(entino.ID(id)).WithChildren(func(q *ent.InodeQuery) {
		q.Order(ent.Asc(entino.FieldName)).Select(entino.FieldID)
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

func (db *EntDatabase) Workflow(ctx context.Context, tx Transaction, id uuid.UUID) (*Workflow, error) {

	clients := db.clients(tx)

	wf, err := clients.Workflow.Query().Where(entwf.ID(id)).WithInode(func(q *ent.InodeQuery) {
		q.Select(entino.FieldID)
	}).WithNamespace(func(q *ent.NamespaceQuery) {
		q.Select(entns.FieldID)
	}).WithRefs(func(q *ent.RefQuery) {
		q.Order(ent.Desc(entref.FieldCreatedAt)).WithRevision(func(q *ent.RevisionQuery) {
			q.Select(entrev.FieldID)
		})
	}).WithRevisions(func(q *ent.RevisionQuery) {
		q.Order(ent.Desc(entrev.FieldCreatedAt)).Select(entrev.FieldID, entrev.FieldHash)
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

func (db *EntDatabase) Revision(ctx context.Context, tx Transaction, id uuid.UUID) (*Revision, error) {

	clients := db.clients(tx)

	rev, err := clients.Revision.Query().Where(entrev.ID(id)).WithWorkflow(func(q *ent.WorkflowQuery) {
		q.Select(entwf.FieldID)
	}).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve revision: %v", parent(), err)
		return nil, err
	}

	return entRevision(rev), nil

}

func (db *EntDatabase) Instance(ctx context.Context, tx Transaction, id uuid.UUID) (*Instance, error) {

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

func (db *EntDatabase) InstanceRuntime(ctx context.Context, tx Transaction, id uuid.UUID) (*InstanceRuntime, error) {

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

func (db *EntDatabase) NamespaceAnnotation(ctx context.Context, tx Transaction, nsID uuid.UUID, key string) (*Annotation, error) {

	clients := db.clients(tx)

	annotation, err := clients.Annotation.Query().Where(entnote.HasNamespaceWith(entns.ID(nsID)), entnote.Name(key)).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve namespace annotation: %v", parent(), err)
		return nil, err
	}

	return db.entAnnotation(annotation), nil

}

func (db *EntDatabase) InodeAnnotation(ctx context.Context, tx Transaction, inodeID uuid.UUID, key string) (*Annotation, error) {

	clients := db.clients(tx)

	annotation, err := clients.Annotation.Query().Where(entnote.HasInodeWith(entino.ID(inodeID)), entnote.Name(key)).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve inode annotation: %v", parent(), err)
		return nil, err
	}

	return db.entAnnotation(annotation), nil

}

func (db *EntDatabase) WorkflowAnnotation(ctx context.Context, tx Transaction, wfID uuid.UUID, key string) (*Annotation, error) {

	clients := db.clients(tx)

	annotation, err := clients.Annotation.Query().Where(entnote.HasWorkflowWith(entwf.ID(wfID)), entnote.Name(key)).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve workflow annotation: %v", parent(), err)
		return nil, err
	}

	return db.entAnnotation(annotation), nil

}

func (db *EntDatabase) InstanceAnnotation(ctx context.Context, tx Transaction, instID uuid.UUID, key string) (*Annotation, error) {

	clients := db.clients(tx)

	annotation, err := clients.Annotation.Query().Where(entnote.HasInstanceWith(entinst.ID(instID)), entnote.Name(key)).Only(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve instance annotation: %v", parent(), err)
		return nil, err
	}

	return db.entAnnotation(annotation), nil

}

func (db *EntDatabase) ThreadVariables(ctx context.Context, tx Transaction, instID uuid.UUID) ([]*VarRef, error) {

	clients := db.clients(tx)

	varrefs, err := clients.VarRef.Query().Where(entvar.HasInstanceWith(entinst.ID(instID)), entvar.BehaviourEQ("thread")).WithVardata(func(q *ent.VarDataQuery) {
		q.Select(entvardata.FieldID)
	}).All(ctx)
	if err != nil {
		db.sugar.Debugf("%s failed to resolve instance thread variables: %v", parent(), err)
		return nil, err
	}

	x := make([]*VarRef, 0)

	for _, y := range varrefs {
		x = append(x, db.entVarRef(y))
	}

	return x, nil

}

func (db *EntDatabase) NamespaceVariableRef(ctx context.Context, tx Transaction, nsID uuid.UUID, key string) (*VarRef, error) {

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

func (db *EntDatabase) WorkflowVariableRef(ctx context.Context, tx Transaction, wfID uuid.UUID, key string) (*VarRef, error) {

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

func (db *EntDatabase) InstanceVariableRef(ctx context.Context, tx Transaction, instID uuid.UUID, key string) (*VarRef, error) {

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

func (db *EntDatabase) ThreadVariableRef(ctx context.Context, tx Transaction, instID uuid.UUID, key string) (*VarRef, error) {

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

func (db *EntDatabase) VariableData(ctx context.Context, tx Transaction, id uuid.UUID, load bool) (*VarData, error) {

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

func (db *EntDatabase) Mirror(ctx context.Context, tx Transaction, id uuid.UUID) (*Mirror, error) {

	clients := db.clients(tx)

	mir, err := clients.Mirror.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return entMirror(mir), nil

}
