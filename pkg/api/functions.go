// Package classification Direktiv API
//
// Direktiv Open API Specification
// Direktiv Documentation can be found at https://docs.direktiv.io/
//
// Terms Of Service:
//
//	Schemes: http, https
//	Host: localhost
//	Version: 1.0.0
//	Contact: info@direktiv.io
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Security:
//	- direktiv-token:
//
//	SecurityDefinitions:
//	direktiv-token:
//	     type: apiKey
//	     name: KEY
//	     in: header
//
// swagger:meta
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	igrpc "github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/functions"
	"github.com/direktiv/direktiv/pkg/functions/grpc"
	grpcfunc "github.com/direktiv/direktiv/pkg/functions/grpc"
	"github.com/gorilla/mux"
	"github.com/heroku/docker-registry-client/registry"
)

type functionHandler = flowHandler

func (h *functionHandler) initFunctionsRoutes(r *mux.Router) {
	// swagger:operation GET /api/logs/{pod} podLogs
	// ---
	// description: |
	//    Watches logs of the pods for a service. This can be a namespace service or a workflow service.
	// summary: Watch Pod Logs
	// tags:
	// - "Pod"
	// parameters:
	// - in: path
	//   name: pod
	//   type: string
	//   required: true
	//   description: 'pod name'
	// responses:
	//   '200':
	//     "description": "successfully watching pod logs"
	r.HandleFunc("/logs/pod/{pod}", h.watchPodLogs).Methods(http.MethodGet).Name(RN_WatchPodLogs)

	// namespace

	// swagger:operation GET /api/functions/namespaces/{namespace} getNamespaceServiceList
	// ---
	// description: |
	//   Gets a list of namespace knative services.
	// summary: Get Namespace Services List
	// tags:
	// - "Namespace Services"
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// responses:
	//   '200':
	//     "description": "successfully got services list"
	handlerPair(r, RN_ListNamespaceServices, "/namespaces/{ns}", h.listNamespaceServices, h.listNamespaceServicesSSE)

	// swagger:operation GET /api/functions/namespaces/{namespace}/function/{serviceName}/revisions/{revisionGeneration}/pods listNamespaceServiceRevisionPods
	// ---
	// description: |
	//   List a revisions pods of a namespace scoped knative service.
	//   The target revision generation is the number suffix on a revision.
	//   Example: A revision named 'namespace-direktiv-fast-request-00003' would have the revisionGeneration '00003'.
	// summary: Get Namespace Service Revision Pods List
	// tags:
	// - "Namespace Services"
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: serviceName
	//   type: string
	//   required: true
	//   description: 'target service name'
	// - in: path
	//   name: revisionGeneration
	//   type: string
	//   required: true
	//   description: 'target revision generation'
	// responses:
	//   '200':
	//     "description": "successfully got list of a service revision pods"
	handlerPair(r, RN_ListNamespacePods, "/namespaces/{ns}/function/{svn}/revisions/{rev}/pods", h.listNamespacePods, h.listNamespacePodsSSE)

	// swagger:operation GET /api/functions/namespaces/{namespace}/function/{serviceName} watchNamespaceServiceRevisionList
	// ---
	// description: |
	//   Watch the revision list of a namespace scoped knative service.
	//   Note: This is a Server-Sent-Event endpoint, and will not work with the default swagger client.
	// summary: Watch Namespace Service Revision
	// tags:
	// - "Namespace Services"
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: serviceName
	//   type: string
	//   required: true
	//   description: 'target service name'
	// produces:
	//    - "text/event-stream"
	// responses:
	//   '200':
	//     "description": "successfully watching service"
	r.HandleFunc("/namespaces/{ns}/function/{svn}", h.singleNamespaceServiceSSE).Name(RN_WatchServices).Methods(http.MethodGet).Headers("Accept", "text/event-stream")

	// swagger:operation GET /api/functions/namespaces/{namespace}/function/{serviceName}/revisions watchNamespaceServiceRevisionList
	// ---
	// description: |
	//   Watch the revision list of a namespace scoped knative service.
	//   Note: This is a Server-Sent-Event endpoint, and will not work with the default swagger client.
	// summary: Watch Namespace Service Revision List
	// tags:
	// - "Namespace Services"
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: serviceName
	//   type: string
	//   required: true
	//   description: 'target service name'
	// produces:
	//    - "text/event-stream"
	// responses:
	//   '200':
	//     "description": "successfully watching service revisions"
	r.HandleFunc("/namespaces/{ns}/function/{svn}/revisions", h.watchNamespaceRevisions).Name(RN_WatchRevisions).Methods(http.MethodGet).Headers("Accept", "text/event-stream")

	// swagger:operation GET /api/functions/namespaces/{namespace}/function/{serviceName}/revisions/{revisionGeneration} watchNamespaceServiceRevision
	// ---
	// description: |
	//   Watch a namespace scoped knative service revision.
	//   The target revision generation is the number suffix on a revision.
	//   Example: A revision named 'namespace-direktiv-fast-request-00003' would have the revisionGeneration '00003'.
	//   Note: This is a Server-Sent-Event endpoint, and will not work with the default swagger client.
	// summary: Watch Namespace Service Revision
	// tags:
	// - "Namespace Services"
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: serviceName
	//   type: string
	//   required: true
	//   description: 'target service name'
	// - in: path
	//   name: revisionGeneration
	//   type: string
	//   required: true
	//   description: 'target revision generation'
	// produces:
	//    - "text/event-stream"
	// responses:
	//   '200':
	//     "description": "successfully watching service revision"
	r.HandleFunc("/namespaces/{ns}/function/{svn}/revisions/{rev}", h.watchNamespaceRevision).Name(RN_WatchRevisions).Methods(http.MethodGet).Headers("Accept", "text/event-stream")

	// swagger:operation POST /api/functions/namespaces/{namespace} createNamespaceService
	// ---
	// description: |
	//   Creates namespace scoped knative service.
	//   Service Names are unique on a scope level.
	//   These services can be used as functions in workflows, more about this can be read here:
	//   https://docs.direktiv.io/getting_started/functions-intro/
	// summary: Create Namespace Service
	// tags:
	// - "Namespace Services"
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: body
	//   name: Service
	//   description: Payload that contains information on new service
	//   required: true
	//   schema:
	//     type: object
	//     example:
	//       name: "fast-request"
	//       image: "direktiv/request:v12"
	//       cmd: ""
	//       minScale: 1
	//       size: 0
	//     required:
	//       - name
	//       - image
	//       - cmd
	//       - minScale
	//       - size
	//     properties:
	//       name:
	//         type: string
	//         description: Name of new service
	//       image:
	//         type: string
	//         description: Target image a service will use
	//       cmd:
	//         type: string
	//       minScale:
	//         type: integer
	//         description: Minimum amount of service pods to be live
	//       size:
	//         type: integer
	//         description: Size of created service pods, 0 = small, 1 = medium, 2 = large
	//       envs:
	//         type: object
	//         additionalProperties:
	//           type: string
	// responses:
	//   '200':
	//     "description": "successfully created service"
	r.HandleFunc("/namespaces/{ns}", h.createNamespaceService).Methods(http.MethodPost).Name(RN_CreateNamespaceService)

	// swagger:operation DELETE /api/functions/namespaces/{namespace}/function/{serviceName} deleteNamespaceService
	// ---
	// description: |
	//   Deletes namespace scoped knative service and all its revisions.
	// summary: Delete Namespace Service
	// tags:
	// - "Namespace Services"
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: serviceName
	//   type: string
	//   required: true
	//   description: 'target service name'
	// responses:
	//   '200':
	//     "description": "successfully deleted service"
	r.HandleFunc("/namespaces/{ns}/function/{svn}", h.deleteNamespaceService).Methods(http.MethodDelete).Name(RN_DeleteNamespaceServices)

	// swagger:operation GET /api/functions/namespaces/{namespace}/function/{serviceName} getNamespaceService
	// ---
	// description: |
	//   Get details of a namespace scoped knative service.
	// summary: Get Namespace Service Details
	// tags:
	// - "Namespace Services"
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: serviceName
	//   type: string
	//   required: true
	//   description: 'target service name'
	// responses:
	//   '200':
	//     "description": "successfully got service details"
	r.HandleFunc("/namespaces/{ns}/function/{svn}", h.getNamespaceService).Methods(http.MethodGet).Name(RN_GetNamespaceService)

	// swagger:operation POST /api/functions/namespaces/{namespace}/function/{serviceName} updateNamespaceService
	// ---
	// description: |
	//   Creates a new namespace scoped knative service revision.
	// summary: Create Namespace Service Revision
	// tags:
	// - "Namespace Services"
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: serviceName
	//   type: string
	//   required: true
	//   description: 'target service name'
	// - in: body
	//   name: Service
	//   description: Payload that contains information on service revision
	//   required: true
	//   schema:
	//     type: object
	//     example:
	//       image: "direktiv/request:v10"
	//       cmd: ""
	//       minScale: 1
	//       size: "small"
	//     required:
	//       - image
	//       - cmd
	//       - minScale
	//       - size
	//     properties:
	//       image:
	//         type: string
	//         description: Target image a service will use
	//       cmd:
	//         type: string
	//       minScale:
	//         type: integer
	//         description: Minimum amount of service pods to be live
	//       size:
	//         type: string
	//         description: Size of created service pods
	//         enum:
	//           - small
	//           - medium
	//           - large
	// responses:
	//   '200':
	//     "description": "successfully created service revision"
	r.HandleFunc("/namespaces/{ns}/function/{svn}", h.updateNamespaceService).Methods(http.MethodPost).Name(RN_UpdateNamespaceService)

	// swagger:operation DELETE /api/functions/namespaces/{namespace}/function/{serviceName}/revisions/{revisionGeneration} deleteNamespaceRevision
	// ---
	// description: |
	//   Delete a namespace scoped knative service revision.
	//   The target revision generation is the number suffix on a revision.
	//   Example: A revision named 'namespace-direktiv-fast-request-00003' would have the revisionGeneration '00003'.
	// summary: Delete Namespace Service Revision
	// tags:
	// - "Namespace Services"
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: serviceName
	//   type: string
	//   required: true
	//   description: 'target service name'
	// - in: path
	//   name: revisionGeneration
	//   type: string
	//   required: true
	//   description: 'target revision generation'
	// responses:
	//   '200':
	//     "description": "successfully deleted service revision"
	r.HandleFunc("/namespaces/{ns}/function/{svn}/revisions/{rev}", h.deleteNamespaceServiceRevision).Methods(http.MethodDelete).Name(RN_DeleteNamespaceServiceRevision)

	// workflow

	// swagger:operation GET /api/functions/namespaces/{namespace}/tree/{workflow}?op=services listWorkflowServices
	// ---
	// description: |
	//   Gets a list of workflow knative services.
	// summary: Get Workflow Services List
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
	// tags:
	// - "Workflow Services"
	// responses:
	//   '200':
	//     "description": "successfully got services list"
	pathHandlerPair(r, RN_ListWorkflowServices, "services", h.listWorkflowServices, h.listWorkflowServicesSSE)

	// swagger:operation DELETE /api/functions/namespaces/{namespace}/tree/{workflow}?op=delete-service deleteWorkflowService
	// ---
	// description: |
	//   Deletes workflow scoped knative service.
	// summary: Delete Namespace Service
	// tags:
	// - "Workflow Services"
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
	//   name: svn
	//   type: string
	//   required: true
	//   description: 'target service name'
	// - in: query
	//   name: version
	//   type: string
	//   required: true
	//   description: 'target service version'
	// responses:
	//   '200':
	//     "description": "successfully deleted service"
	pathHandler(r, http.MethodDelete, RN_DeleteWorkflowServices, "delete-service", h.deleteWorkflowServices)

	// swagger:operation GET /api/functions/namespaces/{namespace}/tree/{workflow}?op=pods listWorkflowServiceRevisionPods
	// ---
	// description: |
	//   List a revisions pods of a workflow scoped knative service.
	//   The target revision generation (rev query) is the number suffix on a revision.
	//   Example: A revision named 'workflow-10640097968065193909-get-00001' would have the revisionGeneration '00001'.
	// summary: Get Workflow Service Revision Pods List
	// tags:
	// - "Workflow Services"
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
	//   name: svn
	//   type: string
	//   required: true
	//   description: 'target service name'
	// - in: query
	//   name: rev
	//   type: string
	//   required: true
	//   description: 'target service revison'
	// responses:
	//   '200':
	//     "description": "successfully got list of a service revision pods"
	pathHandlerPair(r, RN_ListWorkflowServices, "pods", h.listWorkflowPods, h.listWorkflowPodsSSE)

	// swagger:operation GET /api/functions/namespaces/{namespace}/tree/{workflow}?op=function getWorkflowService
	// ---
	// description: |
	//   Get a workflow scoped knative service details.
	//   Note: This is a Server-Sent-Event endpoint, and will not work with the default swagger client.
	// summary: Get Workflow Service Details
	// tags:
	// - "Workflow Services"
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
	//   name: svn
	//   type: string
	//   required: true
	//   description: 'target service name'
	// - in: query
	//   name: version
	//   type: string
	//   required: true
	//   description: 'target service version'
	// responses:
	//   '200':
	//     "description": "successfully got service details"
	pathHandlerPair(r, RN_ListWorkflowServices, "function", h.singleWorkflowService, h.singleWorkflowServiceSSE)

	// swagger:operation GET /api/functions/namespaces/{namespace}/tree/{workflow}?op=function-revisions getWorkflowServiceRevisionList
	// ---
	// description: |
	//   Get the revision list of a workflow scoped knative service.
	// summary: Get Workflow Service Revision List
	// tags:
	// - "Workflow Services"
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: query
	//   name: svn
	//   type: string
	//   required: true
	//   description: 'target service name'
	// - in: query
	//   name: version
	//   type: string
	//   required: true
	//   description: 'target service version'
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// responses:
	//   '200':
	//     "description": "successfully got service revisions"
	pathHandlerPair(r, RN_ListWorkflowServices, "function-revisions", h.singleWorkflowServiceRevisions, h.singleWorkflowServiceRevisionsSSE)

	// swagger:operation GET /api/functions/namespaces/{namespace}/tree/{workflow}?op=function-revision getWorkflowServiceRevision
	// ---
	// description: |
	//   Get a workflow scoped knative service revision.
	//   This will return details on a single revision.
	//   The target revision generation (rev query) is the number suffix on a revision.
	//   Example: A revision named 'workflow-10640097968065193909-get-00001' would have the revisionGeneration '00001'.
	// summary: Get Workflow Service Revision
	// tags:
	// - "Workflow Services"
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
	//   name: svn
	//   type: string
	//   required: true
	//   description: 'target service name'
	// - in: query
	//   name: rev
	//   type: string
	//   required: true
	//   description: 'target service revison'
	// responses:
	//   '200':
	//     "description": "successfully got service revision details"
	pathHandlerPair(r, RN_ListWorkflowServices, "function-revision", h.singleWorkflowServiceRevision, h.singleWorkflowServiceRevisionSSE)

	// r.HandleFunc("/namespaces/{ns}/workflow/{wf}", h.createWorkflowService).Methods(http.MethodPost).Name(RN_CreateNamespaceService)
	// r.HandleFunc("/namespaces/{ns}/function/{svn}", h.deleteNamespaceService).Methods(http.MethodDelete).Name(RN_DeleteNamespaceServices)
	// r.HandleFunc("/namespaces/{ns}/function/{svn}", h.getNamespaceService).Methods(http.MethodGet).Name(RN_GetNamespaceService)
	// r.HandleFunc("/namespaces/{ns}/function/{svn}", h.updateNamespaceService).Methods(http.MethodPost).Name(RN_UpdateNamespaceService)
	// r.HandleFunc("/namespaces/{ns}/function/{svn}", h.updateNamespaceServiceTraffic).Methods(http.MethodPatch).Name(RN_UpdateNamespaceServiceTraffic)
	// r.HandleFunc("/namespaces/{ns}/function/{svn}/revisions/{rev}", h.deleteNamespaceRevision).Methods(http.MethodDelete).Name(RN_DeleteNamespaceRevision)

	// Registry ..

	// swagger:operation POST /api/functions/registries/test testRegistry
	// ---
	// description: |
	//   Test a registry with provided url, username and token
	// summary: Test a registry to make sure the connection is okay
	// parameters:
	// - in: body
	//   name: Registry Payload
	//   required: true
	//   description: Payload that contains registry data
	//   schema:
	//     type: object
	//     example:
	//       username: "admin"
	//       url: "https://prod.customreg.io"
	//       password: "8QwFLg%D$qg*"
	//     required:
	//       - url
	//       - username
	//       - password
	//     properties:
	//       password:
	//         type: string
	//         description: "token to authenticate with the registry"
	//       username:
	//         type: string
	//         description: "username to authenticate with the registry"
	//       url:
	//         type: string
	//         description: The url to test if the registry is valid
	// responses:
	//   '200':
	//     "description": "registry is valid"
	//   '401':
	//     "description": "unauthorized to access the registry"
	r.HandleFunc("/registries/test", h.testRegistry).Methods(http.MethodPost).Name(RN_TestRegistry)

	// swagger:operation GET /api/functions/registries/namespaces/{namespace} Registries getRegistries
	// ---
	// description: |
	//   Gets the list of namespace registries.
	// summary: Get List of Namespace Registries
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// responses:
	//   '200':
	//     "description": "successfully got namespace registries"
	r.HandleFunc("/registries/namespaces/{ns}", h.getRegistries).Methods(http.MethodGet).Name(RN_ListRegistries)

	// swagger:operation POST /api/functions/registries/namespaces/{namespace} Registries createRegistry
	// ---
	// description: |
	//   Create a namespace container registry.
	//   This can be used to connect your workflows to private container registries that require tokens.
	//   The data property in the body is made up from the registry user and token. It follows the pattern :
	//   data=USER:TOKEN
	// summary: Create a Namespace Container Registry
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: body
	//   name: Registry Payload
	//   required: true
	//   description: Payload that contains registry data
	//   schema:
	//     type: object
	//     example:
	//       data: "admin:8QwFLg%D$qg*"
	//       reg: "https://prod.customreg.io"
	//     required:
	//       - data
	//       - reg
	//     properties:
	//       data:
	//         type: string
	//         description: "Target registry connection data containing the user and token."
	//       reg:
	//         type: string
	//         description: Target registry URL
	// responses:
	//   '200':
	//     "description": "successfully created namespace registry"
	r.HandleFunc("/registries/namespaces/{ns}", h.createRegistry).Methods(http.MethodPost).Name(RN_CreateRegistry)

	// swagger:operation DELETE /api/functions/registries/namespaces/{namespace} Registries deleteRegistry
	// ---
	// description: |
	//   Delete a namespace container registry
	// summary: Delete a Namespace Container Registry
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: body
	//   name: Registry Payload
	//   required: true
	//   description: Payload that contains registry data
	//   schema:
	//     example:
	//       data: "admin:8QwFLg%D$qg*"
	//       reg: "https://prod.customreg.io"
	//     type: object
	//     required:
	//       - reg
	//     properties:
	//       reg:
	//         type: string
	//         description: Target registry URL
	// responses:
	//   '200':
	//     "description": "successfully delete namespace registry"
	r.HandleFunc("/registries/namespaces/{ns}", h.deleteRegistry).Methods(http.MethodDelete).Name(RN_DeleteRegistry)
}

