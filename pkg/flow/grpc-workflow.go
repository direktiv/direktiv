package flow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	entino "github.com/direktiv/direktiv/pkg/flow/ent/inode"
	entrev "github.com/direktiv/direktiv/pkg/flow/ent/revision"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (flow *flow) ResolveWorkflowUID(ctx context.Context, req *grpc.ResolveWorkflowUIDRequest) (*grpc.WorkflowResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	d, err := flow.reverseTraverseToWorkflow(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	var resp grpc.WorkflowResponse

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Namespace = d.namespace()
	resp.Node.Parent = d.dir
	resp.Node.Path = d.path
	resp.Oid = d.wf.ID.String()

	// resp.EventLogging = d.wf.LogToEvents
	//
	// err = atob(d.rev(), &resp.Revision)
	// if err != nil {
	// 	return nil, err
	// }

	return &resp, nil

}

func (flow *flow) Workflow(ctx context.Context, req *grpc.WorkflowRequest) (*grpc.WorkflowResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	nsc := flow.db.Namespace
	d, err := flow.traverseToRef(ctx, nsc, req.GetNamespace(), req.GetPath(), req.GetRef())
	if err != nil {
		return nil, err
	}

	var resp grpc.WorkflowResponse

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Namespace = d.namespace()
	resp.Node.Parent = d.dir
	resp.Node.Path = d.path
	resp.EventLogging = d.wf.LogToEvents
	resp.Oid = d.wf.ID.String()

	err = atob(d.rev(), &resp.Revision)
	if err != nil {
		return nil, err
	}

	resp.Revision.Name = d.rev().ID.String()

	return &resp, nil

}

func (flow *flow) WorkflowStream(req *grpc.WorkflowRequest, srv grpc.Flow_WorkflowStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	nsc := flow.db.Namespace
	d, err := flow.traverseToRef(ctx, nsc, req.GetNamespace(), req.GetPath(), req.GetRef())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeWorkflow(d.wf)
	defer flow.cleanup(sub.Close)

resend:

	resp := new(grpc.WorkflowResponse)

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Namespace = d.namespace()
	resp.Node.Parent = d.dir
	resp.Node.Path = d.path
	resp.Oid = d.wf.ID.String()
	resp.EventLogging = d.wf.LogToEvents

	err = atob(d.rev(), &resp.Revision)
	if err != nil {
		return err
	}

	resp.Revision.Name = d.rev().ID.String()

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

	ref, err := flow.getRef(ctx, d.wf, req.GetRef())
	if err != nil {
		return err
	}

	d.ref = ref

	goto resend

}

type lookupWorkflowFromParentArgs struct {
	pino *ent.Inode
	name string
}

func (flow *flow) lookupWorkflowFromParent(ctx context.Context, args *lookupWorkflowFromParentArgs) (*ent.Workflow, error) {

	ino, err := flow.lookupInodeFromParent(ctx, &lookupInodeFromParentArgs{
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

	wf.Edges.Inode = ino

	return wf, nil

}

type createWorkflowArgs struct {
	inoc *ent.InodeClient
	wfc  *ent.WorkflowClient
	revc *ent.RevisionClient
	refc *ent.RefClient

	ns    *ent.Namespace
	pino  *ent.Inode
	path  string
	super bool
	data  []byte
}

func (flow *flow) createWorkflow(ctx context.Context, args *createWorkflowArgs) (*ent.Workflow, error) {

	inoc := args.inoc
	wfc := args.wfc
	revc := args.revc
	refc := args.refc

	ns := args.ns
	pino := args.pino
	path := args.path
	dir, base := filepath.Split(args.path)

	data := args.data
	hash, err := computeHash(data)
	if err != nil {
		return nil, err
	}

	if pino.Type != util.InodeTypeDirectory {
		return nil, errors.New("parent inode is not a directory")
	}

	if !args.super && pino.ReadOnly {
		return nil, errors.New("cannot write into read-only directory")
	}

	ino, err := pino.QueryChildren().Where(entino.NameEQ(base)).Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	} else if err == nil {
		if ino.Type != util.InodeTypeWorkflow {
			return nil, os.ErrExist
		}
		wf, err := ino.QueryWorkflow().Only(ctx)
		if err != nil {
			return nil, err
		}
		wf.Edges.Inode = ino
		return wf, os.ErrExist
	}

	ino, err = inoc.Create().SetName(base).SetNamespace(ns).SetParent(pino).SetReadOnly(pino.ReadOnly).SetType(util.InodeTypeWorkflow).Save(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, os.ErrExist
		}
		return nil, err
	}

	wf, err := wfc.Create().SetInode(ino).SetNamespace(ns).Save(ctx)
	if err != nil {
		return nil, err
	}

	rev, err := revc.Create().SetHash(hash).SetSource(data).SetWorkflow(wf).SetMetadata(make(map[string]interface{})).Save(ctx)
	if err != nil {
		return nil, err
	}

	_, err = refc.Create().SetImmutable(false).SetName(latest).SetWorkflow(wf).SetRevision(rev).Save(ctx)
	if err != nil {
		return nil, err
	}

	_, err = pino.Update().SetUpdatedAt(time.Now()).Save(ctx)
	if err != nil {
		return nil, err
	}

	// TODO?
	// err = flow.configureRouter(ctx, tx.Events, &wf, rcfNoPriors,
	// 	func() error {
	// 		return nil
	// 	},
	// 	tx.Commit,
	// )
	// if err != nil {
	// 	return nil, err
	// }

	metricsWf.WithLabelValues(ns.Name, ns.Name).Inc()
	metricsWfUpdated.WithLabelValues(ns.Name, path, ns.Name).Inc()

	flow.logToNamespace(ctx, time.Now(), ns, "Created workflow '%s'.", path)
	flow.pubsub.NotifyInode(ino)

	err = flow.BroadcastWorkflow(ctx, BroadcastEventTypeCreate,
		broadcastWorkflowInput{
			Name:   base,
			Path:   path,
			Parent: dir,
			Live:   true,
		}, ns)

	if err != nil {
		return nil, err
	}

	wf.Edges.Inode = ino

	return wf, nil

}

func (flow *flow) CreateWorkflow(ctx context.Context, req *grpc.CreateWorkflowRequest) (*grpc.CreateWorkflowResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	data := req.GetSource()

	hash, err := computeHash(data)
	if err != nil {
		return nil, err
	}

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}

	defer rollback(tx)

	nsc := tx.Namespace
	path := GetInodePath(req.GetPath())
	dir, base := filepath.Split(path)
	d, err := flow.traverseToInode(ctx, nsc, req.GetNamespace(), dir)
	if err != nil {
		return nil, err
	}

	if d.ino.Type != util.InodeTypeDirectory {
		return nil, errors.New("parent inode is not a directory")
	}

	if d.ino.ReadOnly {
		return nil, errors.New("cannot write into read-only directory")
	}

	inoc := tx.Inode

	ino, err := inoc.Create().SetName(base).SetNamespace(d.ns()).SetParent(d.ino).SetType(util.InodeTypeWorkflow).Save(ctx)
	if err != nil {
		return nil, err
	}

	wfc := tx.Workflow

	wf, err := wfc.Create().SetInode(ino).SetNamespace(d.ns()).Save(ctx)
	if err != nil {
		return nil, err
	}

	revc := tx.Revision

	rev, err := revc.Create().SetHash(hash).SetSource(data).SetWorkflow(wf).SetMetadata(make(map[string]interface{})).Save(ctx)
	if err != nil {
		return nil, err
	}

	refc := tx.Ref

	_, err = refc.Create().SetImmutable(false).SetName(latest).SetWorkflow(wf).SetRevision(rev).Save(ctx)
	if err != nil {
		return nil, err
	}

	_, err = d.ino.Update().SetUpdatedAt(time.Now()).Save(ctx)
	if err != nil {
		return nil, err
	}

	err = flow.configureRouter(ctx, tx.Events, &wf, rcfNoPriors,
		func() error {
			return nil
		},
		tx.Commit,
	)
	if err != nil {
		return nil, err
	}

	// CREATE HERE

	metricsWf.WithLabelValues(d.ns().Name, d.ns().Name).Inc()
	metricsWfUpdated.WithLabelValues(d.ns().Name, path, d.ns().Name).Inc()

	flow.logToNamespace(ctx, time.Now(), d.ns(), "Created workflow '%s'.", path)
	flow.pubsub.NotifyInode(d.ino)

	var resp grpc.CreateWorkflowResponse

	err = atob(ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Namespace = d.namespace()
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
		}, d.ns())

	if err != nil {
		return nil, err
	}

	return &resp, nil

}

