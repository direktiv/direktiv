package direktiv

import (
	"context"
	"net"

	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/health"
)

type healthServer struct {
	health.UnimplementedHealthServer

	config *Config
	grpc   *grpc.Server
}

func newHealthServer(config *Config) *healthServer {
	return &healthServer{
		config: config,
	}
}

func (h *healthServer) start() error {

	log.Infof("health endpoint starting at %s", h.config.HealthAPI.Bind)

	listener, err := net.Listen("tcp", h.config.HealthAPI.Bind)
	if err != nil {
		return err
	}

	h.grpc = grpc.NewServer()

	health.RegisterHealthServer(h.grpc, h)

	go h.grpc.Serve(listener)

	return nil

}

func (h *healthServer) stop() {

	if h.grpc != nil {
		h.grpc.GracefulStop()
	}

}

func (h *healthServer) name() string {
	return "health"
}

func (h *healthServer) Check(ctx context.Context, in *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {

	var resp health.HealthCheckResponse

	resp.Status = health.HealthCheckResponse_SERVING

	return &resp, nil

}

func (h *healthServer) Watch(in *health.HealthCheckRequest, srv health.Health_WatchServer) error {

	var resp health.HealthCheckResponse

	resp.Status = health.HealthCheckResponse_SERVING

	err := srv.Send(&resp)
	if err != nil {
		return err
	}

	return nil

}
