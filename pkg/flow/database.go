package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/vorteil/direktiv/pkg/flow/ent"
	entino "github.com/vorteil/direktiv/pkg/flow/ent/inode"
	entinst "github.com/vorteil/direktiv/pkg/flow/ent/instance"
	entirt "github.com/vorteil/direktiv/pkg/flow/ent/instanceruntime"
	entns "github.com/vorteil/direktiv/pkg/flow/ent/namespace"
	entref "github.com/vorteil/direktiv/pkg/flow/ent/ref"
	entvardata "github.com/vorteil/direktiv/pkg/flow/ent/vardata"
	entvar "github.com/vorteil/direktiv/pkg/flow/ent/varref"
)

const latest = "latest"

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

	return db, nil

}

func rollback(tx *ent.Tx) {

	err := tx.Rollback()
	if err != nil && !strings.Contains(err.Error(), "already been") {
		fmt.Fprintf(os.Stderr, "failed to rollback transaction: %v\n", err)
	}

}

func (srv *server) getNamespace(ctx context.Context, nsc *ent.NamespaceClient, namespace string) (*ent.Namespace, error) {

	query := nsc.Query()
	query = query.Where(entns.NameEQ(namespace))
	ns, err := query.Only(ctx)
	if err != nil {
		return nil, err
	}

	return ns, nil

}

