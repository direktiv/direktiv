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
	"github.com/vorteil/direktiv/pkg/util"
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
	return util.GrpcStart(&fs.grpc, util.TLSFlowComponent, flowBind, func(srv *grpc.Server) {
		flow.RegisterDirektivFlowServer(srv, fs)
	})
}

func (fs *flowServer) ActionLog(ctx context.Context, in *flow.ActionLogRequest) (*emptypb.Empty, error) {

	var resp = new(emptypb.Empty)

	wi, err := fs.engine.db.getWorkflowInstance(ctx, in.GetInstanceId())
	if err != nil {
		return nil, err
	}

	logger, err := (*fs.engine.instanceLogger).LoggerFunc(wi.Edges.Workflow.Edges.Namespace.ID, in.GetInstanceId())
	if err != nil {
		return nil, err
	}
	defer logger.Close()

	msgs := in.GetMsg()

	for _, msg := range msgs {
		logger.Info(msg)
	}

	return resp, nil

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

	go fs.engine.runState(ctx, wli, savedata, wakedata, nil)

	return &resp, nil

}

func (fs *flowServer) Resume(ctx context.Context, in *flow.ResumeRequest) (*emptypb.Empty, error) {

	var resp emptypb.Empty

	ctx, wli, err := fs.engine.loadWorkflowLogicInstance(in.GetInstanceId(), int(in.GetStep()))
	if err != nil {
		return nil, err
	}

	go fs.engine.runState(ctx, wli, nil, nil, nil)

	return &resp, nil

}

const grpcChunkSize = 2 * 1024 * 1024

func (fs *flowServer) GetNamespaceVariable(in *flow.GetNamespaceVariableRequest, out flow.DirektivFlow_GetNamespaceVariableServer) error {

	ctx := out.Context()

	instanceId := in.GetInstanceId()
	if instanceId == "" {
		return errors.New("required instanceId")
	}

	key := in.GetKey()
	if key == "" {
		return errors.New("requires variable key")
	}

	wi, err := fs.engine.server.dbManager.getWorkflowInstance(ctx, instanceId)
	if err != nil {
		return err
	}

	namespace := wi.Edges.Workflow.Edges.Namespace.ID

	r, err := fs.engine.server.variableStorage.Retrieve(ctx, key, namespace)
	if err != nil {
		return err
	}
	defer r.Close()

	// break data into chunks
	var chunks int
	chunkSize := int64(grpcChunkSize)
	totalSize := r.Size()
	for {
		cr := io.LimitReader(r, grpcChunkSize)
		data, err := ioutil.ReadAll(cr)
		if err != nil {
			return err
		}
		if chunks > 0 && len(data) == 0 {
			break
		}
		resp := new(flow.GetNamespaceVariableResponse)
		resp.Value = data
		resp.TotalSize = &totalSize
		resp.ChunkSize = &chunkSize
		err = out.Send(resp)
		if err != nil {
			return err
		}
		chunks++
	}

	err = r.Close()
	if err != nil {
		return err
	}

	return nil

}

func (fs *flowServer) GetWorkflowVariable(in *flow.GetWorkflowVariableRequest, out flow.DirektivFlow_GetWorkflowVariableServer) error {

	ctx := out.Context()

	instanceId := in.GetInstanceId()
	if instanceId == "" {
		return errors.New("required instanceId")
	}

	key := in.GetKey()
	if key == "" {
		return errors.New("requires variable key")
	}

	wi, err := fs.engine.server.dbManager.getWorkflowInstance(ctx, instanceId)
	if err != nil {
		return err
	}

	namespace := wi.Edges.Workflow.Edges.Namespace.ID
	wfId := wi.Edges.Workflow.ID.String()

	r, err := fs.engine.server.variableStorage.Retrieve(ctx, key, namespace, wfId)
	if err != nil {
		return err
	}
	defer r.Close()

	// break data into chunks
	var chunks int
	totalSize := r.Size()
	chunkSize := int64(grpcChunkSize)
	for {
		cr := io.LimitReader(r, grpcChunkSize)
		data, err := ioutil.ReadAll(cr)
		if err != nil {
			return err
		}
		if chunks > 0 && len(data) == 0 {
			break
		}
		resp := new(flow.GetWorkflowVariableResponse)
		resp.Value = data
		resp.TotalSize = &totalSize
		resp.ChunkSize = &chunkSize
		err = out.Send(resp)
		if err != nil {
			return err
		}
		chunks++
	}

	err = r.Close()
	if err != nil {
		return err
	}

	return nil

}

