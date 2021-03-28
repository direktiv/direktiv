package direktiv

import (
	"context"
	"fmt"
	"regexp"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/health"
	"github.com/vorteil/direktiv/pkg/ingress"
	"github.com/vorteil/direktiv/pkg/model"
	"github.com/vorteil/direktiv/pkg/secrets"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ingressServer struct {
	ingress.UnimplementedDirektivIngressServer
	health.UnimplementedHealthServer

	wfServer *WorkflowServer
	grpc     *grpc.Server

	secretsClient secrets.SecretsServiceClient
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

func newIngressServer(s *WorkflowServer) *ingressServer {

	return &ingressServer{
		wfServer: s,
	}

}

func (is *ingressServer) start(s *WorkflowServer) error {

	// get secrets client
	conn, err := getEndpointTLS(s.config, secretsComponent, s.config.SecretsAPI.Endpoint)
	if err != nil {
		return err
	}
	is.grpcConn = conn
	is.secretsClient = secrets.NewSecretsServiceClient(conn)

	return s.grpcStart(&is.grpc, "ingress", s.config.IngressAPI.Bind, func(srv *grpc.Server) {
		ingress.RegisterDirektivIngressServer(srv, is)

		log.Debugf("append health check to ingress service")
		healthServer := newHealthServer(s.config, s.isolateServer)
		health.RegisterHealthServer(srv, healthServer)
		reflection.Register(srv)
	})

}

func (is *ingressServer) AddNamespace(ctx context.Context, in *ingress.AddNamespaceRequest) (*ingress.AddNamespaceResponse, error) {

	// TODO: can go to ent
	var resp ingress.AddNamespaceResponse
	var name string
	name = in.GetName()
	regex := "^[a-z][a-z0-9._-]{1,34}[a-z0-9]$"

	matched, err := regexp.MatchString(regex, name)
	if err != nil {
		log.Errorf("%v", NewInternalError(err))
		return nil, grpcErrInternal
	}

	if !matched {
		return nil, status.Errorf(codes.InvalidArgument, "namespace name must match regex: %s", regex)
	}

	namespace, err := is.wfServer.dbManager.addNamespace(ctx, name)
	if err != nil {
		return nil, grpcDatabaseError(err, "namespace", name)
	}

	log.Debugf("Added namespace: %v", name)

	resp.Name = &name
	resp.CreatedAt = timestamppb.New(namespace.Created)

	return &resp, nil

}

func (is *ingressServer) AddWorkflow(ctx context.Context, in *ingress.AddWorkflowRequest) (*ingress.AddWorkflowResponse, error) {

	var resp ingress.AddWorkflowResponse

	namespace := in.GetNamespace()

	var active bool
	if in.Active != nil {
		active = *in.Active
	}

	var logToEvents string
	if in.LogToEvents != nil {
		logToEvents = *in.LogToEvents
	}

	var workflow model.Workflow
	document := in.GetWorkflow()
	err := workflow.Load(document)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "bad workflow definition: %v", err)
	}

	wf, err := is.wfServer.dbManager.addWorkflow(ctx, namespace, workflow.ID,
		workflow.Description, active, logToEvents, document, workflow.GetStartDefinition())
	if err != nil {
		return nil, grpcDatabaseError(err, "workflow", workflow.ID)
	}

	is.wfServer.tmManager.actionTimerByName(fmt.Sprintf("cron:%s", wf.ID.String()), deleteTimerAction)
	if active {
		def := workflow.GetStartDefinition()
		if def.GetType() == model.StartTypeScheduled {
			scheduled := def.(*model.ScheduledStart)
			is.wfServer.tmManager.addCron(fmt.Sprintf("cron:%s", wf.ID.String()), wfCron, scheduled.Cron, []byte(wf.ID.String()))
		}
	}

	uid := wf.ID.String()
	revision := int32(wf.Revision)

	log.Debugf("Added workflow %s/%s: %s", namespace, workflow.ID, uid)

	resp.Uid = &uid
	resp.Id = &wf.Name
	resp.Revision = &revision
	resp.Active = &wf.Active
	resp.CreatedAt = timestamppb.New(wf.Created)

	return &resp, nil

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

	err = is.wfServer.handleEvent(*in.Namespace, event)

	return &resp, err

}