func (h *functionHandler) deleteRegistry(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	// Get and Validate namespace exists
	nsResp, err := h.client.Namespace(r.Context(), &igrpc.NamespaceRequest{
		Name: mux.Vars(r)["ns"],
	})
	if err != nil {
		respond(w, nil, err)
		return
	}

	d := make(map[string]string)

	err = json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		respond(w, nil, err)
	}
	reg := d["reg"]

	resp, err := h.functionsClient.DeleteRegistry(r.Context(), &grpc.FunctionsDeleteRegistryRequest{
		Namespace: &nsResp.Namespace.Name,
		Name:      &reg,
	})

	respond(w, resp, err)
}

func (h *functionHandler) createRegistry(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	// Get and Validate namespace exists
	nsResp, err := h.client.Namespace(r.Context(), &igrpc.NamespaceRequest{
		Name: mux.Vars(r)["ns"],
	})
	if err != nil {
		respond(w, nil, err)
		return
	}

	d := make(map[string]string)

	err = json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		respond(w, nil, err)
	}
	reg := d["reg"]

	resp, err := h.functionsClient.StoreRegistry(r.Context(), &grpc.FunctionsStoreRegistryRequest{
		Namespace: &nsResp.Namespace.Name,
		Name:      &reg,
		Data:      []byte(d["data"]),
	})

	respond(w, resp, err)
}

