package flow

import (
	"log/slog"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	libengine "github.com/direktiv/direktiv/pkg/refactor/engine"
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
