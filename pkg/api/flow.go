package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
	"github.com/vorteil/direktiv/pkg/util"
	"go.uber.org/zap"
)

type flowHandler struct {
	logger *zap.SugaredLogger
	client grpc.FlowClient
}

func newFlowHandler(logger *zap.SugaredLogger, router *mux.Router, addr string) (*flowHandler, error) {

	flowAddr := fmt.Sprintf("%s:6666", addr)
	logger.Infof("connecting to flow %s", flowAddr)

	conn, err := util.GetEndpointTLS(flowAddr)
	if err != nil {
		logger.Errorf("can not connect to direktiv flows: %v", err)
		return nil, err
	}

	h := &flowHandler{
		logger: logger,
		client: grpc.NewFlowClient(conn),
	}

	h.initRoutes(router)

	return h, nil

}

func (h *flowHandler) initRoutes(r *mux.Router) {

	handlerPair(r, RN_ListNamespaces, "/namespaces", h.Namespaces, h.NamespacesSSE)
	r.HandleFunc("/namespaces", h.CreateNamespace).Name(RN_AddNamespace).Methods(http.MethodPost)
	r.HandleFunc("/namespaces/{ns}", h.DeleteNamespace).Name(RN_DeleteNamespace).Methods(http.MethodDelete)

	r.HandleFunc("/jq", h.JQ).Name(RN_JQPlayground).Methods(http.MethodPost)
	handlerPair(r, RN_GetServerLogs, "/logs", h.ServerLogs, h.ServerLogsSSE)
	handlerPair(r, RN_GetNamespaceLogs, "/namespaces/{ns}/logs", h.NamespaceLogs, h.NamespaceLogsSSE)
	handlerPair(r, RN_GetInstanceLogs, "/namespaces/{ns}/instances/{in}/logs", h.InstanceLogs, h.InstanceLogsSSE)

	r.HandleFunc("/namespaces/{ns}/vars/{var}", h.NamespaceVariable).Name(RN_GetNamespaceVariable).Methods(http.MethodGet)
	r.HandleFunc("/namespaces/{ns}/vars/{var}", h.DeleteNamespaceVariable).Name(RN_SetNamespaceVariable).Methods(http.MethodDelete)
	r.HandleFunc("/namespaces/{ns}/vars/{var}", h.SetNamespaceVariable).Name(RN_SetNamespaceVariable).Methods(http.MethodPut)
	handlerPair(r, RN_ListNamespaceVariables, "/namespaces/{ns}/vars", h.NamespaceVariables, h.NamespaceVariablesSSE)

	r.HandleFunc("/namespaces/{ns}/instances/{instance}/vars/{var}", h.InstanceVariable).Name(RN_GetInstanceVariable).Methods(http.MethodGet)
	r.HandleFunc("/namespaces/{ns}/instances/{instance}/vars/{var}", h.DeleteInstanceVariable).Name(RN_SetInstanceVariable).Methods(http.MethodDelete)
	r.HandleFunc("/namespaces/{ns}/instances/{instance}/vars/{var}", h.SetInstanceVariable).Name(RN_SetInstanceVariable).Methods(http.MethodPut)
	handlerPair(r, RN_ListInstanceVariables, "/namespaces/{ns}/instances/{instance}/vars", h.InstanceVariables, h.InstanceVariablesSSE)

	pathHandlerPair(r, RN_ListWorkflowVariables, "vars", h.WorkflowVariables, h.WorkflowVariablesSSE)
	pathHandler(r, http.MethodPut, RN_SetWorkflowVariable, "set-var", h.SetWorkflowVariable)
	pathHandler(r, http.MethodDelete, RN_SetWorkflowVariable, "delete-var", h.DeleteWorkflowVariable)
	pathHandler(r, http.MethodGet, RN_GetWorkflowVariable, "var", h.WorkflowVariable)

	handlerPair(r, RN_ListSecrets, "/namespaces/{ns}/secrets", h.Secrets, h.SecretsSSE)
	r.HandleFunc("/namespaces/{ns}/secrets/{secret}", h.SetSecret).Name(RN_CreateSecret).Methods(http.MethodPut)
	r.HandleFunc("/namespaces/{ns}/secrets/{secret}", h.DeleteSecret).Name(RN_DeleteSecret).Methods(http.MethodDelete)

	handlerPair(r, RN_GetInstance, "/namespaces/{ns}/instances/{instance}", h.Instance, h.InstanceSSE)
	handlerPair(r, RN_ListInstances, "/namespaces/{ns}/instances", h.Instances, h.InstancesSSE)
	r.HandleFunc("/namespaces/{ns}/instances/{instance}/input", h.InstanceInput).Name(RN_GetInstance).Methods(http.MethodGet)
	r.HandleFunc("/namespaces/{ns}/instances/{instance}/output", h.InstanceOutput).Name(RN_GetInstance).Methods(http.MethodGet)
	r.HandleFunc("/namespaces/{ns}/instances/{instance}/cancel", h.InstanceCancel).Name(RN_CancelInstance).Methods(http.MethodPost)

	r.HandleFunc("/namespaces/{ns}/broadcast", h.BroadcastCloudevent).Name(RN_NamespaceEvent).Methods(http.MethodPost)

	pathHandlerPair(r, RN_GetWorkflowLogs, "logs", h.WorkflowLogs, h.WorkflowLogsSSE)
	pathHandler(r, http.MethodPut, RN_CreateDirectory, "create-directory", h.CreateDirectory)
	pathHandler(r, http.MethodPut, RN_CreateWorkflow, "create-workflow", h.CreateWorkflow)
	pathHandler(r, http.MethodPost, RN_UpdateWorkflow, "update-workflow", h.UpdateWorkflow)
	pathHandler(r, http.MethodPost, RN_SaveWorkflow, "save-workflow", h.SaveWorkflow)
	pathHandler(r, http.MethodPost, RN_DiscardWorkflow, "discard-workflow", h.DiscardWorkflow)
	pathHandler(r, http.MethodDelete, RN_DeleteNode, "delete-node", h.DeleteNode)
	pathHandlerPair(r, RN_GetWorkflowTags, "tags", h.GetTags, h.GetTagsSSE)
	pathHandlerPair(r, RN_GetWorkflowRefs, "refs", h.GetRefs, h.GetRefsSSE)
	pathHandlerPair(r, RN_GetWorkflowRefs, "revisions", h.GetRevisions, h.GetRevisionsSSE)
	pathHandler(r, http.MethodPost, RN_DeleteRevision, "delete-revision", h.DeleteRevision)
	pathHandler(r, http.MethodPost, RN_Tag, "tag", h.Tag)
	pathHandler(r, http.MethodPost, RN_Untag, "untag", h.Untag)
	pathHandler(r, http.MethodPost, RN_Retag, "retag", h.Retag)
	pathHandlerPair(r, RN_GetWorkflowRouter, "router", h.Router, h.RouterSSE)
	pathHandler(r, http.MethodPost, RN_EditWorkflowRouter, "edit-router", h.EditRouter)
	pathHandler(r, http.MethodPost, RN_ValidateRef, "validate-ref", h.ValidateRef)
	pathHandler(r, http.MethodPost, RN_ValidateRouter, "validate-router", h.ValidateRouter)

	pathHandler(r, http.MethodPost, RN_UpdateWorkflow, "set-workflow-event-logging", h.SetWorkflowEventLogging)
	pathHandler(r, http.MethodPost, RN_UpdateWorkflow, "toggle", h.ToggleWorkflow)

	pathHandler(r, http.MethodPost, RN_ExecuteWorkflow, "execute", h.ExecuteWorkflow)

	pathHandlerPair(r, RN_GetNode, "", h.GetNode, h.GetNodeSSE)

}

