package direktiv

import (
	"context"
	"encoding/base64"
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

	var resp emptypb.Empty

	ctx, wli, err := fs.engine.loadWorkflowLogicInstance(in.GetInstanceId(), int(in.GetStep()))
	if err != nil {
		return nil, err
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

	var savedata []byte

	if wli.rec.Memory != "" {

		savedata, err = base64.StdEncoding.DecodeString(wli.rec.Memory)
		if err != nil {
			wli.Close()
			err = fmt.Errorf("cannot decode the savedata: %v", err)
			log.Error(err)
			return nil, err
		}

		// TODO: ?
		// wli.rec, err = wli.rec.Update().SetNillableMemory(nil).Save(ctx)
		// if err != nil {
		// 	log.Errorf("cannot update savedata information: %v", err)
		// 	return
		// }

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
