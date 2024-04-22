package flow

import (
	"errors"
	"log/slog"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	libengine "github.com/direktiv/direktiv/pkg/refactor/engine"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (internal *internal) FileVariableParcels(req *grpc.VariableInternalRequest, srv grpc.Internal_FileVariableParcelsServer) error {
	slog.Debug("Handling gRPC request", "this", this())

	ctx := srv.Context()

	inst, err := internal.getInstance(ctx, req.GetInstance())
	if err != nil {
		return err
	}

	tx, err := internal.beginSqlTx(ctx)
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
