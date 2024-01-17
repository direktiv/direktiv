package flow

import (
	"context"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

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

	rev, err := tx.FileStore().ForFile(file).GetRevision(ctx)
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

	// flow.logger.Infof(ctx, file.ID, cached.GetAttributes(recipient.Workflow), "Deleted workflow revision: %s.", cached.Revision.ID.String())
	// flow.pubsub.NotifyWorkflow(cached.Workflow)

	var resp emptypb.Empty

	return &resp, nil
}