func (h *flowHandler) Namespaces(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()

	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.NamespacesRequest{
		Pagination: p,
	}

	resp, err := h.client.Namespaces(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) NamespacesSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()

	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.NamespacesRequest{
		Pagination: p,
	}

	resp, err := h.client.NamespacesStream(ctx, in)
	if err != nil {
		respond(w, resp, err)
		return
	}

	ch := make(chan interface{}, 1)

	defer func() {

		_ = resp.CloseSend()

		for {
			_, more := <-ch
			if !more {
				return
			}
		}

	}()

	go func() {

		defer close(ch)

		for {

			x, err := resp.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x

		}

	}()

	sse(w, ch)

}

func (h *flowHandler) CreateNamespace(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()

	in := &grpc.CreateNamespaceRequest{}

	err := unmarshalBody(r, in)
	if err != nil {
		badRequest(w, err)
		return
	}

	resp, err := h.client.CreateNamespace(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) DeleteNamespace(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]

	in := &grpc.DeleteNamespaceRequest{
		Name: namespace,
	}

	resp, err := h.client.DeleteNamespace(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) ServerLogs(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()

	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.ServerLogsRequest{
		Pagination: p,
	}

	resp, err := h.client.ServerLogs(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) ServerLogsSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()

	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.ServerLogsRequest{
		Pagination: p,
	}

	resp, err := h.client.ServerLogsParcels(ctx, in)
	if err != nil {
		respond(w, resp, err)
		return
	}

	ch := make(chan interface{}, 1)

	defer func() {

		_ = resp.CloseSend()

		for {
			_, more := <-ch
			if !more {
				return
			}
		}

	}()

	go func() {

		defer close(ch)

		for {

			x, err := resp.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x

		}

	}()

	sse(w, ch)

}

func (h *flowHandler) NamespaceLogs(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]

	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.NamespaceLogsRequest{
		Pagination: p,
		Namespace:  namespace,
	}

	resp, err := h.client.NamespaceLogs(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) NamespaceLogsSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]

	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.NamespaceLogsRequest{
		Pagination: p,
		Namespace:  namespace,
	}

	resp, err := h.client.NamespaceLogsParcels(ctx, in)
	if err != nil {
		respond(w, resp, err)
		return
	}

	ch := make(chan interface{}, 1)

	defer func() {

		_ = resp.CloseSend()

		for {
			_, more := <-ch
			if !more {
				return
			}
		}

	}()

	go func() {

		defer close(ch)

		for {

			x, err := resp.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x

		}

	}()

	sse(w, ch)

}

