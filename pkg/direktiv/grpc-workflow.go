package direktiv

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"

	hash "github.com/mitchellh/hashstructure/v2"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/ingress"
	"github.com/vorteil/direktiv/pkg/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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

	// create knative services if they are not global or namespace
	err = createKnativeFunctions(is.wfServer.engine.isolateClient, workflow, namespace)
	if err != nil {
		// this can be delayed till the actual call if it fails
		log.Errorf("can not create knative functions: %v", err)
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

	go inst.start()

	if in.GetWait() {
		log.Debugf("waiting for response %v", inst.id)
		<-done
		log.Debugf("got response %v", inst.id)

		// query results here
		wfi, err := is.wfServer.dbManager.getWorkflowInstance(ctx, inst.id)
		if err != nil {
			return nil, fmt.Errorf("can not fetch instance id %v for wait request: %v", inst.id, err)
		}

		// passing the error codes back to caller
		if len(wfi.ErrorCode) > 0 || len(wfi.ErrorMessage) > 0 {
			resp.ErrorCode = &wfi.ErrorCode
			resp.ErrorMsg = &wfi.ErrorMessage
		}

		resp.Output = []byte(wfi.Output)
	}

	return &resp, nil

}

// calculates has over functions defined in workflow
func hashForFunctions(workflow model.Workflow) string {

	csnew := md5.New()
	for _, f := range workflow.GetFunctions() {
		if f.GetType() != model.ReusableContainerFunctionType {
			continue
		}
		fn, _ := json.Marshal(f)
		csnew.Write(fn)
	}
	return fmt.Sprintf("%x", csnew.Sum(nil))

}

func (is *ingressServer) UpdateWorkflow(ctx context.Context, in *ingress.UpdateWorkflowRequest) (*ingress.UpdateWorkflowResponse, error) {

	var (
		resp                   ingress.UpdateWorkflowResponse
		workflow, workflowPrev model.Workflow
	)

	uid := in.GetUid()

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

	// knative services per workflow have no revisions, we delete and recreate them
	// we need to get the checksum for the functions of the workflow
	// and used it top compare against the new version.
	fwf, _ := is.wfServer.dbManager.getWorkflowByUid(ctx, uid)
	if fwf != nil {

		// previous definition. can not have errors
		workflowPrev.Load(fwf.Workflow)

		// if the functyions have changed or the workflow has been renamed
		if workflowPrev.ID != workflow.ID ||
			hashForFunctions(workflow) != hashForFunctions(workflowPrev) {

			log.Debugf("recreating knative workflows")
			err = deleteKnativeFunctions(is.wfServer.engine.isolateClient, fwf.Edges.Namespace.ID, workflowPrev.ID, "")
			if err != nil {
				log.Errorf("can not delete knative functions: %v", err)
			}
			err = createKnativeFunctions(is.wfServer.engine.isolateClient, workflow, fwf.Edges.Namespace.ID)
			if err != nil {
				log.Errorf("can not create knative functions: %v", err)
			}
		}

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

	// get secrets and var references
	if in.GetGetReferences() {
		resp.References = &ingress.GetWorkflowByNameResponse_References{
			Secrets:   make([]*ingress.GetWorkflowByNameResponse_References_Secret, 0),
			Variables: make([]*ingress.GetWorkflowByNameResponse_References_Variable, 0),
		}

		// Load workflow
		var workflow model.Workflow
		err := workflow.Load(wf.Workflow)
		if err != nil {
			return nil, grpcDatabaseError(fmt.Errorf("invalid: %w", err), "workflow", uid)
		}

		// Get References
		secretRefs := workflow.GetSecretReferences()
		varRefs := workflow.GetVariableReferences()

		// Set References
		for i := range secretRefs {
			resp.References.Secrets = append(resp.References.Secrets, &ingress.GetWorkflowByNameResponse_References_Secret{
				Key: &secretRefs[i],
			})
		}

		for i := range varRefs {
			resp.References.Variables = append(resp.References.Variables, &ingress.GetWorkflowByNameResponse_References_Variable{
				Key:        &varRefs[i].Key,
				Scope:      &varRefs[i].Scope,
				Operations: varRefs[i].Operation,
			})
		}
	}

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

	// get secrets and var references
	if in.GetGetReferences() {
		resp.References = &ingress.GetWorkflowByUidResponse_References{
			Secrets:   make([]*ingress.GetWorkflowByUidResponse_References_Secret, 0),
			Variables: make([]*ingress.GetWorkflowByUidResponse_References_Variable, 0),
		}

		// Load workflow
		var workflow model.Workflow
		err := workflow.Load(wf.Workflow)
		if err != nil {
			return nil, grpcDatabaseError(fmt.Errorf("invalid: %w", err), "workflow", uid)
		}

		// Get References
		secretRefs := workflow.GetSecretReferences()
		varRefs := workflow.GetVariableReferences()

		// Set References
		for i := range secretRefs {
			resp.References.Secrets = append(resp.References.Secrets, &ingress.GetWorkflowByUidResponse_References_Secret{
				Key: &secretRefs[i],
			})
		}

		for i := range varRefs {
			resp.References.Variables = append(resp.References.Variables, &ingress.GetWorkflowByUidResponse_References_Variable{
				Key:        &varRefs[i].Key,
				Scope:      &varRefs[i].Scope,
				Operations: varRefs[i].Operation,
			})
		}
	}

	return &resp, nil

}
