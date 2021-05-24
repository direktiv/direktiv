package direktiv

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"encoding/base64"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	hash "github.com/mitchellh/hashstructure/v2"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/ent"
	"github.com/vorteil/direktiv/pkg/health"
	"github.com/vorteil/direktiv/pkg/ingress"
	"github.com/vorteil/direktiv/pkg/model"
	secretsgrpc "github.com/vorteil/direktiv/pkg/secrets/grpc"
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
	conn, err := GetEndpointTLS(secretsEndpoint, false)
	if err != nil {
		return err
	}
	is.grpcConn = conn
	is.secretsClient = secretsgrpc.NewSecretsServiceClient(conn)

	is.cronPoll()
	go is.cronPoller()

	return GrpcStart(&is.grpc, "ingress", s.config.IngressAPI.Bind, func(srv *grpc.Server) {
		ingress.RegisterDirektivIngressServer(srv, is)

		log.Debugf("append health check to ingress service")
		healthServer := newHealthServer()
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

	is.wfServer.tmManager.deleteTimerByName("", "", fmt.Sprintf("cron:%s", wf.ID.String()))
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

	var (
		resp ingress.DeleteWorkflowResponse
	)
	uid := in.GetUid()

	err := is.wfServer.dbManager.deleteWorkflow(ctx, uid)
	if err != nil {
		return nil, grpcDatabaseError(err, "workflow", uid)
	}

	err = is.wfServer.tmManager.deleteTimerByName("", "", fmt.Sprintf("cron:%s", uid))
	if err != nil {
		log.Error(err)
	}

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

func (is *ingressServer) GetWorkflowByName(ctx context.Context, in *ingress.GetWorkflowByNameRequest) (*ingress.GetWorkflowByNameResponse, error) {

	var resp ingress.GetWorkflowByNameResponse

	namespace := in.GetNamespace()
	name := in.GetName()

	wf, err := is.wfServer.dbManager.getWorkflowByName(ctx, namespace, name)
	if err != nil {
		return nil, grpcDatabaseError(err, "workflow", fmt.Sprintf("%s/%s", namespace, name))
	}

	uid := wf.ID.String()
	revision := int32(wf.Revision)

	resp.Uid = &uid
	resp.Name = &wf.Name
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

	err := is.wfServer.engine.hardCancelInstance(in.GetId(), "direktiv.cancels.api", "cancelled by api request")
	if err != nil {
		log.Errorf("error cancelling instance: %v", err)
	}
	return &emptypb.Empty{}, nil

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
	if wfID, err := inst.QueryWorkflow().FirstID(ctx); err == nil {
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

	resp.ErrorCode = &inst.ErrorCode
	resp.ErrorMessage = &inst.ErrorMessage

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
			Timestamp: timestamppb.New(time.Unix(0, l.Timestamp)),
			Message:   &l.Message,
			Context:   l.Context,
		})

	}

	return &resp, nil

}