func (h *flowHandler) WorkflowLogs(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.WorkflowLogsRequest{
		Pagination: p,
		Namespace:  namespace,
		Path:       path,
	}

	resp, err := h.client.WorkflowLogs(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) WorkflowLogsSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.WorkflowLogsRequest{
		Pagination: p,
		Namespace:  namespace,
		Path:       path,
	}

	resp, err := h.client.WorkflowLogsParcels(ctx, in)
	if err != nil {
		respond(w, resp, err)
		return
	}

	ch := make(chan interface{}, 1)

	defer func() {

		_ = resp.CloseSend()

		for {
			_, more := <-ch
			if !more {
				return
			}
		}

	}()

	go func() {

		defer close(ch)

		for {

			x, err := resp.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x

		}

	}()

	sse(w, ch)

}

func (h *flowHandler) InstanceLogs(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	instance := mux.Vars(r)["in"]

	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.InstanceLogsRequest{
		Pagination: p,
		Namespace:  namespace,
		Instance:   instance,
	}

	resp, err := h.client.InstanceLogs(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) InstanceLogsSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	instance := mux.Vars(r)["in"]

	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.InstanceLogsRequest{
		Pagination: p,
		Namespace:  namespace,
		Instance:   instance,
	}

	resp, err := h.client.InstanceLogsParcels(ctx, in)
	if err != nil {
		respond(w, resp, err)
		return
	}

	ch := make(chan interface{}, 1)

	defer func() {

		_ = resp.CloseSend()

		for {
			_, more := <-ch
			if !more {
				return
			}
		}

	}()

	go func() {

		defer close(ch)

		for {

			x, err := resp.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x

		}

	}()

	sse(w, ch)

}

func (h *flowHandler) GetNode(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, ref := pathAndRef(r)

	var resp interface{}
	var err error
	var p *grpc.Pagination

	if ref != "" {
		goto workflow
	}

	p, err = pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	{
		node, err := h.client.Node(ctx, &grpc.NodeRequest{
			Namespace: namespace,
			Path:      path,
		})
		if err != nil {
			respond(w, node, err)
			return
		}

		switch node.Node.Type {
		case "directory":
			goto directory
		case "workflow":
			goto workflow
		}
	}

directory:

	resp, err = h.client.Directory(ctx, &grpc.DirectoryRequest{
		Pagination: p,
		Namespace:  namespace,
		Path:       path,
	})
	respond(w, resp, err)
	return

workflow:

	resp, err = h.client.Workflow(ctx, &grpc.WorkflowRequest{
		Namespace: namespace,
		Path:      path,
		Ref:       ref,
	})

	respond(w, resp, err)
	return

}

