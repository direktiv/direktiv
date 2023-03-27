// TODO: yassir, need refactor.
package flow

import (
	"context"
	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (flow *flow) Tags(ctx context.Context, req *grpc.TagsRequest) (*grpc.TagsResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	fStore, err := flow.fStore.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer fStore.Rollback(ctx)

	file, err := flow.fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}
	_, err = flow.fStore.ForFile(file).GetAllRevisions(ctx)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.TagsResponse)
	resp.Namespace = ns.Name
	resp.PageInfo = nil
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	//TODO, yassir, fix here.
	resp.Results = []*grpc.Ref{
		{Name: "latest"},
		{Name: "rev1"},
		{Name: "rev2"},
	}

	if err = fStore.Commit(ctx); err != nil {
		return nil, err
	}
	return resp, nil
}

func (flow *flow) Refs(ctx context.Context, req *grpc.RefsRequest) (*grpc.RefsResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	file, err := flow.fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}
	_, err = flow.fStore.ForFile(file).GetAllRevisions(ctx)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.RefsResponse)
	resp.Namespace = ns.Name
	resp.PageInfo = nil
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	//TODO, yassir, fix here.
	resp.Results = []*grpc.Ref{
		{Name: "latest"},
		{Name: "rev1"},
		{Name: "rev2"},
	}

	return resp, nil
}

func (flow *flow) Tag(ctx context.Context, req *grpc.TagRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	file, err := flow.fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}
	revID, err := uuid.Parse(req.GetRef())
	if err != nil {
		return nil, err
	}
	revision, err := flow.fStore.ForFile(file).GetRevision(ctx, revID)
	if err != nil {
		return nil, err
	}
	revision, err = flow.fStore.ForRevision(revision).SetTags(ctx, revision.Tags.AddTag(req.GetTag()))
	if err != nil {
		return nil, err
	}

	//TODO: yassir, fix this.
	//flow.logToWorkflow(ctx, time.Now(), cached, "Tagged workflow: %s -> %s.", req.GetTag(), revision.ID.String())
	flow.pubsub.NotifyWorkflowID(file.ID)

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) Untag(ctx context.Context, req *grpc.UntagRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	file, err := flow.fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}
	revision, err := flow.fStore.ForFile(file).GetRevisionByTag(ctx, req.GetTag())
	if err != nil {
		return nil, err
	}
	revision, err = flow.fStore.ForRevision(revision).SetTags(ctx, revision.Tags.RemoveTag(req.GetTag()))
	if err != nil {
		return nil, err
	}

	//TODO: yassir, fix this.
	//flow.logToWorkflow(ctx, time.Now(), cached, "Deleted workflow tag: %s.", req.GetTag())
	flow.pubsub.NotifyWorkflowID(file.ID)

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) Retag(ctx context.Context, req *grpc.RetagRequest) (*emptypb.Empty, error) {
	panic("")
}
func (flow *flow) ValidateRef(ctx context.Context, req *grpc.ValidateRefRequest) (*grpc.ValidateRefResponse, error) {
	panic("")
}