func (is *ingressServer) DeleteNamespace(ctx context.Context, in *ingress.DeleteNamespaceRequest) (*ingress.DeleteNamespaceResponse, error) {

	var resp ingress.DeleteNamespaceResponse
	var name string
	name = in.GetName()

	err := is.wfServer.dbManager.deleteNamespace(ctx, name)
	if err != nil {
		return nil, grpcDatabaseError(err, "namespace", name)
	}

	log.Debugf("Deleted namespace: %v", name)

	resp.Name = &name

	return &resp, nil

}

func (is *ingressServer) DeleteWorkflow(ctx context.Context, in *ingress.DeleteWorkflowRequest) (*ingress.DeleteWorkflowResponse, error) {

	var resp ingress.DeleteWorkflowResponse

	uid := in.GetUid()

	err := is.wfServer.dbManager.deleteWorkflow(ctx, uid)
	if err != nil {
		return nil, grpcDatabaseError(err, "workflow", uid)
	}

	is.wfServer.tmManager.actionTimerByName(fmt.Sprintf("cron:%s", uid), deleteTimerAction)

	log.Debugf("Deleted workflow: %s", uid)

	resp.Uid = &uid

	return &resp, nil

}

func (is *ingressServer) GetNamespaces(ctx context.Context, in *ingress.GetNamespacesRequest) (*ingress.GetNamespacesResponse, error) {

	var resp ingress.GetNamespacesResponse
	offset := in.GetOffset()
	limit := in.GetLimit()

	namespaces, err := is.wfServer.dbManager.getNamespaces(ctx, int(offset), int(limit))
	if err != nil {
		return nil, grpcDatabaseError(err, "namespace", "")
	}

	resp.Offset = &offset
	resp.Limit = &limit

	for _, namespace := range namespaces {

		name := namespace.ID
		createdAt := namespace.Created

		resp.Namespaces = append(resp.Namespaces, &ingress.GetNamespacesResponse_Namespace{
			Name:      &name,
			CreatedAt: timestamppb.New(createdAt),
		})

	}

	return &resp, nil

}

func (is *ingressServer) GetWorkflowById(ctx context.Context, in *ingress.GetWorkflowByIdRequest) (*ingress.GetWorkflowByIdResponse, error) {

	var resp ingress.GetWorkflowByIdResponse

	namespace := in.GetNamespace()
	id := in.GetId()

	wf, err := is.wfServer.dbManager.getWorkflowById(ctx, namespace, id)
	if err != nil {
		return nil, grpcDatabaseError(err, "workflow", fmt.Sprintf("%s/%s", namespace, id))
	}

	uid := wf.ID.String()
	revision := int32(wf.Revision)

	resp.Uid = &uid
	resp.Id = &wf.Name
	resp.Revision = &revision
	resp.Active = &wf.Active
	resp.CreatedAt = timestamppb.New(wf.Created)
	resp.Description = &wf.Description
	resp.Workflow = wf.Workflow
	resp.LogToEvents = &wf.LogToEvents

	return &resp, nil

}

func (is *ingressServer) GetWorkflowByUid(ctx context.Context, in *ingress.GetWorkflowByUidRequest) (*ingress.GetWorkflowByUidResponse, error) {

	var resp ingress.GetWorkflowByUidResponse

	uid := in.GetUid()

	wf, err := is.wfServer.dbManager.getWorkflowByUid(ctx, uid)
	if err != nil {
		return nil, grpcDatabaseError(err, "workflow", uid)
	}

	revision := int32(wf.Revision)

	resp.Uid = &uid
	resp.Id = &wf.Name
	resp.Revision = &revision
	resp.Active = &wf.Active
	resp.CreatedAt = timestamppb.New(wf.Created)
	resp.Description = &wf.Description
	resp.Workflow = wf.Workflow
	resp.LogToEvents = &wf.LogToEvents

	return &resp, nil

}

