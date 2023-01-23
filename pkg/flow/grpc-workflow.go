package flow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/entwrapper"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entino "github.com/direktiv/direktiv/pkg/flow/ent/inode"
	entref "github.com/direktiv/direktiv/pkg/flow/ent/ref"
	entrev "github.com/direktiv/direktiv/pkg/flow/ent/revision"
	entwf "github.com/direktiv/direktiv/pkg/flow/ent/workflow"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (srv *server) traverseToWorkflow(ctx context.Context, tx database.Transaction, namespace, path string) (*database.CacheData, error) {

	cached, err := srv.traverseToInode(ctx, tx, namespace, path)
	if err != nil {
		return nil, err
	}

	err = srv.database.Workflow(ctx, tx, cached, cached.Inode().Workflow)
	if err != nil {
		return nil, err
	}

	return cached, nil

}

func (srv *server) reverseTraverseToWorkflow(ctx context.Context, tx database.Transaction, workflow string) (*database.CacheData, error) {

	id, err := uuid.Parse(workflow)
	if err != nil {
		return nil, err
	}

	cached := new(database.CacheData)

	err = srv.database.Workflow(ctx, tx, cached, id)
	if err != nil {
		return nil, err
	}

	return cached, nil

}

type lookupRefAndRevArgs struct {
	wf        *database.Workflow
	reference string
}

func (srv *server) traverseToRef(ctx context.Context, tx database.Transaction, namespace, path, reference string) (*database.CacheData, error) {

	if reference == "" {
		reference = latest
	}

	cached, err := srv.traverseToWorkflow(ctx, tx, namespace, path)
	if err != nil {
		srv.sugar.Debugf("%s failed to resolve workflow: %v", parent(), err)
		return nil, err
	}

	var ref *database.Ref

	for i := range cached.Workflow.Refs {
		x := cached.Workflow.Refs[i]
		if x.Name == reference {
			ref = x
			break
		}
	}

	cached.Ref = ref

	err = srv.database.Revision(ctx, tx, cached, ref.Revision)
	if err != nil {
		return nil, err
	}

	return cached, nil

}

func (flow *flow) ResolveWorkflowUID(ctx context.Context, req *grpc.ResolveWorkflowUIDRequest) (*grpc.WorkflowResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, err
	}

	cached := new(database.CacheData)
	err = flow.database.Workflow(ctx, nil, cached, id)
	if err != nil {
		return nil, err
	}

	var resp grpc.WorkflowResponse

	err = atob(cached.Inode(), &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Namespace = cached.Namespace.Name
	resp.Node.Parent = cached.Dir()
	resp.Node.Path = cached.Path()
	resp.Oid = cached.Workflow.ID.String()

	return &resp, nil

}

