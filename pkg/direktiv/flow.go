package direktiv

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/flow"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *WorkflowServer) grpcFlowStart() error {

	// TODO: make port configurable
	// TODO: save listener somewhere so that it can be shutdown
	// TODO: save grpc somewhere so that it can be shutdown
	log.Infof("flow endpoint starting at %v", s.config.FlowAPI.Bind)
	listener, err := net.Listen("tcp", s.config.FlowAPI.Bind)
	if err != nil {
		return err
	}

	s.grpcFlow = grpc.NewServer()

	flow.RegisterDirektivFlowServer(s.grpcFlow, s)

	go s.grpcFlow.Serve(listener)

	return nil

}

func (s *WorkflowServer) ReportActionResults(ctx context.Context, in *flow.ReportActionResultsRequest) (*emptypb.Empty, error) {

	var resp emptypb.Empty

	ctx, wli, err := s.engine.loadWorkflowLogicInstance(in.GetInstanceId(), int(in.GetStep()))
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

	go s.engine.runState(ctx, wli, savedata, wakedata)

	return &resp, nil

}

func (s *WorkflowServer) Resume(ctx context.Context, in *flow.ResumeRequest) (*emptypb.Empty, error) {

	var resp emptypb.Empty

	ctx, wli, err := s.engine.loadWorkflowLogicInstance(in.GetInstanceId(), int(in.GetStep()))
	if err != nil {
		return nil, err
	}

	go s.engine.runState(ctx, wli, nil, nil)

	return &resp, nil

}
