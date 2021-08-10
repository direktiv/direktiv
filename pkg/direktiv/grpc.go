package direktiv

import (
	"context"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/ent"
	"github.com/vorteil/direktiv/pkg/health"
	"github.com/vorteil/direktiv/pkg/ingress"
	"github.com/vorteil/direktiv/pkg/model"
	secretsgrpc "github.com/vorteil/direktiv/pkg/secrets/grpc"
	"github.com/vorteil/direktiv/pkg/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ingressServer struct {
	ingress.UnimplementedDirektivIngressServer
	health.UnimplementedHealthServer

	wfServer *WorkflowServer
	grpc     *grpc.Server

	secretsClient secretsgrpc.SecretsServiceClient
	grpcConn      *grpc.ClientConn
}

func (is *ingressServer) stop() {

	if is.grpc != nil {
		is.grpc.GracefulStop()
	}

	if is.grpcConn != nil {
		is.grpcConn.Close()
	}

	// stop engine client
	for _, c := range is.wfServer.engine.grpcConns {
		c.Close()
	}

}

func (is *ingressServer) name() string {
	return "ingress"
}

func newIngressServer(s *WorkflowServer) (*ingressServer, error) {

	return &ingressServer{
		wfServer: s,
	}, nil

}

func (is *ingressServer) start(s *WorkflowServer) error {

	// get secrets client
	conn, err := util.GetEndpointTLS(util.TLSSecretsComponent)
	if err != nil {
		return err
	}
	is.grpcConn = conn
	is.secretsClient = secretsgrpc.NewSecretsServiceClient(conn)

	is.cronPoll()
	go is.cronPoller()

	return util.GrpcStart(&is.grpc, util.TLSIngressComponent, ingressBind, func(srv *grpc.Server) {
		ingress.RegisterDirektivIngressServer(srv, is)

		log.Debugf("append health check to ingress service")
		healthServer := newHealthServer(s)
		health.RegisterHealthServer(srv, healthServer)
		reflection.Register(srv)
	})

}

func (is *ingressServer) cronPoller() {
	for {
		time.Sleep(time.Minute * 15)
		is.cronPoll()
	}
}

func (is *ingressServer) cronPoll() {

	wfs, err := is.wfServer.dbManager.getAllWorkflows()
	if err != nil {
		log.Error(err)
		return
	}

	for _, x := range wfs {
		wf, err := is.wfServer.dbManager.getWorkflowByID(x.ID)
		if err != nil {
			log.Error(err)
		}
		is.cronPollerWorkflow(wf)
	}

}

func (is *ingressServer) cronPollerWorkflow(wf *ent.Workflow) {

	var workflow model.Workflow
	err := workflow.Load(wf.Workflow)
	if err != nil {
		log.Error(err)
	}

	is.wfServer.tmManager.deleteTimerByName("", is.wfServer.hostname, fmt.Sprintf("cron:%s", wf.ID.String()))
	if wf.Active {
		def := workflow.GetStartDefinition()
		if def.GetType() == model.StartTypeScheduled {
			scheduled := def.(*model.ScheduledStart)
			is.wfServer.tmManager.addCronNoBroadcast(fmt.Sprintf("cron:%s", wf.ID.String()), wfCron, scheduled.Cron, []byte(wf.ID.String()))
		}
	}

}

func (is *ingressServer) BroadcastEvent(ctx context.Context, in *ingress.BroadcastEventRequest) (*emptypb.Empty, error) {

	var resp emptypb.Empty

	namespace := in.GetNamespace()
	rawevent := in.GetCloudevent()

	event := new(cloudevents.Event)
	err := event.UnmarshalJSON(rawevent)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid cloudevent: %v", err)
	}

	log.Debugf("Broadcasting event on namespace '%s': %s/%s", namespace, event.Type(), event.Source())
	dlogger, err := is.wfServer.instanceLogger.NamespaceLogger(namespace)
	if err != nil {
		return nil, err
	}

	dlogger.Info(fmt.Sprintf("Broadcasting event: type=%s, source=%s", event.Type(), event.Source()))

	err = is.wfServer.handleEvent(*in.Namespace, event)

	dlogger.Close()

	return &resp, err

}
