package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	protocol "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	prometheus "github.com/prometheus/client_golang/api"
	"go.uber.org/zap"
)

type flowHandler struct {
	logger     *zap.SugaredLogger
	client     grpc.FlowClient
	prometheus prometheus.Client

	apiV2Address string
}

func newSingleHostReverseProxy(patchReq func(req *http.Request) *http.Request) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		req = patchReq(req)
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "")
		}
	}

	return &httputil.ReverseProxy{
		Director: director,
	}
}

func newFlowHandler(logger *zap.SugaredLogger, base *mux.Router, router *mux.Router, conf *core.Config) (*flowHandler, error) {
	flowAddr := fmt.Sprintf("localhost:%d", conf.GrpcPort)
	logger.Debugf("connecting to flow %s", flowAddr)

	flowConn, err := util.GetEndpointTLS(flowAddr)
	if err != nil {
		logger.Errorf("can not connect to direktiv flows: %v", err)
		return nil, err
	}

	h := &flowHandler{
		logger:       logger,
		client:       grpc.NewFlowClient(flowConn),
		apiV2Address: fmt.Sprintf("localhost:%d", conf.ApiV2Port),
	}

	prometheusAddr := fmt.Sprintf("http://%s", conf.Prometheus)
	logger.Debugf("connecting to prometheus %s", prometheusAddr)
	h.prometheus, err = prometheus.NewClient(prometheus.Config{
		Address: prometheusAddr,
	})
	if err != nil {
		return nil, err
	}

	h.initRoutes(router)

	proxy := newSingleHostReverseProxy(func(req *http.Request) *http.Request {
		req.Host = ""
		req.URL.Host = h.apiV2Address
		req.URL.Scheme = "http"

		return req
	})
	router.PathPrefix("/v2").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}))

	base.PathPrefix("/gw").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}))

	base.PathPrefix("/ns").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}))

	return h, nil
}