func (h *functionHandler) testRegistry(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	d := make(map[string]string)

	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		respond(w, nil, err)
		return
	}

	if d["url"] == "" || d["username"] == "" || d["password"] == "" {
		respond(w, nil, errors.New("url, username and password need to be provided"))
		return
	}

	_, err = registry.NewInsecure(d["url"], d["username"], d["password"])
	if err != nil {
		respond(w, nil, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, strings.NewReader("{}"))
	if err != nil {
		h.logger.Errorf("Failed to write response: %v.", err)
	}
}

func (h *functionHandler) getRegistries(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	// Get and Validate namespace exists
	nsResp, err := h.client.Namespace(r.Context(), &igrpc.NamespaceRequest{
		Name: mux.Vars(r)["ns"],
	})
	if err != nil {
		respond(w, nil, err)
		return
	}

	var resp *grpc.FunctionsGetRegistriesResponse
	resp, err = h.functionsClient.GetRegistries(r.Context(), &grpc.FunctionsGetRegistriesRequest{
		Namespace: &nsResp.Namespace.Name,
	})

	respond(w, resp, err)
}

func (h *functionHandler) listNamespaceServices(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()

	resp, err := h.client.Namespace(ctx, &igrpc.NamespaceRequest{
		Name: mux.Vars(r)["ns"],
	})
	if err != nil {
		respond(w, nil, err)
		return
	}

	annotations := make(map[string]string)
	annotations[functions.ServiceHeaderNamespaceID] = resp.Namespace.GetOid()
	annotations[functions.ServiceHeaderScope] = functions.PrefixNamespace
	h.listServices(annotations, w, r)
}

