package flow

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	"github.com/google/uuid"
)

const latest = "latest"

// TODO: use escapes on these key functions
func nsCacheKey(namespace string) string {
	return fmt.Sprintf("ns:%s", namespace)
}

type nsData struct {
	ID     uuid.UUID
	Name   string
	Config string
}

func (srv *server) nsCacheDataUnmarshal(data []byte) *nsData {

	d := new(nsData)

	err := json.Unmarshal(data, d)
	if err != nil {
		srv.sugar.Debugf("%s failed to unmarshal namespace cache data: %v", parent(), err)
		return nil
	}

	return d

}

func (nsd *nsData) Bytes() []byte {
	data, _ := json.Marshal(nsd)
	return data
}

func initDatabase(ctx context.Context, addr string) (*ent.Client, error) {

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

	return db, nil

}

func rollback(tx *ent.Tx) {

	err := tx.Rollback()
	if err != nil && !strings.Contains(err.Error(), "already been") {
		fmt.Fprintf(os.Stderr, "failed to rollback transaction: %v\n", err)
	}

}

// GetInodePath returns the exact path to a inode.
func GetInodePath(path string) string {
	path = strings.TrimSuffix(path, "/")
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	path = filepath.Clean(path)
	return path
}

/*
func (srv *server) traverseToInode(ctx context.Context, tx Transaction, namespace, path string) (*Inode, error) {

	ns, err := srv.database.NamespaceByName(ctx, tx, namespace)
	if err != nil {
		srv.sugar.Debugf("%s failed to resolve namespace: %v", parent(), err)
		return nil, err
	}

	nd, err := srv.getInode(ctx, nil, ns, path, false)
	if err != nil {
		srv.sugar.Debugf("%s failed to resolve inode: %v", parent(), err)
		return nil, err
	}

	return nd, nil

}

func (srv *server) reverseTraverseToInode(ctx context.Context, tx Transaction, id string) (*Inode, error) {

	uid, err := uuid.Parse(id)
	if err != nil {
		srv.sugar.Debugf("%s failed to parse UUID: %v", parent(), err)
		return nil, err
	}

	d := new(Inode)

	ino, err := srv.database.Inode(ctx, tx, uid)
	if err != nil {
		srv.sugar.Debugf("%s failed to resolve inode: %v", parent(), err)
		return nil, err
	}

	cached.Inode() = ino
	cached.Path() = ino.Name
	cached.Dir() = ""
	cached.Inode().Name = ino.Name

	var recurser func(ino *Inode) error

	recurser = func(ino *Inode) error {

		pino, err := srv.database.Inode(ctx, tx, ino.Edges.Parent.ID)
		if derrors.IsNotFound(err) || pino == nil {
			cached.Dir() = "/" + cached.Dir()
			cached.Path() = "/" + cached.Path()
			return nil
		}
		if err != nil {
			return err
		}

		ino.Edges.Parent = pino
		if pino.Name != "" {
			cached.Path() = pino.Name + "/" + cached.Path()
			cached.Dir() = pino.Name + "/" + pino.Name
		}

		return recurser(pino)

	}

	err = recurser(ino)
	if err != nil {
		srv.sugar.Debugf("%s failed to resolve parent(s): %v", parent(), err)
		return nil, err
	}

	return d, nil

}

func (srv *server) getInode(ctx context.Context, tx Transaction, ns *Namespace, path string, createParents bool) (*Inode, error) {

	clients := srv.entClients(tx)

	elems := strings.Split(path, "/")
	if elems[0] == "" {
		elems = elems[1:]
	}

	var descend func(*Inode, []string, string) (*Inode, error)
	descend = func(ino *Inode, elems []string, path string) (*Inode, error) {

		if len(elems) == 0 || elems[0] == "" {
			return ino, nil
		}

		if !strings.HasSuffix(path, "/") {
			path = path + "/"
		}
		path = path + elems[0]

		child, err := srv.database.InodeByParent(ctx, tx, ino, elems[0])
		if err != nil {
			if derrors.IsNotFound(err) {

				if createParents && clients.Inode != nil && len(elems) > 1 {
					x, err := clients.Inode.Create().SetName(elems[0]).SetNamespaceID(ino.Edges.Namespace.ID).SetParentID(ino.ID).SetType(util.InodeTypeDirectory).Save(ctx)
					if err != nil {
						return nil, err
					}

					child = entInode(x)
					child.Edges.Namespace = ino.Edges.Namespace
					child.Edges.Parent = &Inode{
						ID: ino.ID,
					}

				} else {
					err = &derrors.NotFoundError{
						Label: fmt.Sprintf("inode not found at '%s'", path),
					}
				}

			}
			if err != nil {
				return nil, err
			}
		}
		child.Edges.Parent = ino

		elems = elems[1:]

		ino, err = descend(child, elems, path)
		if err != nil {
			return nil, err
		}

		return ino, nil

	}

	ino, err := descend(ns.Edges.Root, elems, path)
	if err != nil {
		srv.sugar.Debugf("%s failed to resolve inode: %v", parent(), err)
		return nil, err
	}

	return ino, nil

}

func (srv *server) getWorkflow(ctx context.Context, ino *ent.Inode) (*Workflow, error) {

	if ino.Type != util.InodeTypeWorkflow {
		srv.sugar.Debugf("%s inode isn't a workflow", parent())
		return nil, ErrNotWorkflow
	}

	wf, err := ino.QueryWorkflow().Only(ctx)
	if err != nil {
		srv.sugar.Debugf("%s failed to query inode's workflow: %v", parent(), err)
		return nil, err
	}

	return entWorkflow(wf), nil

}

func (srv *server) getRef(ctx context.Context, wf *Workflow, reference string) (*ent.Ref, error) {

	ref, err := wf.QueryRefs().Where(entref.NameEQ(reference)).Only(ctx)
	if err != nil {
		srv.sugar.Debugf("%s failed to query workflow ref: %v", parent(), err)
		return nil, err
	}

	return ref, nil

}

func (srv *server) getRevision(ctx context.Context, ref *ent.Ref) (*ent.Revision, error) {

	rev, err := ref.Revision(ctx)
	if err != nil {
		srv.sugar.Debugf("%s failed to query ref's revision: %v", parent(), err)
		return nil, err
	}

	ref.Edges.Revision = rev

	return ref.Edges.Revision, nil

}

func (srv *server) reverseTraverseToWorkflow(ctx context.Context, id string) (*wfData, error) {

	uid, err := uuid.Parse(id)
	if err != nil {
		srv.sugar.Debugf("%s failed to parse workflow UUID: %v", parent(), err)
		return nil, err
	}

	wf, err := srv.db.Workflow.Get(ctx, uid)
	if err != nil {
		srv.sugar.Debugf("%s failed to query workflow: %v", parent(), err)
		return nil, err
	}

	ino, err := wf.Inode(ctx)
	if err != nil {
		srv.sugar.Debugf("%s failed to query workflow's inode: %v", parent(), err)
		return nil, err
	}

	nd, err := srv.reverseTraverseToInode(ctx, srv.db.Inode, ino.ID.String())
	if err != nil {
		srv.sugar.Debugf("%s failed to resolve inode's parent(s): %v", parent(), err)
		return nil, err
	}

	wf.Edges.Inode = nd.ino

	wfd := new(wfData)
	wfd.wf = wf
	wfd.nodeData = nd

	return wfd, nil

}

type instData struct {
	in     *ent.Instance
	inoded *nodeData
}

func (d *instData) namespace() string {
	return cached.Namespace.Name
}

func (srv *server) getInstance(ctx context.Context, tx Transaction, namespace, instance string, load bool) (*instData, error) {

	id, err := uuid.Parse(instance)
	if err != nil {
		srv.sugar.Debugf("%s failed to parse UUID: %v", parent(), err)
		return nil, err
	}

	ns, err := srv.database.NamespaceByName(ctx, tx, namespace)
	if err != nil {
		srv.sugar.Debugf("%s failed to resolve namespace: %v", parent(), err)
		return nil, err
	}

	clients := srv.entClients(tx)

	query := clients.Instance.Query().Where(entinst.HasNamespaceWith(entns.ID(ns.ID))).Where(entinst.IDEQ(id))
	if load {
		query = query.WithRuntime()
	}
	in, err := query.Only(ctx)
	if err != nil {
		srv.sugar.Debugf("%s failed to query instance: %v", parent(), err)
		return nil, err
	}

	if load && in.Edges.Runtime == nil {
		err = &derrors.NotFoundError{
			Label: "instance runtime not found",
		}
		srv.sugar.Debugf("%s failed to query instance runtime: %v", parent(), err)
		return nil, err
	}

	d := new(instData)
	d.in = in
	cached.Namespace = entNamespace(in.Edges.Namespace)

	return d, nil

}

func (srv *server) fastGetInstance(ctx context.Context, d *instData) (*instData, error) {

	query := srv.db.Instance.Query().Where(entinst.IDEQ(d.in.ID)).WithRuntime()

	in, err := query.Only(ctx)
	if err != nil {
		srv.sugar.Debugf("%s failed to query instance: %v", parent(), err)
		return nil, err
	}

	if in.Edges.Runtime == nil {
		err = &derrors.NotFoundError{
			Label: "instance runtime not found",
		}
		srv.sugar.Debugf("%s failed to query instance runtime: %v", parent(), err)
		return nil, err
	}

	d.in = in

	return d, nil

}

func (internal *internal) getInstance(ctx context.Context, Instance *ent.InstanceClient, instance string, load bool) (*instData, error) {

	id, err := uuid.Parse(instance)
	if err != nil {
		internal.sugar.Debugf("%s failed to parse UUID: %v", parent(), err)
		return nil, err
	}

	query := Instance.Query().Where(entinst.IDEQ(id)).WithNamespace().WithWorkflow(func(q *ent.WorkflowQuery) {
		q.WithInode()
	})
	if load {
		query = query.WithRuntime()
	}
	in, err := query.Only(ctx)
	if err != nil {
		internal.sugar.Debugf("%s failed to query instance: %v", parent(), err)
		return nil, err
	}

	if in.Edges.Namespace == nil {
		err = &derrors.NotFoundError{
			Label: "instance namespace not found",
		}
		internal.sugar.Debugf("%s failed to query instance namespace: %v", parent(), err)
		return nil, err
	}

	if in.Edges.Workflow == nil {
		err = &derrors.NotFoundError{
			Label: "instance workflow not found",
		}
		internal.sugar.Debugf("%s failed to query instance workflow: %v", parent(), err)
		return nil, err
	}

	if in.Edges.Workflow.Edges.Inode == nil {
		err = &derrors.NotFoundError{
			Label: "instance workflow's inode not found",
		}
		internal.sugar.Debugf("%s failed to query workflow inode: %v", parent(), err)
		return nil, err
	}

	if load && in.Edges.Runtime == nil {
		err = &derrors.NotFoundError{
			Label: "instance runtime not found",
		}
		internal.sugar.Debugf("%s failed to query instance runtime: %v", parent(), err)
		return nil, err
	}

	d := new(instData)
	d.in = in

	return d, nil

}

type nsvarData struct {
	vref  *ent.VarRef
	vdata *ent.VarData
}

func (d *nsvarData) ns() *ent.Namespace {
	return vref.Edges.Namespace
}

func (srv *server) traverseToNamespaceVariable(ctx context.Context, tx Transaction, namespace, key string, load bool) (*nsvarData, error) {

	ns, err := srv.database.NamespaceByName(ctx, tx, namespace)
	if err != nil {
		srv.sugar.Debugf("%s failed to resolve namespace: %v", parent(), err)
		return nil, err
	}

	clients := srv.entClients(tx)

	query := clients.VarRef.Query().Where(entvar.HasNamespaceWith(entns.ID(ns.ID))).Where(entvar.NameEQ(key))
	if load {
		query = query.WithVardata()
	}

	vref, err := query.Only(ctx)
	if err != nil {
		srv.sugar.Debugf("%s failed to query variable ref: %v", parent(), err)
		return nil, err
	}

	if load && vref.Edges.Vardata == nil {
		err = &derrors.NotFoundError{
			Label: "variable data not found",
		}
		srv.sugar.Debugf("%s failed to query variable data: %v", parent(), err)
		return nil, err
	}

	if !load {
		vdata, err := vref.QueryVardata().Select(entvardata.FieldCreatedAt, entvardata.FieldHash, entvardata.FieldSize, entvardata.FieldUpdatedAt).Only(ctx)
		if err != nil {
			srv.sugar.Debugf("%s failed to query variable metadata: %v", parent(), err)
			return nil, err
		}
		vref.Edges.Vardata = vdata
	}

	d := new(nsvarData)
	vref = vref
	vdata = vref.Edges.Vardata

	return d, nil

}

type wfvarData struct {
	wfd   *wfData
	vref  *ent.VarRef
	vdata *ent.VarData
}

func (srv *server) traverseToWorkflowVariable(ctx context.Context, tx Transaction, namespace, path, key string, load bool) (*wfvarData, error) {

	wd, err := srv.traverseToWorkflow(ctx, tx, namespace, path)
	if err != nil {
		srv.sugar.Debugf("%s failed to resolve workflow: %v", parent(), err)
		return nil, err
	}

	query := wd.wf.QueryVars().Where(entvar.NameEQ(key))
	if load {
		query = query.WithVardata()
	}

	vref, err := query.Only(ctx)
	if err != nil {
		srv.sugar.Debugf("%s failed to query variable ref: %v", parent(), err)
		return nil, err
	}

	if load && vref.Edges.Vardata == nil {
		err = &derrors.NotFoundError{
			Label: "variable data not found",
		}
		srv.sugar.Debugf("%s failed to query variable data: %v", parent(), err)
		return nil, err
	}

	if !load {
		vdata, err := vref.QueryVardata().Select(entvardata.FieldCreatedAt, entvardata.FieldHash, entvardata.FieldSize, entvardata.FieldUpdatedAt).Only(ctx)
		if err != nil {
			srv.sugar.Debugf("%s failed to query variable metadata: %v", parent(), err)
			return nil, err
		}
		vref.Edges.Vardata = vdata
	}

	vref.Edges.Workflow = wd.wf

	d := new(wfvarData)
	d.wfData = wd
	vref = vref
	vdata = vref.Edges.Vardata

	return d, nil

}

type instvarData struct {
	*instData
	vref  *ent.VarRef
	vdata *ent.VarData
}

func (srv *server) traverseToInstanceVariable(ctx context.Context, tx Transaction, namespace, instance, key string, load bool) (*instvarData, error) {

	wd, err := srv.getInstance(ctx, tx, namespace, instance, false)
	if err != nil {
		srv.sugar.Debugf("%s failed to resolve instance: %v", parent(), err)
		return nil, err
	}

	query := wd.in.QueryVars().Where(entvar.NameEQ(key), entvar.BehaviourIsNil())
	if load {
		query = query.WithVardata()
	}

	vref, err := query.Only(ctx)
	if err != nil {
		srv.sugar.Debugf("%s failed to query variable ref: %v", parent(), err)
		return nil, err
	}

	if load && vref.Edges.Vardata == nil {
		err = &derrors.NotFoundError{
			Label: "variable data not found",
		}
		srv.sugar.Debugf("%s failed to query variable data: %v", parent(), err)
		return nil, err
	}

	if !load {
		vdata, err := vref.QueryVardata().Select(entvardata.FieldCreatedAt, entvardata.FieldHash, entvardata.FieldSize, entvardata.FieldUpdatedAt).Only(ctx)
		if err != nil {
			srv.sugar.Debugf("%s failed to query variable metadata: %v", parent(), err)
			return nil, err
		}
		vref.Edges.Vardata = vdata
	}

	vref.Edges.Instance = wd.in

	d := new(instvarData)
	d.instData = wd
	vref = vref
	vdata = vref.Edges.Vardata

	return d, nil

}

func (srv *server) traverseToThreadVariable(ctx context.Context, Namespace *ent.NamespaceClient, namespace, instance, key string, load bool) (*instvarData, error) {

	wd, err := srv.getInstance(ctx, Namespace, namespace, instance, false)
	if err != nil {
		srv.sugar.Debugf("%s failed to resolve instance: %v", parent(), err)
		return nil, err
	}

	query := wd.in.QueryVars().Where(entvar.NameEQ(key), entvar.BehaviourEQ("thread"))
	if load {
		query = query.WithVardata()
	}

	vref, err := query.Only(ctx)
	if err != nil {
		srv.sugar.Debugf("%s failed to query variable ref: %v", parent(), err)
		return nil, err
	}

	if load && vref.Edges.Vardata == nil {
		err = &derrors.NotFoundError{
			Label: "variable data not found",
		}
		srv.sugar.Debugf("%s failed to query variable data: %v", parent(), err)
		return nil, err
	}

	if !load {
		vdata, err := vref.QueryVardata().Select(entvardata.FieldCreatedAt, entvardata.FieldHash, entvardata.FieldSize, entvardata.FieldUpdatedAt).Only(ctx)
		if err != nil {
			srv.sugar.Debugf("%s failed to query variable metadata: %v", parent(), err)
			return nil, err
		}
		vref.Edges.Vardata = vdata
	}

	vref.Edges.Instance = wd.in

	d := new(instvarData)
	d.instData = wd
	vref = vref
	vdata = vref.Edges.Vardata

	return d, nil

}

type nsAnnotationData struct {
	annotation *ent.Annotation
}

func (d *nsAnnotationData) ns() *ent.Namespace {
	return annotation.Edges.Namespace
}

func (srv *server) traverseToNamespaceAnnotation(ctx context.Context, tx Transaction, namespace, key string) (*nsAnnotationData, error) {

	ns, err := srv.database.NamespaceByName(ctx, tx, namespace)
	if err != nil {
		srv.sugar.Debugf("%s failed to resolve namespace: %v", parent(), err)
		return nil, err
	}

	clients := srv.entClients(tx)

	query := clients.Annotation.Query().Where(entnote.HasNamespaceWith(entns.ID(ns.ID))).Where(entnote.NameEQ(key))

	annotation, err := query.Only(ctx)
	if err != nil {
		srv.sugar.Debugf("%s failed to query annotation: %v", parent(), err)
		return nil, err
	}

	d := new(nsAnnotationData)
	annotation = annotation

	return d, nil

}

type wfAnnotationData struct {
	wfd        *wfData
	annotation *ent.Annotation
}

func (srv *server) traverseToWorkflowAnnotation(ctx context.Context, tx Transaction, namespace, path, key string) (*wfAnnotationData, error) {

	wd, err := srv.traverseToWorkflow(ctx, tx, namespace, path)
	if err != nil {
		srv.sugar.Debugf("%s failed to resolve workflow: %v", parent(), err)
		return nil, err
	}

	query := wd.wf.QueryAnnotations().Where(entnote.NameEQ(key))

	annotation, err := query.Only(ctx)
	if err != nil {
		srv.sugar.Debugf("%s failed to query annotation: %v", parent(), err)
		return nil, err
	}

	annotation.Edges.Workflow = wd.wf

	d := new(wfAnnotationData)
	d.wfData = wd
	annotation = annotation

	return d, nil

}

type instAnnotationData struct {
	cached     *CacheData
	annotation *ent.Annotation
}

func (srv *server) traverseToInstanceAnnotation(ctx context.Context, tx Transaction, namespace, instance, key string) (*instAnnotationData, error) {

	wd, err := srv.getInstance(ctx, tx, namespace, instance, false)
	if err != nil {
		srv.sugar.Debugf("%s failed to resolve instance: %v", parent(), err)
		return nil, err
	}

	query := wd.in.QueryAnnotations().Where(entnote.NameEQ(key))

	annotation, err := query.Only(ctx)
	if err != nil {
		srv.sugar.Debugf("%s failed to query annotation: %v", parent(), err)
		return nil, err
	}

	annotation.Edges.Instance = wd.in

	d := new(instAnnotationData)
	d.instData = wd
	annotation = annotation

	return d, nil

}

type inodeAnnotationData struct {
	cached     *CacheData
	annotation *ent.Annotation
}

*/