func (is *ingressServer) CancelWorkflowInstance(ctx context.Context, in *ingress.CancelWorkflowInstanceRequest) (*emptypb.Empty, error) {

	_ = is.wfServer.engine.hardCancelInstance(in.GetId(), "direktiv.cancels.api", "cancelled by api request")

	return nil, nil

}

func (is *ingressServer) GetWorkflowInstance(ctx context.Context, in *ingress.GetWorkflowInstanceRequest) (*ingress.GetWorkflowInstanceResponse, error) {

	var resp ingress.GetWorkflowInstanceResponse

	id := in.GetId()

	inst, err := is.wfServer.dbManager.getWorkflowInstance(ctx, id)
	if err != nil {
		return nil, grpcDatabaseError(err, "instance", id)
	}

	rev := int32(inst.Revision)

	var invokedBy string
	if wfID, err := inst.QueryWorkflow().FirstID(context.Background()); err == nil {
		invokedBy = wfID.String()
	} else {
		return nil, grpcDatabaseError(err, "workflow instance", id)
	}

	resp.Id = &id
	resp.Status = &inst.Status
	resp.InvokedBy = &invokedBy
	resp.BeginTime = timestamppb.New(inst.BeginTime)
	resp.Revision = &rev

	zt := time.Time{}
	if inst.EndTime != zt {
		resp.EndTime = timestamppb.New(inst.EndTime)
	}

	resp.Flow = inst.Flow
	resp.Input = []byte(inst.Input)
	resp.Output = []byte(inst.Output)

	return &resp, nil

}

func (is *ingressServer) GetWorkflowInstanceLogs(ctx context.Context, in *ingress.GetWorkflowInstanceLogsRequest) (*ingress.GetWorkflowInstanceLogsResponse, error) {

	var resp ingress.GetWorkflowInstanceLogsResponse

	instance := in.GetInstanceId()
	offset := in.GetOffset()
	limit := in.GetLimit()

	logs, err := is.wfServer.instanceLogger.QueryLogs(ctx, instance, int(limit), int(offset))
	if err != nil {
		return nil, grpcDatabaseError(err, "instance", instance)
	}

	resp.Offset = &offset
	resp.Limit = &limit

	for i := range logs.Logs {

		l := &logs.Logs[i]

		resp.WorkflowInstanceLogs = append(resp.WorkflowInstanceLogs, &ingress.GetWorkflowInstanceLogsResponse_WorkflowInstanceLog{
			Timestamp: timestamppb.New(time.Unix(l.Timestamp, 0)),
			Message:   &l.Message,
			Context:   l.Context,
		})

	}

	return &resp, nil

}

func (is *ingressServer) GetWorkflowInstances(ctx context.Context, in *ingress.GetWorkflowInstancesRequest) (*ingress.GetWorkflowInstancesResponse, error) {

	var resp ingress.GetWorkflowInstancesResponse

	namespace := in.GetNamespace()
	offset := in.GetOffset()
	limit := in.GetLimit()

	instances, err := is.wfServer.dbManager.getWorkflowInstances(ctx, namespace, int(offset), int(limit))
	if err != nil {
		return nil, grpcDatabaseError(err, "namespace", "")
	}

	resp.Offset = &offset
	resp.Limit = &limit

	for _, inst := range instances {

		resp.WorkflowInstances = append(resp.WorkflowInstances, &ingress.GetWorkflowInstancesResponse_WorkflowInstance{
			Id:        &inst.InstanceID,
			BeginTime: timestamppb.New(inst.BeginTime),
			Status:    &inst.Status,
		})

	}

	return &resp, nil

}

