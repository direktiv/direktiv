package direktiv

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

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

func (fs *flowServer) GetNamespaceVariable(ctx context.Context, in *flow.GetNamespaceVariableRequest) (*flow.GetNamespaceVariableResponse, error) {

	resp := new(flow.GetNamespaceVariableResponse)

	instanceId := in.GetInstanceId()
	if instanceId == "" {
		return nil, errors.New("required instanceId")
	}

	key := in.GetKey()
	if key == "" {
		return nil, errors.New("requires variable key")
	}

	wi, err := fs.engine.server.dbManager.getWorkflowInstance(ctx, instanceId)
	if err != nil {
		return nil, err
	}

	namespace := wi.Edges.Workflow.Edges.Namespace.ID

	r, err := fs.engine.server.variableStorage.Retrieve(ctx, key, namespace)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	err = r.Close()
	if err != nil {
		return nil, err
	}

	resp.Value = data

	return resp, nil

}

func (fs *flowServer) GetWorkflowVariable(ctx context.Context, in *flow.GetWorkflowVariableRequest) (*flow.GetWorkflowVariableResponse, error) {

	resp := new(flow.GetWorkflowVariableResponse)

	instanceId := in.GetInstanceId()
	if instanceId == "" {
		return nil, errors.New("required instanceId")
	}

	key := in.GetKey()
	if key == "" {
		return nil, errors.New("requires variable key")
	}

	wi, err := fs.engine.server.dbManager.getWorkflowInstance(ctx, instanceId)
	if err != nil {
		return nil, err
	}

	namespace := wi.Edges.Workflow.Edges.Namespace.ID
	wfId := wi.Edges.Workflow.ID.String()

	r, err := fs.engine.server.variableStorage.Retrieve(ctx, key, namespace, wfId)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	err = r.Close()
	if err != nil {
		return nil, err
	}

	resp.Value = data

	return resp, nil

}

func (fs *flowServer) GetInstanceVariable(ctx context.Context, in *flow.GetInstanceVariableRequest) (*flow.GetInstanceVariableResponse, error) {

	resp := new(flow.GetInstanceVariableResponse)

	instanceId := in.GetInstanceId()
	if instanceId == "" {
		return nil, errors.New("required instanceId")
	}

	key := in.GetKey()
	if key == "" {
		return nil, errors.New("requires variable key")
	}

	wi, err := fs.engine.server.dbManager.getWorkflowInstance(ctx, instanceId)
	if err != nil {
		return nil, err
	}

	namespace := wi.Edges.Workflow.Edges.Namespace.ID
	wfId := wi.Edges.Workflow.ID.String()

	r, err := fs.engine.server.variableStorage.Retrieve(ctx, key, namespace, wfId, instanceId)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	err = r.Close()
	if err != nil {
		return nil, err
	}

	resp.Value = data

	return resp, nil

}

func (fs *flowServer) SetNamespaceVariable(ctx context.Context, in *flow.SetNamespaceVariableRequest) (*emptypb.Empty, error) {

	var resp emptypb.Empty

	instanceId := in.GetInstanceId()
	if instanceId == "" {
		return nil, errors.New("required instanceId")
	}

	key := in.GetKey()
	if key == "" {
		return nil, errors.New("requires variable key")
	}

	data := in.GetValue()

	wi, err := fs.engine.server.dbManager.getWorkflowInstance(ctx, instanceId)
	if err != nil {
		return nil, err
	}

	namespace := wi.Edges.Workflow.Edges.Namespace.ID

	if len(data) == 0 {
		err = fs.engine.server.variableStorage.Delete(ctx, key, namespace)
		if err != nil {
			return nil, err
		}
	} else {
		w, err := fs.engine.server.variableStorage.Store(ctx, key, namespace)
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(w, bytes.NewReader(data))
		if err != nil {
			return nil, err
		}

		err = w.Close()
		if err != nil {
			return nil, err
		}

		// TODO: resolve edge-case where namespace or workflow is deleted during this process
	}

	return &resp, nil

}

func (fs *flowServer) SetWorkflowVariable(ctx context.Context, in *flow.SetWorkflowVariableRequest) (*emptypb.Empty, error) {

	var resp emptypb.Empty

	instanceId := in.GetInstanceId()
	if instanceId == "" {
		return nil, errors.New("required instanceId")
	}

	key := in.GetKey()
	if key == "" {
		return nil, errors.New("requires variable key")
	}

	data := in.GetValue()

	wi, err := fs.engine.server.dbManager.getWorkflowInstance(ctx, instanceId)
	if err != nil {
		return nil, err
	}

	namespace := wi.Edges.Workflow.Edges.Namespace.ID
	wfId := wi.Edges.Workflow.ID.String()

	if len(data) == 0 {
		err = fs.engine.server.variableStorage.Delete(ctx, key, namespace, wfId)
		if err != nil {
			return nil, err
		}
	} else {
		w, err := fs.engine.server.variableStorage.Store(ctx, key, namespace, wfId)
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(w, bytes.NewReader(data))
		if err != nil {
			return nil, err
		}

		err = w.Close()
		if err != nil {
			return nil, err
		}

		// TODO: resolve edge-case where namespace or workflow is deleted during this process
	}

	return &resp, nil

}

func (fs *flowServer) SetInstanceVariable(ctx context.Context, in *flow.SetInstanceVariableRequest) (*emptypb.Empty, error) {

	var resp emptypb.Empty

	instanceId := in.GetInstanceId()
	if instanceId == "" {
		return nil, errors.New("required instanceId")
	}

	key := in.GetKey()
	if key == "" {
		return nil, errors.New("requires variable key")
	}

	data := in.GetValue()

	wi, err := fs.engine.server.dbManager.getWorkflowInstance(ctx, instanceId)
	if err != nil {
		return nil, err
	}

	namespace := wi.Edges.Workflow.Edges.Namespace.ID
	wfId := wi.Edges.Workflow.ID.String()

	if len(data) == 0 {
		err = fs.engine.server.variableStorage.Delete(ctx, key, namespace, wfId, instanceId)
		if err != nil {
			return nil, err
		}
	} else {
		w, err := fs.engine.server.variableStorage.Store(ctx, key, namespace, wfId, instanceId)
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(w, bytes.NewReader(data))
		if err != nil {
			return nil, err
		}

		err = w.Close()
		if err != nil {
			return nil, err
		}

		// TODO: resolve edge-case where namespace or workflow is deleted during this process
	}

	return &resp, nil

}
