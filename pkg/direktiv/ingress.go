package direktiv

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/ingress"
	"github.com/vorteil/direktiv/pkg/model"
	"github.com/vorteil/direktiv/pkg/secrets"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *WorkflowServer) grpcIngressStart() error {

	// TODO: make port configurable
	// TODO: save listener somewhere so that it can be shutdown
	// TODO: save grpc somewhere so that it can be shutdown
	log.Infof("ingress api starting at %v", s.config.IngressAPI.Bind)

	listener, err := net.Listen("tcp", s.config.IngressAPI.Bind)
	if err != nil {
		return err
	}

	s.grpcIngress = grpc.NewServer()

	ingress.RegisterDirektivIngressServer(s.grpcIngress, s)

	go s.grpcIngress.Serve(listener)

	return nil

}

func (s *WorkflowServer) AddNamespace(ctx context.Context, in *ingress.AddNamespaceRequest) (*ingress.AddNamespaceResponse, error) {

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

	namespace, err := s.dbManager.addNamespace(ctx, name)
	if err != nil {
		return nil, grpcDatabaseError(err, "namespace", name)
	}

	log.Debugf("Added namespace: %v", name)

	resp.Name = &name
	resp.CreatedAt = timestamppb.New(namespace.Created)

	return &resp, nil

}

