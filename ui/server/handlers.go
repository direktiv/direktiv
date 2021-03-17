package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/vorteil/direktiv/pkg/ingress"
)

const (
	// GRPCCommandTimeout : timeout for grpc function calls
	GRPCCommandTimeout = 30 * time.Second
)

func respond(w http.ResponseWriter, out interface{}) {
	b, _ := json.Marshal(out)
	io.Copy(w, bytes.NewReader(b))
}

// namespacesHandler
func (g *grpcClient) namespacesHandler(w http.ResponseWriter, r *http.Request) {
	gCTX := context.Background()
	gCTX, cancel := context.WithDeadline(gCTX, time.Now().Add(GRPCCommandTimeout))
	defer cancel()

	resp, err := g.client.GetNamespaces(gCTX, &ingress.GetNamespacesRequest{})
	if err != nil {
		errResponse(w, err)
		return
	}

	respond(w, resp)
}

// workflowsHandler
func (g *grpcClient) workflowsHandler(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]

	gCTX := context.Background()
	gCTX, cancel := context.WithDeadline(gCTX, time.Now().Add(GRPCCommandTimeout))
	defer cancel()

	resp, err := g.client.GetWorkflows(gCTX, &ingress.GetWorkflowsRequest{
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

	gCTX := context.Background()
	gCTX, cancel := context.WithDeadline(gCTX, time.Now().Add(GRPCCommandTimeout))
	defer cancel()

	resp, err := g.client.GetWorkflowById(gCTX, &ingress.GetWorkflowByIdRequest{
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

	gCTX := context.Background()
	gCTX, cancel := context.WithDeadline(gCTX, time.Now().Add(GRPCCommandTimeout))
	defer cancel()

	resp, err := g.client.GetWorkflowInstances(gCTX, &ingress.GetWorkflowInstancesRequest{
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

	i := fmt.Sprintf("%s/%s/%s", mux.Vars(r)["namespace"], mux.Vars(r)["workflowID"], mux.Vars(r)["id"])

	gCTX := context.Background()
	gCTX, cancel := context.WithDeadline(gCTX, time.Now().Add(GRPCCommandTimeout))
	defer cancel()

	resp, err := g.client.GetWorkflowInstance(gCTX, &ingress.GetWorkflowInstanceRequest{
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

	i := fmt.Sprintf("%s/%s/%s", mux.Vars(r)["namespace"], mux.Vars(r)["workflowID"], mux.Vars(r)["id"])

	var limit int
	if x, ok := r.URL.Query()["limit"]; ok && len(x) > 0 {
		limit, _ = strconv.Atoi(x[0])
	}

	l := int32(limit)

	gCTX := context.Background()
	gCTX, cancel := context.WithDeadline(gCTX, time.Now().Add(GRPCCommandTimeout))
	defer cancel()

	resp, err := g.client.GetWorkflowInstanceLogs(gCTX, &ingress.GetWorkflowInstanceLogsRequest{
		InstanceId: &i,
		Limit:      &l,
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

	gCTX := context.Background()
	gCTX, cancel := context.WithDeadline(gCTX, time.Now().Add(GRPCCommandTimeout))
	defer cancel()

	resp, err := g.client.InvokeWorkflow(gCTX, &ingress.InvokeWorkflowRequest{
		Namespace:  &n,
		WorkflowId: &wf,
	})
	if err != nil {
		errResponse(w, err)
		return
	}

	respond(w, resp)
}

// createNamespaceHandler
func (g *grpcClient) createNamespaceHandler(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]

	gCTX := context.Background()
	gCTX, cancel := context.WithDeadline(gCTX, time.Now().Add(GRPCCommandTimeout))
	defer cancel()

	resp, err := g.client.AddNamespace(gCTX, &ingress.AddNamespaceRequest{
		Name: &n,
	})
	if err != nil {
		errResponse(w, err)
		return
	}

	respond(w, resp)
}

// createWorkflowHandler
func (g *grpcClient) createWorkflowHandler(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]
	active := true

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errResponse(w, err)
		return
	}

	gCTX := context.Background()
	gCTX, cancel := context.WithDeadline(gCTX, time.Now().Add(GRPCCommandTimeout))
	defer cancel()

	resp, err := g.client.AddWorkflow(gCTX, &ingress.AddWorkflowRequest{
		Active:    &active,
		Namespace: &n,
		Workflow:  b,
	})
	if err != nil {
		errResponse(w, err)
		return
	}

	respond(w, resp)
}