func (h *flowHandler) initRoutes(r *mux.Router) {
	// swagger:operation GET /api/namespaces Namespaces getNamespaces
	// ---
	// description: |
	//   Gets the list of namespaces.
	// parameters:
	// - "": "#/parameters/PaginationQuery/order.field"
	// - "": "#/parameters/PaginationQuery/order.direction"
	// - "": "#/parameters/PaginationQuery/filter.field"
	// - "": "#/parameters/PaginationQuery/filter.type"
	// summary: Gets the list of namespaces
	// responses:
	//   '200':
	//     "description": "successfully got list of namespaces"
	//   error:
	//	   "description": "an error has occurred"
	handlerPair(r, RN_ListNamespaces, "/namespaces", h.Namespaces, h.NamespacesSSE)

	// swagger:operation PUT /api/namespaces/{namespace} Namespaces createNamespace
	// ---
	// summary: Creates a namespace
	// description: |
	//   Creates a new namespace.
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace to create'
	// responses:
	//   200:
	//     produces: application/json
	//     description: "namespace has been successfully created"
	//     schema:
	//       "$ref": '#/definitions/OkBody'
	//   default:
	//     produces: application/json
	//     description: an error has occurred
	//     schema:
	//       "$ref": '#/definitions/ErrorResponse'
	r.HandleFunc("/namespaces/{ns}", h.CreateNamespace).Name(RN_AddNamespace).Methods(http.MethodPut)

	r.HandleFunc("/namespaces/{ns}/lint", h.NamespaceLint).Name(RN_GetNamespaceLogs).Methods(http.MethodGet)

	// swagger:operation DELETE /api/namespaces/{namespace} Namespaces deleteNamespace
	// ---
	// description: |
	//   Delete a namespace.
	//   A namespace will not delete by default if it has any child resources (workflows, etc...).
	//   Deleting the namespace with all its children can be done using the `recursive` query parameter.
	// summary: Delete a namespace
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace to delete'
	// - in: query
	//   name: recursive
	//   type: boolean
	//   required: false
	//   description: 'recursively deletes all child resources'
	// responses:
	//   200:
	//     produces: application/json
	//     description: "namespace has been successfully deleted"
	//     schema:
	//       "$ref": '#/definitions/OkBody'
	//   default:
	//     produces: application/json
	//     description: an error has occurred
	//     schema:
	//       "$ref": '#/definitions/ErrorResponse'
	r.HandleFunc("/namespaces/{ns}", h.DeleteNamespace).Name(RN_DeleteNamespace).Methods(http.MethodDelete)

	// swagger:operation POST /api/jq Other jqPlayground
	// ---
	// description: |
	//   JQ Playground is a sandbox where you can test jq queries with custom data.
	// summary: JQ Playground api to test jq queries
	// parameters:
	// - in: body
	//   name: JQ payload
	//   required: true
	//   description: Payload that contains both the JSON data to manipulate and jq query.
	//   schema:
	//     example:
	//       data: "eyJhIjogMSwgImIiOiAyLCAiYyI6IDQsICJkIjogN30="
	//       query: "map(select(. >= 2))"
	//     type: object
	//     required:
	//       - data
	//       - query
	//     properties:
	//       data:
	//         type: string
	//         description: JSON data encoded in base64
	//       query:
	//         type: string
	//         description: jq query to manipulate JSON data
	// responses:
	//   '500':
	//     "description": "an unexpected internal error occurred"
	//   '400':
	//     "description": "the request was invalid"
	//   '200':
	//     "description": "jq query was successful"
	r.HandleFunc("/jq", h.JQ).Name(RN_JQPlayground).Methods(http.MethodPost)

	// swagger:operation GET /api/logs Logs serverLogs
	// ---
	// description: |
	//   Gets Direktiv Server Logs.
	// summary: Get Direktiv Server Logs
	// parameters:
	// - "": "#/parameters/PaginationQuery/order.field"
	// - "": "#/parameters/PaginationQuery/order.direction"
	// - "": "#/parameters/PaginationQuery/filter.field"
	// - "": "#/parameters/PaginationQuery/filter.type"
	// responses:
	//   200:
	//     produces: application/json
	//     description: "successfully got server logs"
	//     schema:
	//       "$ref": '#/definitions/OkBody'
	//   default:
	//     produces: application/json
	//     description: an error has occurred
	//     schema:
	//       "$ref": '#/definitions/ErrorResponse'
	handlerPair(r, RN_GetServerLogs, "/logs", h.ServerLogs, h.ServerLogsSSE)

	// swagger:operation GET /api/namespaces/{namespace}/logs Logs namespaceLogs
	// ---
	// description: |
	//   Gets Namespace Level Logs.
	// summary: Gets Namespace Level Logs
	// parameters:
	// - "": "#/parameters/PaginationQuery/order.field"
	// - "": "#/parameters/PaginationQuery/order.direction"
	// - "": "#/parameters/PaginationQuery/filter.field"
	// - "": "#/parameters/PaginationQuery/filter.type"
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// responses:
	//   '200':
	//     "description": "successfully got namespace logs"
	handlerPair(r, RN_GetNamespaceLogs, "/namespaces/{ns}/logs", h.NamespaceLogs, h.NamespaceLogsSSE)

	// swagger:operation GET /api/namespaces/{namespace}/instances/{instance}/logs Logs instanceLogs
	// ---
	// description: |
	//   Gets the logs of an executed instance.
	// summary: Gets Instance Logs
	// parameters:
	// - "": "#/parameters/PaginationQuery/order.field"
	// - "": "#/parameters/PaginationQuery/order.direction"
	// - "": "#/parameters/PaginationQuery/filter.field"
	// - "": "#/parameters/PaginationQuery/filter.type"
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: instance
	//   type: string
	//   required: true
	//   description: 'target instance id'
	// responses:
	//   '200':
	//     "description": "successfully got instance logs"
	//     schema:
	//       "$ref": '#/definitions/OkBody'
	//   default:
	//     produces: application/json
	//     description: an error has occurred
	//     schema:
	//       "$ref": '#/definitions/ErrorResponse'
	handlerPair(r, RN_GetInstanceLogs, "/namespaces/{ns}/instances/{instance}/logs", h.InstanceLogs, h.InstanceLogsSSE)

	// swagger:operation GET /api/namespaces/{namespace}/tree/{workflow}?op=metrics-invoked Metrics workflowMetricsInvoked
	// ---
	// description: |
	//   Get metrics of invoked workflow instances.
	// summary: Gets Invoked Workflow Metrics
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// responses:
	//   '200':
	//     "description": "successfully got workflow metrics"
	pathHandler(r, http.MethodGet, RN_GetWorkflowMetrics, "metrics-invoked", h.WorkflowMetricsInvoked)

	// swagger:operation GET /api/namespaces/{namespace}/tree/{workflow}?op=metrics-successful Metrics workflowMetricsSuccessful
	// ---
	// description: |
	//   Get metrics of a workflow, where the instance was successful.
	// summary: Gets Successful Workflow Metrics
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// responses:
	//   '200':
	//     "description": "successfully got workflow metrics"
	pathHandler(r, http.MethodGet, RN_GetWorkflowMetrics, "metrics-successful", h.WorkflowMetricsSuccessful)

	// swagger:operation GET /api/namespaces/{namespace}/tree/{workflow}?op=metrics-failed Metrics workflowMetricsFailed
	// ---
	// description: |
	//   Get metrics of a workflow, where the instance failed.
	// summary: Gets Failed Workflow Metrics
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// responses:
	//   '200':
	//     "description": "successfully got workflow metrics"
	pathHandler(r, http.MethodGet, RN_GetWorkflowMetrics, "metrics-failed", h.WorkflowMetricsFailed)

	// swagger:operation GET /api/namespaces/{namespace}/tree/{workflow}?op=metrics-failed Metrics workflowMetricsMilliseconds
	// ---
	// description: |
	//   Get the timing metrics of a workflow's instance.
	//   This returns a total sum of the milliseconds a workflow has been executed for.
	// summary: Gets Workflow Time Metrics
	// parameters:
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// responses:
	//   '200':
	//     "description": "successfully got workflow metrics"
	pathHandler(r, http.MethodGet, RN_GetWorkflowMetrics, "metrics-milliseconds", h.WorkflowMetricsMilliseconds)

	// swagger:operation GET /api/namespaces/{namespace}/tree/{workflow}?op=metrics-state-milliseconds Metrics workflowMetricsStateMilliseconds
	// ---
	// description: |
	//   Get the state timing metrics of a workflow's instance.
	//   This returns the timing of individual states in a workflow.
	// summary: Gets a Workflow State Time Metrics
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// responses:
	//   '200':
	//     "description": "successfully got workflow metrics"
	pathHandler(r, http.MethodGet, RN_GetWorkflowMetrics, "metrics-state-milliseconds", h.WorkflowMetricsStateMilliseconds)

	// swagger:operation GET /api/namespaces/{namespace}/metrics/invoked Metrics namespaceMetricsInvoked
	// ---
	// description: |
	//   Get metrics of invoked workflows in the targeted namespace.
	// summary: Gets Namespace Invoked Workflow Metrics
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// responses:
	//   '200':
	//     "description": "successfully got namespace metrics"
	r.HandleFunc("/namespaces/{ns}/metrics/invoked", h.NamespaceMetricsInvoked).Name(RN_GetNamespaceMetrics).Methods(http.MethodGet)

	// swagger:operation GET /api/namespaces/{namespace}/metrics/successful Metrics namespaceMetricsSuccessful
	// ---
	// description: |
	//   Get metrics of successful workflows in the targeted namespace.
	// summary: Gets Namespace Successful Workflow Instances Metrics
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// responses:
	//   '200':
	//     "description": "successfully got namespace metrics"
	r.HandleFunc("/namespaces/{ns}/metrics/successful", h.NamespaceMetricsSuccessful).Name(RN_GetNamespaceMetrics).Methods(http.MethodGet)

	// swagger:operation GET /api/namespaces/{namespace}/metrics/failed Metrics namespaceMetricsFailed
	// ---
	// description: |
	//   Get metrics of failed workflows in the targeted namespace.
	// summary: Gets Namespace Failed Workflow Instances Metrics
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// responses:
	//   '200':
	//     "description": "successfully got namespace metrics"
	r.HandleFunc("/namespaces/{ns}/metrics/failed", h.NamespaceMetricsFailed).Name(RN_GetNamespaceMetrics).Methods(http.MethodGet)

	// swagger:operation GET /api/namespaces/{namespace}/metrics/milliseconds Metrics namespaceMetricsMilliseconds
	// ---
	// description: |
	//   Get timing metrics of workflows in the targeted namespace.
	// summary: Gets Namespace Workflow Timing Metrics
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// responses:
	//   '200':
	//     "description": "successfully got namespace metrics"
	r.HandleFunc("/namespaces/{ns}/metrics/milliseconds", h.NamespaceMetricsMilliseconds).Name(RN_GetNamespaceMetrics).Methods(http.MethodGet)

	// swagger:operation GET /api/namespaces/{namespace}/tree/{workflow}?op=metrics-sankey Metrics workflowMetricsSankey
	// ---
	// description: |
	//   Get Sankey metrics of a workflow revision.
	//   If ref query is not provided, metrics for the latest revision
	//   will be retrieved.
	// summary: Get Sankey metrics of a workflow revision.
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// - in: query
	//   name: ref
	//   type: string
	//   required: false
	//   description: 'target workflow revision reference'
	// responses:
	//   '200':
	//     "description": "successfully got workflow metrics"
	pathHandler(r, http.MethodGet, RN_GetWorkflowMetrics, "metrics-sankey", h.MetricsSankey)

	// swagger:operation GET /api/namespaces/{namespace}/vars/{variable} Variables getNamespaceVariable
	// ---
	// description: |
	//   Get the value sorted in a namespace variable.
	// summary: Get a Namespace Variable
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: variable
	//   type: string
	//   required: true
	//   description: 'target variable'
	// responses:
	//   '200':
	//     "description": "successfully got namespace variable"
	r.HandleFunc("/namespaces/{ns}/vars/{var}", h.NamespaceVariable).Name(RN_GetNamespaceVariable).Methods(http.MethodGet)

	// swagger:operation DELETE /api/namespaces/{namespace}/vars/{variable} Variables deleteNamespaceVariable
	// ---
	// description: |
	//   Delete a namespace variable.
	// summary: Delete a Namespace Variable
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: variable
	//   type: string
	//   required: true
	//   description: 'target variable'
	// responses:
	//   '200':
	//     "description": "successfully deleted namespace variable"
	r.HandleFunc("/namespaces/{ns}/vars/{var}", h.DeleteNamespaceVariable).Name(RN_SetNamespaceVariable).Methods(http.MethodDelete)

	// swagger:operation PUT /api/namespaces/{namespace}/vars/{variable} Variables setNamespaceVariable
	// ---
	// description: |
	//   Set the value sorted in a namespace variable.
	//   If the target variable does not exists, it will be created.
	//   Variable data can be anything.
	// summary: Set a Namespace Variable
	// consumes:
	// - text/plain
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: variable
	//   type: string
	//   required: true
	//   description: 'target variable'
	// - in: body
	//   name: data
	//   required: true
	//   description: "Payload that contains variable data."
	//   schema:
	//     example: Data to Store
	//     type: string
	// responses:
	//   '200':
	//     "description": "successfully set namespace variable"
	r.HandleFunc("/namespaces/{ns}/vars/{var}", h.SetNamespaceVariable).Name(RN_SetNamespaceVariable).Methods(http.MethodPut)

	// swagger:operation GET /api/namespaces/{namespace}/vars Variables getNamespaceVariables
	// ---
	// description: |
	//   Gets a list of variables in a namespace.
	// summary: Get Namespace Variable List
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// responses:
	//   '200':
	//     "description": "successfully got namespace variables"
	handlerPair(r, RN_ListNamespaceVariables, "/namespaces/{ns}/vars", h.NamespaceVariables, h.NamespaceVariablesSSE)

	// swagger:operation GET /api/namespaces/{namespace}/instances/{instance}/vars/{variable} Variables getInstanceVariable
	// ---
	// description: |
	//   Get the value sorted in a instance variable.
	// summary: Get a Instance Variable
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: variable
	//   type: string
	//   required: true
	//   description: 'target variable'
	// - in: path
	//   name: instance
	//   type: string
	//   required: true
	//   description: 'target instance'
	// responses:
	//   '200':
	//     "description": "successfully got instance variable"
	r.HandleFunc("/namespaces/{ns}/instances/{instance}/vars/{var}", h.InstanceVariable).Name(RN_GetInstanceVariable).Methods(http.MethodGet)

	// swagger:operation DELETE /api/namespaces/{namespace}/instances/{instance}/vars/{variable} Variables deleteInstanceVariable
	// ---
	// description: |
	//   Delete a instance variable.
	// summary: Delete a Instance Variable
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: variable
	//   type: string
	//   required: true
	//   description: 'target variable'
	// - in: path
	//   name: instance
	//   type: string
	//   required: true
	//   description: 'target instance'
	// responses:
	//   '200':
	//     "description": "successfully deleted instance variable"
	r.HandleFunc("/namespaces/{ns}/instances/{instance}/vars/{var}", h.DeleteInstanceVariable).Name(RN_SetInstanceVariable).Methods(http.MethodDelete)

	// swagger:operation PUT /api/namespaces/{namespace}/instances/{instance}/vars/{variable} Variables setInstanceVariable
	// ---
	// description: |
	//   Set the value sorted in a instance variable.
	//   If the target variable does not exists, it will be created.
	//   Variable data can be anything.
	// summary: Set a Instance Variable
	// consumes:
	// - text/plain
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: variable
	//   type: string
	//   required: true
	//   description: 'target variable'
	// - in: path
	//   name: instance
	//   type: string
	//   required: true
	//   description: 'target instance'
	// - in: body
	//   name: data
	//   required: true
	//   description: "Payload that contains variable data."
	//   schema:
	//     example: "Data to Store"
	//     type: string
	// responses:
	//   '200':
	//     "description": "successfully set instance variable"
	r.HandleFunc("/namespaces/{ns}/instances/{instance}/vars/{var}", h.SetInstanceVariable).Name(RN_SetInstanceVariable).Methods(http.MethodPut)

	// swagger:operation GET /api/namespaces/{namespace}/instances/{instance}/vars Variables getInstanceVariables
	// ---
	// description: |
	//   Gets a list of variables in a instance.
	// summary: Get List of Instance Variable
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: instance
	//   type: string
	//   required: true
	//   description: 'target instance'
	// responses:
	//   '200':
	//     "description": "successfully got instance variables"
	handlerPair(r, RN_ListInstanceVariables, "/namespaces/{ns}/instances/{instance}/vars", h.InstanceVariables, h.InstanceVariablesSSE)

	// swagger:operation GET /api/namespaces/{namespace}/tree/{workflow}?op=var Variables getWorkflowVariable
	// ---
	// description: |
	//   Get the value sorted in a workflow variable.
	// summary: Get a Workflow Variable
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: query
	//   name: var
	//   type: string
	//   required: true
	//   description: 'target variable'
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// responses:
	//   '200':
	//     "description": "successfully got workflow variable"
	pathHandler(r, http.MethodGet, RN_GetWorkflowVariable, "var", h.WorkflowVariable)

	// swagger:operation DELETE /api/namespaces/{namespace}/tree/{workflow}?op=delete-var Variables deleteWorkflowVariable
	// ---
	// description: |
	//   Delete a workflow variable.
	// summary: Delete a Workflow Variable
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: query
	//   name: var
	//   type: string
	//   required: true
	//   description: 'target variable'
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// responses:
	//   '200':
	//     "description": "successfully deleted workflow variable"
	pathHandler(r, http.MethodDelete, RN_SetWorkflowVariable, "delete-var", h.DeleteWorkflowVariable)

	// swagger:operation PUT /api/namespaces/{namespace}/tree/{workflow}?op=set-var Variables setWorkflowVariable
	// ---
	// description: |
	//   Set the value sorted in a workflow variable.
	//   If the target variable does not exists, it will be created.
	//   Variable data can be anything.
	// summary: Set a Workflow Variable
	// consumes:
	// - text/plain
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: query
	//   name: var
	//   type: string
	//   required: true
	//   description: 'target variable'
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// - in: body
	//   name: data
	//   required: true
	//   description: "Payload that contains variable data."
	//   schema:
	//     example: "Data to Store"
	//     type: string
	// responses:
	//   '200':
	//     "description": "successfully set workflow variable"
	pathHandler(r, http.MethodPut, RN_SetWorkflowVariable, "set-var", h.SetWorkflowVariable)

	// swagger:operation GET /api/namespaces/{namespace}/tree/{workflow}?op=vars Variables getWorkflowVariables
	// ---
	// description: |
	//   Gets a list of variables in a workflow.
	// summary: Get List of Workflow Variables
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// responses:
	//   '200':
	//     "description": "successfully got workflow variables"
	pathHandlerPair(r, RN_ListWorkflowVariables, "vars", h.WorkflowVariables, h.WorkflowVariablesSSE)

	// swagger:operation GET /api/namespaces/{namespace}/secrets Secrets getSecrets
	// ---
	// description: |
	//   Gets the list of namespace secrets. Also can use for search by setting query param op=search and term=<searchingName>
	// summary: Get List of Namespace Secrets or Search for Namespace Secrets by given name
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// responses:
	//   200:
	//     produces: application/json
	//     description: "successfully got namespace nodes"
	//     schema:
	//       "$ref": '#/definitions/OkBody'
	//   default:
	//     produces: application/json
	//     description: an error has occurred
	//     schema:
	//       "$ref": '#/definitions/ErrorResponse'
	handlerPair(r, RN_ListSecrets, "/namespaces/{ns}/secrets", h.Secrets, h.SecretsSSE)

	// swagger:operation GET /api/namespaces/{namespace}/secrets/{folder}/ Secrets getSecretsInsideFolder
	// ---
	// description: |
	//   Gets the list of namespace secrets and folders inside specific folder.
	// summary: Get List of Namespace nodes inside Folder
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: folder
	//   type: string
	//   required: true
	//   description: 'target folder path'
	// responses:
	//   200:
	//     produces: application/json
	//     description: "successfully got namespace nodes inside sepcific folder"
	//     schema:
	//       "$ref": '#/definitions/OkBody'
	//   default:
	//     produces: application/json
	//     description: an error has occurred
	//     schema:
	//       "$ref": '#/definitions/ErrorResponse'
	handlerPair(r, RN_ListSecrets, "/namespaces/{ns}/secrets/{folder:.*}", h.Secrets, h.SecretsSSE)

	// swagger:operation PUT /api/namespaces/{namespace}/secrets/{secret} Secrets createSecret
	// ---
	// description: |
	//   Create a namespace secret.
	// summary: Create a Namespace Secret
	// consumes:
	// - text/plain
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: secret
	//   type: string
	//   required: true
	//   description: 'target secret'
	// - in: body
	//   name: Secret Payload
	//   required: true
	//   description: "Payload that contains secret data."
	//   schema:
	//     example: 7F8E7B0124ACB2BD20B383DE0756C7C0
	//     type: string
	// responses:
	//   200:
	//     produces: application/json
	//     description: "namespace secret has been successfully created"
	//     schema:
	//       "$ref": '#/definitions/OkBody'
	//   default:
	//     produces: application/json
	//     description: an error has occurred
	//     schema:
	//       "$ref": '#/definitions/ErrorResponse'
	r.HandleFunc("/namespaces/{ns}/secrets/{secret:.*[^/]$}", h.SetSecret).Name(RN_CreateSecret).Methods(http.MethodPut)

	// swagger:operation DELETE /api/namespaces/{namespace}/secrets/{secret} Secrets deleteSecret
	// ---
	// description: |
	//   Delete a namespace secret.
	// summary: Delete a Namespace Secret
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: secret
	//   type: string
	//   required: true
	//   description: 'target secret'
	// responses:
	//   200:
	//     produces: application/json
	//     description: "namespace secret has been successfully deleted"
	//     schema:
	//       "$ref": '#/definitions/OkBody'
	//   default:
	//     produces: application/json
	//     description: secret not found
	//     schema:
	//       "$ref": '#/definitions/ErrorResponse'
	r.HandleFunc("/namespaces/{ns}/secrets/{secret:.*[^/]$}", h.DeleteSecret).Name(RN_DeleteSecret).Methods(http.MethodDelete)

	// swagger:operation DELETE /api/namespaces/{namespace}/secrets/{folder}/ Secrets deleteFolder
	// ---
	// description: |
	//   Delete a namespace folder.
	// summary: Delete a Namespace Folder
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: folder
	//   type: string
	//   required: true
	//   description: 'target folder'
	// responses:
	//   200:
	//     produces: application/json
	//     description: "namespace folder has been successfully deleted"
	//     schema:
	//       "$ref": '#/definitions/OkBody'
	//   default:
	//     produces: application/json
	//     description: folder not found
	//     schema:
	//       "$ref": '#/definitions/ErrorResponse'
	r.HandleFunc("/namespaces/{ns}/secrets/{folder:.*[/]$}", h.DeleteSecretsFolder).Name(RN_DeleteSecretsFolder).Methods(http.MethodDelete)

	// swagger:operation PUT /api/namespaces/{namespace}/secrets/{folder}/ Secrets createFolder
	// ---
	// description: |
	//   Create a namespace folder.
	// summary: Create a Namespace Folder
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: folder
	//   type: string
	//   required: true
	//   description: 'target secret'
	// responses:
	//   200:
	//     produces: application/json
	//     description: "namespace folder has been successfully created"
	//     schema:
	//       "$ref": '#/definitions/OkBody'
	//   default:
	//     produces: application/json
	//     description: an error has occurred
	//     schema:
	//       "$ref": '#/definitions/ErrorResponse'
	r.HandleFunc("/namespaces/{ns}/secrets/{folder:.*[/]$}", h.CreateSecretsFolder).Name(RN_CreateSecretsFolder).Methods(http.MethodPut)

	// swagger:operation PATCH /api/namespaces/{namespace}/secrets/{secret} Secrets overwriteAndSearchSecret
	// ---
	// description: |
	//   Overwrite a namespace secret
	// summary: Overwrite a Namespace Secret
	// consumes:
	// - text/plain
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: secret
	//   type: string
	//   required: true
	//   description: 'target secret'
	// - in: body
	//   name: Secret Payload
	//   required: true
	//   description: "Payload that contains secret data"
	//   schema:
	//     example: 7F8E7B0124ACB2BD20B383DE0756C7C0
	//     type: string
	// responses:
	//   200:
	//     produces: application/json
	//     description: "namespace has been successfully overwritten"
	//     schema:
	//       "$ref": '#/definitions/OkBody'
	//   default:
	//     produces: application/json
	//     description: secret not found
	//     schema:
	//       "$ref": '#/definitions/ErrorResponse'
	r.HandleFunc("/namespaces/{ns}/secrets/{secret:.*[^/]$}", h.OverwriteSecret).Name(RN_OverwriteSecret).Methods(http.MethodPatch)

	// swagger:operation GET /api/namespaces/{namespace}/instances/{instance} Instances getInstance
	// ---
	// description: |
	//   Gets the details of a executed workflow instance in this namespace.
	// summary: Get a Instance
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: instance
	//   type: string
	//   required: true
	//   description: 'target instance'
	// responses:
	//   200:
	//     produces: application/json
	//     description: "successfully got instance"
	//     schema:
	//       "$ref": '#/definitions/OkBody'
	//   default:
	//     produces: application/json
	//     description: an error has occurred
	//     schema:
	//       "$ref": '#/definitions/ErrorResponse'
	handlerPair(r, RN_GetInstance, "/namespaces/{ns}/instances/{instance}", h.Instance, h.InstanceSSE)

	// swagger:operation GET /api/namespaces/{namespace}/instances Instances getInstanceList
	// ---
	// description: |
	//   Gets a list of instances in a namespace.
	// summary: Get List Instances
	// parameters:
	// - "": "#/parameters/PaginationQuery/order.field"
	// - "": "#/parameters/PaginationQuery/order.direction"
	// - "": "#/parameters/PaginationQuery/filter.field"
	// - "": "#/parameters/PaginationQuery/filter.type"
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// responses:
	//   '200':
	//     "description": "successfully got namespace instances"
	handlerPair(r, RN_ListInstances, "/namespaces/{ns}/instances", h.Instances, h.InstancesSSE)

	// swagger:operation GET /api/namespaces/{namespace}/instances/{instance}/input Instances getInstanceInput
	// ---
	// description: |
	//   Gets the input an instance was provided when executed.
	// summary: Get a Instance Input
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: instance
	//   type: string
	//   required: true
	//   description: 'target instance'
	// responses:
	//   '200':
	//     "description": "successfully got instance input"
	r.HandleFunc("/namespaces/{ns}/instances/{instance}/input", h.InstanceInput).Name(RN_GetInstance).Methods(http.MethodGet)

	// swagger:operation GET /api/namespaces/{namespace}/instances/{instance}/output Instances getInstanceOutput
	// ---
	// description: |
	//   Gets the output an instance was provided when executed.
	// summary: Get a Instance Output
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: instance
	//   type: string
	//   required: true
	//   description: 'target instance'
	// responses:
	//   '200':
	//     "description": "successfully got instance output"
	r.HandleFunc("/namespaces/{ns}/instances/{instance}/output", h.InstanceOutput).Name(RN_GetInstance).Methods(http.MethodGet)

	// swagger:operation GET /api/namespaces/{namespace}/instances/{instance}/metadata Instances getInstanceMetadata
	// ---
	// description: |
	//   Gets the metadata of an instance.
	// summary: Get a Instance Metadata
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: instance
	//   type: string
	//   required: true
	//   description: 'target instance'
	// responses:
	//   '200':
	//     "description": "successfully got instance metadata"
	r.HandleFunc("/namespaces/{ns}/instances/{instance}/metadata", h.InstanceMetadata).Name(RN_GetInstance).Methods(http.MethodGet)

	// swagger:operation POST /api/namespaces/{namespace}/instances/{instance}/cancel Instances cancelInstance
	// ---
	// description: |
	//   Cancel a currently pending instance.
	// summary: Cancel a Pending Instance
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: instance
	//   type: string
	//   required: true
	//   description: 'target instance'
	// responses:
	//   '200':
	//     "description": "successfully cancelled instance"
	r.HandleFunc("/namespaces/{ns}/instances/{instance}/cancel", h.InstanceCancel).Name(RN_CancelInstance).Methods(http.MethodPost)

	// swagger:operation POST /api/namespaces/{namespace}/broadcast Other broadcastCloudevent
	// ---
	// description: |
	//   Broadcast a cloud event to a namespace.
	//   Cloud events posted to this api will be picked up by any workflows listening to the same event type on the namescape.
	//   The body of this request should follow the cloud event core specification defined at https://github.com/cloudevents/spec .
	// summary: Broadcast Cloud Event
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: body
	//   name: cloudevent
	//   required: true
	//   description: Cloud Event request to be sent.
	//   schema:
	//     type: object
	// responses:
	//   '200':
	//     "description": "successfully sent cloud event"
	r.HandleFunc("/namespaces/{ns}/broadcast", h.BroadcastCloudevent).Name(RN_NamespaceEvent).Methods(http.MethodPost)

	// swagger:operation GET /api/namespaces/{namespace}/tree/{workflow}?op=logs Logs getWorkflowLogs
	// ---
	// description: |
	//   Get workflow level logs.
	// summary: Get Workflow Level Logs
	// parameters:
	// - "": "#/parameters/PaginationQuery/order.field"
	//   enum:
	//     - CREATED
	//     - UPDATED
	// - "": "#/parameters/PaginationQuery/order.direction"
	// - "": "#/parameters/PaginationQuery/filter.field"
	// - "": "#/parameters/PaginationQuery/filter.type"
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// responses:
	//   '200':
	//     "description": "successfully got workflow logs"
	pathHandlerPair(r, RN_GetWorkflowLogs, "logs", h.WorkflowLogs, h.WorkflowLogsSSE)

	// swagger:operation PUT /api/namespaces/{namespace}/tree/{directory}?op=create-directory Directory createDirectory
	// ---
	// description: |
	//   Creates a directory at the target path.
	// summary: Create a Directory
	// parameters:
	// - in: query
	//   name: op
	//   default: create-directory
	//   type: string
	//   required: true
	//   description: the operation for the api
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: directory
	//   type: string
	//   required: true
	//   description: 'path to target directory'
	// responses:
	//   200:
	//     produces: application/json
	//     description: "directory has been created"
	//     schema:
	//       "$ref": '#/definitions/OkBody'
	//   default:
	//     produces: application/json
	//     description: an error has occurred
	//     schema:
	//       "$ref": '#/definitions/ErrorResponse'
	pathHandler(r, http.MethodPut, RN_CreateDirectory, "create-directory", h.CreateDirectory)

	// swagger:operation PUT /api/namespaces/{namespace}/tree/{workflow}?op=create-workflow Workflows createWorkflow
	// ---
	// description: |
	//   Creates a workflow at the target path.
	//   The body of this request should contain the workflow yaml.
	// summary: Create a Workflow
	// consumes:
	// - text/plain
	// parameters:
	// - in: query
	//   name: op
	//   default: create-workflow
	//   type: string
	//   required: true
	//   description: the operation for the api
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// - in: body
	//   name: workflow data
	//   description: Payload that contains the direktiv workflow yaml to create.
	//   schema:
	//     type: string
	//     example: |
	//       description: A simple no-op state that returns Hello world!
	//       states:
	//       - id: helloworld
	//         type: noop
	//         transform:
	//           result: Hello world!
	// responses:
	//   200:
	//     produces: application/json
	//     description: "successfully created workflow"
	//     schema:
	//       "$ref": '#/definitions/OkBody'
	//   default:
	//     produces: application/json
	//     description: an error has occurred
	//     schema:
	//       "$ref": '#/definitions/ErrorResponse'
	pathHandler(r, http.MethodPut, RN_CreateWorkflow, "create-workflow", h.CreateWorkflow)

	// TODO: SWAGGER-SPEC
	pathHandler(r, http.MethodPut, RN_CreateWorkflow, "create-file", h.CreateFile)

	// swagger:operation POST /api/namespaces/{namespace}/tree/{workflow}?op=update-workflow Workflows updateWorkflow
	// ---
	// description: |
	//   Updates a workflow at the target path.
	//   The body of this request should contain the workflow yaml you want to update to.
	// summary: Update a Workflow
	// consumes:
	// - text/plain
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// - in: body
	//   name: workflow data
	//   description: Payload that contains the updated direktiv workflow yaml.
	//   schema:
	//     type: string
	//     example: |
	//       description: A simple no-op state that returns Hello world Updated !!!
	//       states:
	//       - id: helloworld
	//         type: noop
	//         transform:
	//           result: Hello world Updated !!!
	// responses:
	//   '200':
	//     "description": "successfully updated workflow"
	pathHandler(r, http.MethodPost, RN_UpdateWorkflow, "update-workflow", h.UpdateWorkflow)

	// TODO: SWAGGER-SPEC
	pathHandler(r, http.MethodPost, RN_UpdateWorkflow, "update-file", h.UpdateFile)

	// swagger:operation DELETE /api/namespaces/{namespace}/tree/{node}?op=delete-node Node deleteNode
	// ---
	// description: |
	//   Creates a directory at the target path.
	// summary: Delete a node
	// parameters:
	// - in: query
	//   name: op
	//   default: delete-node
	//   type: string
	//   required: true
	//   description: the operation for the api
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: node
	//   type: string
	//   required: true
	//   description: 'path to target node'
	// - in: query
	//   name: recursive
	//   type: boolean
	//   required: false
	//   description: 'whether to recursively delete child nodes'
	// responses:
	//   200:
	//     produces: application/json
	//     description: "node has been deleted"
	//     schema:
	//       "$ref": '#/definitions/OkBody'
	//   default:
	//     produces: application/json
	//     description: an error has occurred
	//     schema:
	//       "$ref": '#/definitions/ErrorResponse'
	pathHandler(r, http.MethodDelete, RN_DeleteNode, "delete-node", h.DeleteNode)

	// TODO: SWAGGER_SPEC
	pathHandler(r, http.MethodPost, RN_RenameNode, "rename-node", h.RenameNode)

	// TODO: SWAGGER_SPEC
	pathHandler(r, http.MethodPost, RN_UpdateMirror, "update-mirror", h.UpdateMirror)
	pathHandler(r, http.MethodPost, RN_LockMirror, "lock-mirror", h.LockMirror)
	pathHandler(r, http.MethodPost, RN_LockMirror, "unlock-mirror", h.UnlockMirror)
	pathHandler(r, http.MethodPost, RN_SyncMirror, "sync-mirror", h.SyncMirror)
	pathHandlerPair(r, RN_GetMirrorInfo, "mirror-info", h.MirrorInfo, h.MirrorInfoSSE)
	handlerPair(r, RN_GetMirrorActivityLogs, "/namespaces/{ns}/activities/{activity}/logs", h.MirrorActivityLogs, h.MirrorActivityLogsSSE)
	r.HandleFunc("/namespaces/{ns}/activities/{activity}/cancel", h.MirrorActivityCancel).Name(RN_CancelMirrorActivity).Methods(http.MethodPost)

	// swagger:operation GET /api/namespaces/{namespace}/event-listeners Events getEventListeners
	// ---
	// description: |
	//   Get current event listeners.
	// summary: Get current event listeners.
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// responses:
	//   '200':
	//     "description": "successfully got event listeners"
	handlerPair(r, RN_EventListeners, "/namespaces/{ns}/event-listeners", h.EventListeners, h.EventListenersSSE)

	// swagger:operation GET /api/namespaces/{namespace}/events Events getEventHistory
	// ---
	// description: |
	//   Get recent events history.
	// summary: Get events history.
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// responses:
	//   '200':
	//     "description": "successfully got events history"
	handlerPair(r, RN_EventHistory, "/namespaces/{ns}/events", h.EventHistory, h.EventHistorySSE)

	// swagger:operation POST /api/namespaces/{namespace}/events/{event}/replay Other replayCloudevent
	// ---
	// description: |
	//   Replay a cloud event to a namespace.
	// summary: Replay Cloud Event
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: event
	//   type: string
	//   required: true
	//   description: 'target cloudevent'
	// responses:
	//   '200':
	//     "description": "successfully replayed cloud event"
	r.HandleFunc("/namespaces/{ns}/events/{event:.*}/replay", h.ReplayEvent).Name(RN_NamespaceEvent).Methods(http.MethodPost)

	// swagger:operation POST /api/namespaces/{namespace}/tree/{workflow}?op=execute Workflows executeWorkflow
	// ---
	// description: |
	//   Executes a workflow with optionally some input provided in the request body as json.
	// summary: Execute a Workflow
	// parameters:
	// - in: query
	//   name: op
	//   default: execute
	//   type: string
	//   required: true
	//   description: the operation for the api
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// - in: body
	//   name: Workflow Input
	//   required: true
	//   description: The input of this workflow instance
	//   schema:
	//     example:
	//       animals:
	//         - dog
	//         - cat
	//         - snake
	//     type: object
	//     properties:
	// responses:
	//   200:
	//     produces: application/json
	//     description: "node has been deleted"
	//     schema:
	//       "$ref": '#/definitions/OkBody'
	//   default:
	//     produces: application/json
	//     description: an error has occurred
	//     schema:
	//       "$ref": '#/definitions/ErrorResponse'
	pathHandler(r, http.MethodPost, RN_ExecuteWorkflow, "execute", h.ExecuteWorkflow)

	// swagger:operation POST /api/namespaces/{namespace}/tree/{workflow}?op=wait Workflows awaitExecuteWorkflowBody
	// ---
	// description: |
	//   Executes a workflow with optionally some input provided in the request body as json.
	//   This path will wait until the workflow execution has completed and return the instance output.
	//   NOTE: Input can also be provided with the `input.X` query parameters; Where `X` is the json key.
	//   Only top level json keys are supported when providing input with query parameters.
	//   Input query parameters are only read if the request has no body.
	// summary: Await Execute a Workflow With Body
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// - in: query
	//   name: ctype
	//   type: string
	//   description: "Manually set the Content-Type response header instead of auto-detected. This doesn't change the body of the response in any way."
	//   required: false
	// - in: query
	//   name: field
	//   type: string
	//   required: false
	//   description: 'If provided, instead of returning the entire output json the response body will contain the single top-level json field'
	// - in: query
	//   name: raw-output
	//   type: boolean
	//   required: false
	//   description: "If set to true, will return an empty output as null, encoded base64 data as decoded binary data, and quoted json strings as a escaped string."
	// - in: body
	//   name: Workflow Input
	//   required: true
	//   description: The input of this workflow instance
	//   schema:
	//     example:
	//       animals:
	//         - dog
	//         - cat
	//         - snake
	//     type: object
	//     properties:
	// responses:
	//   '200':
	//     "description": "successfully executed workflow"
	pathHandler(r, http.MethodPost, RN_ExecuteWorkflow, "wait", h.WaitWorkflow)

	// swagger:operation GET /api/namespaces/{namespace}/tree/{workflow}?op=wait Workflows awaitExecuteWorkflow
	// ---
	// description: |
	//   Executes a workflow. This path will wait until the workflow execution has completed and return the instance output.
	//   NOTE: Input can also be provided with the `input.X` query parameters; Where `X` is the json key.
	//   Only top level json keys are supported when providing input with query parameters.
	// summary: Await Execute a Workflow
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// - in: query
	//   name: ctype
	//   type: string
	//   required: false
	//   description: "Manually set the Content-Type response header instead of auto-detected. This doesn't change the body of the response in any way."
	// - in: query
	//   name: field
	//   type: string
	//   required: false
	//   description: 'If provided, instead of returning the entire output json the response body will contain the single top-level json field'
	// - in: query
	//   name: raw-output
	//   type: boolean
	//   required: false
	//   description: "If set to true, will return an empty output as null, encoded base64 data as decoded binary data, and quoted json strings as a escaped string."
	// responses:
	//   '200':
	//     "description": "successfully executed workflow"
	pathHandler(r, http.MethodGet, RN_ExecuteWorkflow, "wait", h.WaitWorkflow)

	// swagger:operation GET /api/namespaces/{namespace}/tree/{nodePath} Node getNodes
	// ---
	// description: |
	//   Gets Workflow and Directory Nodes at nodePath.
	// summary: Get List of Namespace Nodes
	// tags:
	// - "Node"
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: nodePath
	//   type: string
	//   required: true
	//   description: 'target path in tree'
	// responses:
	//   200:
	//     produces: application/json
	//     description: "successfully got namespace nodes"
	//     schema:
	//       "$ref": '#/definitions/OkBody'
	//   default:
	//     produces: application/json
	//     description: an error has occurred
	//     schema:
	//       "$ref": '#/definitions/ErrorResponse'
	pathHandlerPair(r, RN_GetNode, "", h.GetNode, h.GetNodeSSE)
}

