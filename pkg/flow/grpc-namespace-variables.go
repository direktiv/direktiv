package flow

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (flow *flow) NamespaceVariable(ctx context.Context, req *grpc.NamespaceVariableRequest) (*grpc.NamespaceVariableResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	_, store, _, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	list, err := store.RuntimeVariables().ListByNamespaceID(ctx, ns.ID)
	if err != nil {
		return nil, err
	}
	item := list.FilterByName(req.GetKey())
	if item == nil {
		return nil, status.Error(codes.NotFound, "variable key is not found")
	}

	var resp grpc.NamespaceVariableResponse

	resp.Namespace = ns.Name
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

func (flow *flow) NamespaceVariableParcels(req *grpc.NamespaceVariableRequest, srv grpc.Flow_NamespaceVariableParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	resp, err := flow.NamespaceVariable(ctx, &grpc.NamespaceVariableRequest{
		Namespace: req.GetNamespace(),
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

func (flow *flow) NamespaceVariables(ctx context.Context, req *grpc.NamespaceVariablesRequest) (*grpc.NamespaceVariablesResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	_, store, _, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	list, err := store.RuntimeVariables().ListByNamespaceID(ctx, ns.ID)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.NamespaceVariablesResponse)
	resp.Namespace = ns.Name
	resp.Variables = new(grpc.Variables)
	resp.Variables.PageInfo = nil

	resp.Variables.Results = bytedata.ConvertRuntimeVariablesToGrpcVariableList(list)

	return resp, nil
}

func (flow *flow) NamespaceVariablesStream(req *grpc.NamespaceVariablesRequest, srv grpc.Flow_NamespaceVariablesStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())
	ctx := srv.Context()

	resp, err := flow.NamespaceVariables(ctx, req)
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

func (flow *flow) SetNamespaceVariable(ctx context.Context, req *grpc.SetNamespaceVariableRequest) (*grpc.SetNamespaceVariableResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	_, store, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	newVar, err := store.RuntimeVariables().Set(ctx, &core.RuntimeVariable{
		ID:          uuid.New(),
		NamespaceID: ns.ID,
		Name:        req.GetKey(),
		// TODO: check this.
		Scope:    "namespace",
		Data:     req.GetData(),
		MimeType: req.GetMimeType(),
	})
	if err != nil {
		return nil, err
	}

	if err = commit(ctx); err != nil {
		return nil, err
	}

	// TODO: Alex, please fix here.

	// flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Set namespace variable '%s'.", req.GetKey())
	// flow.pubsub.NotifyNamespaceVariables(cached.Namespace)

	var resp grpc.SetNamespaceVariableResponse

	resp.Namespace = ns.Name
	resp.Key = req.GetKey()
	resp.CreatedAt = timestamppb.New(newVar.CreatedAt)
	resp.UpdatedAt = timestamppb.New(newVar.UpdatedAt)
	resp.Checksum = newVar.Hash
	resp.TotalSize = int64(newVar.Size)
	resp.MimeType = newVar.MimeType

	return &resp, nil
}

func (flow *flow) SetNamespaceVariableParcels(srv grpc.Flow_SetNamespaceVariableParcelsServer) error {
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
	resp, err := flow.SetNamespaceVariable(ctx, firstReq)
	if err != nil {
		return err
	}
	err = srv.SendAndClose(resp)
	if err != nil {
		return err
	}

	return nil
}

func (flow *flow) DeleteNamespaceVariable(ctx context.Context, req *grpc.DeleteNamespaceVariableRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	_, store, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	list, err := store.RuntimeVariables().ListByNamespaceID(ctx, ns.ID)
	if err != nil {
		return nil, err
	}
	item := list.FilterByName(req.GetKey())
	if item == nil {
		return nil, status.Error(codes.NotFound, "variable key is not found")
	}
	err = store.RuntimeVariables().Delete(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	if err = commit(ctx); err != nil {
		return nil, err
	}

	// TODO: nned fix here.
	// flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Deleted namespace variable '%s'.", vref.Name)
	// flow.pubsub.NotifyNamespaceVariables(cached.Namespace)

	// Broadcast Event
	//broadcastInput := broadcastVariableInput{
	//	WorkflowPath: "",
	//	Key:          req.GetKey(),
	//	TotalSize:    int64(vdata.Size),
	//	Scope:        BroadcastEventScopeNamespace,
	//}
	//err = flow.BroadcastVariable(ctx, BroadcastEventTypeDelete, BroadcastEventScopeNamespace, broadcastInput, cached.Namespace)
	//if err != nil {
	//	return nil, err
	//}

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) RenameNamespaceVariable(ctx context.Context, req *grpc.RenameNamespaceVariableRequest) (*grpc.RenameNamespaceVariableResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	_, store, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	list, err := store.RuntimeVariables().ListByNamespaceID(ctx, ns.ID)
	if err != nil {
		return nil, err
	}
	item := list.FilterByName(req.GetOld())
	if item == nil {
		return nil, status.Error(codes.NotFound, "variable key is not found")
	}
	updated, err := store.RuntimeVariables().SetName(ctx, item.ID, req.GetNew())
	if err != nil {
		return nil, err
	}
	if err = commit(ctx); err != nil {
		return nil, err
	}

	// TODO: need fix.
	// flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Renamed namespace variable from '%s' to '%s'.", req.GetOld(), req.GetNew())
	// flow.pubsub.NotifyNamespaceVariables(cached.Namespace)

	var resp grpc.RenameNamespaceVariableResponse

	resp.Checksum = updated.Hash
	resp.CreatedAt = timestamppb.New(updated.CreatedAt)
	resp.Key = updated.Name
	resp.Namespace = ns.Name
	resp.TotalSize = int64(updated.Size)
	resp.UpdatedAt = timestamppb.New(updated.UpdatedAt)
	resp.MimeType = updated.MimeType

	return &resp, nil
}