func getInodePath(path string) string {
	if strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

type nodeData struct {
	ino             *ent.Inode
	path, dir, base string
}

func (d *nodeData) ns() *ent.Namespace {
	return d.ino.Edges.Namespace
}

func (d *nodeData) namespace() string {
	return d.ns().Name
}

func (srv *server) traverseToInode(ctx context.Context, nsc *ent.NamespaceClient, namespace, path string) (*nodeData, error) {

	ns, err := srv.getNamespace(ctx, nsc, namespace)
	if err != nil {
		return nil, err
	}

	nd, err := srv.getInode(ctx, nil, ns, path, false)
	if err != nil {
		return nil, err
	}

	return nd, nil

}

func (srv *server) reverseTraverseToInode(ctx context.Context, id string) (*nodeData, error) {

	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	d := new(nodeData)

	ino, err := srv.db.Inode.Get(ctx, uid)
	if err != nil {
		return nil, err
	}

	ns, err := ino.Namespace(ctx)
	if err != nil {
		return nil, err
	}

	ino.Edges.Namespace = ns

	d.ino = ino
	d.path = ino.Name
	d.dir = ""
	d.base = ino.Name

	var recurser func(ino *ent.Inode) error

	recurser = func(ino *ent.Inode) error {

		pino, err := ino.Parent(ctx)
		if IsNotFound(err) || pino == nil {
			d.dir = "/" + d.dir
			d.path = "/" + d.path
			return nil
		}
		if err != nil {
			return err
		}

		pino.Edges.Namespace = ns
		ino.Edges.Parent = pino
		if pino.Name != "" {
			d.path = pino.Name + "/" + d.path
			d.dir = pino.Name + "/" + pino.Name
		}

		return recurser(pino)

	}

	err = recurser(ino)
	if err != nil {
		return nil, err
	}

	return d, nil

}

func (srv *server) getInode(ctx context.Context, inoc *ent.InodeClient, ns *ent.Namespace, path string, createParents bool) (*nodeData, error) {

	d := new(nodeData)
	d.path = getInodePath(path)
	d.dir, d.base = filepath.Split(d.path)

	query := ns.QueryInodes()
	query = query.Where(entino.NameIsNil())
	rootino, err := query.Only(ctx)
	if err != nil {
		return nil, err
	}

	elems := strings.Split(path, "/")
	if elems[0] == "" {
		elems = elems[1:]
	}
	path = "/"

	var descend func(*ent.Inode, []string, string) (*ent.Inode, error)
	descend = func(ino *ent.Inode, elems []string, path string) (*ent.Inode, error) {

		ino.Edges.Namespace = ns

		if len(elems) == 0 || elems[0] == "" {
			return ino, nil
		}

		if !strings.HasSuffix(path, "/") {
			path = path + "/"
		}
		path = path + elems[0]

		query := ino.QueryChildren()
		query = query.Where(entino.NameEQ(elems[0]))
		child, err := query.Only(ctx)
		if err != nil {
			if IsNotFound(err) {

				if createParents && inoc != nil && len(elems) > 1 {
					child, err = inoc.Create().SetName(elems[0]).SetNamespace(ns).SetParent(ino).SetType("directory").Save(ctx)
				} else {
					err = &NotFoundError{
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

	ino, err := descend(rootino, elems, path)
	if err != nil {
		return nil, err
	}

	d.ino = ino
	d.ino.Edges.Namespace = ns

	return d, nil

}

func (srv *server) getWorkflow(ctx context.Context, ino *ent.Inode) (*ent.Workflow, error) {

	if ino.Type != "workflow" {
		return nil, ErrNotWorkflow
	}

	wf, err := ino.QueryWorkflow().Only(ctx)
	if err != nil {
		return nil, err
	}

	return wf, nil

}

func (srv *server) getRef(ctx context.Context, wf *ent.Workflow, reference string) (*ent.Ref, error) {

	ref, err := wf.QueryRefs().Where(entref.NameEQ(reference)).Only(ctx)
	if err != nil {
		return nil, err
	}

	return ref, nil

}

func (srv *server) getRevision(ctx context.Context, ref *ent.Ref) (*ent.Revision, error) {

	rev, err := ref.Revision(ctx)
	if err != nil {
		return nil, err
	}

	ref.Edges.Revision = rev

	return ref.Edges.Revision, nil

}

type wfData struct {
	*nodeData
	wf *ent.Workflow
}

func (srv *server) traverseToWorkflow(ctx context.Context, nsc *ent.NamespaceClient, namespace, path string) (*wfData, error) {

	nd, err := srv.traverseToInode(ctx, nsc, namespace, path)
	if err != nil {
		return nil, err
	}

	wd := new(wfData)
	wd.nodeData = nd

	wf, err := srv.getWorkflow(ctx, wd.ino)
	if err != nil {
		return nil, err
	}

	wd.wf = wf

	wd.ino.Edges.Namespace = wd.ns()
	// NOTE: can't do this due to cycle: wd.ino.Edges.Workflow = wf
	wf.Edges.Inode = wd.ino
	wf.Edges.Namespace = wd.ns()

	return wd, nil

}

func (srv *server) reverseTraverseToWorkflow(ctx context.Context, id string) (*wfData, error) {

	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	wf, err := srv.db.Workflow.Get(ctx, uid)
	if err != nil {
		return nil, err
	}

	ino, err := wf.Inode(ctx)
	if err != nil {
		return nil, err
	}

	nd, err := srv.reverseTraverseToInode(ctx, ino.ID.String())
	if err != nil {
		return nil, err
	}

	wf.Edges.Inode = nd.ino
	wf.Edges.Namespace = nd.ino.Edges.Namespace

	d := new(wfData)
	d.wf = wf
	d.nodeData = nd

	return d, nil

}

type refData struct {
	*wfData
	ref *ent.Ref
}

func (d *refData) rev() *ent.Revision {
	return d.ref.Edges.Revision
}

func (d *refData) reference() string {
	return d.ref.Name
}

func (srv *server) traverseToRef(ctx context.Context, nsc *ent.NamespaceClient, namespace, path, reference string) (*refData, error) {

	wd, err := srv.traverseToWorkflow(ctx, nsc, namespace, path)
	if err != nil {
		return nil, err
	}

	rd := new(refData)

	if reference == "" {
		reference = latest
	}

	ref, err := srv.getRef(ctx, wd.wf, reference)
	if err != nil {
		return nil, err
	}

	rd.wfData = wd
	rd.ref = ref

	rev, err := srv.getRevision(ctx, ref)
	if err != nil {
		return nil, err
	}

	ref.Edges.Revision = rev
	ref.Edges.Workflow = wd.wf
	// NOTE: can't do this due to cycle: rev.Edges.Workflow = wd.wf

	return rd, nil

}

type instData struct {
	in *ent.Instance
	*nodeData
}

func (d *instData) ns() *ent.Namespace {
	return d.in.Edges.Namespace
}

func (d *instData) namespace() string {
	return d.in.Edges.Namespace.Name
}

func (srv *server) getInstance(ctx context.Context, nsc *ent.NamespaceClient, namespace, instance string, load bool) (*instData, error) {

	id, err := uuid.Parse(instance)
	if err != nil {
		return nil, err
	}

	ns, err := srv.getNamespace(ctx, nsc, namespace)
	if err != nil {
		return nil, err
	}

	query := ns.QueryInstances().Where(entinst.IDEQ(id))
	if load {
		query = query.WithRuntime()
	}
	in, err := query.Only(ctx)
	if err != nil {
		return nil, err
	}

	in.Edges.Namespace = ns

	d := new(instData)
	d.in = in

	return d, nil

}

func (srv *server) traverseToInstance(ctx context.Context, nsc *ent.NamespaceClient, namespace, instance string) (*instData, error) {

	d, err := srv.getInstance(ctx, nsc, namespace, instance, false)
	if err != nil {
		return nil, err
	}

	rt, err := d.in.QueryRuntime().Select(entirt.FieldFlow).WithCaller().Only(ctx)
	if err != nil {
		return nil, err
	}
	d.in.Edges.Runtime = rt

	rev, err := d.in.QueryRevision().Only(ctx)
	if err != nil {
		return nil, err
	}
	d.in.Edges.Revision = rev

	wf, err := d.in.QueryWorkflow().Only(ctx)
	if err != nil {
		return nil, err
	}
	d.in.Edges.Workflow = wf

	ino, err := wf.QueryInode().Only(ctx)
	if err != nil {
		return nil, err
	}
	wf.Edges.Inode = ino

	elems := make([]string, 0)

	var recurser func(x *ent.Inode) error

	recurser = func(x *ent.Inode) error {

		parent, err := x.QueryParent().Only(ctx)

		if err != nil {

			if IsNotFound(err) {
				return nil
			}

			return err

		}

		x.Edges.Parent = parent

		err = recurser(parent)
		if err != nil {
			return err
		}

		elems = append(elems, parent.Name)

		return nil

	}

	err = recurser(ino)
	if err != nil {
		return nil, err
	}

	nd := new(nodeData)
	d.nodeData = nd
	d.ino = ino
	d.dir = filepath.Join(elems...)
	d.base = ino.Name
	d.path = filepath.Join(d.dir, d.base)

	return d, nil

}

func (internal *internal) getInstance(ctx context.Context, inc *ent.InstanceClient, instance string, load bool) (*instData, error) {

	id, err := uuid.Parse(instance)
	if err != nil {
		return nil, err
	}

	query := inc.Query().Where(entinst.IDEQ(id)).WithNamespace().WithWorkflow(func(q *ent.WorkflowQuery) {
		q.WithInode()
	})
	if load {
		query = query.WithRuntime()
	}
	in, err := query.Only(ctx)
	if err != nil {
		return nil, err
	}

	d := new(instData)
	d.in = in

	// TODO: check if this is a good place to put this
	// d.nodeData, err = internal.reverseTraverseToInode(ctx, in.Edges.Workflow.Edges.Inode.ID.String())
	// if err != nil {
	// 	return nil, err
	// }

	return d, nil

}

type nsvarData struct {
	vref  *ent.VarRef
	vdata *ent.VarData
}

func (d *nsvarData) ns() *ent.Namespace {
	return d.vref.Edges.Namespace
}

func (srv *server) traverseToNamespaceVariable(ctx context.Context, nsc *ent.NamespaceClient, namespace, key string, load bool) (*nsvarData, error) {

	ns, err := srv.getNamespace(ctx, nsc, namespace)
	if err != nil {
		return nil, err
	}

	query := ns.QueryVars().Where(entvar.NameEQ(key))
	if load {
		query = query.WithVardata()
	}

	vref, err := query.Only(ctx)
	if err != nil {
		return nil, err
	}

	if !load {
		vdata, err := vref.QueryVardata().Select(entvardata.FieldCreatedAt, entvardata.FieldHash, entvardata.FieldSize, entvardata.FieldUpdatedAt).Only(ctx)
		if err != nil {
			return nil, err
		}
		vref.Edges.Vardata = vdata
	}

	vref.Edges.Namespace = ns

	d := new(nsvarData)
	d.vref = vref
	d.vdata = vref.Edges.Vardata

	return d, nil

}

type wfvarData struct {
	*wfData
	vref  *ent.VarRef
	vdata *ent.VarData
}

func (srv *server) traverseToWorkflowVariable(ctx context.Context, nsc *ent.NamespaceClient, namespace, path, key string, load bool) (*wfvarData, error) {

	wd, err := srv.traverseToWorkflow(ctx, nsc, namespace, path)
	if err != nil {
		return nil, err
	}

	query := wd.wf.QueryVars().Where(entvar.NameEQ(key))
	if load {
		query = query.WithVardata()
	}

	vref, err := query.Only(ctx)
	if err != nil {
		return nil, err
	}

	if !load {
		vdata, err := vref.QueryVardata().Select(entvardata.FieldCreatedAt, entvardata.FieldHash, entvardata.FieldSize, entvardata.FieldUpdatedAt).Only(ctx)
		if err != nil {
			return nil, err
		}
		vref.Edges.Vardata = vdata
	}

	vref.Edges.Workflow = wd.wf

	d := new(wfvarData)
	d.wfData = wd
	d.vref = vref
	d.vdata = vref.Edges.Vardata

	return d, nil

}

type instvarData struct {
	*instData
	vref  *ent.VarRef
	vdata *ent.VarData
}

func (srv *server) traverseToInstanceVariable(ctx context.Context, nsc *ent.NamespaceClient, namespace, instance, key string, load bool) (*instvarData, error) {

	wd, err := srv.getInstance(ctx, nsc, namespace, instance, false)
	if err != nil {
		return nil, err
	}

	query := wd.in.QueryVars().Where(entvar.NameEQ(key))
	if load {
		query = query.WithVardata()
	}

	vref, err := query.Only(ctx)
	if err != nil {
		return nil, err
	}

	if !load {
		vdata, err := vref.QueryVardata().Select(entvardata.FieldCreatedAt, entvardata.FieldHash, entvardata.FieldSize, entvardata.FieldUpdatedAt).Only(ctx)
		if err != nil {
			return nil, err
		}
		vref.Edges.Vardata = vdata
	}

	vref.Edges.Instance = wd.in

	d := new(instvarData)
	d.instData = wd
	d.vref = vref
	d.vdata = vref.Edges.Vardata

	return d, nil

}

func (engine *engine) SetMemory(ctx context.Context, im *instanceMemory, x interface{}) error {

	im.SetMemory(x)

	data, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	s := string(data)

	ir, err := im.in.Edges.Runtime.Update().SetMemory(s).Save(ctx)
	if err != nil {
		return NewInternalError(err)
	}

	ir.Edges = im.in.Edges.Runtime.Edges
	im.in.Edges.Runtime = ir

	return nil

}