func (h *flowHandler) EventListeners(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}
	namespace := mux.Vars(r)["ns"]

	in := &grpc.EventListenersRequest{
		Pagination: p,
		Namespace:  namespace,
	}

	resp, err := h.client.EventListeners(ctx, in)
	respond(w, resp, err)
}

func (h *flowHandler) EventListenersSSE(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())
	namespace := mux.Vars(r)["ns"]

	ctx := r.Context()
	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.EventListenersRequest{
		Pagination: p,
		Namespace:  namespace,
	}

	resp, err := h.client.EventListenersStream(ctx, in)
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

func (h *flowHandler) EventHistory(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}
	namespace := mux.Vars(r)["ns"]

	in := &grpc.EventHistoryRequest{
		Pagination: p,
		Namespace:  namespace,
	}

	resp, err := h.client.EventHistory(ctx, in)
	respond(w, resp, err)
}

func (h *flowHandler) EventHistorySSE(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())
	namespace := mux.Vars(r)["ns"]

	ctx := r.Context()
	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.EventHistoryRequest{
		Pagination: p,
		Namespace:  namespace,
	}

	resp, err := h.client.EventHistoryStream(ctx, in)
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
	namespace := mux.Vars(r)["ns"]

	data, err := loadRawBody(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	if len(data) == 0 {
		in := &grpc.CreateNamespaceRequest{
			Name: namespace,
		}

		resp, err := h.client.CreateNamespace(ctx, in)
		respond(w, resp, err)
		return
	} else {
		settings := new(grpc.MirrorSettings)
		err = json.Unmarshal(data, settings)
		if err != nil {
			respond(w, nil, err)
			return
		}

		resp, err := h.client.CreateNamespaceMirror(ctx, &grpc.CreateNamespaceMirrorRequest{
			Name:     namespace,
			Settings: settings,
		})
		respond(w, resp, err)
		return
	}
}

