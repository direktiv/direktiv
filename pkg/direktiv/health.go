package direktiv

import (
	"context"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/vorteil/direktiv/pkg/health"
)

type healthServer struct {
	health.UnimplementedHealthServer

	config        *Config
	grpc          *grpc.Server
	isolateServer *isolateServer
	engine        *workflowEngine
}

func newHealthServer(config *Config, isolate *isolateServer, engine *workflowEngine) *healthServer {
	return &healthServer{
		config:        config,
		isolateServer: isolate,
		engine:        engine,
	}
}

func (hs *healthServer) name() string {
	return "health"
}

func (hs *healthServer) Check(ctx context.Context, in *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {

	var resp health.HealthCheckResponse
	resp.Status = health.HealthCheckResponse_SERVING

	if hs.isolateServer != nil {

		// log.Debugf("running isolate health check")
		//
		// dummy := ""
		// img := "vorteil/debug:v1"
		// var s int32
		//
		// actionID := "testAction"
		// instanceID := "testInstance"
		//
		// data := `{}`
		// cmd := "/debug"
		//
		// ir := &isolate.RunIsolateRequest{
		// 	ActionId:   &actionID,
		// 	Namespace:  &dummy,
		// 	InstanceId: &instanceID,
		// 	Image:      &img,
		// 	Size:       &s,
		// 	Data:       []byte(data),
		// 	Command:    &cmd,
		// }
		//
		// err := hs.isolateServer.runAction(ir, true)
		// if err != nil {
		// 	log.Errorf("health check failed: %v", err)
		// 	resp.Status = health.HealthCheckResponse_NOT_SERVING
		// }

	}

	if hs.engine != nil {
		// log.Debugf("running flow health check")
	}

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
