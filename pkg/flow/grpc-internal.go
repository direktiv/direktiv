package flow

import (
	"context"
	"encoding/json"
	"net"
	"time"

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

	internal.sugar.Debugf("Handling gRPC request: %s", this())

	t := time.Now()

	inc := internal.db.Instance

	d, err := internal.getInstance(ctx, inc, req.GetInstanceId(), false)
	if err != nil {
		internal.sugar.Error(err)
		return nil, err
	}

	for _, msg := range req.GetMsg() {
		//internal.logToInstanceRaw(ctx, t, d.in, d.tags(), msg)
		internal.logToInstanceRaw(ctx, t, d, msg)
	}

	var resp emptypb.Empty

	return &resp, nil

}
