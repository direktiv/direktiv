package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vorteil/direktiv/pkg/ingress"
)

func respond(w http.ResponseWriter, out interface{}) {
	b, _ := json.Marshal(out)
	io.Copy(w, bytes.NewReader(b))
}

// namespacesHandler
func (g *grpcClient) namespacesHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := g.client.GetNamespaces(r.Context(), &ingress.GetNamespacesRequest{})
	if err != nil {
		errResponse(w, err)
		return
	}

	respond(w, resp)
}

// workflowsHandler
func (g *grpcClient) workflowsHandler(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]

	resp, err := g.client.GetWorkflows(r.Context(), &ingress.GetWorkflowsRequest{
		Namespace: &n,
	})
	if err != nil {
		errResponse(w, err)
		return
	}

	respond(w, resp)
}

// workflowsHandler
func (g *grpcClient) getWorkflowHandler(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]
	wf := mux.Vars(r)["workflow"]

	resp, err := g.client.GetWorkflowById(r.Context(), &ingress.GetWorkflowByIdRequest{
		Namespace: &n,
		Id:        &wf,
	})
	if err != nil {
		errResponse(w, err)
		return
	}

	respond(w, resp)
}

// instancesHandler
func (g *grpcClient) instancesHandler(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]

	resp, err := g.client.GetWorkflowInstances(r.Context(), &ingress.GetWorkflowInstancesRequest{
		Namespace: &n,
	})
	if err != nil {
		errResponse(w, err)
		return
	}

	respond(w, resp)
}

// instanceHandler
func (g *grpcClient) instanceHandler(w http.ResponseWriter, r *http.Request) {

	i := mux.Vars(r)["instance"]

	resp, err := g.client.GetWorkflowInstance(r.Context(), &ingress.GetWorkflowInstanceRequest{
		Id: &i,
	})
	if err != nil {
		errResponse(w, err)
		return
	}

	respond(w, resp)
}

// instanceLogsHandler
func (g *grpcClient) instanceLogsHandler(w http.ResponseWriter, r *http.Request) {

	i := mux.Vars(r)["instance"]

	resp, err := g.client.GetWorkflowInstanceLogs(r.Context(), &ingress.GetWorkflowInstanceLogsRequest{
		InstanceId: &i,
	})
	if err != nil {
		errResponse(w, err)
		return
	}

	respond(w, resp)
}

// executeWorkflowHandler
func (g *grpcClient) executeWorkflowHandler(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]
	wf := mux.Vars(r)["workflow"]

	resp, err := g.client.InvokeWorkflow(r.Context(), &ingress.InvokeWorkflowRequest{
		Namespace:  &n,
		WorkflowId: &wf,
	})
	if err != nil {
		errResponse(w, err)
		return
	}

	respond(w, resp)
}
