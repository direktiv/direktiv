package flow

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (flow *flow) WorkflowVariable(ctx context.Context, req *grpc.WorkflowVariableRequest) (*grpc.WorkflowVariableResponse, error) {
	slog.Debug("Handling gRPC request", "this", this())

	tx, err := flow.beginSQLTx(ctx)
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

	item, err := tx.DataStore().RuntimeVariables().GetForWorkflow(ctx, ns.Name, file.Path, req.GetKey())
	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			t := time.Now()

			return &grpc.WorkflowVariableResponse{
				Namespace: ns.Name,
				Path:      file.Path,
				Key:       req.GetKey(),
				CreatedAt: timestamppb.New(t),
				UpdatedAt: timestamppb.New(t),
				TotalSize: int64(0),
				MimeType:  "",
				Data:      make([]byte, 0),
			}, nil
		}

		return nil, err
	}

	var resp grpc.WorkflowVariableResponse

	resp.Namespace = ns.Name
	resp.Path = file.Path
	resp.Key = item.Name
	resp.CreatedAt = timestamppb.New(item.CreatedAt)
	resp.UpdatedAt = timestamppb.New(item.UpdatedAt)
	resp.TotalSize = int64(item.Size)
	resp.MimeType = item.MimeType

	if resp.GetTotalSize() > parcelSize {
		return nil, status.Error(codes.ResourceExhausted, "variable too large to return without using the parcelling API")
	}
	data, err := tx.DataStore().RuntimeVariables().LoadData(ctx, item.ID)
	if err != nil {
		return nil, err
	}

	resp.Data = data

	return &resp, nil
}

func (flow *flow) SetWorkflowVariable(ctx context.Context, req *grpc.SetWorkflowVariableRequest) (*grpc.SetWorkflowVariableResponse, error) {
	slog.Debug("Handling gRPC request", "this", this())

	tx, err := flow.beginSQLTx(ctx)
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

	newVar, err := tx.DataStore().RuntimeVariables().Set(ctx, &datastore.RuntimeVariable{
		Namespace:    ns.Name,
		WorkflowPath: file.Path,
		Name:         req.GetKey(),
		Data:         req.GetData(),
		MimeType:     req.GetMimeType(),
	})
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	// TODO: need fix here.
	// flow.logger.Infof(ctx, file.ID, database.GetAttributes(recipient.Workflow, ns, fileAttributes(*file)), "Set workflow variable '%s'.", key)
	// flow.pubsub.NotifyWorkflowVariables(file.ID)

	var resp grpc.SetWorkflowVariableResponse

	resp.Namespace = ns.Name
	resp.Path = file.Path
	resp.Key = newVar.Name
	resp.CreatedAt = timestamppb.New(newVar.CreatedAt)
	resp.UpdatedAt = timestamppb.New(newVar.UpdatedAt)
	resp.TotalSize = int64(newVar.Size)
	resp.MimeType = newVar.MimeType

	return &resp, nil
}
