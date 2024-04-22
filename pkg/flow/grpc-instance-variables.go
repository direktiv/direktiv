package flow

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	libengine "github.com/direktiv/direktiv/pkg/refactor/engine"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	parcelSize = 0x100000
)

func (internal *internal) InstanceVariableParcels(req *grpc.VariableInternalRequest, srv grpc.Internal_InstanceVariableParcelsServer) error {
	slog.Debug("Handling gRPC request", "this", this())

	ctx := srv.Context()

	inst, err := internal.getInstance(ctx, req.GetInstance())
	if err != nil {
		return err
	}

	resp, err := internal.flow.InstanceVariable(ctx, &grpc.InstanceVariableRequest{
		Namespace: inst.TelemetryInfo.NamespaceName,
		Instance:  inst.Instance.ID.String(),
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

type setInstanceVariableParcelsTranslator struct {
	internal *internal
	inst     *libengine.Instance
	grpc.Internal_SetInstanceVariableParcelsServer
}

func (srv *setInstanceVariableParcelsTranslator) SendAndClose(resp *grpc.SetInstanceVariableResponse) error {
	var inst string
	if srv.inst != nil {
		inst = srv.inst.Instance.ID.String()
	}

	return srv.Internal_SetInstanceVariableParcelsServer.SendAndClose(&grpc.SetVariableInternalResponse{
		Instance:  inst,
		Key:       resp.GetKey(),
		CreatedAt: resp.GetCreatedAt(),
		UpdatedAt: resp.GetUpdatedAt(),
		Checksum:  resp.GetChecksum(),
		TotalSize: resp.GetTotalSize(),
		MimeType:  resp.GetMimeType(),
	})
}

func (srv *setInstanceVariableParcelsTranslator) Recv() (*grpc.SetInstanceVariableRequest, error) {
	req, err := srv.Internal_SetInstanceVariableParcelsServer.Recv()
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

	return &grpc.SetInstanceVariableRequest{
		Namespace: srv.inst.TelemetryInfo.NamespaceName,
		Instance:  srv.inst.Instance.ID.String(),
		Key:       req.GetKey(),
		TotalSize: req.GetTotalSize(),
		Data:      req.GetData(),
		MimeType:  req.GetMimeType(),
	}, nil
}

func (internal *internal) SetInstanceVariableParcels(srv grpc.Internal_SetInstanceVariableParcelsServer) error {
	slog.Debug("Handling gRPC request", "this", this())

	fsrv := &setInstanceVariableParcelsTranslator{
		internal: internal,
		Internal_SetInstanceVariableParcelsServer: srv,
	}

	err := internal.flow.SetInstanceVariableParcels(fsrv)
	if err != nil {
		return err
	}

	return nil
}

func (flow *flow) InstanceVariable(ctx context.Context, req *grpc.InstanceVariableRequest) (*grpc.InstanceVariableResponse, error) {
	slog.Debug("Handling gRPC request", "this", this())

	inst, err := flow.getInstance(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return nil, err
	}

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	item, err := tx.DataStore().RuntimeVariables().GetForInstance(ctx, inst.Instance.ID, req.GetKey())
	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			t := time.Now()

			return &grpc.InstanceVariableResponse{
				Namespace: req.GetNamespace(),
				Instance:  inst.Instance.ID.String(),
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

	var resp grpc.InstanceVariableResponse

	resp.Namespace = req.GetNamespace()
	resp.Instance = inst.Instance.ID.String()
	resp.Key = item.Name
	resp.CreatedAt = timestamppb.New(item.CreatedAt)
	resp.UpdatedAt = timestamppb.New(item.UpdatedAt)
	resp.TotalSize = int64(item.Size)
	resp.MimeType = item.MimeType

	if resp.TotalSize > parcelSize {
		return nil, status.Error(codes.ResourceExhausted, "variable too large to return without using the parcelling API")
	}
	data, err := tx.DataStore().RuntimeVariables().LoadData(ctx, item.ID)
	if err != nil {
		return nil, err
	}

	resp.Data = data

	return &resp, nil
}

func (flow *flow) InstanceVariableParcels(req *grpc.InstanceVariableRequest, srv grpc.Flow_InstanceVariableParcelsServer) error {
	slog.Debug("Handling gRPC request", "this", this())

	ctx := srv.Context()

	resp, err := flow.InstanceVariable(ctx, &grpc.InstanceVariableRequest{
		Namespace: req.GetNamespace(),
		Instance:  req.GetInstance(),
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

func (flow *flow) InstanceVariables(ctx context.Context, req *grpc.InstanceVariablesRequest) (*grpc.InstanceVariablesResponse, error) {
	slog.Debug("Handling gRPC request", "this", this())

	instance, err := flow.getInstance(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return nil, err
	}

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	list, err := tx.DataStore().RuntimeVariables().ListForInstance(ctx, instance.Instance.ID)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.InstanceVariablesResponse)
	resp.Namespace = instance.TelemetryInfo.NamespaceName
	resp.Instance = instance.Instance.ID.String()
	resp.Variables = new(grpc.Variables)
	resp.Variables.PageInfo = nil

	resp.Variables.Results = bytedata.ConvertRuntimeVariablesToGrpcVariableList(list)

	return resp, nil
}

func (flow *flow) InstanceVariablesStream(req *grpc.InstanceVariablesRequest, srv grpc.Flow_InstanceVariablesStreamServer) error {
	slog.Debug("Handling gRPC request", "this", this())
	ctx := srv.Context()

	resp, err := flow.InstanceVariables(ctx, req)
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

func (flow *flow) SetInstanceVariable(ctx context.Context, req *grpc.SetInstanceVariableRequest) (*grpc.SetInstanceVariableResponse, error) {
	slog.Debug("Handling gRPC request", "this", this())

	inst, err := flow.getInstance(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return nil, err
	}

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	newVar, err := tx.DataStore().RuntimeVariables().Set(ctx, &datastore.RuntimeVariable{
		Namespace:  inst.Instance.Namespace,
		InstanceID: inst.Instance.ID,
		Name:       req.GetKey(),
		Data:       req.GetData(),
		MimeType:   req.GetMimeType(),
	})
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	// TODO: Alex, please fix here.

	// flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Set namespace variable '%s'.", req.GetKey())
	// flow.pubsub.NotifyNamespaceVariables(cached.Namespace)

	var resp grpc.SetInstanceVariableResponse

	resp.Namespace = inst.TelemetryInfo.NamespaceName
	resp.Instance = inst.Instance.ID.String()
	resp.Key = req.GetKey()
	resp.CreatedAt = timestamppb.New(newVar.CreatedAt)
	resp.UpdatedAt = timestamppb.New(newVar.UpdatedAt)
	resp.TotalSize = int64(newVar.Size)
	resp.MimeType = newVar.MimeType

	return &resp, nil
}

func (flow *flow) SetInstanceVariableParcels(srv grpc.Flow_SetInstanceVariableParcelsServer) error {
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
	resp, err := flow.SetInstanceVariable(ctx, firstReq)
	if err != nil {
		return err
	}
	err = srv.SendAndClose(resp)
	if err != nil {
		return err
	}

	return nil
}
