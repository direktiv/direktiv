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
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (internal *internal) FileVariableParcels(req *grpc.VariableInternalRequest, srv grpc.Internal_FileVariableParcelsServer) error {
	slog.Debug("Handling gRPC request", "this", this())

	ctx := srv.Context()

	inst, err := internal.getInstance(ctx, req.GetInstance())
	if err != nil {
		return err
	}

	tx, err := internal.beginSQLTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var data []byte
	var path, checksum, mime string
	var createdAt, updatedAt *timestamppb.Timestamp

	path, err = filestore.SanitizePath(req.GetKey())
	if err != nil {
		return err
	}

	file, err := tx.FileStore().ForNamespace(inst.Instance.Namespace).GetFile(ctx, req.GetKey())
	if err == nil {
		path = file.Path
		checksum = file.Checksum
		createdAt = timestamppb.New(file.CreatedAt)
		updatedAt = timestamppb.New(file.UpdatedAt)
		mime = file.MIMEType

		data, err = tx.FileStore().ForFile(file).GetData(ctx)
		if err != nil {
			return err
		}
	} else {
		if errors.Is(err, filestore.ErrNotFound) {
			data = make([]byte, 0)
			checksum = bytedata.Checksum(data)
			createdAt = timestamppb.New(time.Now())
			updatedAt = createdAt
		} else {
			return err
		}
	}

	tx.Rollback()

	iresp := &grpc.VariableInternalResponse{
		Instance:  inst.Instance.ID.String(),
		Key:       path,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Checksum:  checksum,
		TotalSize: int64(len(data)),
		Data:      data,
		MimeType:  mime,
	}

	err = srv.Send(iresp)
	if err != nil {
		return err
	}

	return nil
}

func (internal *internal) NamespaceVariableParcels(req *grpc.VariableInternalRequest, srv grpc.Internal_NamespaceVariableParcelsServer) error {
	slog.Debug("Handling gRPC request", "this", this())

	ctx := srv.Context()

	inst, err := internal.getInstance(ctx, req.GetInstance())
	if err != nil {
		return err
	}

	resp, err := internal.flow.NamespaceVariable(ctx, &grpc.NamespaceVariableRequest{
		Namespace: inst.TelemetryInfo.NamespaceName,
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

type setNamespaceVariableParcelsTranslator struct {
	internal *internal
	inst     *libengine.Instance
	grpc.Internal_SetNamespaceVariableParcelsServer
}

func (srv *setNamespaceVariableParcelsTranslator) SendAndClose(resp *grpc.SetNamespaceVariableResponse) error {
	var inst string
	if srv.inst != nil {
		inst = srv.inst.Instance.ID.String()
	}

	return srv.Internal_SetNamespaceVariableParcelsServer.SendAndClose(&grpc.SetVariableInternalResponse{
		Instance:  inst,
		Key:       resp.GetKey(),
		CreatedAt: resp.GetCreatedAt(),
		UpdatedAt: resp.GetUpdatedAt(),
		Checksum:  resp.GetChecksum(),
		TotalSize: resp.GetTotalSize(),
		MimeType:  resp.GetMimeType(),
	})
}

func (srv *setNamespaceVariableParcelsTranslator) Recv() (*grpc.SetNamespaceVariableRequest, error) {
	req, err := srv.Internal_SetNamespaceVariableParcelsServer.Recv()
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

	return &grpc.SetNamespaceVariableRequest{
		Namespace: srv.inst.TelemetryInfo.NamespaceName,
		Key:       req.GetKey(),
		TotalSize: req.GetTotalSize(),
		Data:      req.GetData(),
		MimeType:  req.GetMimeType(),
	}, nil
}

func (internal *internal) SetNamespaceVariableParcels(srv grpc.Internal_SetNamespaceVariableParcelsServer) error {
	slog.Debug("Handling gRPC request", "this", this())

	fsrv := &setNamespaceVariableParcelsTranslator{
		internal: internal,
		Internal_SetNamespaceVariableParcelsServer: srv,
	}

	err := internal.flow.SetNamespaceVariableParcels(fsrv)
	if err != nil {
		return err
	}

	return nil
}

func (flow *flow) NamespaceVariable(ctx context.Context, req *grpc.NamespaceVariableRequest) (*grpc.NamespaceVariableResponse, error) {
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

	item, err := tx.DataStore().RuntimeVariables().GetForNamespace(ctx, ns.Name, req.GetKey())
	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			t := time.Now()

			return &grpc.NamespaceVariableResponse{
				Namespace: ns.Name,
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

	var resp grpc.NamespaceVariableResponse

	resp.Namespace = ns.Name
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

func (flow *flow) SetNamespaceVariable(ctx context.Context, req *grpc.SetNamespaceVariableRequest) (*grpc.SetNamespaceVariableResponse, error) {
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

	newVar, err := tx.DataStore().RuntimeVariables().Set(ctx, &datastore.RuntimeVariable{
		Namespace: ns.Name,
		Name:      req.GetKey(),
		Data:      req.GetData(),
		MimeType:  req.GetMimeType(),
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

	var resp grpc.SetNamespaceVariableResponse

	resp.Namespace = ns.Name
	resp.Key = req.GetKey()
	resp.CreatedAt = timestamppb.New(newVar.CreatedAt)
	resp.UpdatedAt = timestamppb.New(newVar.UpdatedAt)
	resp.TotalSize = int64(newVar.Size)
	resp.MimeType = newVar.MimeType

	return &resp, nil
}

//nolint:dupl
func (flow *flow) SetNamespaceVariableParcels(srv grpc.Flow_SetNamespaceVariableParcelsServer) error {
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
