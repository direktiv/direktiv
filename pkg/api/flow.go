package api

import (
	"errors"
	"fmt"
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

	handlerPair(r, RN_GetServerLogs, "/logs/server", h.ServerLogs, h.ServerLogsSSE)
	handlerPair(r, RN_GetNamespaceLogs, "/logs/namespaces/{ns}", h.NamespaceLogs, h.NamespaceLogsSSE)
	handlerPair(r, RN_GetWorkflowLogs, "/logs/namespaces/{ns}/tree/{path:.*}", h.WorkflowLogs, h.WorkflowLogsSSE)
	handlerPair(r, RN_GetInstanceLogs, "/logs/namespaces/{ns}/instances/{in}", h.InstanceLogs, h.InstanceLogsSSE)

	handlerPair(r, RN_GetNode, "/namespaces/{ns}/tree", h.GetNode, h.GetNodeSSE)
	handlerPair(r, RN_GetNode, "/namespaces/{ns}/tree/{path:.*}", h.GetNode, h.GetNodeSSE)
	r.HandleFunc("/namespaces/{ns}/tree/{path:.*}", h.CreateNode).Name(RN_CreateNode).Methods(http.MethodPut)
	r.HandleFunc("/namespaces/{ns}/tree/{path:.*}", h.DeleteNode).Name(RN_DeleteNode).Methods(http.MethodDelete)

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
	path := mux.Vars(r)["path"]

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
	path := mux.Vars(r)["path"]

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

type node struct {
	Type string `json:"type"`
	Data []byte `json:"data,omitempty"`
}

func (h *flowHandler) CreateNode(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	n := new(node)
	err := unmarshalBody(r, n)
	if err != nil {
		badRequest(w, err)
		return
	}

	switch n.Type {

	case "directory":

		in := &grpc.CreateDirectoryRequest{
			Namespace: namespace,
			Path:      path,
		}

		resp, err := h.client.CreateDirectory(ctx, in)
		respond(w, resp, err)
		return

	case "workflow":

		in := &grpc.CreateWorkflowRequest{
			Namespace: namespace,
			Path:      path,
			Source:    n.Data,
		}

		resp, err := h.client.CreateWorkflow(ctx, in)
		respond(w, resp, err)
		return

	default:
		badRequest(w, errors.New("bad node type should be: 'directory' or 'workflow'"))
		return
	}

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
