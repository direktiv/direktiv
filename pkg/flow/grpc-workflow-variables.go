package flow

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	libengine "github.com/direktiv/direktiv/pkg/refactor/engine"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (internal *internal) WorkflowVariableParcels(req *grpc.VariableInternalRequest, srv grpc.Internal_WorkflowVariableParcelsServer) error {
	slog.Debug("Handling gRPC request", "this", this())

	ctx := srv.Context()

	inst, err := internal.getInstance(ctx, req.GetInstance())
	if err != nil {
		return err
	}

	resp, err := internal.flow.WorkflowVariable(ctx, &grpc.WorkflowVariableRequest{
		Namespace: inst.TelemetryInfo.NamespaceName,
		Path:      inst.Instance.WorkflowPath,
		Key:       req.GetKey(),
	})
	if err != nil {
		return err
	}

	iresp := &grpc.VariableInternalResponse{
		Instance:  inst.Instance.ID.String(),
		Key:       resp.GetKey(),
		CreatedAt: resp.GetCreatedAt(),
		UpdatedAt: resp.GetUpdatedAt(),
		Checksum:  resp.GetChecksum(),
		TotalSize: resp.GetTotalSize(),
		Data:      resp.GetData(),
		MimeType:  resp.GetMimeType(),
	}

	err = srv.Send(iresp)
	if err != nil {
		return err
	}

	return nil
}

type setWorkflowVariableParcelsTranslator struct {
	internal *internal
	inst     *libengine.Instance
	grpc.Internal_SetWorkflowVariableParcelsServer
}

func (srv *setWorkflowVariableParcelsTranslator) SendAndClose(resp *grpc.SetWorkflowVariableResponse) error {
	var inst string
	if srv.inst != nil {
		inst = srv.inst.Instance.ID.String()
	}

	return srv.Internal_SetWorkflowVariableParcelsServer.SendAndClose(&grpc.SetVariableInternalResponse{
		Instance:  inst,
		Key:       resp.GetKey(),
		CreatedAt: resp.GetCreatedAt(),
		UpdatedAt: resp.GetUpdatedAt(),
		Checksum:  resp.GetChecksum(),
		TotalSize: resp.GetTotalSize(),
		MimeType:  resp.GetMimeType(),
	})
}

func (srv *setWorkflowVariableParcelsTranslator) Recv() (*grpc.SetWorkflowVariableRequest, error) {
	req, err := srv.Internal_SetWorkflowVariableParcelsServer.Recv()
	if err != nil {
		return nil, err
	}

	if srv.inst == nil {
		ctx := srv.Context()

		srv.inst, err = srv.internal.getInstance(ctx, req.GetInstance())
		if err != nil {
			return nil, err
		}
	}

	return &grpc.SetWorkflowVariableRequest{
		Namespace: srv.inst.TelemetryInfo.NamespaceName,
		Path:      srv.inst.Instance.WorkflowPath,
		Key:       req.GetKey(),
		TotalSize: req.GetTotalSize(),
		Data:      req.GetData(),
		MimeType:  req.GetMimeType(),
	}, nil
}

func (internal *internal) SetWorkflowVariableParcels(srv grpc.Internal_SetWorkflowVariableParcelsServer) error {
	slog.Debug("Handling gRPC request", "this", this())

	fsrv := &setWorkflowVariableParcelsTranslator{
		internal: internal,
		Internal_SetWorkflowVariableParcelsServer: srv,
	}

	err := internal.flow.SetWorkflowVariableParcels(fsrv)
	if err != nil {
		return err
	}

	return nil
}

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

//nolint:dupl
func (flow *flow) SetWorkflowVariableParcels(srv grpc.Flow_SetWorkflowVariableParcelsServer) error {
	slog.Debug("Handling gRPC request", "this", this())
	ctx := srv.Context()

	req, err := srv.Recv()
	if err != nil {
		return err
	}

	firstReq := req

	totalSize := int(req.GetTotalSize())

	buf := new(bytes.Buffer)

	for {
		_, err = io.Copy(buf, bytes.NewReader(req.GetData()))
		if err != nil {
			return err
		}

		if req.GetTotalSize() <= 0 {
			if buf.Len() >= totalSize {
				break
			}
		}

		req, err = srv.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return err
		}

		if req.GetTotalSize() <= 0 {
			if buf.Len() >= totalSize {
				break
			}
		} else {
			if req == nil {
				break
			}
		}

		if int(req.GetTotalSize()) != totalSize {
			return errors.New("totalSize changed mid stream")
		}
	}

	if buf.Len() > totalSize {
		return errors.New("received more data than expected")
	}

	firstReq.Data = buf.Bytes()
	resp, err := flow.SetWorkflowVariable(ctx, firstReq)
	if err != nil {
		return err
	}
	err = srv.SendAndClose(resp)
	if err != nil {
		return err
	}

	return nil
}

func (flow *flow) DeleteWorkflowVariable(ctx context.Context, req *grpc.DeleteWorkflowVariableRequest) (*emptypb.Empty, error) {
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
		return nil, err
	}

	err = tx.DataStore().RuntimeVariables().Delete(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) RenameWorkflowVariable(ctx context.Context, req *grpc.RenameWorkflowVariableRequest) (*grpc.RenameWorkflowVariableResponse, error) {
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
	item, err := tx.DataStore().RuntimeVariables().GetForWorkflow(ctx, ns.Name, file.Path, req.GetOld())
	if err != nil {
		return nil, err
	}

	newName := req.GetNew()
	updated, err := tx.DataStore().RuntimeVariables().Patch(ctx, item.ID, &datastore.RuntimeVariablePatch{
		Name: &newName,
	})
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	var resp grpc.RenameWorkflowVariableResponse

	resp.CreatedAt = timestamppb.New(updated.CreatedAt)
	resp.Key = updated.Name
	resp.Namespace = ns.Name
	resp.TotalSize = int64(updated.Size)
	resp.UpdatedAt = timestamppb.New(updated.UpdatedAt)
	resp.MimeType = updated.MimeType

	return &resp, nil
}
