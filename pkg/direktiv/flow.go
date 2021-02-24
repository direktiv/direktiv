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
	grpccred "google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/emptypb"
)

type flowServer struct {
	flow.UnimplementedDirektivFlowServer

	config   *Config
	engine   *workflowEngine
	grpc     *grpc.Server
	grpcConn *grpc.ClientConn
}

func newFlowServer(config *Config, engine *workflowEngine) *flowServer {
	return &flowServer{
		config: config,
		engine: engine,
	}
}

func (f *flowServer) stop() {

	if f.grpc != nil {
		f.grpc.GracefulStop()
	}

}

func (f *flowServer) name() string {
	return "flow"
}

func (f *flowServer) setClient(wfs *WorkflowServer) error {

	conn, err := getEndpointTLS(wfs.config, flowComponent, wfs.config.FlowAPI.Endpoint)
	if err != nil {
		return err
	}

	wfs.componentAPIs.flowClient = flow.NewDirektivFlowClient(conn)
	wfs.componentAPIs.conns = append(wfs.componentAPIs.conns, conn)

	return nil

}

func (f *flowServer) start() error {

	log.Infof("flow endpoint starting at %v", f.config.FlowAPI.Bind)

	tls, err := tlsForGRPC(f.config.Certs.Directory, flowComponent,
		serverType, (f.config.Certs.Secure != 1))
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", f.config.FlowAPI.Bind)
	if err != nil {
		return err
	}

	f.grpc = grpc.NewServer(grpc.Creds(grpccred.NewTLS(tls)))

	flow.RegisterDirektivFlowServer(f.grpc, f)

	go f.grpc.Serve(listener)

	return nil

}

func (f *flowServer) ReportActionResults(ctx context.Context, in *flow.ReportActionResultsRequest) (*emptypb.Empty, error) {

	var resp emptypb.Empty

	ctx, wli, err := f.engine.loadWorkflowLogicInstance(in.GetInstanceId(), int(in.GetStep()))
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

	go f.engine.runState(ctx, wli, savedata, wakedata)

	return &resp, nil

}

func (f *flowServer) Resume(ctx context.Context, in *flow.ResumeRequest) (*emptypb.Empty, error) {

	var resp emptypb.Empty

	ctx, wli, err := f.engine.loadWorkflowLogicInstance(in.GetInstanceId(), int(in.GetStep()))
	if err != nil {
		return nil, err
	}

	go f.engine.runState(ctx, wli, nil, nil)

	return &resp, nil

}