func (h *flowHandler) GetNodeSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, ref := pathAndRef(r)

	var err error
	var p *grpc.Pagination
	var ch chan interface{}
	var dirc grpc.Flow_DirectoryStreamClient
	var wfc grpc.Flow_WorkflowStreamClient

	p, err = pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	{
		node, err := h.client.Node(ctx, &grpc.NodeRequest{
			Namespace: namespace,
			Path:      path,
		})
		if err != nil {
			respond(w, node, err)
			return
		}

		switch node.Node.Type {
		case "directory":
			goto directory
		case "workflow":
			goto workflow
		}
	}

directory:

	dirc, err = h.client.DirectoryStream(ctx, &grpc.DirectoryRequest{
		Pagination: p,
		Namespace:  namespace,
		Path:       path,
	})

	if err != nil {
		respond(w, nil, err)
		return
	}

	ch = make(chan interface{}, 1)

	defer func() {

		_ = dirc.CloseSend()

		for {
			_, more := <-ch
			if !more {
				return
			}
		}

	}()

	go func() {

		defer close(ch)

		for {

			x, err := dirc.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x

		}

	}()

	sse(w, ch)
	return

workflow:

	wfc, err = h.client.WorkflowStream(ctx, &grpc.WorkflowRequest{
		Namespace: namespace,
		Path:      path,
		Ref:       ref,
	})
	if err != nil {
		respond(w, nil, err)
		return
	}

	ch = make(chan interface{}, 1)

	defer func() {

		_ = wfc.CloseSend()

		for {
			_, more := <-ch
			if !more {
				return
			}
		}

	}()

	go func() {

		defer close(ch)

		for {

			x, err := wfc.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x

		}

	}()

	sse(w, ch)
	return

}

func (h *flowHandler) CreateDirectory(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	in := &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      path,
	}

	resp, err := h.client.CreateDirectory(ctx, in)
	respond(w, resp, err)
	return

}

func (h *flowHandler) CreateWorkflow(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	data, err := loadRawBody(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      path,
		Source:    data,
	}

	resp, err := h.client.CreateWorkflow(ctx, in)
	respond(w, resp, err)
	return

}

