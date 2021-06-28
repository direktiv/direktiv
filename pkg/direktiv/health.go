package direktiv

import (
	"context"

	"google.golang.org/grpc"

	"github.com/vorteil/direktiv/pkg/health"
)

type healthServer struct {
	health.UnimplementedHealthServer
	grpc *grpc.Server

	s *WorkflowServer
}

func newHealthServer(s *WorkflowServer) *healthServer {
	return &healthServer{
		s: s,
	}
}

func (hs *healthServer) name() string {
	return "health"
}

func (hs *healthServer) Check(ctx context.Context, in *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {

	var resp health.HealthCheckResponse
	resp.Status = health.HealthCheckResponse_SERVING

	_, err := hs.s.dbManager.getNamespaces(context.Background(), int(0), int(10))
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
