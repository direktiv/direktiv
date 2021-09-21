package flow

import (
	"context"
	"encoding/json"
	"net"
	"time"

	"github.com/vorteil/direktiv/pkg/flow/grpc"
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

	internal.srv = libgrpc.NewServer(
		libgrpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *libgrpc.UnaryServerInfo, handler libgrpc.UnaryHandler) (resp interface{}, err error) {
			resp, err = handler(ctx, req)
			if err != nil {
				return nil, translateError(err)
			}
			return resp, nil
		}),
		libgrpc.StreamInterceptor(func(srv interface{}, ss libgrpc.ServerStream, info *libgrpc.StreamServerInfo, handler libgrpc.StreamHandler) error {
			err := handler(srv, ss)
			if err != nil {
				return translateError(err)
			}
			return nil
		}),
	)

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

	wakedata, err := json.Marshal(&actionResultPayload{
		ActionID:     req.GetActionId(),
		ErrorCode:    req.GetErrorCode(),
		ErrorMessage: req.GetErrorMessage(),
		Output:       req.GetOutput(),
	})
	if err != nil {
		internal.sugar.Error(err)
		return nil, err
	}

	ctx, im, err := internal.engine.loadInstanceMemory(req.GetInstanceId(), int(req.GetStep()))
	if err != nil {
		internal.sugar.Error(err)
		return nil, err
	}

	internal.sugar.Debugf("Handling report action results: %s", this())

	go internal.engine.runState(ctx, im, wakedata, nil)

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
		internal.logToInstance(ctx, t, d.in, msg)
	}

	var resp emptypb.Empty

	return &resp, nil

}
