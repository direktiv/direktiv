package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gorilla/mux"
	prometheus "github.com/prometheus/client_golang/api"
	"github.com/vorteil/direktiv/pkg/flow"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
	"github.com/vorteil/direktiv/pkg/util"
	"go.uber.org/zap"
)

type flowHandler struct {
	logger     *zap.SugaredLogger
	client     grpc.FlowClient
	prometheus prometheus.Client
}

func newFlowHandler(logger *zap.SugaredLogger, router *mux.Router, conf *util.Config) (*flowHandler, error) {

	flowAddr := fmt.Sprintf("%s:6666", conf.FlowService)
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

	prometheusAddr := fmt.Sprintf("http://%s", conf.PrometheusBackend)
	logger.Infof("connecting to prometheus %s", prometheusAddr)
	h.prometheus, err = prometheus.NewClient(prometheus.Config{
		Address: prometheusAddr,
	})
	if err != nil {
		return nil, err
	}

	h.initRoutes(router)

	return h, nil

}

func (h *flowHandler) initRoutes(r *mux.Router) {

	// swagger:operation GET /api/namespaces Namespaces getNamespaces
	// Gets the list of namespaces
	// ---
	// summary: Gets the list of namespaces
	// responses:
	//   '200':
	//     "description": "successfully got list of namespaces"
	handlerPair(r, RN_ListNamespaces, "/namespaces", h.Namespaces, h.NamespacesSSE)

	// swagger:operation PUT /api/namespaces/{namespace} Namespaces createNamespace
	// Creates a new namespace
	// ---
	// summary: Creates a namespace
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace to create'
	// responses:
	//   '200':
	//     "description": "namespace has been successfully created"
	r.HandleFunc("/namespaces/{ns}", h.CreateNamespace).Name(RN_AddNamespace).Methods(http.MethodPut)

	// swagger:operation DELETE /api/namespaces/{namespace} Namespaces deleteNamespace
	// Delete a namespace
	// ---
	// summary: Delete a namespace
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace to delete'
	// responses:
	//   '200':
	//     "description": "namespace has been successfully deleted"
	r.HandleFunc("/namespaces/{ns}", h.DeleteNamespace).Name(RN_DeleteNamespace).Methods(http.MethodDelete)

	// swagger:operation POST /api/jq Other jqPlayground
	// JQ Playground is a sandbox where
	// you can test jq queries with custom data
	// ---
	// summary: JQ Playground api to test jq queries
	// parameters:
	// - in: body
	//   name: JQ payload
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
	//   '200':
	//     "description": "jq query was successful"
	r.HandleFunc("/jq", h.JQ).Name(RN_JQPlayground).Methods(http.MethodPost)

	// swagger:operation GET /api/logs Logs serverLogs
	// Gets Direktiv Server Logs
	// ---
	// summary: Get Direktiv Server Logs
	// responses:
	//   '200':
	//     "description": "successfully got server logs"
	handlerPair(r, RN_GetServerLogs, "/logs", h.ServerLogs, h.ServerLogsSSE)

	// swagger:operation GET /api/namespaces/{namespace}/logs Logs namespaceLogs
	// Gets Namespace Level Logs
	// ---
	// summary: Gets Namespace Level Logs
	// parameters:
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
	// Gets the logs of an executed instance
	// ---
	// summary: Gets Instance Logs
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
	//   description: 'target instance id'
	// responses:
	//   '200':
	//     "description": "successfully got instance logs"
	handlerPair(r, RN_GetInstanceLogs, "/namespaces/{ns}/instances/{in}/logs", h.InstanceLogs, h.InstanceLogsSSE)

	// swagger:operation GET /api/namespaces/{namespace}/tree/{workflow}?op=metrics-invoked Metrics workflowMetricsInvoked
	// Get metrics of invoked workflow instances
	// ---
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
	// Get metrics of a workflow, where the instance was successful
	// ---
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
	// Get metrics of a workflow, where the instance failed
	// ---
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
	// Get the timing metrics of a workflow's instance
	// This returns a total sum of the milliseconds a workflow has been executed for
	// ---
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
	// Get the state timing metrics of a workflow's instance
	// The returns the timing of a individual states in a workflow
	// ---
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
	// Get metrics of invoked workflows in the targeted namespace
	// ---
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
	// Get metrics of successful workflows in the targeted namespace
	// ---
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
	// Get metrics of failed workflows in the targeted namespace
	// ---
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
	// Get timing metrics of workflows in the targeted namespace
	// ---
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

	pathHandler(r, http.MethodGet, RN_GetWorkflowMetrics, "metrics-sankey", h.MetricsSankey)

	// swagger:operation GET /api/namespaces/{namespace}/vars/{variable} Variables getNamespaceVariable
	// Get the value sorted in a namespace variable
	// ---
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
	// Delete a namespace variable
	// ---
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
	// Set the value sorted in a namespace variable
	// If the target variable does not exists, it will be created
	// Variable data can be anything so long as it can be represented as a string
	// ---
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
	//   description: "Payload that contains variable data."
	//   schema:
	//     example:
	//       counter: 0
	//     type: string
	// responses:
	//   '200':
	//     "description": "successfully set namespace variable"
	r.HandleFunc("/namespaces/{ns}/vars/{var}", h.SetNamespaceVariable).Name(RN_SetNamespaceVariable).Methods(http.MethodPut)

	// swagger:operation GET /api/namespaces/{namespace}/vars Variables getNamespaceVariables
	// Gets a list of variables in a namespace
	// ---
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
	// Get the value sorted in a instance variable
	// ---
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
	// Delete a instance variable
	// ---
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
	// Set the value sorted in a instance variable
	// If the target variable does not exists, it will be created
	// Variable data can be anything so long as it can be represented as a string
	// ---
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
	//   description: "Payload that contains variable data."
	//   schema:
	//     example:
	//       counter: 0
	//     type: string
	// responses:
	//   '200':
	//     "description": "successfully set instance variable"
	r.HandleFunc("/namespaces/{ns}/instances/{instance}/vars/{var}", h.SetInstanceVariable).Name(RN_SetInstanceVariable).Methods(http.MethodPut)

	// swagger:operation GET /api/namespaces/{namespace}/instances/{instance}/vars Variables getInstanceVariables
	// Gets a list of variables in a instance
	// ---
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
	// Get the value sorted in a workflow variable
	// ---
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
	// Delete a workflow variable
	// ---
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
	// Set the value sorted in a workflow variable
	// If the target variable does not exists, it will be created
	// Variable data can be anything so long as it can be represented as a string
	// ---
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
	//   description: "Payload that contains variable data."
	//   schema:
	//     example:
	//       counter: 0
	//     type: string
	// responses:
	//   '200':
	//     "description": "successfully set workflow variable"
	pathHandler(r, http.MethodPut, RN_SetWorkflowVariable, "set-var", h.SetWorkflowVariable)

	// swagger:operation GET /api/namespaces/{namespace}/tree/{workflow}?op=vars Variables getWorkflowVariables
	// Gets a list of variables in a workflow
	// ---
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
	// Gets the list of namespace secrets
	// ---
	// summary: Get List of Namespace Secrets
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// responses:
	//   '200':
	//     "description": "successfully got namespace secrets"
	handlerPair(r, RN_ListSecrets, "/namespaces/{ns}/secrets", h.Secrets, h.SecretsSSE)

	// swagger:operation PUT /api/namespaces/{namespace}/secrets/{secret} Secrets createSecret
	// Create a namespace secret
	// ---
	// summary: Create a Namespace Secret
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
	//   description: Payload that contains secret data
	//   schema:
	//     example:
	//       data: "8QwFLg%D$qg*3r++`{D<BAp~4mB49^"
	//     type: object
	//     required:
	//       - data
	//     properties:
	//       data:
	//         type: string
	//         description: Secret data to be set
	// responses:
	//   '200':
	//     "description": "successfully created namespace secret"
	r.HandleFunc("/namespaces/{ns}/secrets/{secret}", h.SetSecret).Name(RN_CreateSecret).Methods(http.MethodPut)

	// swagger:operation DELETE /api/namespaces/{namespace}/secrets/{secret} Secrets deleteSecret
	// Delete a namespace secret
	// ---
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
	//   '200':
	//     "description": "successfully deleted namespace secret"
	r.HandleFunc("/namespaces/{ns}/secrets/{secret}", h.DeleteSecret).Name(RN_DeleteSecret).Methods(http.MethodDelete)

	// swagger:operation GET /api/namespaces/{namespace}/instances/{instance} Instances getInstance
	// Gets the details of a executed workflow instance in this namespace
	// ---
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
	//   '200':
	//     "description": "successfully got instance"
	handlerPair(r, RN_GetInstance, "/namespaces/{ns}/instances/{instance}", h.Instance, h.InstanceSSE)

	// swagger:operation GET /api/namespaces/{namespace}/instances Instances getInstanceList
	// Gets a list of instances in a namespace
	// ---
	// summary: Get List Instances
	// parameters:
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
	// Gets the input an instance was provided when executed
	// ---
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
	// Gets the output an instance was provided when executed
	// ---
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

	// swagger:operation POST /api/namespaces/{namespace}/instances/{instance}/cancel Instances cancelInstance
	// Cancel a currently pending instance
	// ---
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
	// Broadcast a cloud event to a namespace
	// Cloud events posted to this api will be picked up by any workflows listening to the same event type on the namescape.
	// The body of this request should follow the cloud event core specification defined at https://github.com/cloudevents/spec
	// ---
	// summary: Broadcast Cloud Event
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: body
	//   name: cloudevent
	//   description: Cloud Event request to be sent.
	//   schema:
	//     type: object
	// responses:
	//   '200':
	//     "description": "successfully sent cloud event"
	r.HandleFunc("/namespaces/{ns}/broadcast", h.BroadcastCloudevent).Name(RN_NamespaceEvent).Methods(http.MethodPost)

	// swagger:operation GET /api/namespaces/{namespace}/tree/{workflow}?op=logs Logs getWorkflowLogs
	// Get workflow level logs
	// ---
	// summary: Get Workflow Level Logs
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
	//     "description": "successfully got workflow logs"
	pathHandlerPair(r, RN_GetWorkflowLogs, "logs", h.WorkflowLogs, h.WorkflowLogsSSE)

	// swagger:operation PUT /api/namespaces/{namespace}/tree/{directory}?op=create-directory Directory createDirectory
	// Creates a directory at the target path
	// ---
	// summary: Create a Directory
	// parameters:
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
	//   '200':
	//     "description": "successfully created directory"
	pathHandler(r, http.MethodPut, RN_CreateDirectory, "create-directory", h.CreateDirectory)

	// swagger:operation PUT /api/namespaces/{namespace}/tree/{workflow}?op=create-workflow Workflows createWorkflow
	// Creates a workflow at the target path
	// The body of this request should contain the workflow yaml
	// ---
	// summary: Create a Workflow
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
	//   '200':
	//     "description": "successfully created workflow"
	pathHandler(r, http.MethodPut, RN_CreateWorkflow, "create-workflow", h.CreateWorkflow)

	// swagger:operation POST /api/namespaces/{namespace}/tree/{workflow}?op=update-workflow Workflows updateWorkflow
	// Updates a workflow at the target path
	// The body of this request should contain the workflow yaml you want to update to
	// ---
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
	pathHandler(r, http.MethodPost, RN_SaveWorkflow, "save-workflow", h.SaveWorkflow)
	// TODO: SWAGGER-SPEC
	pathHandler(r, http.MethodPost, RN_DiscardWorkflow, "discard-workflow", h.DiscardWorkflow)
	// TODO: SWAGGER-SPEC
	pathHandler(r, http.MethodDelete, RN_DeleteNode, "delete-node", h.DeleteNode)

	// TODO: SWAGGER-SPEC
	pathHandler(r, http.MethodPut, RN_CreateNodeAttributes, "create-node-attributes", h.CreateNodeAttributes)
	// TODO: SWAGGER-SPEC
	pathHandler(r, http.MethodDelete, RN_DeleteNodeAttributes, "delete-node-attributes", h.DeleteNodeAttributes)

	// TODO: SWAGGER-SPEC
	pathHandlerPair(r, RN_GetWorkflowTags, "tags", h.GetTags, h.GetTagsSSE)
	// TODO: SWAGGER-SPEC
	pathHandlerPair(r, RN_GetWorkflowRefs, "refs", h.GetRefs, h.GetRefsSSE)
	// TODO: SWAGGER-SPEC
	pathHandlerPair(r, RN_GetWorkflowRefs, "revisions", h.GetRevisions, h.GetRevisionsSSE)
	// TODO: SWAGGER-SPEC
	pathHandler(r, http.MethodPost, RN_DeleteRevision, "delete-revision", h.DeleteRevision)
	// TODO: SWAGGER-SPEC
	pathHandler(r, http.MethodPost, RN_Tag, "tag", h.Tag)
	// TODO: SWAGGER-SPEC
	pathHandler(r, http.MethodPost, RN_Untag, "untag", h.Untag)
	// TODO: SWAGGER-SPEC
	pathHandler(r, http.MethodPost, RN_Retag, "retag", h.Retag)
	// TODO: SWAGGER-SPEC
	pathHandlerPair(r, RN_GetWorkflowRouter, "router", h.Router, h.RouterSSE)
	// TODO: SWAGGER-SPEC
	pathHandler(r, http.MethodPost, RN_EditWorkflowRouter, "edit-router", h.EditRouter)
	// TODO: SWAGGER-SPEC
	pathHandler(r, http.MethodPost, RN_ValidateRef, "validate-ref", h.ValidateRef)
	// TODO: SWAGGER-SPEC
	pathHandler(r, http.MethodPost, RN_ValidateRouter, "validate-router", h.ValidateRouter)

	// swagger:operation POST /api/namespaces/{namespace}/tree/{workflow}?op=set-workflow-event-logging Workflows setWorkflowCloudEventLogs
	// Set Cloud Event for Workflow to Log to
	// When configured type `direktiv.instanceLog` cloud events will be generated with the `logger` parameter set to the
	// conifgured value
	// Workflows can be configured to generate cloud events on their namespace
	// anything the log parameter produces data. Please find more information on this topic below:
	// https://docs.direktiv.io/docs/examples/logging.html
	// ---
	// summary: Set Cloud Event for Workflow to Log to
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
	//   name: Cloud Event Logger
	//   description: Cloud event logger to target
	//   schema:
	//     example:
	//       logger: "mylog"
	//     type: object
	//     required:
	//       - logger
	//     properties:
	//       logger:
	//         type: string
	//         description: Target Cloud Event
	// responses:
	//   '200':
	//     "description": "successfully update workflow"
	pathHandler(r, http.MethodPost, RN_UpdateWorkflow, "set-workflow-event-logging", h.SetWorkflowEventLogging)

	// swagger:operation POST /api/namespaces/{namespace}/tree/{workflow}?op=toggle Workflows toggleWorkflow
	// Toggle's whether or not a workflow is active
	// Disabled workflows cannot be invoked
	// ---
	// summary: Set Cloud Event for Workflow to Log to
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
	//   name: Workflow Live Status
	//   description: Whether or not the workflow is alive or disabled
	//   schema:
	//     example:
	//       live: false
	//     type: object
	//     required:
	//       - live
	//     properties:
	//       live:
	//         type: boolean
	//         description: Workflow live status
	// responses:
	//   '200':
	//     "description": "successfully updated workflow live status"
	pathHandler(r, http.MethodPost, RN_UpdateWorkflow, "toggle", h.ToggleWorkflow)

	// swagger:operation POST /api/namespaces/{namespace}/tree/{workflow}?op=execute Workflows executeWorkflow
	// Executes a workflow with optionally some input provided in the request body as json
	// ---
	// summary: Execute a Workflow
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
	//   name: Workflow Input
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
	pathHandler(r, http.MethodPost, RN_ExecuteWorkflow, "execute", h.ExecuteWorkflow)

	// swagger:operation POST /api/namespaces/{namespace}/tree/{workflow}?op=wait Workflows awaitExecuteWorkflowBody
	// Executes a workflow with optionally some input provided in the request body as json
	// This path will wait until the workflow execution has completed and return the instance output
	// NOTE: Input can also be provided with the `input.X` query parameters; Where `X` is the json
	// key. Only top level json keys are supported when providing input with query parameters. Input query
	// parameters are only read if the request has not body
	// ---
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
	// Executes a workflow with optionally some input provided in the request body as json
	// This path will wait until the workflow execution has completed and return the instance output
	// NOTE: Input can also be provided with the `input.X` query parameters; Where `X` is the json
	// key. Only top level json keys are supported when providing input with query parameters
	// ---
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

	// swagger:operation GET /api/namespaces/{namespace}/tree/{nodePath} Registries getNodes
	// Gets Workflow and Directory Nodes at nodePath
	// ---
	// summary: Get List of Namespace Nodes
	// tags:
	// - "Directory"
	// - "Workflows"
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
	//   '200':
	//     "description": "successfully got namespace nodes"
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
	namespace := mux.Vars(r)["ns"]

	in := &grpc.CreateNamespaceRequest{
		Name: namespace,
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

func (h *flowHandler) CreateNodeAttributes(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	in := new(grpc.CreateNodeAttributesRequest)

	err := unmarshalBody(r, in)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in.Namespace = namespace
	in.Path = path

	resp, err := h.client.CreateNodeAttributes(ctx, in)
	respond(w, resp, err)
}

func (h *flowHandler) DeleteNodeAttributes(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	in := new(grpc.DeleteNodeAttributesRequest)

	err := unmarshalBody(r, in)
	if err != nil {
		respond(w, nil, err)
		return
	}

	in.Namespace = namespace
	in.Path = path

	resp, err := h.client.DeleteNodeAttributes(ctx, in)
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

func (h *flowHandler) WaitWorkflow(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, ref := pathAndRef(r)

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
		if len(m) > 0 {
			input, err = json.Marshal(m)
			if err != nil {
				respond(w, nil, err)
				return
			}
		}
	}

	in := &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      path,
		Ref:       ref,
		Input:     input,
	}

	resp, err := h.client.StartWorkflow(ctx, in)
	if err != nil {
		respond(w, nil, err)
		return
	}

	w.Header().Set("Direktiv-Instance-Id", resp.Instance)

	c, err := h.client.InstanceStream(ctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		respond(w, nil, err)
		return
	}
	defer c.CloseSend()

	for {
		status, err := c.Recv()
		if err != nil {
			respond(w, nil, err)
			return
		}

		if s := status.Instance.GetStatus(); s == flow.StatusComplete {

			_ = c.CloseSend()

			output, err := h.client.InstanceOutput(ctx, &grpc.InstanceOutputRequest{
				Namespace: namespace,
				Instance:  resp.Instance,
			})

			if err != nil {
				respond(w, nil, err)
				return
			}

			data := output.Data

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
					data, _ = json.Marshal(x)
				} else {
					data, _ = json.Marshal(nil)
				}
			}

			var x interface{}
			err = json.Unmarshal(data, &x)
			if err != nil {
				respond(w, nil, err)
				return
			}

			rawo := r.URL.Query().Get("raw-output")
			if rawo == "true" {

				if x == nil {
					data = make([]byte, 0)
				} else if str, ok := x.(string); ok {
					data = []byte(str)
					b64, err := base64.StdEncoding.DecodeString(str)
					if err == nil {
						data = []byte(b64)
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

			_, _ = io.Copy(w, bytes.NewReader(data))
			return

		} else if s == flow.StatusFailed {
			w.Header().Set("Direktiv-Instance-Error-Code", status.Instance.ErrorCode)
			w.Header().Set("Direktiv-Instance-Error-Message", status.Instance.ErrorMessage)
			code := http.StatusInternalServerError
			http.Error(w, fmt.Sprintf("An error occurred executing instance %s: %s: %s", resp.Instance, status.Instance.ErrorCode, status.Instance.ErrorMessage), code)
			return
		} else if s == flow.StatusCrashed {
			code := http.StatusInternalServerError
			http.Error(w, fmt.Sprintf("An internal error occurred executing instance: %s", resp.Instance), code)
			return
		} else {
			continue
		}

	}

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
	key := r.URL.Query().Get("var")

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