func (flow *flow) Workflow(ctx context.Context, req *grpc.WorkflowRequest) (*grpc.WorkflowResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, err := flow.traverseToRef(ctx, nil, req.GetNamespace(), req.GetPath(), req.GetRef())
	if err != nil {
		return nil, err
	}

	var resp grpc.WorkflowResponse

	err = atob(cached.Inode(), &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Namespace = cached.Namespace.Name
	resp.Node.Parent = cached.Dir()
	resp.Node.Path = cached.Path()
	resp.EventLogging = cached.Workflow.LogToEvents
	resp.Oid = cached.Workflow.ID.String()

	err = atob(cached.Revision, &resp.Revision)
	if err != nil {
		return nil, err
	}

	resp.Revision.Name = cached.Revision.ID.String()

	return &resp, nil

}

func (flow *flow) WorkflowStream(req *grpc.WorkflowRequest, srv grpc.Flow_WorkflowStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	cached, err := flow.traverseToRef(ctx, nil, req.GetNamespace(), req.GetPath(), req.GetRef())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeWorkflow(cached)
	defer flow.cleanup(sub.Close)

resend:

	resp := new(grpc.WorkflowResponse)

	err = atob(cached.Inode(), &resp.Node)
	if err != nil {
		return err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Namespace = cached.Namespace.Name
	resp.Node.Parent = cached.Dir()
	resp.Node.Path = cached.Path()
	resp.Oid = cached.Workflow.ID.String()
	resp.EventLogging = cached.Workflow.LogToEvents

	err = atob(cached.Revision, &resp.Revision)
	if err != nil {
		return err
	}

	resp.Revision.Name = cached.Revision.ID.String()

	nhash = checksum(resp)
	if nhash != phash {
		err = srv.Send(resp)
		if err != nil {
			return err
		}
	}
	phash = nhash

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	cached, err = flow.traverseToRef(ctx, nil, cached.Namespace.Name, cached.Path(), req.GetRef())
	if err != nil {
		return err
	}

	goto resend

}

type lookupWorkflowFromParentArgs struct {
	pino *database.Inode
	name string
}

func (flow *flow) lookupWorkflowFromParent(ctx context.Context, tx database.Transaction, args *lookupWorkflowFromParentArgs) (*database.Workflow, error) {

	ino, err := flow.lookupInodeFromParent(ctx, tx, &lookupInodeFromParentArgs{
		pino: args.pino,
		name: args.name,
	})
	if err != nil {
		return nil, err
	}

	wf, err := ino.QueryWorkflow().Only(ctx)
	if err != nil {
		return nil, err
	}

	return entwrapper.EntWorkflow(wf), nil

}

type createWorkflowArgs struct {
	ns         *database.Namespace
	pino       *database.Inode
	path       string
	super      bool
	data       []byte
	noValidate bool
}

func (flow *flow) createWorkflow(ctx context.Context, tx database.Transaction, args *createWorkflowArgs) (*database.Workflow, *database.Inode, error) {

	ns := args.ns
	pino := args.pino
	path := args.path
	dir, base := filepath.Split(args.path)

	data := args.data
	hash, err := computeHash(data)
	if err != nil {
		return nil, nil, err
	}

	if pino.Type != util.InodeTypeDirectory {
		return nil, nil, errors.New("parent inode is not a directory")
	}

	if !args.super && pino.ReadOnly {
		return nil, nil, errors.New("cannot write into read-only directory")
	}

	clients := flow.edb.Clients(tx)

	ino, err := clients.Inode.Query().Where(entino.HasParentWith(entino.ID(pino.ID))).Where(entino.NameEQ(base)).Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, nil, err
	} else if err == nil {
		if ino.Type != util.InodeTypeWorkflow {
			return nil, nil, os.ErrExist
		}
		wf, err := ino.QueryWorkflow().Only(ctx)
		if err != nil {
			return nil, nil, err
		}
		return entwrapper.EntWorkflow(wf), entwrapper.EntInode(ino), os.ErrExist
	}

	ino, err = clients.Inode.Create().SetName(base).SetNamespaceID(ns.ID).SetParentID(pino.ID).SetReadOnly(pino.ReadOnly).SetType(util.InodeTypeWorkflow).Save(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, nil, os.ErrExist
		}
		return nil, nil, err
	}

	wf, err := clients.Workflow.Create().SetInodeID(ino.ID).SetNamespaceID(ns.ID).Save(ctx)
	if err != nil {
		return nil, nil, err
	}

	rev, err := clients.Revision.Create().SetHash(hash).SetSource(data).SetWorkflow(wf).SetMetadata(make(map[string]interface{})).Save(ctx)
	if err != nil {
		return nil, nil, err
	}

	_, err = clients.Ref.Create().SetImmutable(false).SetName(latest).SetWorkflow(wf).SetRevision(rev).Save(ctx)
	if err != nil {
		return nil, nil, err
	}

	_, err = clients.Inode.UpdateOneID(pino.ID).SetUpdatedAt(time.Now()).Save(ctx)
	if err != nil {
		return nil, nil, err
	}

	flags := rcfNoPriors
	if args.noValidate {
		flags |= rcfNoValidate
	}

	cached, err := flow.traverseToRef(ctx, tx, ns.Name, args.path, latest)
	if err != nil {
		return nil, nil, err
	}

	err = flow.configureRouter(ctx, tx, cached, flags,
		func() error {
			return nil
		},
		func() error {
			return nil
		},
		//tx.Commit,
	)
	if err != nil {
		return nil, nil, err
	}

	metricsWf.WithLabelValues(ns.Name, ns.Name).Inc()
	metricsWfUpdated.WithLabelValues(ns.Name, path, ns.Name).Inc()

	flow.logToNamespace(ctx, time.Now(), cached, "Created workflow '%s'.", path)
	flow.pubsub.NotifyInode(cached.Inode())

	err = flow.BroadcastWorkflow(ctx, BroadcastEventTypeCreate,
		broadcastWorkflowInput{
			Name:   base,
			Path:   path,
			Parent: dir,
			Live:   true,
		}, cached)

	if err != nil {
		return nil, nil, err
	}

	return entwrapper.EntWorkflow(wf), entwrapper.EntInode(ino), nil

}

