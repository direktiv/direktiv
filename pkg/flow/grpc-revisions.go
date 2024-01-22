package flow

import (
	"context"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (flow *flow) Revisions(ctx context.Context, req *grpc.RevisionsRequest) (*grpc.RevisionsResponse, error) {
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

	revisions := []*grpc.Ref{}
	for idx, rev := range revs {
		if idx > 0 {
			revisions = append(revisions, &grpc.Ref{
				Name: rev.ID.String(),
			})
		}
	}

	resp := new(grpc.RevisionsResponse)
	resp.Namespace = ns.Name
	resp.Results = revisions
	resp.PageInfo = &grpc.PageInfo{
		Total: int32(len(resp.Results)),
	}
	resp.Node = bytedata.ConvertFileToGrpcNode(file)

	return resp, nil
}

func (flow *flow) RevisionsStream(req *grpc.RevisionsRequest, srv grpc.Flow_RevisionsStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	resp, err := flow.Revisions(ctx, req)
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

func (flow *flow) DeleteRevision(ctx context.Context, req *grpc.DeleteRevisionRequest) (*emptypb.Empty, error) {
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

	rev, err := tx.FileStore().ForFile(file).GetRevision(ctx, req.GetRevision())
	if err != nil {
		return nil, err
	}

	err = tx.FileStore().ForRevision(rev).Delete(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	// // flow.logger.Infof(ctx, file.ID, cached.GetAttributes(recipient.Workflow), "Deleted workflow revision: %s.", cached.Revision.ID.String())
	// flow.pubsub.NotifyWorkflow(cached.Workflow)

	var resp emptypb.Empty

	return &resp, nil
}