func (s *WorkflowServer) AddWorkflow(ctx context.Context, in *ingress.AddWorkflowRequest) (*ingress.AddWorkflowResponse, error) {

	var resp ingress.AddWorkflowResponse

	namespace := in.GetNamespace()

	var active bool
	if in.Active != nil {
		active = *in.Active
	}

	var workflow model.Workflow
	document := in.GetWorkflow()
	err := workflow.Load(document)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "bad workflow definition: %v", err)
	}

	wf, err := s.dbManager.addWorkflow(ctx, namespace, workflow.ID,
		workflow.Description, active, document, workflow.GetStartDefinition())
	if err != nil {
		return nil, grpcDatabaseError(err, "workflow", workflow.ID)
	}

	s.tmManager.actionTimerByName(fmt.Sprintf("cron:%s", wf.ID.String()), deleteTimerAction)
	if active {
		def := workflow.GetStartDefinition()
		if def.GetType() == model.StartTypeScheduled {
			scheduled := def.(*model.ScheduledStart)
			s.tmManager.addCron(fmt.Sprintf("cron:%s", wf.ID.String()), wfCron, scheduled.Cron, []byte(wf.ID.String()))
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

func (s *WorkflowServer) BroadcastEvent(ctx context.Context, in *ingress.BroadcastEventRequest) (*emptypb.Empty, error) {

	var resp emptypb.Empty

	namespace := in.GetNamespace()
	rawevent := in.GetCloudevent()

	event := new(cloudevents.Event)
	err := event.UnmarshalJSON(rawevent)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid cloudevent: %v", err)
	}

	log.Debugf("Broadcasting event on namespace '%s': %s/%s", namespace, event.Type(), event.Source())

	err = s.handleEvent(event)

	return &resp, err

}

func (s *WorkflowServer) DeleteNamespace(ctx context.Context, in *ingress.DeleteNamespaceRequest) (*ingress.DeleteNamespaceResponse, error) {

	var resp ingress.DeleteNamespaceResponse
	var name string
	name = in.GetName()

	err := s.dbManager.deleteNamespace(ctx, name)
	if err != nil {
		return nil, grpcDatabaseError(err, "namespace", name)
	}

	log.Debugf("Deleted namespace: %v", name)

	resp.Name = &name

	return &resp, nil

}

func (s *WorkflowServer) DeleteWorkflow(ctx context.Context, in *ingress.DeleteWorkflowRequest) (*ingress.DeleteWorkflowResponse, error) {

	var resp ingress.DeleteWorkflowResponse

	uid := in.GetUid()

	err := s.dbManager.deleteWorkflow(ctx, uid)
	if err != nil {
		return nil, grpcDatabaseError(err, "workflow", uid)
	}

	s.tmManager.actionTimerByName(fmt.Sprintf("cron:%s", uid), deleteTimerAction)

	log.Debugf("Deleted workflow: %s", uid)

	resp.Uid = &uid

	return &resp, nil

}

func (s *WorkflowServer) GetNamespaces(ctx context.Context, in *ingress.GetNamespacesRequest) (*ingress.GetNamespacesResponse, error) {

	var resp ingress.GetNamespacesResponse
	offset := in.GetOffset()
	limit := in.GetLimit()

	namespaces, err := s.dbManager.getNamespaces(ctx, int(offset), int(limit))
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

func (s *WorkflowServer) GetWorkflowById(ctx context.Context, in *ingress.GetWorkflowByIdRequest) (*ingress.GetWorkflowByIdResponse, error) {

	var resp ingress.GetWorkflowByIdResponse

	namespace := in.GetNamespace()
	id := in.GetId()

	wf, err := s.dbManager.getWorkflowById(ctx, namespace, id)
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

	return &resp, nil

}

func (s *WorkflowServer) GetWorkflowByUid(ctx context.Context, in *ingress.GetWorkflowByUidRequest) (*ingress.GetWorkflowByUidResponse, error) {

	var resp ingress.GetWorkflowByUidResponse

	uid := in.GetUid()

	wf, err := s.dbManager.getWorkflowByUid(ctx, uid)
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

	return &resp, nil

}

func (s *WorkflowServer) CancelWorkflowInstance(ctx context.Context, in *ingress.CancelWorkflowInstanceRequest) (*emptypb.Empty, error) {

	_ = s.engine.hardCancelInstance(in.GetId(), "direktiv.cancels.api", "cancelled by api request")

	return nil, nil

}

func (s *WorkflowServer) GetWorkflowInstance(ctx context.Context, in *ingress.GetWorkflowInstanceRequest) (*ingress.GetWorkflowInstanceResponse, error) {

	var resp ingress.GetWorkflowInstanceResponse

	id := in.GetId()

	inst, err := s.dbManager.getWorkflowInstance(ctx, id)
	if err != nil {
		return nil, grpcDatabaseError(err, "instance", id)
	}

	rev := int32(inst.Revision)

	resp.Id = &id
	resp.Status = &inst.Status
	resp.InvokedBy = &inst.InvokedBy
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

func (s *WorkflowServer) GetWorkflowInstanceLogs(ctx context.Context, in *ingress.GetWorkflowInstanceLogsRequest) (*ingress.GetWorkflowInstanceLogsResponse, error) {

	var resp ingress.GetWorkflowInstanceLogsResponse

	instance := in.GetInstanceId()
	offset := in.GetOffset()
	limit := in.GetLimit()

	logs, err := s.instanceLogger.QueryLogs(ctx, instance, int(limit), int(offset))
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

func (s *WorkflowServer) GetWorkflowInstances(ctx context.Context, in *ingress.GetWorkflowInstancesRequest) (*ingress.GetWorkflowInstancesResponse, error) {

	var resp ingress.GetWorkflowInstancesResponse

	namespace := in.GetNamespace()
	offset := in.GetOffset()
	limit := in.GetLimit()

	instances, err := s.dbManager.getWorkflowInstances(ctx, namespace, int(offset), int(limit))
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

func (s *WorkflowServer) GetWorkflows(ctx context.Context, in *ingress.GetWorkflowsRequest) (*ingress.GetWorkflowsResponse, error) {

	var resp ingress.GetWorkflowsResponse

	namespace := in.GetNamespace()
	offset := in.GetOffset()
	limit := in.GetLimit()

	workflows, err := s.dbManager.getWorkflows(ctx, namespace, int(offset), int(limit))
	if err != nil {
		return nil, grpcDatabaseError(err, "namespace", "")
	}

	wfC, err := s.dbManager.getWorkflowCount(ctx, namespace, int(offset), int(limit))
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
		})

	}

	return &resp, nil

}

func (s *WorkflowServer) InvokeWorkflow(ctx context.Context, in *ingress.InvokeWorkflowRequest) (*ingress.InvokeWorkflowResponse, error) {

	var resp ingress.InvokeWorkflowResponse

	namespace := in.GetNamespace()
	workflow := in.GetWorkflowId()
	input := in.GetInput()

	instID, err := s.engine.DirectInvoke(namespace, workflow, input)
	if err != nil {
		return nil, grpcDatabaseError(err, "instance", fmt.Sprintf("%s/%s", namespace, workflow))
	}

	log.Debugf("Invoked workflow %s/%s: %s", namespace, workflow, instID)

	resp.InstanceId = &instID

	return &resp, nil

}

func (s *WorkflowServer) UpdateWorkflow(ctx context.Context, in *ingress.UpdateWorkflowRequest) (*ingress.UpdateWorkflowResponse, error) {

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

	wf, err := s.dbManager.updateWorkflow(ctx, uid, checkRevision, workflow.ID,
		workflow.Description, in.Active, document, workflow.GetStartDefinition())
	if err != nil {
		return nil, grpcDatabaseError(err, "workflow", workflow.ID)
	}

	s.tmManager.actionTimerByName(fmt.Sprintf("cron:%s", wf.ID.String()), deleteTimerAction)
	if *in.Active {
		def := workflow.GetStartDefinition()
		if def.GetType() == model.StartTypeScheduled {
			scheduled := def.(*model.ScheduledStart)
			s.tmManager.addCron(fmt.Sprintf("cron:%s", wf.ID.String()), wfCron, scheduled.Cron, []byte(wf.ID.String()))
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

func (s *WorkflowServer) DeleteSecret(ctx context.Context, in *ingress.DeleteSecretRequest) (*emptypb.Empty, error) {

	stype := secrets.SecretTypes_SECRET

	_, err := s.secrets.DeleteSecret(ctx, &secrets.SecretsDeleteRequest{
		Namespace: in.Namespace,
		Name:      in.Name,
		Stype:     &stype,
	})

	return &emptypb.Empty{}, err

}

func (s *WorkflowServer) DeleteRegistry(ctx context.Context, in *ingress.DeleteRegistryRequest) (*emptypb.Empty, error) {

	stype := secrets.SecretTypes_REGISTRY

	_, err := s.secrets.DeleteSecret(ctx, &secrets.SecretsDeleteRequest{
		Namespace: in.Namespace,
		Name:      in.Name,
		Stype:     &stype,
	})

	return &emptypb.Empty{}, err

}

func (s *WorkflowServer) GetSecrets(ctx context.Context, in *ingress.GetSecretsRequest) (*ingress.GetSecretsResponse, error) {

	stype := secrets.SecretTypes_SECRET

	output, err := s.secrets.GetSecrets(ctx, &secrets.GetSecretsRequest{
		Namespace: in.Namespace,
		Stype:     &stype,
	})

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

func (s *WorkflowServer) GetRegistries(ctx context.Context, in *ingress.GetRegistriesRequest) (*ingress.GetRegistriesResponse, error) {

	stype := secrets.SecretTypes_REGISTRY

	output, err := s.secrets.GetSecrets(ctx, &secrets.GetSecretsRequest{
		Namespace: in.Namespace,
		Stype:     &stype,
	})

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

func (s *WorkflowServer) StoreSecret(ctx context.Context, in *ingress.StoreSecretRequest) (*emptypb.Empty, error) {

	var resp emptypb.Empty

	ns, err := s.dbManager.getNamespace(in.GetNamespace())
	if err != nil {
		return nil, err
	}

	encryptedBytes, err := encryptData(ns.Key, in.GetData())
	if err != nil {
		return &resp, err
	}

	stype := secrets.SecretTypes_SECRET

	return s.secrets.StoreSecret(ctx, &secrets.SecretsStoreRequest{
		Namespace: in.Namespace,
		Name:      in.Name,
		Data:      encryptedBytes,
		Stype:     &stype,
	})

}

func (s *WorkflowServer) StoreRegistry(ctx context.Context, in *ingress.StoreRegistryRequest) (*emptypb.Empty, error) {

	var resp emptypb.Empty

	ns, err := s.dbManager.getNamespace(in.GetNamespace())
	if err != nil {
		return nil, err
	}

	encryptedBytes, err := encryptData(ns.Key, in.GetData())
	if err != nil {
		return &resp, err
	}

	stype := secrets.SecretTypes_REGISTRY

	return s.secrets.StoreSecret(ctx, &secrets.SecretsStoreRequest{
		Namespace: in.Namespace,
		Name:      in.Name,
		Data:      encryptedBytes,
		Stype:     &stype,
	})

}