func (flow *flow) CreateWorkflow(ctx context.Context, req *grpc.CreateWorkflowRequest) (*grpc.CreateWorkflowResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	data := req.GetSource()

	hash, err := computeHash(data)
	if err != nil {
		return nil, err
	}

	tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}

	defer rollback(tx)

	path := GetInodePath(req.GetPath())
	dir, base := filepath.Split(path)

	cached, err := flow.traverseToInode(ctx, tx, req.GetNamespace(), dir)
	if err != nil {
		return nil, err
	}

	if cached.Inode().Type != util.InodeTypeDirectory {
		return nil, errors.New("parent inode is not a directory")
	}

	if cached.Inode().ReadOnly {
		return nil, errors.New("cannot write into read-only directory")
	}

	clients := flow.edb.Clients(tx)

	ino, err := clients.Inode.Create().SetName(base).SetNamespaceID(cached.Namespace.ID).SetParentID(cached.Inode().ID).SetType(util.InodeTypeWorkflow).Save(ctx)
	if err != nil {
		return nil, err
	}

	wf, err := clients.Workflow.Create().SetInodeID(ino.ID).SetNamespaceID(cached.Namespace.ID).Save(ctx)
	if err != nil {
		return nil, err
	}

	rev, err := clients.Revision.Create().SetHash(hash).SetSource(data).SetWorkflow(wf).SetMetadata(make(map[string]interface{})).Save(ctx)
	if err != nil {
		return nil, err
	}

	_, err = clients.Ref.Create().SetImmutable(false).SetName(latest).SetWorkflow(wf).SetRevision(rev).Save(ctx)
	if err != nil {
		return nil, err
	}

	_, err = clients.Inode.UpdateOneID(cached.Inode().ID).SetUpdatedAt(time.Now()).Save(ctx)
	if err != nil {
		return nil, err
	}

	err = flow.configureRouter(ctx, tx, cached, rcfNoPriors,
		func() error {
			return nil
		},
		tx.Commit,
	)
	if err != nil {
		return nil, err
	}

	// CREATE HERE

	metricsWf.WithLabelValues(cached.Namespace.Name, cached.Namespace.Name).Inc()
	metricsWfUpdated.WithLabelValues(cached.Namespace.Name, path, cached.Namespace.Name).Inc()

	flow.logToNamespace(ctx, time.Now(), cached, "Created workflow '%s'.", path)
	flow.pubsub.NotifyInode(cached.Inode())

	var resp grpc.CreateWorkflowResponse

	err = atob(ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Namespace = cached.Namespace.Name
	resp.Node.Parent = dir
	resp.Node.Path = path

	err = atob(rev, &resp.Revision)
	if err != nil {
		return nil, err
	}

	resp.Revision.Name = rev.ID.String()

	err = flow.BroadcastWorkflow(ctx, BroadcastEventTypeCreate,
		broadcastWorkflowInput{
			Name:   resp.Node.Name,
			Path:   resp.Node.Path,
			Parent: resp.Node.Parent,
			Live:   true,
		}, cached)

	if err != nil {
		return nil, err
	}

	return &resp, nil

}

type updateWorkflowArgs struct {
	cached     *database.CacheData
	path       string
	super      bool
	data       []byte
	noValidate bool
}

