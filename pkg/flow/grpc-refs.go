package flow

import (
	"context"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"google.golang.org/protobuf/types/known/emptypb"
)

func loadSource(rev *filestore.Revision) (*model.Workflow, error) {
	workflow := new(model.Workflow)

	err := workflow.Load(rev.Data)
	if err != nil {
		return nil, err
	}

	return workflow, nil
}

func (flow *flow) Tags(ctx context.Context, req *grpc.TagsRequest) (*grpc.TagsResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	file, err := tx.FileStore().ForNamespace(ns.Name).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}
	revs, err := tx.FileStore().ForFile(file).GetAllRevisions(ctx)
	if err != nil {
		return nil, err
	}

	tags := []*grpc.Ref{
		{
			Name: filestore.Latest,
		},
	}
	for _, rev := range revs {
		revTags := rev.Tags.List()
		for _, revTag := range revTags {
			tags = append(tags, &grpc.Ref{
				Name: revTag,
			})
		}
	}

	resp := &grpc.TagsResponse{}
	resp.Namespace = ns.Name
	resp.Results = tags
	resp.PageInfo = &grpc.PageInfo{
		Total: int32(len(resp.Results)),
	}
	resp.Node = bytedata.ConvertFileToGrpcNode(file)

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

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	file, err := tx.FileStore().ForNamespace(ns.Name).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}
	revs, err := tx.FileStore().ForFile(file).GetAllRevisions(ctx)
	if err != nil {
		return nil, err
	}

	refs := []*grpc.Ref{
		{
			Name: filestore.Latest,
		},
	}
	for _, rev := range revs {
		revTags := rev.Tags.List()
		for _, revTag := range revTags {
			refs = append(refs, &grpc.Ref{
				Name: revTag,
			})
		}
		if !rev.IsCurrent {
			refs = append(refs, &grpc.Ref{
				Name: rev.ID.String(),
			})
		}
	}

	resp := &grpc.RefsResponse{}
	resp.Namespace = ns.Name
	resp.Results = refs
	resp.PageInfo = &grpc.PageInfo{
		Total: int32(len(resp.Results)),
	}
	resp.Node = bytedata.ConvertFileToGrpcNode(file)

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

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	file, err := tx.FileStore().ForNamespace(ns.Name).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	revision, err := tx.FileStore().ForFile(file).GetRevision(ctx, req.GetRef())
	if err != nil {
		return nil, err
	}
	err = tx.FileStore().ForRevision(revision).SetTags(ctx, revision.Tags.AddTag(req.GetTag()))
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	// flow.logger.Infof(ctx, file.ID, database.GetAttributes(recipient.Workflow, ns, fileAttributes(*file)), "Tagged workflow: %s -> %s.", req.GetTag(), revision.ID.String())

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) Untag(ctx context.Context, req *grpc.UntagRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	file, err := tx.FileStore().ForNamespace(ns.Name).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}
	revision, err := tx.FileStore().ForFile(file).GetRevision(ctx, req.GetTag())
	if err != nil {
		return nil, err
	}
	err = tx.FileStore().ForRevision(revision).SetTags(ctx, revision.Tags.RemoveTag(req.GetTag()))
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	// flow.logger.Infof(ctx, file.ID, database.GetAttributes(recipient.Workflow, ns, fileAttributes(*file)), "Deleted workflow tag: %s.", req.GetTag())

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) Retag(ctx context.Context, req *grpc.RetagRequest) (*emptypb.Empty, error) {
	// TODO: yassir, low priority. we might remove this feature.
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) ValidateRef(ctx context.Context, req *grpc.ValidateRefRequest) (*grpc.ValidateRefResponse, error) {
	// TODO: yassir, low priority.
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	var resp grpc.ValidateRefResponse

	return &resp, nil
}
