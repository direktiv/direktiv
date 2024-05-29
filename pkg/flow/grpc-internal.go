package flow

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/flow/nohome/recipient"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	enginerefactor "github.com/direktiv/direktiv/pkg/refactor/engine"
	libengine "github.com/direktiv/direktiv/pkg/refactor/engine"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	libgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type internal struct {
	*server
	listener net.Listener
	srv      *libgrpc.Server
	grpc.UnsafeInternalServer
}

func initInternalServer(ctx context.Context, srv *server) (*internal, error) {
	var err error

	internal := &internal{server: srv}

	internal.listener, err = net.Listen("tcp", ":7777") //nolint:gosec
	if err != nil {
		return nil, err
	}

	opts := util.GrpcServerOptions(unaryInterceptor, streamInterceptor)

	internal.srv = libgrpc.NewServer(opts...)

	grpc.RegisterInternalServer(internal.srv, internal)
	reflection.Register(internal.srv)

	go func() {
		<-ctx.Done()
		internal.srv.Stop()
	}()

	return internal, nil
}

func (internal *internal) Run() error {
	err := internal.srv.Serve(internal.listener)
	if err != nil {
		return err
	}

	return nil
}

func (internal *internal) ReportActionResults(ctx context.Context, req *grpc.ReportActionResultsRequest) (*emptypb.Empty, error) {
	payload := &actionResultPayload{
		ActionID:     req.GetActionId(),
		ErrorCode:    req.GetErrorCode(),
		ErrorMessage: req.GetErrorMessage(),
		Output:       req.GetOutput(),
	}

	uid, err := uuid.Parse(req.GetInstanceId())
	if err != nil {
		slog.Debug("Error returned to gRPC request", "this", this(), "error", err)
		return nil, err
	}

	err = internal.engine.enqueueInstanceMessage(ctx, uid, "action", payload)
	if err != nil {
		slog.Debug("Error returned to gRPC request", "this", this(), "error", err)
		return nil, err
	}

	var resp emptypb.Empty

	return &resp, nil
}

func (internal *internal) ActionLog(ctx context.Context, req *grpc.ActionLogRequest) (*emptypb.Empty, error) {
	slog.Debug("Handling gRPC request", "this", this())

	instance, err := internal.getInstance(ctx, req.GetInstanceId())
	if err != nil {
		slog.Error("get instance", "error", err)
		return nil, err
	}

	flow := instance.RuntimeInfo.Flow
	stateID := flow[len(flow)-1]

	tags := instance.GetAttributes(recipient.Instance)
	tags["loop-index"] = fmt.Sprintf("%d", req.GetIterator())
	tags["state-id"] = stateID
	tags["state-type"] = "action"
	loggingCtx := enginerefactor.AddTag(ctx, "state", stateID)
	loggingCtx = enginerefactor.AddTag(loggingCtx, "branch", req.GetIterator())
	loggingCtx = enginerefactor.WithTrack(loggingCtx, enginerefactor.BuildInstanceTrack(instance))
	loggingCtx = enginerefactor.AddTag(loggingCtx, "namespace", instance.Instance.Namespace)
	loggingCtx = instance.WithTags(loggingCtx)
	for _, msg := range req.GetMsg() {
		res := truncateLogsMsg(msg, 1024)
		slog.Info(res, enginerefactor.GetSlogAttributesWithStatus(loggingCtx, core.LogRunningStatus)...)
	}
	var resp emptypb.Empty

	return &resp, nil
}

func truncateLogsMsg(msg string,
	length int,
) string {
	res := ""
	if len(msg) <= 1 {
		return msg
	}
	m := strings.Split(msg, "\n")
	for i, v := range m {
		//nolint:copyloopvar
		truncated := v
		if len(v) > length {
			truncated = v[:length]
		}
		if i == len(m)-1 {
			res += truncated
		} else {
			res += truncated + "\n"
		}
	}

	return res
}

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

func (internal *internal) getInstance(ctx context.Context, instanceID string) (*enginerefactor.Instance, error) {
	id, err := uuid.Parse(instanceID)
	if err != nil {
		return nil, err
	}

	tx, err := internal.flow.beginSQLTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	idata, err := tx.InstanceStore().ForInstanceID(id).GetSummary(ctx)
	if err != nil {
		return nil, err
	}

	instance, err := enginerefactor.ParseInstanceData(idata)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

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