func (h *functionHandler) listNamespaceServicesSSE(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()

	resp, err := h.client.Namespace(ctx, &igrpc.NamespaceRequest{
		Name: mux.Vars(r)["ns"],
	})
	if err != nil {
		respond(w, nil, err)
		return
	}

	annotations := make(map[string]string)
	annotations[functions.ServiceHeaderNamespaceID] = resp.Namespace.GetOid()
	annotations[functions.ServiceHeaderScope] = functions.PrefixNamespace

	h.listServicesSSE(annotations, w, r)
}

func (h *functionHandler) listWorkflowServices(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())
	annotations := make(map[string]string)
	wf := bytedata.ShortChecksum("/" + mux.Vars(r)["path"])
	annotations[functions.ServiceHeaderWorkflowID] = wf
	annotations[functions.ServiceHeaderNamespaceName] = mux.Vars(r)["ns"]
	h.listServices(annotations, w, r)
}

func (h *functionHandler) listWorkflowServicesSSE(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())
	annotations := make(map[string]string)
	wf := bytedata.ShortChecksum("/" + mux.Vars(r)["path"])
	annotations[functions.ServiceHeaderWorkflowID] = wf
	annotations[functions.ServiceHeaderNamespaceName] = mux.Vars(r)["ns"]
	h.listServicesSSE(annotations, w, r)
}