func (flow *flow) updateWorkflow(ctx context.Context, tx database.Transaction, args *updateWorkflowArgs) (*database.Revision, error) {

	data := args.data

	hash, err := computeHash(data)
	if err != nil {
		return nil, err
	}

	if !args.super && args.cached.Inode().ReadOnly {
		return nil, errors.New("cannot write into read-only directory")
	}

	var ref *database.Ref

	for i := range args.cached.Workflow.Refs {
		x := args.cached.Workflow.Refs[i]
		if x.Name == latest {
			ref = x
			break
		}
	}

	args.cached.Ref = ref

	err = flow.database.Revision(ctx, tx, args.cached, ref.Revision)
	if err != nil {
		return nil, err
	}

	oldrev := args.cached.Revision

	var k int
	var rev *database.Revision

	if oldrev.Hash == hash {
		// gracefully abort if hash matches latest
		return oldrev, nil
	}

	// flags := rcfNoPriors
	flags := rcfBreaking
	if args.noValidate {
		flags |= rcfNoValidate
	}

	err = flow.configureRouter(ctx, tx, args.cached, flags,
		func() error {

			clients := flow.edb.Clients(tx)

			x, err := clients.Revision.Create().SetHash(hash).SetSource(data).SetWorkflowID(args.cached.Workflow.ID).SetMetadata(make(map[string]interface{})).Save(ctx)
			if err != nil {
				return err
			}

			rev := entwrapper.EntRevision(x)

			// change latest tag
			err = clients.Ref.UpdateOneID(ref.ID).SetRevisionID(rev.ID).Exec(ctx)
			if err != nil {
				return err
			}

			k, err = clients.Ref.Query().Where(entref.HasRevisionWith(entrev.ID(oldrev.ID))).Count(ctx)
			if err != nil {
				return err
			}

			// ??? if hash matches non-latest

			if k == 0 {
				// delete previous latest if untagged
				err = clients.Revision.DeleteOneID(oldrev.ID).Exec(ctx)
				if err != nil {
					return err
				}
			}

			return nil

		},
		func() error { return nil },
	)
	if err != nil {
		return nil, err
	}

	metricsWfUpdated.WithLabelValues(args.cached.Namespace.Name, args.cached.Path(), args.cached.Namespace.Name).Inc()

	flow.logToWorkflow(ctx, time.Now(), args.cached, "Updated workflow.")
	flow.pubsub.NotifyWorkflow(args.cached.Workflow)

	err = flow.BroadcastWorkflow(ctx, BroadcastEventTypeUpdate,
		broadcastWorkflowInput{
			Name:   args.cached.Inode().Name,
			Path:   args.cached.Path(),
			Parent: args.cached.Dir(),
			Live:   args.cached.Workflow.Live,
		}, args.cached)

	if err != nil {
		return nil, err
	}

	return rev, nil

}

