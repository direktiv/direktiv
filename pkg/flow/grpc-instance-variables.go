package flow

import (
	"log/slog"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	libengine "github.com/direktiv/direktiv/pkg/refactor/engine"
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