func (fs *flowServer) GetInstanceVariable(in *flow.GetInstanceVariableRequest, out flow.DirektivFlow_GetInstanceVariableServer) error {

	ctx := out.Context()

	instanceId := in.GetInstanceId()
	if instanceId == "" {
		return errors.New("required instanceId")
	}

	key := in.GetKey()
	if key == "" {
		return errors.New("requires variable key")
	}

	wi, err := fs.engine.server.dbManager.getWorkflowInstance(ctx, instanceId)
	if err != nil {
		return err
	}

	namespace := wi.Edges.Workflow.Edges.Namespace.ID
	wfId := wi.Edges.Workflow.ID.String()

	r, err := fs.engine.server.variableStorage.Retrieve(ctx, key, namespace, wfId, instanceId)
	if err != nil {
		return err
	}
	defer r.Close()

	// break data into chunks
	var chunks int
	totalSize := r.Size()
	chunkSize := int64(grpcChunkSize)
	for {
		cr := io.LimitReader(r, grpcChunkSize)
		data, err := ioutil.ReadAll(cr)
		if err != nil {
			return err
		}
		if chunks > 0 && len(data) == 0 {
			break
		}
		resp := new(flow.GetInstanceVariableResponse)
		resp.Value = data
		resp.TotalSize = &totalSize
		resp.ChunkSize = &chunkSize
		err = out.Send(resp)
		if err != nil {
			return err
		}
		chunks++
	}

	err = r.Close()
	if err != nil {
		return err
	}

	return nil

}

func (fs *flowServer) SetNamespaceVariable(srv flow.DirektivFlow_SetNamespaceVariableServer) error {

	ctx := srv.Context()

	in, err := srv.Recv()
	if err != nil {
		return err
	}

	instanceId := in.GetInstanceId()
	if instanceId == "" {
		return errors.New("required instanceId")
	}

	key := in.GetKey()
	if key == "" {
		return errors.New("requires variable key")
	}

	totalSize := in.GetTotalSize()

	wi, err := fs.engine.server.dbManager.getWorkflowInstance(ctx, instanceId)
	if err != nil {
		return err
	}

	namespace := wi.Edges.Workflow.Edges.Namespace.ID

	if totalSize == 0 {
		err = fs.engine.server.variableStorage.Delete(ctx, key, namespace)
		if err != nil {
			return err
		}
	} else {
		w, err := fs.engine.server.variableStorage.Store(ctx, key, namespace)
		if err != nil {
			return err
		}

		var totalRead int64

		for {
			data := in.GetValue()
			k, err := io.Copy(w, bytes.NewReader(data))
			if err != nil {
				return err
			}
			totalRead += k
			if totalRead >= totalSize {
				break
			}
			in, err = srv.Recv()
			if err != nil {
				return err
			}
		}

		err = w.Close()
		if err != nil {
			return err
		}

		// TODO: resolve edge-case where namespace or workflow is deleted during this process
	}

	return nil

}

func (fs *flowServer) SetWorkflowVariable(srv flow.DirektivFlow_SetWorkflowVariableServer) error {

	ctx := srv.Context()

	in, err := srv.Recv()
	if err != nil {
		return err
	}

	instanceId := in.GetInstanceId()
	if instanceId == "" {
		return errors.New("required instanceId")
	}

	key := in.GetKey()
	if key == "" {
		return errors.New("requires variable key")
	}

	totalSize := in.GetTotalSize()

	wi, err := fs.engine.server.dbManager.getWorkflowInstance(ctx, instanceId)
	if err != nil {
		return err
	}

	namespace := wi.Edges.Workflow.Edges.Namespace.ID
	wfId := wi.Edges.Workflow.ID.String()

	if totalSize == 0 {
		err = fs.engine.server.variableStorage.Delete(ctx, key, namespace, wfId)
		if err != nil {
			return err
		}
	} else {
		w, err := fs.engine.server.variableStorage.Store(ctx, key, namespace, wfId)
		if err != nil {
			return err
		}

		var totalRead int64

		for {
			data := in.GetValue()
			k, err := io.Copy(w, bytes.NewReader(data))
			if err != nil {
				return err
			}
			totalRead += k
			if totalRead >= totalSize {
				break
			}
			in, err = srv.Recv()
			if err != nil {
				return err
			}
		}

		err = w.Close()
		if err != nil {
			return err
		}

		// TODO: resolve edge-case where namespace or workflow is deleted during this process
	}

	return nil

}

func (fs *flowServer) SetInstanceVariable(srv flow.DirektivFlow_SetInstanceVariableServer) error {

	ctx := srv.Context()

	in, err := srv.Recv()
	if err != nil {
		return err
	}

	instanceId := in.GetInstanceId()
	if instanceId == "" {
		return errors.New("required instanceId")
	}

	key := in.GetKey()
	if key == "" {
		return errors.New("requires variable key")
	}

	totalSize := in.GetTotalSize()

	wi, err := fs.engine.server.dbManager.getWorkflowInstance(ctx, instanceId)
	if err != nil {
		return err
	}

	namespace := wi.Edges.Workflow.Edges.Namespace.ID
	wfId := wi.Edges.Workflow.ID.String()

	if totalSize == 0 {
		err = fs.engine.server.variableStorage.Delete(ctx, key, namespace, wfId, instanceId)
		if err != nil {
			return err
		}
	} else {
		w, err := fs.engine.server.variableStorage.Store(ctx, key, namespace, wfId, instanceId)
		if err != nil {
			return err
		}

		var totalRead int64

		for {
			data := in.GetValue()
			k, err := io.Copy(w, bytes.NewReader(data))
			if err != nil {
				return err
			}
			totalRead += k
			if totalRead >= totalSize {
				break
			}
			in, err = srv.Recv()
			if err != nil {
				return err
			}
		}

		err = w.Close()
		if err != nil {
			return err
		}

		// TODO: resolve edge-case where namespace or workflow is deleted during this process
	}

	return nil

}
