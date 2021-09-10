package flow

import (
	"context"
	"path/filepath"
	"time"

	"github.com/vorteil/direktiv/pkg/flow/ent"
	entrev "github.com/vorteil/direktiv/pkg/flow/ent/revision"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
)

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

	resp.Namespace = d.namespace()
	resp.Node.Parent = d.dir
	resp.Node.Path = d.path

	err = atob(d.rev(), &resp.Revision)
	if err != nil {
		return nil, err
	}

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

	resp.Namespace = d.namespace()
	resp.Node.Parent = d.dir
	resp.Node.Path = d.path

	err = atob(d.rev(), &resp.Revision)
	if err != nil {
		return err
	}

	nhash = checksum(resp)
	if nhash != phash {
		err = srv.Send(resp)
		if err != nil {
			return err
		}
	}
	phash = nhash

	more := sub.Wait()
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
	path := getInodePath(req.GetPath())
	dir, base := filepath.Split(path)
	d, err := flow.traverseToInode(ctx, nsc, req.GetNamespace(), dir)
	if err != nil {
		return nil, err
	}

	inoc := tx.Inode

	ino, err := inoc.Create().SetName(base).SetNamespace(d.ns()).SetParent(d.ino).SetType("workflow").Save(ctx)
	if err != nil {
		return nil, err
	}

	wfc := tx.Workflow

	wf, err := wfc.Create().SetInode(ino).SetNamespace(d.ns()).Save(ctx)
	if err != nil {
		return nil, err
	}

	revc := tx.Revision

	rev, err := revc.Create().SetHash(hash).SetSource(data).SetWorkflow(wf).Save(ctx)
	if err != nil {
		return nil, err
	}

	refc := tx.Ref

	_, err = refc.Create().SetImmutable(false).SetName(latest).SetWorkflow(wf).SetRevision(rev).Save(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToNamespace(ctx, time.Now(), d.ns(), "Created workflow '%s'.", path)
	flow.pubsub.NotifyInode(d.ino)

	var resp grpc.CreateWorkflowResponse

	err = atob(ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	resp.Namespace = d.namespace()
	resp.Node.Parent = dir
	resp.Node.Path = path

	err = atob(rev, &resp.Revision)
	if err != nil {
		return nil, err
	}

	return &resp, nil

}

func (flow *flow) UpdateWorkflow(ctx context.Context, req *grpc.UpdateWorkflowRequest) (*grpc.UpdateWorkflowResponse, error) {

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
	d, err := flow.traverseToRef(ctx, nsc, req.GetNamespace(), req.GetPath(), "")
	if err != nil {
		return nil, err
	}

	oldrev := d.rev()

	var k int
	var rev *ent.Revision
	revc := tx.Revision

	if oldrev.Hash == hash {
		// gracefully abort if hash matches latest
		rollback(tx)
		rev = oldrev
		goto respond
	}

	rev, err = revc.Create().SetHash(hash).SetSource(data).SetWorkflow(d.wf).Save(ctx)
	if err != nil {
		return nil, err
	}

	// change latest tag
	err = d.ref.Update().SetRevision(rev).Exec(ctx)
	if err != nil {
		return nil, err
	}

	k, err = oldrev.QueryRefs().Count(ctx)
	if err != nil {
		return nil, err
	}

	// ??? if hash matches non-latest

	if k == 0 {
		// delete previous latest if untagged
		err = revc.DeleteOne(oldrev).Exec(ctx)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToWorkflow(ctx, time.Now(), d.wf, "Updated workflow.")
	flow.pubsub.NotifyWorkflow(d.wf)

respond:

	var resp grpc.UpdateWorkflowResponse

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	resp.Namespace = d.namespace()
	resp.Node.Parent = d.dir
	resp.Node.Path = d.path

	err = atob(rev, &resp.Revision)
	if err != nil {
		return nil, err
	}

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

	if k > 1 {
		// already saved, gracefully back out
		rollback(tx)
		goto respond
	}

	err = refc.Create().SetImmutable(true).SetName(d.rev().ID.String()).SetRevision(d.rev()).SetWorkflow(d.wf).Exec(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToWorkflow(ctx, time.Now(), d.wf, "Saved workflow: %s.", d.rev().ID.String())
	flow.pubsub.NotifyWorkflow(d.wf)

respond:

	var resp grpc.SaveHeadResponse

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	resp.Namespace = d.namespace()
	resp.Node.Parent = d.dir
	resp.Node.Path = d.path

	err = atob(d.rev(), &resp.Revision)
	if err != nil {
		return nil, err
	}

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
	var rev, prevrev *ent.Revision

	if revcount == 1 || refcount > 1 {
		// already saved, or not discardable, gracefully back out
		rollback(tx)
		goto respond
	}

	prevrev, err = d.wf.QueryRevisions().Order(ent.Desc(entrev.FieldCreatedAt)).Offset(1).Limit(1).Only(ctx)
	if err != nil {
		return nil, err
	}

	err = d.ref.Update().SetRevision(prevrev).Exec(ctx)
	if err != nil {
		return nil, err
	}

	rev = d.rev()
	err = revc.DeleteOne(rev).Exec(ctx)
	if err != nil {
		return nil, err
	}

	rev = prevrev

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToWorkflow(ctx, time.Now(), d.wf, "Discard unsaved changes to workflow.")
	flow.pubsub.NotifyWorkflow(d.wf)

respond:

	var resp grpc.DiscardHeadResponse

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	resp.Namespace = d.namespace()
	resp.Node.Parent = d.dir
	resp.Node.Path = d.path

	err = atob(rev, &resp.Revision)
	if err != nil {
		return nil, err
	}

	return &resp, nil

}