func (h *functionHandler) singleNamespaceServiceSSE(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	annotations := make(map[string]string)
	annotations[functions.ServiceHeaderScope] = functions.PrefixNamespace
	annotations[functions.ServiceHeaderName] = mux.Vars(r)["svn"]

	h.listServicesSSE(annotations, w, r)
}

func (h *functionHandler) singleWorkflowService(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	http.Error(w, "text/event-stream only", http.StatusBadRequest)
}

func (h *functionHandler) singleWorkflowServiceSSE(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	vers := r.URL.Query().Get("version")

	hash, err := strconv.ParseUint(vers, 10, 64)
	if err != nil {
		respond(w, nil, err)
		return
	}

	svc := functions.AssembleWorkflowServiceName(hash)

	annotations := make(map[string]string)

	annotations[functions.ServiceKnativeHeaderName] = svc

	h.listServicesSSE(annotations, w, r)
}

func (h *functionHandler) listServicesSSE(
	annotations map[string]string, w http.ResponseWriter, r *http.Request,
) {
	grpcReq := grpcfunc.FunctionsWatchFunctionsRequest{
		Annotations: annotations,
	}

	client, err := h.functionsClient.WatchFunctions(r.Context(), &grpcReq)
	if err != nil {
		respond(w, nil, err)
		return
	}

	ch := make(chan interface{}, 1)

	defer func() {
		_ = client.CloseSend()

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
			x, err := client.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x
		}
	}()

	sse(w, ch)
}

func (h *functionHandler) listServices(
	annotations map[string]string, w http.ResponseWriter, r *http.Request,
) {
	grpcReq := grpcfunc.FunctionsListFunctionsRequest{
		Annotations: annotations,
	}

	resp, err := h.functionsClient.ListFunctions(r.Context(), &grpcReq)
	respond(w, resp, err)
}

// sse

func (h *functionHandler) deleteWorkflowServices(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	path, _ := pathAndRef(r)

	ver := r.URL.Query().Get("version")
	svn := r.URL.Query().Get("svn")

	if ver == "" {
		respond(w, nil, errors.New("version is missing in queries"))
		return
	} else if svn == "" {
		respond(w, nil, errors.New("svn is missing in queries"))
		return
	}

	annotations := make(map[string]string)
	wf := bytedata.ShortChecksum("/" + path)
	annotations[functions.ServiceHeaderWorkflowID] = wf
	annotations[functions.ServiceHeaderNamespaceName] = mux.Vars(r)["ns"]
	annotations[functions.ServiceHeaderName] = svn
	annotations[functions.ServiceHeaderRevision] = ver

	h.deleteService(annotations, w, r)
}

func (h *functionHandler) deleteNamespaceService(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	annotations := make(map[string]string)
	annotations[functions.ServiceHeaderScope] = functions.PrefixNamespace
	annotations[functions.ServiceHeaderName] = mux.Vars(r)["svn"]
	annotations[functions.ServiceHeaderNamespaceName] = mux.Vars(r)["ns"]

	h.deleteService(annotations, w, r)
}

