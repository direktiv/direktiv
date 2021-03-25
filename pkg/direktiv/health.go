package direktiv

import (
	"context"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/vorteil/direktiv/pkg/health"
	"github.com/vorteil/direktiv/pkg/isolate"
)

type healthServer struct {
	health.UnimplementedHealthServer

	config        *Config
	grpc          *grpc.Server
	isolateServer *isolateServer
}

func newHealthServer(config *Config, isolate *isolateServer) *healthServer {
	return &healthServer{
		config:        config,
		isolateServer: isolate,
	}
}

func (hs *healthServer) name() string {
	return "health"
}

func (hs *healthServer) Check(ctx context.Context, in *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {

	var resp health.HealthCheckResponse
	resp.Status = health.HealthCheckResponse_SERVING

	if hs.isolateServer != nil {

		dummy := ""
		img := "vorteil/debug:v1"
		var s int32

		actionID := "testAction"
		instanceID := "testInstance"

		data := `{}`

		ir := &isolate.RunIsolateRequest{
			ActionId:   &actionID,
			Namespace:  &dummy,
			InstanceId: &instanceID,
			Image:      &img,
			Size:       &s,
			Data:       []byte(data),
		}

		err := hs.isolateServer.runAction(ir, true)
		if err != nil {
			log.Errorf("health check failed: %v", err)
			resp.Status = health.HealthCheckResponse_NOT_SERVING
		}

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