func (is *ingressServer) GetWorkflows(ctx context.Context, in *ingress.GetWorkflowsRequest) (*ingress.GetWorkflowsResponse, error) {

	var resp ingress.GetWorkflowsResponse

	namespace := in.GetNamespace()
	offset := in.GetOffset()
	limit := in.GetLimit()

	workflows, err := is.wfServer.dbManager.getWorkflows(ctx, namespace, int(offset), int(limit))
	if err != nil {
		return nil, grpcDatabaseError(err, "namespace", "")
	}

	wfC, err := is.wfServer.dbManager.getWorkflowCount(ctx, namespace, int(offset), int(limit))
	if err != nil {
		return nil, grpcDatabaseError(err, "namespace", "")
	}

	count := int32(wfC)

	resp.Offset = &offset
	resp.Limit = &limit
	resp.Total = &count

	for _, wf := range workflows {

		uid := wf.ID.String()
		revision := int32(wf.Revision)

		resp.Workflows = append(resp.Workflows, &ingress.GetWorkflowsResponse_Workflow{
			Uid:         &uid,
			Id:          &wf.Name,
			Revision:    &revision,
			Description: &wf.Description,
			Active:      &wf.Active,
			CreatedAt:   timestamppb.New(wf.Created),
			LogToEvents: &wf.LogToEvents,
		})

	}

	return &resp, nil

}

func (is *ingressServer) InvokeWorkflow(ctx context.Context, in *ingress.InvokeWorkflowRequest) (*ingress.InvokeWorkflowResponse, error) {

	var resp ingress.InvokeWorkflowResponse

	namespace := in.GetNamespace()
	workflow := in.GetWorkflowId()
	input := in.GetInput()

	instID, err := is.wfServer.engine.DirectInvoke(namespace, workflow, input)
	if err != nil {
		return nil, grpcDatabaseError(err, "instance", fmt.Sprintf("%s/%s", namespace, workflow))
	}

	log.Debugf("Invoked workflow %s/%s: %s", namespace, workflow, instID)

	resp.InstanceId = &instID

	return &resp, nil

}

func (is *ingressServer) UpdateWorkflow(ctx context.Context, in *ingress.UpdateWorkflowRequest) (*ingress.UpdateWorkflowResponse, error) {

	var resp ingress.UpdateWorkflowResponse

	uid := in.GetUid()

	var workflow model.Workflow
	document := in.GetWorkflow()
	err := workflow.Load(document)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "bad workflow definition: %v", err)
	}

	var checkRevisionVal int
	var checkRevision *int
	if in.Revision != nil {
		checkRevisionVal = int(in.GetRevision())
		checkRevision = &checkRevisionVal
	}

	wf, err := is.wfServer.dbManager.updateWorkflow(ctx, uid, checkRevision, workflow.ID,
		workflow.Description, in.Active, in.LogToEvents, document, workflow.GetStartDefinition())
	if err != nil {
		return nil, grpcDatabaseError(err, "workflow", workflow.ID)
	}

	is.wfServer.tmManager.actionTimerByName(fmt.Sprintf("cron:%s", wf.ID.String()), deleteTimerAction)
	if *in.Active {
		def := workflow.GetStartDefinition()
		if def.GetType() == model.StartTypeScheduled {
			scheduled := def.(*model.ScheduledStart)
			is.wfServer.tmManager.addCron(fmt.Sprintf("cron:%s", wf.ID.String()), wfCron, scheduled.Cron, []byte(wf.ID.String()))
		}
	}

	revision := int32(wf.Revision)

	log.Debugf("Updated workflow: %s", uid)

	resp.Uid = &uid
	resp.Id = &wf.Name
	resp.Revision = &revision
	resp.Active = &wf.Active
	resp.CreatedAt = timestamppb.New(wf.Created)

	return &resp, nil

}