func (h *functionHandler) deleteService(annotations map[string]string,
	w http.ResponseWriter, r *http.Request,
) {
	grpcReq := grpcfunc.FunctionsListFunctionsRequest{
		Annotations: annotations,
	}

	resp, err := h.functionsClient.DeleteFunctions(r.Context(), &grpcReq)
	respond(w, resp, err)
}

type getFunctionResponse struct {
	Name      string                        `json:"name,omitempty"`
	Namespace string                        `json:"namespace,omitempty"`
	Workflow  string                        `json:"workflow,omitempty"`
	Config    *grpc.FunctionsConfig         `json:"config,omitempty"`
	Revisions []getFunctionResponseRevision `json:"revisions,omitempty"`
	Scope     string                        `json:"scope,omitempty"`
}

type getFunctionResponseRevision struct {
	Name       string                     `json:"name,omitempty"`
	Image      string                     `json:"image,omitempty"`
	Cmd        string                     `json:"cmd,omitempty"`
	Size       int32                      `json:"size,omitempty"`
	MinScale   int32                      `json:"minScale,omitempty"`
	Generation int64                      `json:"generation,omitempty"`
	Created    int64                      `json:"created,omitempty"`
	Status     string                     `json:"status,omitempty"`
	Conditions []*grpc.FunctionsCondition `json:"conditions,omitempty"`
	Revision   string                     `json:"revision,omitempty"`
}

func (h *functionHandler) getNamespaceService(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	svcName := mux.Vars(r)["svn"]
	nsName := mux.Vars(r)["ns"]

	svn, _, _ := functions.GenerateServiceName(&grpcfunc.FunctionsBaseInfo{
		NamespaceName: &nsName,
		Name:          &svcName,
	})

	h.getService(svn, w, r)
}

/*
func (h *functionHandler) getServiceSSE(annotations map[string]string,
	w http.ResponseWriter, r *http.Request) {

	grpcReq := &grpcfunc.WatchFunctionsRequest{
		Annotations: annotations,
	}

	functionsClient, err := h.functionsClient.WatchFunctions(r.Context(), grpcReq)
	if err != nil {
		respond(w, nil, err)
		return
	}
	ch := make(chan interface{}, 1)

	defer func() {

		_ = functionsClient.CloseSend()

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
			x, err := functionsClient.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x

		}

	}()

	sse(w, ch)
}
*/

func (h *functionHandler) getService(svn string, w http.ResponseWriter, r *http.Request) {
	grpcReq := new(grpc.FunctionsGetFunctionRequest)
	grpcReq.ServiceName = &svn

	resp, err := h.functionsClient.GetFunction(r.Context(), grpcReq)
	if err != nil {
		respond(w, resp, err)
		return
	}

	out := &getFunctionResponse{
		Name:      resp.GetName(),
		Namespace: resp.GetNamespace(),
		Workflow:  resp.GetWorkflow(),
		Revisions: make([]getFunctionResponseRevision, 0),
		Config:    resp.GetConfig(),
		Scope:     resp.GetScope(),
	}

	for _, rev := range resp.GetRevisions() {
		out.Revisions = append(out.Revisions, getFunctionResponseRevision{
			Name:       rev.GetName(),
			Image:      rev.GetImage(),
			Cmd:        rev.GetCmd(),
			Size:       rev.GetSize(),
			MinScale:   rev.GetMinScale(),
			Generation: rev.GetGeneration(),
			Created:    rev.GetCreated(),
			Status:     rev.GetStatus(),
			Conditions: rev.GetConditions(),
			Revision:   rev.GetRev(),
		})
	}

	respondStruct(w, out, http.StatusOK, nil)
}

type createNamespaceServiceRequest struct {
	Cmd      string            `json:"cmd,omitempty"`
	Image    string            `json:"image,omitempty"`
	Name     string            `json:"name,omitempty"`
	Size     int32             `json:"size,omitempty"`
	MinScale int32             `json:"minScale,omitempty"`
	Envs     map[string]string `json:"envs"`

	Namespace    string
	NamespaceOID string
	WorkflowPath string
	Workflow     string
}

func (h *functionHandler) createNamespaceService(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()

	nsName := mux.Vars(r)["ns"]

	// fetch namespace uuid
	resp, err := h.client.Namespace(ctx, &igrpc.NamespaceRequest{
		Name: nsName,
	})
	if err != nil {
		respond(w, nil, err)
		return
	}

	var cr createNamespaceServiceRequest
	err = json.NewDecoder(r.Body).Decode(&cr)
	if err != nil {
		respond(w, nil, err)
		return
	}

	cr.Namespace = nsName
	cr.NamespaceOID = resp.Namespace.GetOid()

	h.createService(cr, r, w)
}

func (h *functionHandler) createService(cr createNamespaceServiceRequest, r *http.Request, w http.ResponseWriter) {
	grpcReq := new(grpcfunc.FunctionsCreateFunctionRequest)
	grpcReq.Info = &grpc.FunctionsBaseInfo{
		Name:          &cr.Name,
		Namespace:     &cr.NamespaceOID,
		Workflow:      &cr.Workflow,
		Image:         &cr.Image,
		Cmd:           &cr.Cmd,
		Size:          &cr.Size,
		MinScale:      &cr.MinScale,
		NamespaceName: &cr.Namespace,
		Path:          &cr.WorkflowPath,
		Envs:          cr.Envs,
	}

	// returns an empty body
	resp, err := h.functionsClient.CreateFunction(r.Context(), grpcReq)
	respond(w, resp, err)
}