func (flow *flow) UpdateWorkflow(ctx context.Context, req *grpc.UpdateWorkflowRequest) (*grpc.UpdateWorkflowResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, err := flow.traverseToWorkflow(ctx, tx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	rev, err := flow.updateWorkflow(ctx, tx, &updateWorkflowArgs{
		cached: cached,
		path:   cached.Path(),
		data:   req.GetSource(),
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	var resp grpc.UpdateWorkflowResponse

	err = atob(cached.Inode(), &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Namespace = cached.Namespace.Name
	resp.Node.Parent = cached.Dir()
	resp.Node.Path = cached.Path()

	err = atob(rev, &resp.Revision)
	if err != nil {
		return nil, err
	}

	resp.Revision.Name = rev.ID.String()

	return &resp, nil

}

func (flow *flow) SaveHead(ctx context.Context, req *grpc.SaveHeadRequest) (*grpc.SaveHeadResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, err := flow.traverseToRef(ctx, tx, req.GetNamespace(), req.GetPath(), "")
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(tx)

	k, err := clients.Ref.Query().Where(entref.HasRevisionWith(entrev.ID(cached.Revision.ID))).Count(ctx)
	if err != nil {
		return nil, err
	}

	metadata := req.GetMetadata()

	if k > 1 {
		// already saved, gracefully back out
		rollback(tx)
		goto respond
	}

	if len(metadata) != 0 {
		obj := make(map[string]interface{})
		err := unmarshal(string(metadata), &obj)
		if err != nil {
			return nil, err
		}

		_, err = clients.Revision.UpdateOneID(cached.Revision.ID).SetMetadata(obj).Save(ctx)
		if err != nil {
			return nil, err
		}
	}

	err = clients.Ref.Create().SetImmutable(true).SetName(cached.Revision.ID.String()).SetRevisionID(cached.Revision.ID).SetWorkflowID(cached.Workflow.ID).Exec(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToWorkflow(ctx, time.Now(), cached, "Saved workflow: %s.", cached.Revision.ID.String())
	flow.pubsub.NotifyWorkflow(cached.Workflow)

respond:

	var resp grpc.SaveHeadResponse

	err = atob(cached.Inode(), &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Namespace = cached.Namespace.Name
	resp.Node.Parent = cached.Dir()
	resp.Node.Path = cached.Path()

	err = atob(cached.Revision, &resp.Revision)
	if err != nil {
		return nil, err
	}

	resp.Revision.Name = cached.Revision.ID.String()

	return &resp, nil

}

func (flow *flow) DiscardHead(ctx context.Context, req *grpc.DiscardHeadRequest) (*grpc.DiscardHeadResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, err := flow.traverseToRef(ctx, tx, req.GetNamespace(), req.GetPath(), "")
	if err != nil {
		return nil, err
	}

	var prevrev []*ent.Revision

	clients := flow.edb.Clients(tx)

	if len(cached.Workflow.Revisions) == 1 || len(cached.Workflow.Refs) > 1 {
		// already saved, or not discardable, gracefully back out
		rollback(tx)
		goto respond
	}

	prevrev, err = clients.Revision.Query().Where(entrev.HasWorkflowWith(entwf.ID(cached.Workflow.ID))).Order(ent.Desc(entrev.FieldCreatedAt)).Offset(1).Limit(1).All(ctx)
	if err != nil {
		return nil, err
	}

	if len(prevrev) != 1 {
		return nil, errors.New("revisions list returned more than one")
	}

	err = flow.configureRouter(ctx, tx, cached, rcfBreaking,
		func() error {

			err = clients.Ref.UpdateOneID(cached.Ref.ID).SetRevision(prevrev[0]).Exec(ctx)
			if err != nil {
				return err
			}

			err = clients.Revision.DeleteOneID(cached.Revision.ID).Exec(ctx)
			if err != nil {
				return err
			}

			cached.Revision = entwrapper.EntRevision(prevrev[0])

			return nil

		},
		tx.Commit,
	)
	if err != nil {
		return nil, err
	}

	metricsWfUpdated.WithLabelValues(cached.Namespace.Name, cached.Path(), cached.Namespace.Name).Inc()

	flow.logToWorkflow(ctx, time.Now(), cached, "Discard unsaved changes to workflow.")
	flow.pubsub.NotifyWorkflow(cached.Workflow)

respond:

	var resp grpc.DiscardHeadResponse

	err = atob(cached.Inode(), &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Namespace = cached.Namespace.Name
	resp.Node.Parent = cached.Dir()
	resp.Node.Path = cached.Path()

	err = atob(cached.Revision, &resp.Revision)
	if err != nil {
		return nil, err
	}

	resp.Revision.Name = cached.Revision.ID.String()

	return &resp, nil

}

func (flow *flow) ToggleWorkflow(ctx context.Context, req *grpc.ToggleWorkflowRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, err := flow.traverseToWorkflow(ctx, tx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	var resp emptypb.Empty

	if cached.Workflow.Live == req.GetLive() {
		rollback(tx)
		return &resp, nil
	}

	clients := flow.edb.Clients(tx)

	err = flow.configureRouter(ctx, tx, cached, rcfBreaking,
		func() error {
			wf, err := clients.Workflow.UpdateOneID(cached.Workflow.ID).SetLive(req.GetLive()).Save(ctx)
			if err != nil {
				return err
			}
			cached.Workflow.Live = wf.Live
			return nil
		},
		tx.Commit,
	)
	if err != nil {
		return nil, err
	}

	live := "disabled"
	if cached.Workflow.Live {
		live = "enabled"
	}

	err = flow.BroadcastWorkflow(ctx, BroadcastEventTypeUpdate,
		broadcastWorkflowInput{
			Name:   cached.Inode().Name,
			Path:   cached.Path(),
			Parent: cached.Dir(),
			Live:   cached.Workflow.Live,
		}, cached)

	if err != nil {
		return nil, err
	}

	flow.logToWorkflow(ctx, time.Now(), cached, "Workflow is now %s", live)
	flow.pubsub.NotifyWorkflow(cached.Workflow)

	return &resp, nil

}

func (flow *flow) SetWorkflowEventLogging(ctx context.Context, req *grpc.SetWorkflowEventLoggingRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, err := flow.traverseToWorkflow(ctx, tx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(tx)

	_, err = clients.Workflow.UpdateOneID(cached.Workflow.ID).SetLogToEvents(req.GetLogger()).Save(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToWorkflow(ctx, time.Now(), cached, "Workflow now logging to cloudevents: %s", req.GetLogger())
	flow.pubsub.NotifyWorkflow(cached.Workflow)
	var resp emptypb.Empty

	return &resp, nil

}