func (is *ingressServer) GetInstancesByWorkflow(ctx context.Context, in *ingress.GetInstancesByWorkflowRequest) (*ingress.GetInstancesByWorkflowResponse, error) {

	var resp ingress.GetInstancesByWorkflowResponse

	namespace := in.GetNamespace()
	workflow := in.GetWorkflow()
	offset := in.GetOffset()
	limit := in.GetLimit()

	workflowUID, err := is.wfServer.dbManager.getWorkflowByName(ctx, namespace, workflow)
	if err != nil {
		return nil, err
	}

	instances, err := is.wfServer.dbManager.getWorkflowInstancesByWFID(ctx, workflowUID.ID, int(offset), int(limit))
	if err != nil {
		return nil, err
	}

	resp.Offset = &offset
	resp.Limit = &limit

	for _, inst := range instances {

		resp.WorkflowInstances = append(resp.WorkflowInstances, &ingress.GetInstancesByWorkflowResponse_WorkflowInstance{
			Id:        &inst.InstanceID,
			BeginTime: timestamppb.New(inst.BeginTime),
			Status:    &inst.Status,
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
	workflow := in.GetName()
	input := in.GetInput()

	inst, err := is.wfServer.engine.PrepareInvoke(ctx, namespace, workflow, input)
	if err != nil {
		return nil, grpcDatabaseError(err, "instance", fmt.Sprintf("%s/%s", namespace, workflow))
	}

	log.Debugf("Invoked workflow %s/%s: %s", namespace, workflow, inst.id)

	resp.InstanceId = &inst.id

	done := make(chan bool)
	defer close(done)

	// the workflow started, check if we need to wait
	// wait sends to chan -> sub ready
	if in.GetWait() {
		h, _ := hash.Hash(fmt.Sprintf("%s", inst.id), hash.FormatV2, nil)
		go syncAPIWait(is.wfServer.config.Database.DB, fmt.Sprintf("api:%d", h), done)
		<-done
	}

	go inst.start(ctx)

	if in.GetWait() {
		log.Debugf("waiting for response %v", inst.id)
		<-done
		log.Debugf("got response %v", inst.id)

		// query results here
		wfi, err := is.wfServer.dbManager.getWorkflowInstance(ctx, inst.id)
		if err != nil {
			return nil, fmt.Errorf("can not fetch instance id %v for wait request: %v", inst.id, err)
		}
		resp.Output = []byte(wfi.Output)
	}

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

	is.wfServer.tmManager.deleteTimerByName("", "", fmt.Sprintf("cron:%s", wf.ID.String()))
	if wf.Active {
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

func (is *ingressServer) DeleteSecret(ctx context.Context, in *ingress.DeleteSecretRequest) (*emptypb.Empty, error) {

	namespace := in.GetNamespace()
	name := in.GetName()

	_, err := is.secretsClient.DeleteSecret(ctx, &secretsgrpc.SecretsDeleteRequest{
		Namespace: &namespace,
		Name:      &name,
	})

	return &emptypb.Empty{}, err

}

func (is *ingressServer) DeleteRegistry(ctx context.Context, in *ingress.DeleteRegistryRequest) (*emptypb.Empty, error) {
	var resp emptypb.Empty

	err := kubernetesDeleteSecret(in.GetName(), in.GetNamespace())

	return &resp, err
}

func (is *ingressServer) fetchSecrets(ctx context.Context, ns string) (*secretsgrpc.GetSecretsResponse, error) {

	return is.secretsClient.GetSecrets(ctx, &secretsgrpc.GetSecretsRequest{
		Namespace: &ns,
	})

}

func (is *ingressServer) GetSecrets(ctx context.Context, in *ingress.GetSecretsRequest) (*ingress.GetSecretsResponse, error) {

	output, err := is.fetchSecrets(ctx, in.GetNamespace())
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

	resp := new(ingress.GetRegistriesResponse)

	regs, err := kubernetesListRegistries(in.GetNamespace())

	if err != nil {
		return resp, err
	}

	for _, reg := range regs {
		split := strings.SplitN(reg, "###", 2)

		if len(split) != 2 {
			return nil, fmt.Errorf("invalid registry format")
		}

		resp.Registries = append(resp.Registries, &ingress.GetRegistriesResponse_Registry{
			Name: &split[0],
			Id:   &split[1],
		})
	}

	return resp, nil

}

type storeEncryptedRequest interface {
	GetNamespace() string
	GetName() string
	GetData() []byte
}

func (is *ingressServer) StoreSecret(ctx context.Context, in *ingress.StoreSecretRequest) (*emptypb.Empty, error) {
	var resp emptypb.Empty

	namespace := in.GetNamespace()
	name := in.GetName()

	_, err := is.secretsClient.StoreSecret(ctx, &secretsgrpc.SecretsStoreRequest{
		Namespace: &namespace,
		Name:      &name,
		Data:      in.GetData(),
	})

	return &resp, err
}

func (is *ingressServer) StoreRegistry(ctx context.Context, in *ingress.StoreRegistryRequest) (*emptypb.Empty, error) {
	var resp emptypb.Empty

	// create secret data, needs to be attached to service account
	userToken := strings.SplitN(string(in.Data), ":", 2)
	if len(userToken) != 2 {
		return nil, fmt.Errorf("invalid username/token format")
	}

	tmpl := `{
	"auths": {
		"%s": {
			"username": "%s",
			"password": "%s",
			"auth": "%s"
		}
	}
	}`

	auth := fmt.Sprintf(tmpl, in.GetName(), userToken[0], userToken[1],
		base64.StdEncoding.EncodeToString(in.Data))

	err := kubernetesAddSecret(in.GetName(), in.GetNamespace(), []byte(auth))
	if err != nil {
		return nil, err
	}

	return &resp, nil

}

func (is *ingressServer) SetNamespaceVariable(srv ingress.DirektivIngress_SetNamespaceVariableServer) error {

	ctx := srv.Context()

	in, err := srv.Recv()
	if err != nil {
		return err
	}

	namespace := in.GetNamespace()
	if namespace == "" {
		return errors.New("required namespace")
	}

	key := in.GetKey()
	if key == "" {
		return errors.New("requires variable key")
	}

	totalSize := in.GetTotalSize()

	_, err = is.wfServer.dbManager.getNamespace(namespace)
	if err != nil {
		return err
	}

	if totalSize == 0 {
		err = is.wfServer.variableStorage.Delete(ctx, key, namespace)
		if err != nil {
			return err
		}
	} else {
		w, err := is.wfServer.variableStorage.Store(ctx, key, namespace)
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

func (is *ingressServer) GetNamespaceVariable(in *ingress.GetNamespaceVariableRequest, out ingress.DirektivIngress_GetNamespaceVariableServer) error {

	ctx := out.Context()

	namespace := in.GetNamespace()
	if namespace == "" {
		return errors.New("required namespace")
	}

	key := in.GetKey()
	if key == "" {
		return errors.New("requires variable key")
	}

	r, err := is.wfServer.variableStorage.Retrieve(ctx, key, namespace)
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
		resp := new(ingress.GetNamespaceVariableResponse)
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

func (is *ingressServer) ListNamespaceVariables(ctx context.Context, in *ingress.ListNamespaceVariablesRequest) (*ingress.ListNamespaceVariablesResponse, error) {

	resp := new(ingress.ListNamespaceVariablesResponse)

	namespace := in.GetNamespace()
	if namespace == "" {
		return nil, errors.New("required namespace")
	}

	list, err := is.wfServer.variableStorage.List(ctx, namespace)
	if err != nil {
		return nil, err
	}

	var names []string
	var sizes []int64

	for i, tuple := range list {
		names = append(names, tuple.Key())
		sizes = append(sizes, tuple.Size())
		resp.Variables = append(resp.Variables, &ingress.ListNamespaceVariablesResponse_Variable{
			Name: &(names[i]),
			Size: &(sizes[i]),
		})
	}

	return resp, nil

}

func (is *ingressServer) SetWorkflowVariable(srv ingress.DirektivIngress_SetWorkflowVariableServer) error {

	ctx := srv.Context()

	in, err := srv.Recv()
	if err != nil {
		return err
	}

	workflow := in.GetWorkflowUid()
	if workflow == "" {
		return errors.New("required workflow uid")
	}

	key := in.GetKey()
	if key == "" {
		return errors.New("requires variable key")
	}

	totalSize := in.GetTotalSize()

	wf, err := is.wfServer.dbManager.getWorkflowByUid(ctx, workflow)
	if err != nil {
		return err
	}

	namespace := wf.Edges.Namespace.ID
	wfId := wf.ID.String()

	if totalSize == 0 {
		err = is.wfServer.variableStorage.Delete(ctx, key, namespace, wfId)
		if err != nil {
			return err
		}
	} else {
		w, err := is.wfServer.variableStorage.Store(ctx, key, namespace, wfId)
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

func (is *ingressServer) GetWorkflowVariable(in *ingress.GetWorkflowVariableRequest, out ingress.DirektivIngress_GetWorkflowVariableServer) error {

	ctx := out.Context()

	workflow := in.GetWorkflowUid()
	if workflow == "" {
		return errors.New("required workflow uid")
	}

	key := in.GetKey()
	if key == "" {
		return errors.New("requires variable key")
	}

	wf, err := is.wfServer.dbManager.getWorkflowByUid(ctx, workflow)
	if err != nil {
		return err
	}

	ns := wf.Edges.Namespace

	r, err := is.wfServer.variableStorage.Retrieve(ctx, key, ns.ID, workflow)
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
		resp := new(ingress.GetWorkflowVariableResponse)
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

func (is *ingressServer) ListWorkflowVariables(ctx context.Context, in *ingress.ListWorkflowVariablesRequest) (*ingress.ListWorkflowVariablesResponse, error) {

	resp := new(ingress.ListWorkflowVariablesResponse)

	workflow := in.GetWorkflowUid()
	if workflow == "" {
		return nil, errors.New("required workflow uid")
	}

	wf, err := is.wfServer.dbManager.getWorkflowByUid(ctx, workflow)
	if err != nil {
		return nil, err
	}

	ns := wf.Edges.Namespace

	list, err := is.wfServer.variableStorage.List(ctx, ns.ID, workflow)
	if err != nil {
		return nil, err
	}

	var names []string
	var sizes []int64

	for i, tuple := range list {
		names = append(names, tuple.Key())
		sizes = append(sizes, tuple.Size())
		resp.Variables = append(resp.Variables, &ingress.ListWorkflowVariablesResponse_Variable{
			Name: &(names[i]),
			Size: &(sizes[i]),
		})
	}

	return resp, nil

}
