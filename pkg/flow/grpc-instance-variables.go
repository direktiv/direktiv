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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	parcelSize = 0x100000
)

func (flow *flow) InstanceVariable(ctx context.Context, req *grpc.InstanceVariableRequest) (*grpc.InstanceVariableResponse, error) {
	slog.Debug("Handling gRPC request", "this", this())

	inst, err := flow.getInstance(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return nil, err
	}

	tx, err := flow.beginSQLTx(ctx)
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

func (flow *flow) SetInstanceVariable(ctx context.Context, req *grpc.SetInstanceVariableRequest) (*grpc.SetInstanceVariableResponse, error) {
	slog.Debug("Handling gRPC request", "this", this())

	inst, err := flow.getInstance(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return nil, err
	}

	tx, err := flow.beginSQLTx(ctx)
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
	//  flow.pubsub.NotifyNamespaceVariables(cached.Namespace)

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

//nolint:dupl
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