func (h *flowHandler) UpdateWorkflow(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	data, err := loadRawBody(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := &grpc.UpdateWorkflowRequest{
		Namespace: namespace,
		Path:      path,
		Source:    data,
	}

	resp, err := h.client.UpdateWorkflow(ctx, in)
	respond(w, resp, err)
	return

}

func (h *flowHandler) SaveWorkflow(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	in := &grpc.SaveHeadRequest{
		Namespace: namespace,
		Path:      path,
	}

	resp, err := h.client.SaveHead(ctx, in)
	respond(w, resp, err)
	return

}

func (h *flowHandler) DiscardWorkflow(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	in := &grpc.DiscardHeadRequest{
		Namespace: namespace,
		Path:      path,
	}

	resp, err := h.client.DiscardHead(ctx, in)
	respond(w, resp, err)
	return

}

func (h *flowHandler) DeleteNode(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	in := &grpc.DeleteNodeRequest{
		Namespace: namespace,
		Path:      path,
	}

	resp, err := h.client.DeleteNode(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) GetTags(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.TagsRequest{
		Pagination: p,
		Namespace:  namespace,
		Path:       path,
	}

	resp, err := h.client.Tags(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) GetTagsSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.TagsRequest{
		Pagination: p,
		Namespace:  namespace,
		Path:       path,
	}

	resp, err := h.client.TagsStream(ctx, in)
	if err != nil {
		respond(w, resp, err)
		return
	}

	ch := make(chan interface{}, 1)

	defer func() {

		_ = resp.CloseSend()

		for {
			_, more := <-ch
			if !more {
				return
			}
		}

	}()

	go func() {

		defer close(ch)

		for {

			x, err := resp.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x

		}

	}()

	sse(w, ch)

}

func (h *flowHandler) GetRefs(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.RefsRequest{
		Pagination: p,
		Namespace:  namespace,
		Path:       path,
	}

	resp, err := h.client.Refs(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) GetRefsSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.RefsRequest{
		Pagination: p,
		Namespace:  namespace,
		Path:       path,
	}

	resp, err := h.client.RefsStream(ctx, in)
	if err != nil {
		respond(w, resp, err)
		return
	}

	ch := make(chan interface{}, 1)

	defer func() {

		_ = resp.CloseSend()

		for {
			_, more := <-ch
			if !more {
				return
			}
		}

	}()

	go func() {

		defer close(ch)

		for {

			x, err := resp.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x

		}

	}()

	sse(w, ch)

}

func (h *flowHandler) GetRevisions(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.RevisionsRequest{
		Pagination: p,
		Namespace:  namespace,
		Path:       path,
	}

	resp, err := h.client.Revisions(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) GetRevisionsSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.RevisionsRequest{
		Pagination: p,
		Namespace:  namespace,
		Path:       path,
	}

	resp, err := h.client.RevisionsStream(ctx, in)
	if err != nil {
		respond(w, resp, err)
		return
	}

	ch := make(chan interface{}, 1)

	defer func() {

		_ = resp.CloseSend()

		for {
			_, more := <-ch
			if !more {
				return
			}
		}

	}()

	go func() {

		defer close(ch)

		for {

			x, err := resp.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x

		}

	}()

	sse(w, ch)

}

func (h *flowHandler) DeleteRevision(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, ref := pathAndRef(r)

	in := &grpc.DeleteRevisionRequest{
		Namespace: namespace,
		Path:      path,
		Revision:  ref,
	}

	resp, err := h.client.DeleteRevision(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) Tag(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, ref := pathAndRef(r)

	tag := r.URL.Query().Get("tag")

	in := &grpc.TagRequest{
		Namespace: namespace,
		Path:      path,
		Ref:       ref,
		Tag:       tag,
	}

	resp, err := h.client.Tag(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) Untag(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, ref := pathAndRef(r)

	in := &grpc.UntagRequest{
		Namespace: namespace,
		Path:      path,
		Tag:       ref,
	}

	resp, err := h.client.Untag(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) Retag(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, ref := pathAndRef(r)

	tag := r.URL.Query().Get("tag")

	in := &grpc.RetagRequest{
		Namespace: namespace,
		Path:      path,
		Ref:       ref,
		Tag:       tag,
	}

	resp, err := h.client.Retag(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) ValidateRef(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, ref := pathAndRef(r)

	in := &grpc.ValidateRefRequest{
		Namespace: namespace,
		Path:      path,
		Ref:       ref,
	}

	resp, err := h.client.ValidateRef(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) ValidateRouter(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	in := &grpc.ValidateRouterRequest{
		Namespace: namespace,
		Path:      path,
	}

	resp, err := h.client.ValidateRouter(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) EditRouter(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	in := new(grpc.EditRouterRequest)

	err := unmarshalBody(r, in)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in.Namespace = namespace
	in.Path = path

	resp, err := h.client.EditRouter(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) Router(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	in := &grpc.RouterRequest{
		Namespace: namespace,
		Path:      path,
	}

	resp, err := h.client.Router(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) RouterSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	in := &grpc.RouterRequest{
		Namespace: namespace,
		Path:      path,
	}

	resp, err := h.client.RouterStream(ctx, in)
	if err != nil {
		respond(w, resp, err)
		return
	}

	ch := make(chan interface{}, 1)

	defer func() {

		_ = resp.CloseSend()

		for {
			_, more := <-ch
			if !more {
				return
			}
		}

	}()

	go func() {

		defer close(ch)

		for {

			x, err := resp.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x

		}

	}()

	sse(w, ch)

}

func (h *flowHandler) Secrets(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]

	p, err := pagination(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := &grpc.SecretsRequest{
		Namespace:  namespace,
		Pagination: p,
	}

	resp, err := h.client.Secrets(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) SecretsSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]

	p, err := pagination(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := &grpc.SecretsRequest{
		Namespace:  namespace,
		Pagination: p,
	}

	resp, err := h.client.SecretsStream(ctx, in)
	if err != nil {
		respond(w, resp, err)
		return
	}

	ch := make(chan interface{}, 1)

	defer func() {

		_ = resp.CloseSend()

		for {
			_, more := <-ch
			if !more {
				return
			}
		}

	}()

	go func() {

		defer close(ch)

		for {

			x, err := resp.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x

		}

	}()

	sse(w, ch)

}

func (h *flowHandler) SetSecret(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	secret := mux.Vars(r)["secret"]

	in := new(grpc.SetSecretRequest)

	data, err := loadRawBody(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in.Namespace = namespace
	in.Key = secret
	in.Data = data

	resp, err := h.client.SetSecret(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) DeleteSecret(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	secret := mux.Vars(r)["secret"]

	in := new(grpc.DeleteSecretRequest)
	in.Namespace = namespace
	in.Key = secret

	resp, err := h.client.DeleteSecret(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) Instance(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	instance := mux.Vars(r)["instance"]

	in := &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  instance,
	}

	resp, err := h.client.Instance(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) InstanceSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	instance := mux.Vars(r)["instance"]

	in := &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  instance,
	}

	resp, err := h.client.InstanceStream(ctx, in)
	if err != nil {
		respond(w, resp, err)
		return
	}

	ch := make(chan interface{}, 1)

	defer func() {

		_ = resp.CloseSend()

		for {
			_, more := <-ch
			if !more {
				return
			}
		}

	}()

	go func() {

		defer close(ch)

		for {

			x, err := resp.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x

		}

	}()

	sse(w, ch)

}

func (h *flowHandler) Instances(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]

	p, err := pagination(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := &grpc.InstancesRequest{
		Namespace:  namespace,
		Pagination: p,
	}

	resp, err := h.client.Instances(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) InstancesSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]

	p, err := pagination(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := &grpc.InstancesRequest{
		Namespace:  namespace,
		Pagination: p,
	}

	resp, err := h.client.InstancesStream(ctx, in)
	if err != nil {
		respond(w, resp, err)
		return
	}

	ch := make(chan interface{}, 1)

	defer func() {

		_ = resp.CloseSend()

		for {
			_, more := <-ch
			if !more {
				return
			}
		}

	}()

	go func() {

		defer close(ch)

		for {

			x, err := resp.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x

		}

	}()

	sse(w, ch)

}

func (h *flowHandler) InstanceInput(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	instance := mux.Vars(r)["instance"]

	in := &grpc.InstanceInputRequest{
		Namespace: namespace,
		Instance:  instance,
	}

	resp, err := h.client.InstanceInput(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) InstanceOutput(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	instance := mux.Vars(r)["instance"]

	in := &grpc.InstanceOutputRequest{
		Namespace: namespace,
		Instance:  instance,
	}

	resp, err := h.client.InstanceOutput(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) InstanceCancel(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	instance := mux.Vars(r)["instance"]

	in := &grpc.CancelInstanceRequest{
		Namespace: namespace,
		Instance:  instance,
	}

	resp, err := h.client.CancelInstance(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) ExecuteWorkflow(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, ref := pathAndRef(r)

	input, err := loadRawBody(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      path,
		Ref:       ref,
		Input:     input,
	}

	resp, err := h.client.StartWorkflow(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) BroadcastCloudevent(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]

	data, err := loadRawBody(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := &grpc.BroadcastCloudeventRequest{
		Namespace:  namespace,
		Cloudevent: data,
	}

	resp, err := h.client.BroadcastCloudevent(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) JQ(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()

	in := new(grpc.JQRequest)

	err := unmarshalBody(r, in)
	if err != nil {
		respond(w, nil, err)
		return
	}

	resp, err := h.client.JQ(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) NamespaceVariables(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]

	p, err := pagination(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := &grpc.NamespaceVariablesRequest{
		Namespace:  namespace,
		Pagination: p,
	}

	resp, err := h.client.NamespaceVariables(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) NamespaceVariablesSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]

	p, err := pagination(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := &grpc.NamespaceVariablesRequest{
		Namespace:  namespace,
		Pagination: p,
	}

	resp, err := h.client.NamespaceVariablesStream(ctx, in)
	if err != nil {
		respond(w, resp, err)
		return
	}

	ch := make(chan interface{}, 1)

	defer func() {

		_ = resp.CloseSend()

		for {
			_, more := <-ch
			if !more {
				return
			}
		}

	}()

	go func() {

		defer close(ch)

		for {

			x, err := resp.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x

		}

	}()

	sse(w, ch)

}

func (h *flowHandler) NamespaceVariable(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	key := mux.Vars(r)["var"]

	in := &grpc.NamespaceVariableRequest{
		Namespace: namespace,
		Key:       key,
	}

	resp, err := h.client.NamespaceVariableParcels(ctx, in)
	if err != nil {
		respond(w, resp, err)
		return
	}

	msg, err := resp.Recv()
	if err != nil {
		respond(w, resp, err)
		return
	}

	for {

		packet := msg.Data
		if len(packet) == 0 {
			return
		}

		_, err = io.Copy(w, bytes.NewReader(packet))
		if err != nil {
			return
		}

		msg, err = resp.Recv()
		if err != nil {
			return
		}

	}

}

func (h *flowHandler) SetNamespaceVariable(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	key := mux.Vars(r)["var"]

	var rdr io.Reader
	rdr = r.Body

	total := r.ContentLength
	if total <= 0 {
		data, err := loadRawBody(r)
		if err != nil {
			respond(w, nil, err)
			return
		}
		total = int64(len(data))
		rdr = bytes.NewReader(data)
	}

	rdr = io.LimitReader(rdr, total)

	client, err := h.client.SetNamespaceVariableParcels(ctx)
	if err != nil {
		respond(w, nil, err)
		return
	}

	var done int64

	for done < total {

		buf := new(bytes.Buffer)
		k, err := io.CopyN(buf, rdr, 2*1024*1024)
		done += k
		if err != nil && done < total {
			respond(w, nil, err)
			return
		}

		err = client.Send(&grpc.SetNamespaceVariableRequest{
			Namespace: namespace,
			Key:       key,
			TotalSize: total,
			Data:      buf.Bytes(),
		})
		if err != nil {
			respond(w, nil, err)
			return
		}

	}

	resp, err := client.CloseAndRecv()
	respond(w, resp, err)

}

func (h *flowHandler) DeleteNamespaceVariable(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	key := mux.Vars(r)["var"]

	in := &grpc.DeleteNamespaceVariableRequest{
		Namespace: namespace,
		Key:       key,
	}

	resp, err := h.client.DeleteNamespaceVariable(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) InstanceVariables(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	instance := mux.Vars(r)["instance"]

	p, err := pagination(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := &grpc.InstanceVariablesRequest{
		Namespace:  namespace,
		Instance:   instance,
		Pagination: p,
	}

	resp, err := h.client.InstanceVariables(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) InstanceVariablesSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	instance := mux.Vars(r)["instance"]

	p, err := pagination(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := &grpc.InstanceVariablesRequest{
		Namespace:  namespace,
		Instance:   instance,
		Pagination: p,
	}

	resp, err := h.client.InstanceVariablesStream(ctx, in)
	if err != nil {
		respond(w, resp, err)
		return
	}

	ch := make(chan interface{}, 1)

	defer func() {

		_ = resp.CloseSend()

		for {
			_, more := <-ch
			if !more {
				return
			}
		}

	}()

	go func() {

		defer close(ch)

		for {

			x, err := resp.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x

		}

	}()

	sse(w, ch)

}

func (h *flowHandler) InstanceVariable(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	instance := mux.Vars(r)["instance"]
	key := mux.Vars(r)["var"]

	in := &grpc.InstanceVariableRequest{
		Namespace: namespace,
		Instance:  instance,
		Key:       key,
	}

	resp, err := h.client.InstanceVariableParcels(ctx, in)
	if err != nil {
		respond(w, resp, err)
		return
	}

	msg, err := resp.Recv()
	if err != nil {
		respond(w, resp, err)
		return
	}

	for {

		packet := msg.Data
		if len(packet) == 0 {
			return
		}

		_, err = io.Copy(w, bytes.NewReader(packet))
		if err != nil {
			return
		}

		msg, err = resp.Recv()
		if err != nil {
			return
		}

	}

}

func (h *flowHandler) SetInstanceVariable(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	instance := mux.Vars(r)["instance"]
	key := mux.Vars(r)["var"]

	var rdr io.Reader
	rdr = r.Body

	total := r.ContentLength
	if total <= 0 {
		data, err := loadRawBody(r)
		if err != nil {
			respond(w, nil, err)
			return
		}
		total = int64(len(data))
		rdr = bytes.NewReader(data)
	}

	rdr = io.LimitReader(rdr, total)

	client, err := h.client.SetInstanceVariableParcels(ctx)
	if err != nil {
		respond(w, nil, err)
		return
	}

	var done int64

	for done < total {

		buf := new(bytes.Buffer)
		k, err := io.CopyN(buf, rdr, 2*1024*1024)
		done += k
		if err != nil && done < total {
			respond(w, nil, err)
			return
		}

		err = client.Send(&grpc.SetInstanceVariableRequest{
			Namespace: namespace,
			Instance:  instance,
			Key:       key,
			TotalSize: total,
			Data:      buf.Bytes(),
		})
		if err != nil {
			respond(w, nil, err)
			return
		}

	}

	resp, err := client.CloseAndRecv()
	respond(w, resp, err)

}

func (h *flowHandler) DeleteInstanceVariable(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	instance := mux.Vars(r)["instance"]
	key := mux.Vars(r)["var"]

	in := &grpc.DeleteInstanceVariableRequest{
		Namespace: namespace,
		Instance:  instance,
		Key:       key,
	}

	resp, err := h.client.DeleteInstanceVariable(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) WorkflowVariables(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	p, err := pagination(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := &grpc.WorkflowVariablesRequest{
		Namespace:  namespace,
		Path:       path,
		Pagination: p,
	}

	resp, err := h.client.WorkflowVariables(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) WorkflowVariablesSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	p, err := pagination(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := &grpc.WorkflowVariablesRequest{
		Namespace:  namespace,
		Path:       path,
		Pagination: p,
	}

	resp, err := h.client.WorkflowVariablesStream(ctx, in)
	if err != nil {
		respond(w, resp, err)
		return
	}

	ch := make(chan interface{}, 1)

	defer func() {

		_ = resp.CloseSend()

		for {
			_, more := <-ch
			if !more {
				return
			}
		}

	}()

	go func() {

		defer close(ch)

		for {

			x, err := resp.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x

		}

	}()

	sse(w, ch)

}

func (h *flowHandler) WorkflowVariable(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)
	key := mux.Vars(r)["var"]

	in := &grpc.WorkflowVariableRequest{
		Namespace: namespace,
		Path:      path,
		Key:       key,
	}

	resp, err := h.client.WorkflowVariableParcels(ctx, in)
	if err != nil {
		respond(w, resp, err)
		return
	}

	msg, err := resp.Recv()
	if err != nil {
		respond(w, resp, err)
		return
	}

	for {

		packet := msg.Data
		if len(packet) == 0 {
			return
		}

		_, err = io.Copy(w, bytes.NewReader(packet))
		if err != nil {
			return
		}

		msg, err = resp.Recv()
		if err != nil {
			return
		}

	}

}

func (h *flowHandler) SetWorkflowVariable(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)
	key := mux.Vars(r)["var"]

	var rdr io.Reader
	rdr = r.Body

	total := r.ContentLength
	if total <= 0 {
		data, err := loadRawBody(r)
		if err != nil {
			respond(w, nil, err)
			return
		}
		total = int64(len(data))
		rdr = bytes.NewReader(data)
	}

	rdr = io.LimitReader(rdr, total)

	client, err := h.client.SetWorkflowVariableParcels(ctx)
	if err != nil {
		respond(w, nil, err)
		return
	}

	var done int64

	for done < total {

		buf := new(bytes.Buffer)
		k, err := io.CopyN(buf, rdr, 2*1024*1024)
		done += k
		if err != nil && done < total {
			respond(w, nil, err)
			return
		}

		err = client.Send(&grpc.SetWorkflowVariableRequest{
			Namespace: namespace,
			Path:      path,
			Key:       key,
			TotalSize: total,
			Data:      buf.Bytes(),
		})
		if err != nil {
			respond(w, nil, err)
			return
		}

	}

	resp, err := client.CloseAndRecv()
	respond(w, resp, err)

}

func (h *flowHandler) DeleteWorkflowVariable(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)
	key := mux.Vars(r)["var"]

	in := &grpc.DeleteWorkflowVariableRequest{
		Namespace: namespace,
		Path:      path,
		Key:       key,
	}

	resp, err := h.client.DeleteWorkflowVariable(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) SetWorkflowEventLogging(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	in := new(grpc.SetWorkflowEventLoggingRequest)

	err := unmarshalBody(r, in)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in.Namespace = namespace
	in.Path = path

	resp, err := h.client.SetWorkflowEventLogging(ctx, in)
	respond(w, resp, err)

}

func (h *flowHandler) ToggleWorkflow(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	in := new(grpc.ToggleWorkflowRequest)

	err := unmarshalBody(r, in)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in.Namespace = namespace
	in.Path = path

	resp, err := h.client.ToggleWorkflow(ctx, in)
	respond(w, resp, err)

}