// UpdateServiceRequest update service request
//
// swagger:model UpdateServiceRequest
type updateServiceRequest struct {
	// image
	// Required: true
	Image *string `json:"image,omitempty"`
	// cmd
	// Required: true
	Cmd *string `json:"cmd,omitempty"`
	// size
	// Required: true
	Size *int32 `json:"size,omitempty"`
	// minScale
	// Required: true
	MinScale *int32 `json:"minScale,omitempty"`
}

func (h *functionHandler) updateNamespaceService(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	svcName := mux.Vars(r)["svn"]
	nsName := mux.Vars(r)["ns"]

	resp, err := h.client.Namespace(r.Context(), &igrpc.NamespaceRequest{
		Name: nsName,
	})
	if err != nil {
		respond(w, nil, err)
		return
	}

	svn, _, _ := functions.GenerateServiceName(&grpcfunc.FunctionsBaseInfo{
		NamespaceName: &nsName,
		Name:          &svcName,
	})
	h.updateService(svn, svcName, resp.GetNamespace(), w, r)
}

func (h *functionHandler) updateService(svc, name string, ns *igrpc.Namespace, w http.ResponseWriter, r *http.Request) {
	obj := new(updateServiceRequest)
	err := json.NewDecoder(r.Body).Decode(obj)
	if err != nil {
		respond(w, nil, err)
		return
	}

	grpcReq := new(grpcfunc.FunctionsUpdateFunctionRequest)
	grpcReq.ServiceName = &svc

	nsOID := ns.GetOid()
	nsName := ns.GetName()

	grpcReq.Info = &grpc.FunctionsBaseInfo{
		Image:         obj.Image,
		Cmd:           obj.Cmd,
		Size:          obj.Size,
		MinScale:      obj.MinScale,
		Namespace:     &nsOID,
		NamespaceName: &nsName,
		Name:          &name,
	}

	// returns an empty body
	resp, err := h.functionsClient.UpdateFunction(r.Context(), grpcReq)
	respond(w, resp, err)
}

func (h *functionHandler) deleteNamespaceServiceRevision(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	svcName := mux.Vars(r)["svn"]
	nsName := mux.Vars(r)["ns"]

	svn, _, _ := functions.GenerateServiceName(&grpcfunc.FunctionsBaseInfo{
		NamespaceName: &nsName,
		Name:          &svcName,
	})

	h.deleteRevision(fmt.Sprintf("%s-%s",
		svn, mux.Vars(r)["rev"]), w, r)
}

func (h *functionHandler) deleteRevision(rev string,
	w http.ResponseWriter, r *http.Request,
) {
	grpcReq := &grpcfunc.FunctionsDeleteRevisionRequest{
		Revision: &rev,
	}

	resp, err := h.functionsClient.DeleteRevision(r.Context(), grpcReq)
	respond(w, resp, err)
}

func (h *functionHandler) watchNamespaceRevision(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	svcName := mux.Vars(r)["svn"]
	nsName := mux.Vars(r)["ns"]

	svn, _, _ := functions.GenerateServiceName(&grpcfunc.FunctionsBaseInfo{
		NamespaceName: &nsName,
		Name:          &svcName,
	})

	h.watchRevisions(svn, mux.Vars(r)["rev"] /*functions.PrefixNamespace,*/, w, r)
}

func (h *functionHandler) singleWorkflowServiceRevision(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	vers := r.URL.Query().Get("version")
	rev := r.URL.Query().Get("rev")
	if rev == "" {
		rev = "00001"
	}

	hash, err := strconv.ParseUint(vers, 10, 64)
	if err != nil {
		respond(w, nil, err)
		return
	}

	svc := functions.AssembleWorkflowServiceName(hash)

	req := &grpc.FunctionsWatchRevisionsRequest{
		ServiceName:  &svc,
		RevisionName: &rev,
	}

	revisions, err := h.functionsClient.ListRevisions(r.Context(), req)
	if err != nil {
		respond(w, nil, err)
		return
	}

	respond(w, revisions, nil)
}

func (h *functionHandler) singleWorkflowServiceRevisionSSE(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	vers := r.URL.Query().Get("version")

	rev := r.URL.Query().Get("rev")
	if rev == "" {
		rev = "00001"
	}

	hash, err := strconv.ParseUint(vers, 10, 64)
	if err != nil {
		respond(w, nil, err)
		return
	}

	svc := functions.AssembleWorkflowServiceName(hash)

	h.watchRevisions(svc, rev /*functions.PrefixWorkflow,*/, w, r)
}

func (h *functionHandler) watchNamespaceRevisions(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())
	svcName := mux.Vars(r)["svn"]
	nsName := mux.Vars(r)["ns"]

	svn, _, _ := functions.GenerateServiceName(&grpcfunc.FunctionsBaseInfo{
		NamespaceName: &nsName,
		Name:          &svcName,
	})

	h.watchRevisions(svn, "" /*functions.PrefixNamespace,*/, w, r)
}

func (h *functionHandler) singleWorkflowServiceRevisions(w http.ResponseWriter, r *http.Request) {
	vers := r.URL.Query().Get("version")
	hash, err := strconv.ParseUint(vers, 10, 64)
	if err != nil {
		respond(w, nil, err)
		return
	}
	svc := functions.AssembleWorkflowServiceName(hash)
	h.logger.Debugf("Handeling singleWorkflowServiceRevisions for version: %v hash: %v, svc:", vers, hash, svc)
	req := &grpc.FunctionsWatchRevisionsRequest{
		ServiceName: &svc,
	}

	revisions, err := h.functionsClient.ListRevisions(r.Context(), req)
	if err != nil {
		respond(w, nil, err)
		return
	}

	respond(w, revisions, nil)
}

