package direktiv

import (
	"context"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/vorteil/direktiv/pkg/health"
)

type healthServer struct {
	health.UnimplementedHealthServer

	config *Config
	grpc   *grpc.Server
	engine *workflowEngine
}

func newHealthServer(config *Config, engine *workflowEngine) *healthServer {
	return &healthServer{
		config: config,
		engine: engine,
	}
}

func (hs *healthServer) name() string {
	return "health"
}

func (hs *healthServer) Check(ctx context.Context, in *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {

	var resp health.HealthCheckResponse
	resp.Status = health.HealthCheckResponse_SERVING

	log.Debugf("running health check executed")

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