type updateWorkflowArgs struct {
	revc   *ent.RevisionClient
	eventc *ent.EventsClient

	ns    *ent.Namespace
	ino   *ent.Inode
	wf    *ent.Workflow
	path  string
	super bool
	data  []byte
}

func (flow *flow) updateWorkflow(ctx context.Context, args *updateWorkflowArgs) (*ent.Revision, error) {

	data := args.data

	hash, err := computeHash(data)
	if err != nil {
		return nil, err
	}

	if !args.super && args.ino.ReadOnly {
		return nil, errors.New("cannot write into read-only directory")
	}

	path := GetInodePath(args.path)
	dir, base := filepath.Split(path)
	ns := args.ns
	wf := args.wf
	ino := args.ino

	revc := args.revc

	ref, err := flow.lookupRefAndRev(ctx, &lookupRefAndRevArgs{
		wf:        wf,
		reference: "",
	})
	if err != nil {
		return nil, err
	}

	oldrev := ref.Edges.Revision

	var k int
	var rev *ent.Revision

	if oldrev.Hash == hash {
		// gracefully abort if hash matches latest
		return oldrev, nil
	}

	err = flow.configureRouter(ctx, args.eventc, &wf, rcfBreaking,
		func() error {

			rev, err = revc.Create().SetHash(hash).SetSource(data).SetWorkflow(wf).SetMetadata(make(map[string]interface{})).Save(ctx)
			if err != nil {
				return err
			}

			// change latest tag
			err = ref.Update().SetRevision(rev).Exec(ctx)
			if err != nil {
				return err
			}

			k, err = oldrev.QueryRefs().Count(ctx)
			if err != nil {
				return err
			}

			// ??? if hash matches non-latest

			if k == 0 {
				// delete previous latest if untagged
				err = revc.DeleteOne(oldrev).Exec(ctx)
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

	metricsWfUpdated.WithLabelValues(ns.Name, path, ns.Name).Inc()

	d := new(wfData)
	d.nodeData = new(nodeData)
	d.path = path
	d.base = base
	d.dir = dir
	d.ino = ino
	d.wf = wf

	flow.logToWorkflow(ctx, time.Now(), d, "Updated workflow.")
	flow.pubsub.NotifyWorkflow(wf)

	err = flow.BroadcastWorkflow(ctx, BroadcastEventTypeUpdate,
		broadcastWorkflowInput{
			Name:   base,
			Path:   path,
			Parent: dir,
			Live:   wf.Live,
		}, ns)

	if err != nil {
		return nil, err
	}

	return rev, nil

}

func (flow *flow) UpdateWorkflow(ctx context.Context, req *grpc.UpdateWorkflowRequest) (*grpc.UpdateWorkflowResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	d, err := flow.traverseToWorkflow(ctx, tx.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	rev, err := flow.updateWorkflow(ctx, &updateWorkflowArgs{
		revc:   tx.Revision,
		eventc: tx.Events,
		ns:     d.ns(),
		ino:    d.ino,
		wf:     d.wf,
		path:   d.path,
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

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Namespace = d.namespace()
	resp.Node.Parent = d.dir
	resp.Node.Path = d.path

	err = atob(rev, &resp.Revision)
	if err != nil {
		return nil, err
	}

	resp.Revision.Name = rev.ID.String()

	return &resp, nil

}

func (flow *flow) SaveHead(ctx context.Context, req *grpc.SaveHeadRequest) (*grpc.SaveHeadResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	d, err := flow.traverseToRef(ctx, nsc, req.GetNamespace(), req.GetPath(), "")
	if err != nil {
		return nil, err
	}

	k, err := d.rev().QueryRefs().Count(ctx)
	if err != nil {
		return nil, err
	}

	refc := tx.Ref

	metadata := req.GetMetadata()

	if k > 1 {
		// already saved, gracefully back out
		rollback(tx)
		goto respond
	}

	if metadata != nil && len(metadata) != 0 {
		obj := make(map[string]interface{})
		err := unmarshal(string(metadata), &obj)
		if err != nil {
			return nil, err
		}

		_, err = d.rev().Update().SetMetadata(obj).Save(ctx)
		if err != nil {
			return nil, err
		}
	}

	err = refc.Create().SetImmutable(true).SetName(d.rev().ID.String()).SetRevision(d.rev()).SetWorkflow(d.wf).Exec(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToWorkflow(ctx, time.Now(), d.wfData, "Saved workflow: %s.", d.rev().ID.String())
	flow.pubsub.NotifyWorkflow(d.wf)

respond:

	var resp grpc.SaveHeadResponse

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Namespace = d.namespace()
	resp.Node.Parent = d.dir
	resp.Node.Path = d.path

	err = atob(d.rev(), &resp.Revision)
	if err != nil {
		return nil, err
	}

	resp.Revision.Name = d.rev().ID.String()

	return &resp, nil

}

func (flow *flow) DiscardHead(ctx context.Context, req *grpc.DiscardHeadRequest) (*grpc.DiscardHeadResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	d, err := flow.traverseToRef(ctx, nsc, req.GetNamespace(), req.GetPath(), "")
	if err != nil {
		return nil, err
	}

	revcount, err := d.wf.QueryRevisions().Count(ctx)
	if err != nil {
		return nil, err
	}

	refcount, err := d.rev().QueryRefs().Count(ctx)
	if err != nil {
		return nil, err
	}

	revc := tx.Revision
	var rev *ent.Revision
	var prevrev []*ent.Revision

	rev = d.rev()

	if revcount == 1 || refcount > 1 {
		// already saved, or not discardable, gracefully back out
		rollback(tx)
		goto respond
	}

	prevrev, err = d.wf.QueryRevisions().Order(ent.Desc(entrev.FieldCreatedAt)).Offset(1).Limit(1).All(ctx)
	if err != nil {
		return nil, err
	}

	if len(prevrev) != 1 {
		return nil, errors.New("revisions list returned more than one")
	}

	err = flow.configureRouter(ctx, tx.Events, &d.wf, rcfBreaking,
		func() error {

			err = d.ref.Update().SetRevision(prevrev[0]).Exec(ctx)
			if err != nil {
				return err
			}

			rev = d.rev()
			err = revc.DeleteOne(rev).Exec(ctx)
			if err != nil {
				return err
			}

			rev = prevrev[0]

			return nil

		},
		tx.Commit,
	)
	if err != nil {
		return nil, err
	}

	metricsWfUpdated.WithLabelValues(d.ns().Name, d.path, d.ns().Name).Inc()

	flow.logToWorkflow(ctx, time.Now(), d.wfData, "Discard unsaved changes to workflow.")
	flow.pubsub.NotifyWorkflow(d.wf)

respond:

	var resp grpc.DiscardHeadResponse

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Namespace = d.namespace()
	resp.Node.Parent = d.dir
	resp.Node.Path = d.path

	err = atob(rev, &resp.Revision)
	if err != nil {
		return nil, err
	}

	resp.Revision.Name = rev.ID.String()

	return &resp, nil

}

func (flow *flow) ToggleWorkflow(ctx context.Context, req *grpc.ToggleWorkflowRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	d, err := flow.traverseToWorkflow(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	var resp emptypb.Empty

	if d.wf.Live == req.GetLive() {
		rollback(tx)
		return &resp, nil
	}

	wfr := &d.wf

	err = flow.configureRouter(ctx, tx.Events, wfr, rcfBreaking,
		func() error {
			edges := (*wfr).Edges
			wf, err := (*wfr).Update().SetLive(req.GetLive()).Save(ctx)
			if err != nil {
				return err
			}
			wf.Edges = edges
			(*wfr) = wf

			return nil

		},
		tx.Commit,
	)
	if err != nil {
		return nil, err
	}

	live := "disabled"
	if d.wf.Live {
		live = "enabled"
	}

	err = flow.BroadcastWorkflow(ctx, BroadcastEventTypeUpdate,
		broadcastWorkflowInput{
			Name:   d.base,
			Path:   d.path,
			Parent: d.dir,
			Live:   d.wf.Live,
		}, d.ns())

	if err != nil {
		return nil, err
	}

	flow.logToWorkflow(ctx, time.Now(), d, "Workflow is now %s", live)
	flow.pubsub.NotifyWorkflow(d.wf)

	return &resp, nil

}

func (flow *flow) SetWorkflowEventLogging(ctx context.Context, req *grpc.SetWorkflowEventLoggingRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace

	d, err := flow.traverseToWorkflow(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	_, err = d.wf.Update().SetLogToEvents(req.GetLogger()).Save(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToWorkflow(ctx, time.Now(), d, "Workflow now logging to cloudevents: %s", req.GetLogger())
	flow.pubsub.NotifyWorkflow(d.wf)
	var resp emptypb.Empty

	return &resp, nil

}