func (h *functionHandler) singleWorkflowServiceRevisionsSSE(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	vers := r.URL.Query().Get("version")

	hash, err := strconv.ParseUint(vers, 10, 64)
	if err != nil {
		respond(w, nil, err)
		return
	}

	svc := functions.AssembleWorkflowServiceName(hash)

	h.watchRevisions(svc, "" /*functions.PrefixWorkflow,*/, w, r)
}

func (h *functionHandler) watchRevisions(svc, rev /*, scope*/ string,
	w http.ResponseWriter, r *http.Request,
) {
	if rev != "" {
		rev = fmt.Sprintf("%s-%s", svc, rev)
	}

	grpcReq := &grpc.FunctionsWatchRevisionsRequest{
		ServiceName:  &svc,
		RevisionName: &rev,
		// Scope:        &scope,
	}

	client, err := h.functionsClient.WatchRevisions(r.Context(), grpcReq)
	if err != nil {
		respond(w, nil, err)
		return
	}
	ch := make(chan interface{}, 1)

	defer func() {
		_ = client.CloseSend()

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
			x, err := client.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x
		}
	}()

	sse(w, ch)
}

func (h *functionHandler) watchPodLogs(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	sn := mux.Vars(r)["pod"]
	grpcReq := new(grpc.FunctionsWatchLogsRequest)
	grpcReq.PodName = &sn

	client, err := h.functionsClient.WatchLogs(r.Context(), grpcReq)
	if err != nil {
		respond(w, nil, err)
		return
	}

	ch := make(chan interface{}, 1)

	defer func() {
		_ = client.CloseSend()

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
			x, err := client.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x
		}
	}()

	sse(w, ch)
}

func (h *functionHandler) listNamespacePods(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()

	resp, err := h.client.Namespace(ctx, &igrpc.NamespaceRequest{
		Name: mux.Vars(r)["ns"],
	})
	if err != nil {
		respond(w, nil, err)
		return
	}

	ns := resp.Namespace.GetName()
	svn := mux.Vars(r)["svn"]

	annotations := make(map[string]string)

	svn, _, _ = functions.GenerateServiceName(&grpcfunc.FunctionsBaseInfo{
		NamespaceName: &ns,
		Name:          &svn,
	})

	annotations[functions.ServiceKnativeHeaderRevision] = fmt.Sprintf("%s-%s", svn, mux.Vars(r)["rev"])
	annotations[functions.ServiceHeaderNamespaceName] = mux.Vars(r)["ns"]

	h.listPods(annotations, w, r)
}

func (h *functionHandler) listWorkflowPods(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	rev := r.URL.Query().Get("rev")
	if rev == "" {
		rev = "00001"
	}

	vers := r.URL.Query().Get("version")

	hash, err := strconv.ParseUint(vers, 10, 64)
	if err != nil {
		respond(w, nil, err)
		return
	}

	svc := functions.AssembleWorkflowServiceName(hash)

	knrev := fmt.Sprintf("%s-%s", svc, rev)

	annotations := make(map[string]string)
	wf := bytedata.ShortChecksum("/" + mux.Vars(r)["path"])
	annotations[functions.ServiceHeaderWorkflowID] = wf
	annotations[functions.ServiceHeaderNamespaceName] = mux.Vars(r)["ns"]
	annotations[functions.ServiceKnativeHeaderRevision] = knrev

	h.listPods(annotations, w, r)
}

func (h *functionHandler) listNamespacePodsSSE(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	svcName := mux.Vars(r)["svn"]
	nsName := mux.Vars(r)["ns"]

	svn, _, _ := functions.GenerateServiceName(&grpcfunc.FunctionsBaseInfo{
		NamespaceName: &nsName,
		Name:          &svcName,
	})

	rev := fmt.Sprintf("%s-%s", svn, mux.Vars(r)["rev"])

	h.listPodsSSE(svn, rev, w, r)
}

func (h *functionHandler) listWorkflowPodsSSE(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	rev := r.URL.Query().Get("rev")
	if rev == "" {
		rev = "00001"
	}

	vers := r.URL.Query().Get("version")

	hash, err := strconv.ParseUint(vers, 10, 64)
	if err != nil {
		respond(w, nil, err)
		return
	}

	svc := functions.AssembleWorkflowServiceName(hash)

	knrev := fmt.Sprintf("%s-%s", svc, rev)

	h.listPodsSSE(svc, knrev, w, r)
}

func (h *functionHandler) listPodsSSE(svc, rev string,
	w http.ResponseWriter, r *http.Request,
) {
	grpcReq := &grpc.FunctionsWatchPodsRequest{
		ServiceName:  &svc,
		RevisionName: &rev,
	}

	client, err := h.functionsClient.WatchPods(r.Context(), grpcReq)
	if err != nil {
		respond(w, nil, err)
		return
	}
	ch := make(chan interface{}, 1)

	defer func() {
		_ = client.CloseSend()

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
			x, err := client.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x
		}
	}()

	sse(w, ch)
}

func (h *functionHandler) listPods(annotations map[string]string,
	w http.ResponseWriter, r *http.Request,
) {
	grpcReq := grpc.FunctionsListPodsRequest{
		Annotations: annotations,
	}

	resp, err := h.functionsClient.ListPods(r.Context(), &grpcReq)
	respond(w, resp, err)
}
