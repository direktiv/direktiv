package flow

import (
	"context"

	"github.com/vorteil/direktiv/pkg/health"
	"google.golang.org/grpc"
)

type healthServer struct {
	health.UnimplementedHealthServer
	grpc *grpc.Server
	*server
}

func newHealthServer(srv *server) *healthServer {
	return &healthServer{
		server: srv,
	}
}

func (hs *healthServer) name() string {
	return "health"
}

func (hs *healthServer) Check(ctx context.Context, in *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {

	var resp health.HealthCheckResponse
	resp.Status = health.HealthCheckResponse_SERVING

	nsc := hs.db.Namespace
	query := nsc.Query()

	_, err := query.Limit(10).Offset(0).Count(ctx)
	if err != nil {
		return nil, err
	}

	return &resp, nil

}

func (hs *healthServer) Watch(in *health.HealthCheckRequest, srv health.Health_WatchServer) error {

	var resp health.HealthCheckResponse

	resp.Status = health.HealthCheckResponse_SERVING

	err := srv.Send(&resp)
	if err != nil {
		return err
	}

	return nil

}
