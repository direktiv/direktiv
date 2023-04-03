package flow

import (
	"context"
	"time"

	"github.com/google/uuid"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/model"
)

func loadSource(rev *database.Revision) (*model.Workflow, error) {
	workflow := new(model.Workflow)

	err := workflow.Load(rev.Source)
	if err != nil {
		return nil, err
	}

	return workflow, nil
}

func (flow *flow) Tags(ctx context.Context, req *grpc.TagsRequest) (*grpc.TagsResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	fStore, _, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(ctx)

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}
	_, err = fStore.ForFile(file).GetAllRevisions(ctx)
	if err != nil {
		return nil, err
	}

	resp := &grpc.TagsResponse{}
	resp.Namespace = ns.Name
	resp.PageInfo = nil
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	// TODO, yassir, fix here.
	resp.Results = []*grpc.Ref{
		{Name: "latest"},
		{Name: "rev1"},
		{Name: "rev2"},
	}

	if err := commit(ctx); err != nil {
		return nil, err
	}
	return resp, nil
}

func (flow *flow) TagsStream(req *grpc.TagsRequest, srv grpc.Flow_TagsStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	resp, err := flow.Tags(ctx, req)
	if err != nil {
		return err
	}

	// mock streaming response.
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			err = srv.Send(resp)
			if err != nil {
				return err
			}
			time.Sleep(time.Second * 5)
		}
	}
}

func (flow *flow) Refs(ctx context.Context, req *grpc.RefsRequest) (*grpc.RefsResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	fStore, _, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(ctx)

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}
	_, err = fStore.ForFile(file).GetAllRevisions(ctx)
	if err != nil {
		return nil, err
	}

	resp := &grpc.RefsResponse{}
	resp.Namespace = ns.Name
	resp.PageInfo = nil
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	// TODO, yassir, fix here.
	resp.Results = []*grpc.Ref{
		{Name: "latest"},
		{Name: "rev1"},
		{Name: "rev2"},
	}

	if err := commit(ctx); err != nil {
		return nil, err
	}
	return resp, nil
}

func (flow *flow) RefsStream(req *grpc.RefsRequest, srv grpc.Flow_RefsStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	resp, err := flow.Refs(ctx, req)
	if err != nil {
		return err
	}

	// mock streaming response.
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			err = srv.Send(resp)
			if err != nil {
				return err
			}
			time.Sleep(time.Second * 5)
		}
	}
}

func (flow *flow) Tag(ctx context.Context, req *grpc.TagRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	fStore, _, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(ctx)

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}
	revID, err := uuid.Parse(req.GetRef())
	if err != nil {
		return nil, err
	}
	revision, err := fStore.ForFile(file).GetRevision(ctx, revID)
	if err != nil {
		return nil, err
	}
	revision, err = fStore.ForRevision(revision).SetTags(ctx, revision.Tags.AddTag(req.GetTag()))
	if err != nil {
		return nil, err
	}
	if err = commit(ctx); err != nil {
		return nil, err
	}

	// TODO: yassir, fix this.
	// flow.logToWorkflow(ctx, time.Now(), cached, "Tagged workflow: %s -> %s.", req.GetTag(), revision.ID.String())
	// flow.pubsub.NotifyWorkflowID(file.ID)

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) Untag(ctx context.Context, req *grpc.UntagRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	fStore, _, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(ctx)

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}
	revision, err := fStore.ForFile(file).GetRevisionByTag(ctx, req.GetTag())
	if err != nil {
		return nil, err
	}
	revision, err = fStore.ForRevision(revision).SetTags(ctx, revision.Tags.RemoveTag(req.GetTag()))
	if err != nil {
		return nil, err
	}
	if err = commit(ctx); err != nil {
		return nil, err
	}

	// TODO: yassir, fix this.
	// flow.logToWorkflow(ctx, time.Now(), cached, "Deleted workflow tag: %s.", req.GetTag())
	// flow.pubsub.NotifyWorkflowID(file.ID)

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) Retag(ctx context.Context, req *grpc.RetagRequest) (*emptypb.Empty, error) {
	// TODO: yassir, question here.
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) ValidateRef(ctx context.Context, req *grpc.ValidateRefRequest) (*grpc.ValidateRefResponse, error) {
	// TODO: yassir, question here.
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	var resp grpc.ValidateRefResponse

	return &resp, nil
}
