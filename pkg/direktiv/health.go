package direktiv

import (
	"context"
	"net"

	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/health"
)

func (s *WorkflowServer) grpcHealthStart() error {

	// TODO: make port configurable
	// TODO: save listener somewhere so that it can be shutdown
	// TODO: save grpc somewhere so that it can be shutdown

	log.Infof("health endpoint starting at %s", s.config.HealthAPI.Bind)

	listener, err := net.Listen("tcp", s.config.HealthAPI.Bind)
	if err != nil {
		return err
	}

	s.grpcHealth = grpc.NewServer()

	health.RegisterHealthServer(s.grpcHealth, s)

	go s.grpcHealth.Serve(listener)

	return nil

}

func (s *WorkflowServer) Check(ctx context.Context, in *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {

	var resp health.HealthCheckResponse

	resp.Status = health.HealthCheckResponse_SERVING

	// TODO

	return &resp, nil

}

func (s *WorkflowServer) Watch(in *health.HealthCheckRequest, srv health.Health_WatchServer) error {

	var resp health.HealthCheckResponse

	resp.Status = health.HealthCheckResponse_SERVING

	err := srv.Send(&resp)
	if err != nil {
		return err
	}

	// TODO

	return nil

}
