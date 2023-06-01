package flow

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (flow *flow) getWorkflow(ctx context.Context, namespace, path string) (ns *database.Namespace, f *filestore.File, err error) {
	ns, err = flow.edb.NamespaceByName(ctx, namespace)
	if err != nil {
		return
	}

	fStore, _, _, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return
	}
	defer rollback()

	f, err = fStore.ForRootID(ns.ID).GetFile(ctx, path)
	if err != nil {
		return
	}

	if f.Typ != filestore.FileTypeWorkflow {
		err = ErrNotWorkflow
		return
	}

	return
}

func (flow *flow) WorkflowVariable(ctx context.Context, req *grpc.WorkflowVariableRequest) (*grpc.WorkflowVariableResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	fStore, store, _, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	item, err := store.RuntimeVariables().GetByReferenceAndName(ctx, file.ID, req.GetKey())
	if err != nil {
		return nil, err
	}

	var resp grpc.WorkflowVariableResponse

	resp.Namespace = ns.Name
	resp.Path = file.Path
	resp.Key = item.Name
	resp.CreatedAt = timestamppb.New(item.CreatedAt)
	resp.UpdatedAt = timestamppb.New(item.UpdatedAt)
	resp.Checksum = item.Hash
	resp.TotalSize = int64(item.Size)
	resp.MimeType = item.MimeType

	if resp.TotalSize > parcelSize {
		return nil, status.Error(codes.ResourceExhausted, "variable too large to return without using the parcelling API")
	}
	data, err := store.RuntimeVariables().LoadData(ctx, item.ID)
	if err != nil {
		return nil, err
	}

	resp.Data = data

	return &resp, nil
}

func (flow *flow) WorkflowVariableParcels(req *grpc.WorkflowVariableRequest, srv grpc.Flow_WorkflowVariableParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	resp, err := flow.WorkflowVariable(ctx, &grpc.WorkflowVariableRequest{
		Namespace: req.GetNamespace(),
		Path:      req.GetPath(),
		Key:       req.GetKey(),
	})
	if err != nil {
		return err
	}
	err = srv.Send(resp)
	if err != nil {
		return err
	}

	return nil
}

func (flow *flow) WorkflowVariables(ctx context.Context, req *grpc.WorkflowVariablesRequest) (*grpc.WorkflowVariablesResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	fStore, store, _, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	list, err := store.RuntimeVariables().ListByWorkflowID(ctx, file.ID)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.WorkflowVariablesResponse)
	resp.Namespace = ns.Name
	resp.Path = file.Path
	resp.Variables = new(grpc.Variables)
	resp.Variables.PageInfo = nil

	resp.Variables.Results = bytedata.ConvertRuntimeVariablesToGrpcVariableList(list)

	return resp, nil
}

func (flow *flow) WorkflowVariablesStream(req *grpc.WorkflowVariablesRequest, srv grpc.Flow_WorkflowVariablesStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())
	ctx := srv.Context()

	resp, err := flow.WorkflowVariables(ctx, req)
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

func (flow *flow) SetWorkflowVariable(ctx context.Context, req *grpc.SetWorkflowVariableRequest) (*grpc.SetWorkflowVariableResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	fStore, store, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	newVar, err := store.RuntimeVariables().Set(ctx, &core.RuntimeVariable{
		ID:         uuid.New(),
		WorkflowID: file.ID,
		Name:       req.GetKey(),
		// TODO: check this.
		Scope:    "workflow",
		Data:     req.GetData(),
		MimeType: req.GetMimeType(),
	})
	if err != nil {
		return nil, err
	}

	if err = commit(ctx); err != nil {
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
	resp.Checksum = newVar.Hash
	resp.TotalSize = int64(newVar.Size)
	resp.MimeType = newVar.MimeType

	return &resp, nil
}

func (flow *flow) SetWorkflowVariableParcels(srv grpc.Flow_SetWorkflowVariableParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())
	ctx := srv.Context()

	req, err := srv.Recv()
	if err != nil {
		return err
	}

	firstReq := req

	totalSize := int(req.GetTotalSize())

	buf := new(bytes.Buffer)

	for {
		_, err = io.Copy(buf, bytes.NewReader(req.Data))
		if err != nil {
			return err
		}

		if req.TotalSize <= 0 {
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

		if req.TotalSize <= 0 {
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
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	fStore, store, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	item, err := store.RuntimeVariables().GetByReferenceAndName(ctx, file.ID, req.GetKey())
	if err != nil {
		return nil, err
	}

	err = store.RuntimeVariables().Delete(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	if err = commit(ctx); err != nil {
		return nil, err
	}

	// TODO: need fix here.
	//flow.logger.Infof(ctx, file.ID, database.GetAttributes(recipient.Workflow, ns, fileAttributes(*file)), "Deleted workflow variable '%s'.", vref.Name)
	//flow.pubsub.NotifyWorkflowVariables(file.ID)
	//
	//// Broadcast Event
	//broadcastInput := broadcastVariableInput{
	//	WorkflowPath: req.GetPath(),
	//	Key:          req.GetKey(),
	//	TotalSize:    int64(item.Size),
	//	Scope:        BroadcastEventScopeWorkflow,
	//}
	//err = flow.BroadcastVariable(ctx, BroadcastEventTypeDelete, BroadcastEventScopeNamespace, broadcastInput, ns)
	//if err != nil {
	//	return nil, err
	//}

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) RenameWorkflowVariable(ctx context.Context, req *grpc.RenameWorkflowVariableRequest) (*grpc.RenameWorkflowVariableResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	fStore, store, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}
	item, err := store.RuntimeVariables().GetByReferenceAndName(ctx, file.ID, req.GetOld())
	if err != nil {
		return nil, err
	}

	updated, err := store.RuntimeVariables().SetName(ctx, item.ID, req.GetNew())
	if err != nil {
		return nil, err
	}
	if err = commit(ctx); err != nil {
		return nil, err
	}

	// TODO: fix here.
	// flow.logger.Infof(ctx, file.ID, database.GetAttributes(recipient.Workflow, ns, fileAttributes(*file)), "Renamed workflow variable from '%s' to '%s'.", req.GetOld(), req.GetNew())
	// flow.pubsub.NotifyWorkflowVariables(file.ID)

	var resp grpc.RenameWorkflowVariableResponse

	resp.Checksum = updated.Hash
	resp.CreatedAt = timestamppb.New(updated.CreatedAt)
	resp.Key = updated.Name
	resp.Namespace = ns.Name
	resp.TotalSize = int64(updated.Size)
	resp.UpdatedAt = timestamppb.New(updated.UpdatedAt)
	resp.MimeType = updated.MimeType

	return &resp, nil
}