type deleteEncryptedRequest interface {
	GetNamespace() string
	GetName() string
}

func (is *ingressServer) deleteEncrypted(ctx context.Context, in deleteEncryptedRequest, stype secrets.SecretTypes) error {

	namespace := in.GetNamespace()
	name := in.GetName()

	_, err := is.secretsClient.DeleteSecret(ctx, &secrets.SecretsDeleteRequest{
		Namespace: &namespace,
		Name:      &name,
		Stype:     &stype,
	})

	return err

}

func (is *ingressServer) DeleteSecret(ctx context.Context, in *ingress.DeleteSecretRequest) (*emptypb.Empty, error) {
	err := is.deleteEncrypted(ctx, in, secrets.SecretTypes_SECRET)
	return &emptypb.Empty{}, err
}

func (is *ingressServer) DeleteRegistry(ctx context.Context, in *ingress.DeleteRegistryRequest) (*emptypb.Empty, error) {
	err := is.deleteEncrypted(ctx, in, secrets.SecretTypes_REGISTRY)
	return &emptypb.Empty{}, err
}

func (is *ingressServer) fetchSecrets(ctx context.Context, ns string,
	stype secrets.SecretTypes) (*secrets.GetSecretsResponse, error) {

	return is.secretsClient.GetSecrets(ctx, &secrets.GetSecretsRequest{
		Namespace: &ns,
		Stype:     &stype,
	})

}

func (is *ingressServer) GetSecrets(ctx context.Context, in *ingress.GetSecretsRequest) (*ingress.GetSecretsResponse, error) {

	output, err := is.fetchSecrets(ctx, in.GetNamespace(), secrets.SecretTypes_SECRET)
	if err != nil {
		return nil, err
	}

	resp := new(ingress.GetSecretsResponse)
	for i := range output.Secrets {
		resp.Secrets = append(resp.Secrets, &ingress.GetSecretsResponse_Secret{
			Name: output.Secrets[i].Name,
		})
	}

	return resp, nil

}

func (is *ingressServer) GetRegistries(ctx context.Context, in *ingress.GetRegistriesRequest) (*ingress.GetRegistriesResponse, error) {

	output, err := is.fetchSecrets(ctx, in.GetNamespace(), secrets.SecretTypes_REGISTRY)
	if err != nil {
		return nil, err
	}

	resp := new(ingress.GetRegistriesResponse)
	for i := range output.Secrets {
		resp.Registries = append(resp.Registries, &ingress.GetRegistriesResponse_Registry{
			Name: output.Secrets[i].Name,
		})
	}

	return resp, nil

}

type storeEncryptedRequest interface {
	GetNamespace() string
	GetName() string
	GetData() []byte
}

func (is *ingressServer) storeEncrypted(ctx context.Context, in storeEncryptedRequest, stype secrets.SecretTypes) error {

	ns, err := is.wfServer.dbManager.getNamespace(in.GetNamespace())
	if err != nil {
		return err
	}

	encryptedBytes, err := encryptData(ns.Key, in.GetData())
	if err != nil {
		return err
	}

	namespace := in.GetNamespace()
	name := in.GetName()

	_, err = is.secretsClient.StoreSecret(ctx, &secrets.SecretsStoreRequest{
		Namespace: &namespace,
		Name:      &name,
		Data:      encryptedBytes,
		Stype:     &stype,
	})

	return err

}

func (is *ingressServer) StoreSecret(ctx context.Context, in *ingress.StoreSecretRequest) (*emptypb.Empty, error) {
	var resp emptypb.Empty
	return &resp, is.storeEncrypted(ctx, in, secrets.SecretTypes_SECRET)
}

func (is *ingressServer) StoreRegistry(ctx context.Context, in *ingress.StoreRegistryRequest) (*emptypb.Empty, error) {
	var resp emptypb.Empty
	return &resp, is.storeEncrypted(ctx, in, secrets.SecretTypes_REGISTRY)
}
