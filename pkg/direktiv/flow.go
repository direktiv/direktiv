package direktiv

import (
	"context"
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/flow"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type flowServer struct {
	flow.UnimplementedDirektivFlowServer

	config *Config
	engine *workflowEngine
	grpc   *grpc.Server
}

func newFlowServer(config *Config, engine *workflowEngine) *flowServer {
	return &flowServer{
		config: config,
		engine: engine,
	}
}

func (fs *flowServer) stop() {

	if fs.grpc != nil {
		fs.grpc.GracefulStop()
	}

}

func (fs *flowServer) name() string {
	return "flow"
}

func (fs *flowServer) start(s *WorkflowServer) error {
	return GrpcStart(&fs.grpc, "flow", s.config.FlowAPI.Bind, func(srv *grpc.Server) {
		flow.RegisterDirektivFlowServer(srv, fs)
	})
}

func (fs *flowServer) ReportActionResults(ctx context.Context, in *flow.ReportActionResultsRequest) (*emptypb.Empty, error) {

	log.Debugf("action response: %v", in.GetActionId())

	var resp emptypb.Empty

	ctx, wli, err := fs.engine.loadWorkflowLogicInstance(in.GetInstanceId(), int(in.GetStep()))
	if err != nil {
		return nil, err
	}

	if fs.engine.server.hostname == in.GetSource() {
		log.Debugf("different action receiver")
	}

	wakedata, err := json.Marshal(&actionResultPayload{
		ActionID:     in.GetActionId(),
		ErrorCode:    in.GetErrorCode(),
		ErrorMessage: in.GetErrorMessage(),
		Output:       in.GetOutput(),
	})
	if err != nil {
		wli.Close()
		err = fmt.Errorf("cannot marshal the action results payload: %v", err)
		log.Error(err)
		return nil, err
	}

	savedata, err := InstanceMemory(wli.rec)
	if err != nil {
		wli.Close()
		return nil, err
	}

	go fs.engine.runState(ctx, wli, savedata, wakedata)

	return &resp, nil

}

func (fs *flowServer) Resume(ctx context.Context, in *flow.ResumeRequest) (*emptypb.Empty, error) {

	var resp emptypb.Empty

	ctx, wli, err := fs.engine.loadWorkflowLogicInstance(in.GetInstanceId(), int(in.GetStep()))
	if err != nil {
		return nil, err
	}

	go fs.engine.runState(ctx, wli, nil, nil)

	return &resp, nil

}
