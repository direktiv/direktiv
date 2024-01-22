package flow

import (
	"context"
	"encoding/json"
	"net"
	"strings"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	libgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

type internal struct {
	*server
	listener net.Listener
	srv      *libgrpc.Server
	grpc.UnimplementedInternalServer
}

func initInternalServer(ctx context.Context, srv *server) (*internal, error) {
	var err error

	internal := &internal{server: srv}

	internal.listener, err = net.Listen("tcp", ":7777")
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
	internal.sugar.Debugf("Handling gRPC request: %s", this())

	payload := &actionResultPayload{
		ActionID:     req.GetActionId(),
		ErrorCode:    req.GetErrorCode(),
		ErrorMessage: req.GetErrorMessage(),
		Output:       req.GetOutput(),
	}

	wakedata, err := json.Marshal(payload)
	if err != nil {
		internal.sugar.Error(err)
		return nil, err
	}

	ctx2, im, err := internal.engine.loadInstanceMemory(req.GetInstanceId(), int(req.GetStep()))
	if err != nil {
		internal.sugar.Error(err)
		return nil, err
	}

	internal.sugar.Debugf("Handling report action results: %s", this())

	traceActionResult(ctx2, payload)

	go internal.engine.runState(ctx2, im, wakedata, nil)

	var resp emptypb.Empty

	return &resp, nil
}

func (internal *internal) ActionLog(ctx context.Context, req *grpc.ActionLogRequest) (*emptypb.Empty, error) {
	// internal.sugar.Debugf("Handling gRPC request: %s", this())

	// instance, err := internal.getInstance(ctx, req.GetInstanceId())
	// if err != nil {
	// 	internal.sugar.Error(err)
	// 	return nil, err
	// }

	// flow := instance.RuntimeInfo.Flow
	// stateID := flow[len(flow)-1]

	// tags := instance.GetAttributes(recipient.Instance)
	// tags["loop-index"] = fmt.Sprintf("%d", req.Iterator)
	// tags["state-id"] = stateID
	// tags["state-type"] = "action"
	// for _, msg := range req.GetMsg() {
	// 	res := truncateLogsMsg(msg, 1024)
	// 	internal.logger.Infof(ctx, instance.Instance.ID, tags, res)
	// }

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