func (h *flowHandler) UpdateMirror(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	data, err := loadRawBody(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	settings := new(grpc.MirrorSettings)
	err = json.Unmarshal(data, settings)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := &grpc.UpdateMirrorSettingsRequest{
		Namespace: namespace,
		Path:      path,
		Settings:  settings,
	}

	resp, err := h.client.UpdateMirrorSettings(ctx, in)
	respond(w, resp, err)
}

func (h *flowHandler) NamespaceLint(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]

	in := &grpc.NamespaceLintRequest{
		Name: namespace,
	}

	resp, err := h.client.NamespaceLint(ctx, in)
	if err != nil {
		respond(w, resp, err)
		return
	}

	respond(w, resp, nil)
}

func (h *flowHandler) DeleteNamespace(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	params := r.URL.Query()

	recursive := false
	// ignore err if not provided its set by default to false
	recursive, _ = strconv.ParseBool(params.Get("recursive"))

	in := &grpc.DeleteNamespaceRequest{
		Name:      namespace,
		Recursive: recursive,
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
	instance := mux.Vars(r)["instance"]

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
	instance := mux.Vars(r)["instance"]

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

func (h *flowHandler) MirrorActivityLogs(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	activity := mux.Vars(r)["activity"]

	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.MirrorActivityLogsRequest{
		Pagination: p,
		Namespace:  namespace,
		Activity:   activity,
	}

	resp, err := h.client.MirrorActivityLogs(ctx, in)
	respond(w, resp, err)
}

func (h *flowHandler) MirrorActivityLogsSSE(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	activity := mux.Vars(r)["activity"]

	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.MirrorActivityLogsRequest{
		Pagination: p,
		Namespace:  namespace,
		Activity:   activity,
	}

	resp, err := h.client.MirrorActivityLogsParcels(ctx, in)
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
	h.logger.Debugf("method: %s\nuri: %s\n", r.Method, r.URL.String())
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	var resp interface{}
	var err error
	var p *grpc.Pagination

	p, err = pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	resp, err = h.client.Directory(ctx, &grpc.DirectoryRequest{
		Pagination: p,
		Namespace:  namespace,
		Path:       path,
	})
	if err == nil {
		respond(w, resp, err)

		return
	}
	if !strings.Contains(err.Error(), filestore.ErrNotFound.Error()) {
		respond(w, resp, err)

		return
	}

	resp, err = h.client.Workflow(ctx, &grpc.WorkflowRequest{
		Namespace: namespace,
		Path:      path,
	})
	if err == nil {
		respond(w, resp, err)

		return
	}

	resp, err = h.client.File(ctx, &grpc.FileRequest{
		Namespace: namespace,
		Path:      path,
	})

	respond(w, resp, err)
}

func (h *flowHandler) GetNodeSSE(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

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
}

func (h *flowHandler) CreateDirectory(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	data, err := loadRawBody(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	if len(data) == 0 {
		in := &grpc.CreateDirectoryRequest{
			Namespace: namespace,
			Path:      path,
		}

		resp, err := h.client.CreateDirectory(ctx, in)
		respond(w, resp, err)
		return
	} else {
		settings := new(grpc.MirrorSettings)
		err = json.Unmarshal(data, settings)
		if err != nil {
			respond(w, nil, err)
			return
		}

		resp, err := h.client.CreateDirectoryMirror(ctx, &grpc.CreateDirectoryMirrorRequest{
			Namespace: namespace,
			Path:      path,
			Settings:  settings,
		})
		respond(w, resp, err)
		return
	}
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
}

func (h *flowHandler) CreateFile(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	data, err := loadRawBody(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := &grpc.CreateFileRequest{
		Namespace: namespace,
		Path:      path,
		Source:    data,
		MimeType:  r.Header.Get("Content-Type"),
	}

	resp, err := h.client.CreateFile(ctx, in)
	respond(w, resp, err)
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
}

func (h *flowHandler) UpdateFile(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	data, err := loadRawBody(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := &grpc.UpdateFileRequest{
		Namespace: namespace,
		Path:      path,
		Source:    data,
		MimeType:  r.Header.Get("Content-Type"),
	}

	resp, err := h.client.UpdateFile(ctx, in)
	respond(w, resp, err)
}

func (h *flowHandler) LockMirror(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	in := &grpc.LockMirrorRequest{}

	in.Namespace = namespace
	in.Path = path

	resp, err := h.client.LockMirror(ctx, in)
	respond(w, resp, err)
}

func (h *flowHandler) UnlockMirror(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	in := &grpc.UnlockMirrorRequest{}

	in.Namespace = namespace
	in.Path = path

	resp, err := h.client.UnlockMirror(ctx, in)
	respond(w, resp, err)
}

func (h *flowHandler) SyncMirror(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	force, _ := strconv.ParseBool(r.URL.Query().Get("force"))

	if force {
		in := &grpc.HardSyncMirrorRequest{}

		in.Namespace = namespace
		in.Path = path

		resp, err := h.client.HardSyncMirror(ctx, in)
		respond(w, resp, err)
		return
	} else {
		in := &grpc.SoftSyncMirrorRequest{}

		in.Namespace = namespace
		in.Path = path

		resp, err := h.client.SoftSyncMirror(ctx, in)
		respond(w, resp, err)
		return
	}
}

func (h *flowHandler) RenameNode(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	in := &grpc.RenameNodeRequest{}

	data, err := loadRawBody(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	err = json.Unmarshal(data, &in)
	if err != nil {
		respond(w, nil, err)
		return
	}
	in.Namespace = namespace
	in.Old = path

	resp, err := h.client.RenameNode(ctx, in)
	respond(w, resp, err)
}

func (h *flowHandler) DeleteNode(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	recursiveDelete, _ := strconv.ParseBool(r.URL.Query().Get("recursive"))

	in := &grpc.DeleteNodeRequest{
		Namespace: namespace,
		Path:      path,
		Recursive: recursiveDelete,
	}

	resp, err := h.client.DeleteNode(ctx, in)
	respond(w, resp, err)
}

func (h *flowHandler) MirrorInfo(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	p, err := pagination(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := &grpc.MirrorInfoRequest{
		Namespace:  namespace,
		Path:       path,
		Pagination: p,
	}

	resp, err := h.client.MirrorInfo(ctx, in)
	respond(w, resp, err)
}

func (h *flowHandler) MirrorInfoSSE(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	p, err := pagination(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := &grpc.MirrorInfoRequest{
		Namespace:  namespace,
		Path:       path,
		Pagination: p,
	}

	resp, err := h.client.MirrorInfoStream(ctx, in)
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
	folder := mux.Vars(r)["folder"]
	op := r.FormValue("op")
	term := r.FormValue("term")

	if op == "search" {
		p, err := pagination(r)
		if err != nil {
			respond(w, nil, err)
			return
		}

		in := &grpc.SearchSecretRequest{
			Namespace:  namespace,
			Pagination: p,
			Key:        term,
		}

		resp, err := h.client.SearchSecret(ctx, in)
		respond(w, resp, err)
	} else {
		p, err := pagination(r)
		if err != nil {
			respond(w, nil, err)
			return
		}

		in := &grpc.SecretsRequest{
			Namespace:  namespace,
			Pagination: p,
			Key:        folder,
		}

		resp, err := h.client.Secrets(ctx, in)
		respond(w, resp, err)
	}
}

func (h *flowHandler) SecretsSSE(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	folder := mux.Vars(r)["folder"]

	p, err := pagination(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := &grpc.SecretsRequest{
		Namespace:  namespace,
		Pagination: p,
		Key:        folder,
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

func (h *flowHandler) OverwriteSecret(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	secret := mux.Vars(r)["secret"]

	data, err := loadRawBody(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := new(grpc.UpdateSecretRequest)
	in.Namespace = namespace
	in.Key = secret
	in.Data = data

	resp, err := h.client.UpdateSecret(ctx, in)
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

func (h *flowHandler) DeleteSecretsFolder(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	folder := mux.Vars(r)["folder"]

	in := new(grpc.DeleteSecretsFolderRequest)
	in.Namespace = namespace
	in.Key = folder

	resp, err := h.client.DeleteSecretsFolder(ctx, in)
	respond(w, resp, err)
}

func (h *flowHandler) CreateSecretsFolder(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	folder := mux.Vars(r)["folder"]

	in := new(grpc.CreateSecretsFolderRequest)

	in.Namespace = namespace
	in.Key = folder

	resp, err := h.client.CreateSecretsFolder(ctx, in)
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

func (h *flowHandler) InstanceMetadata(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	instance := mux.Vars(r)["instance"]

	in := &grpc.InstanceMetadataRequest{
		Namespace: namespace,
		Instance:  instance,
	}

	resp, err := h.client.InstanceMetadata(ctx, in)
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

func (h *flowHandler) MirrorActivityCancel(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	activity := mux.Vars(r)["activity"]

	in := &grpc.CancelMirrorActivityRequest{
		Namespace: namespace,
		Activity:  activity,
	}

	resp, err := h.client.CancelMirrorActivity(ctx, in)
	respond(w, resp, err)
}

func (h *flowHandler) ExecuteWorkflow(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	input, err := loadRawBody(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in := &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      path,
		Input:     input,
	}

	resp, err := h.client.StartWorkflow(ctx, in)
	respond(w, resp, err)
}

func (h *flowHandler) WaitWorkflow(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	var err error
	var input []byte

	if r.ContentLength != 0 {
		input, err = loadRawBody(r)
		if err != nil {
			respond(w, nil, err)
			return
		}
	} else {
		m := make(map[string]interface{})
		query := r.URL.Query()
		for k, v := range query {
			if strings.HasPrefix(k, "input.") {
				k = k[6:]
				if len(v) == 1 {
					m[k] = v[0]
				} else {
					m[k] = v
				}
			}
		}

		input, err = json.Marshal(m)
		if err != nil {
			respond(w, nil, err)
			return
		}
	}

	in := &grpc.AwaitWorkflowRequest{
		Namespace: namespace,
		Path:      path,
		Input:     input,
	}

	c, err := h.client.AwaitWorkflow(ctx, in)
	if err != nil {
		respond(w, nil, err)
		return
	}
	defer func() {
		err := c.CloseSend()
		if err != nil {
			h.logger.Errorf("Failed to close connection: %v.", err)
		}
	}()

	instanceID := ""

	for {
		status, err := c.Recv()
		if err != nil {
			respond(w, nil, err)
			return
		}
		if instanceID == "" {
			instanceID = status.Instance.Id
			w.Header().Set("Direktiv-Instance-Id", instanceID)
		}

		if s := status.Instance.GetStatus(); s == util.InstanceStatusComplete {
			data := status.Data

			err = c.CloseSend()
			if err != nil {
				h.logger.Errorf("Failed to close connection: %v.", err)
			}

			field := r.URL.Query().Get("field")
			if field != "" {
				m := make(map[string]interface{})
				err = json.Unmarshal(data, &m)
				if err != nil {
					respond(w, nil, err)
					return
				}

				x, exists := m[field]
				if exists {
					data, err = json.Marshal(x)
					if err != nil {
						respond(w, nil, err)
						return
					}
				} else {
					data, err = json.Marshal(nil)
					if err != nil {
						panic(err)
					}
				}
			}

			var x interface{}

			err = json.Unmarshal(data, &x)
			if err != nil {
				respond(w, nil, err)
				return
			}

			rawo, _ := strconv.ParseBool(r.URL.Query().Get("raw-output"))

			if rawo {
				if x == nil {
					data = make([]byte, 0)
				} else if str, ok := x.(string); ok {
					data = []byte(str)
					b64, err := base64.StdEncoding.DecodeString(str)
					if err == nil {
						data = b64
					}
				}
			}

			w.Header().Set("Content-Length", fmt.Sprintf("%v", len(data)))

			ctype := r.URL.Query().Get("ctype")
			if ctype == "" {
				mtype := mimetype.Detect(data)
				ctype = mtype.String()
			}

			w.Header().Set("Content-Type", ctype)

			_, err = io.Copy(w, bytes.NewReader(data))
			if err != nil {
				h.logger.Errorf("Failed to send response: %v.", err)
			}

			return
		} else if s == util.InstanceStatusFailed || s == util.InstanceStatusCancelled {
			w.Header().Set("Direktiv-Instance-Error-Code", status.Instance.ErrorCode)
			w.Header().Set("Direktiv-Instance-Error-Message", status.Instance.ErrorMessage)
			code := http.StatusInternalServerError
			http.Error(w, fmt.Sprintf("An error occurred executing instance %s: %s: %s", instanceID, status.Instance.ErrorCode, status.Instance.ErrorMessage), code)
			return
		} else if s == util.InstanceStatusCrashed {
			code := http.StatusInternalServerError
			http.Error(w, fmt.Sprintf("An internal error occurred executing instance: %s", instanceID), code)
			return
		} else {
			continue
		}
	}
}

func ToGRPCCloudEvents(r *http.Request) ([]cloudevents.Event, error) {
	var events []cloudevents.Event
	ct := r.Header.Get("Content-type")
	oct := ct

	// if batch mode we need to parse the body to multiple events
	if strings.HasPrefix(ct, "application/cloudevents-batch+json") {
		// load body
		data, err := loadRawBody(r)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(data, &events)
		if err != nil {
			return nil, err
		}

		for i := range events {
			ev := events[i]
			if ev.ID() == "" {
				ev.SetID(uuid.New().String())
			}
			err = ev.Validate()
			if err != nil {
				return nil, err
			}
		}

		return events, nil
	}

	if strings.HasPrefix(ct, "application/json") {
		_, err := json.Marshal(r.Header)
		if err != nil {
			return nil, err
		}
		s := r.Header.Get("Ce-Type")
		if s == "" {
			ct = "application/cloudevents+json; charset=UTF-8"
			r.Header.Set("Content-Type", ct)
		}
	}

	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	r.Body = io.NopCloser(bytes.NewReader(bodyData))

	msg := protocol.NewMessageFromHttpRequest(r)
	ev, err := binding.ToEvent(context.Background(), msg)
	if err != nil {
		goto generic
	}

	// validate:
	if ev.ID() == "" {
		ev.SetID(uuid.New().String())
	}
	err = ev.Validate()

	// azure hack for dataschema '#' which is an invalid cloudevent
	if err != nil && strings.HasPrefix(err.Error(), "dataschema: if present") {
		err = ev.Context.SetDataSchema("")
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		goto generic
	}

	events = append(events, *ev)

	return events, nil

generic:

	xerr := err
	unmarshalable := false

	m := make(map[string]interface{})

	if strings.HasPrefix(oct, "application/json") {
		err = json.Unmarshal(bodyData, &m)
		if err == nil {
			unmarshalable = true
		}
	}

	event := cloudevents.NewEvent(cloudevents.VersionV1)
	ev = &event

	uid := uuid.New()
	ev.SetID(uid.String())
	ev.SetType("noncompliant")
	ev.SetSource("unknown")
	ev.SetDataContentType(ct)
	if unmarshalable {
		err = ev.SetData(oct, m)
		if err != nil {
			return events, xerr
		}
	} else {
		err = ev.SetData(oct, bodyData)
		if err != nil {
			return events, xerr
		}
	}

	err = ev.Context.SetExtension("error", xerr.Error())
	if err != nil {
		return events, xerr
	}

	err = ev.Validate()
	if err != nil {
		return events, xerr
	}

	events = append(events, event)

	return events, nil
}

func (h *flowHandler) doBroadcast(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]

	ces, err := ToGRPCCloudEvents(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	for i := range ces {
		d, err := json.Marshal(ces[i])
		if err != nil {
			respond(w, nil, err)
			return
		}

		in := &grpc.BroadcastCloudeventRequest{
			Namespace:  namespace,
			Cloudevent: d,
		}

		resp, err := h.client.BroadcastCloudevent(ctx, in)
		respond(w, resp, err)
	}
}

func (h *flowHandler) BroadcastCloudevent(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	h.doBroadcast(w, r)
}

func (h *flowHandler) ReplayEvent(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	event := mux.Vars(r)["event"]

	in := &grpc.ReplayEventRequest{
		Namespace: namespace,
		Id:        event,
	}

	resp, err := h.client.ReplayEvent(ctx, in)
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

	// Set MimeType
	w.Header().Set("Content-Type", msg.MimeType)

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
			if errors.Is(err, io.EOF) {
				total = 0
				rdr = bytes.NewReader([]byte(""))
			} else {
				respond(w, nil, err)
				return
			}
		} else {
			total = int64(len(data))
			rdr = bytes.NewReader(data)
		}
	}

	rdr = io.LimitReader(rdr, total)

	client, err := h.client.SetNamespaceVariableParcels(ctx)
	if err != nil {
		respond(w, nil, err)
		return
	}

	ctype := r.Header.Get("Content-Type")

	if total == 0 {
		err = client.Send(&grpc.SetNamespaceVariableRequest{
			Namespace: namespace,
			Key:       key,
			TotalSize: 0,
			Data:      []byte{},
			MimeType:  ctype,
		})
		if err != nil {
			respond(w, nil, err)
			return
		}
	} else {
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
				MimeType:  ctype,
			})
			if err != nil {
				respond(w, nil, err)
				return
			}
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

	// Set MimeType
	w.Header().Set("Content-Type", msg.MimeType)

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

	ctype := r.Header.Get("Content-Type")

	if total == 0 {
		err = client.Send(&grpc.SetInstanceVariableRequest{
			Namespace: namespace,
			Instance:  instance,
			Key:       key,
			TotalSize: 0,
			Data:      []byte{},
			MimeType:  ctype,
		})
		if err != nil {
			respond(w, nil, err)
			return
		}
	} else {
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
				MimeType:  ctype,
			})
			if err != nil {
				respond(w, nil, err)
				return
			}
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
	key := r.URL.Query().Get("var")

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

	// Set MimeType
	w.Header().Set("Content-Type", msg.MimeType)

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
	key := r.URL.Query().Get("var")

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

	ctype := r.Header.Get("Content-Type")

	if total == 0 {
		err = client.Send(&grpc.SetWorkflowVariableRequest{
			Namespace: namespace,
			Path:      path,
			Key:       key,
			TotalSize: 0,
			Data:      []byte{},
			MimeType:  ctype,
		})
		if err != nil {
			respond(w, nil, err)
			return
		}
	} else {
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
				MimeType:  ctype,
			})
			if err != nil {
				respond(w, nil, err)
				return
			}
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
	key := r.URL.Query().Get("var")

	in := &grpc.DeleteWorkflowVariableRequest{
		Namespace: namespace,
		Path:      path,
		Key:       key,
	}

	resp, err := h.client.DeleteWorkflowVariable(ctx, in)
	respond(w, resp, err)
}
