// Package classification Direktiv API.
//
// Direktiv Open API Specification
// Direktiv Documentation can be found at https://docs.direktiv.io/
//
// Terms Of Service:
//
//     Schemes: http, https
//     Host: localhost
//     Version: 1.0.0
//     Contact: info@direktiv.io
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Security:
//     - api_key:
//
//     SecurityDefinitions:
//     api_key:
//          type: apiKey
//          name: KEY
//          in: header
//
// swagger:meta
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	igrpc "github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/functions"
	"github.com/direktiv/direktiv/pkg/functions/grpc"
	grpcfunc "github.com/direktiv/direktiv/pkg/functions/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type functionHandler struct {
	srv    *Server
	logger *zap.SugaredLogger
	client grpcfunc.FunctionsServiceClient
}

func newFunctionHandler(srv *Server, logger *zap.SugaredLogger,
	router *mux.Router, addr string) (*functionHandler, error) {

	funcAddr := fmt.Sprintf("%s:5555", addr)
	logger.Infof("connecting to functions %s", funcAddr)

	conn, err := util.GetEndpointTLS(funcAddr)
	if err != nil {
		logger.Errorf("can not connect to direktiv function: %v", err)
		return nil, err
	}

	fh := &functionHandler{
		srv:    srv,
		logger: logger,
		client: grpcfunc.NewFunctionsServiceClient(conn),
	}

	fh.initRoutes(router)

	return fh, err

}

func (h *functionHandler) initRoutes(r *mux.Router) {

	// swagger:operation GET /api/functions getGlobalServiceList
	// ---
	// description: |
	//   Gets a list of global knative services.
	// summary: Get Global Services List
	// tags:
	// - "Global Services"
	// responses:
	//   '200':
	//     "description": "successfully got services list"
	handlerPair(r, RN_ListServices, "", h.listGlobalServices, h.listGlobalServicesSSE)

	// swagger:operation GET /api/functions/{serviceName}/revisions/{revisionGeneration}/pods listGlobalServiceRevisionPods
	// ---
	// description: |
	//   List a revisions pods of a global scoped knative service.
	//   The target revision generation is the number suffix on a revision.
	//   Example: A revisions named 'global-fast-request-00003' would have the revisionGeneration '00003' .
	// summary: Get Global Service Revision Pods List
	// tags:
	// - "Global Services"
	// parameters:
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
	handlerPair(r, RN_ListPods, "/{svn}/revisions/{rev}/pods", h.listGlobalPods, h.listGlobalPodsSSE)

	// swagger:operation GET /api/functions/{serviceName} watchGlobalServiceRevisionList
	// ---
	// description: |
	//   Watch the revision list of a global scoped knative service.
	//   Note: This is a Server-Sent-Event endpoint, and will not work with the default swagger client.
	// summary: Watch Global Service Revision
	// tags:
	// - "Global Services"
	// parameters:
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
	r.HandleFunc("/{svn}", h.singleGlobalServiceSSE).Name(RN_WatchServices).Methods(http.MethodGet).Headers("Accept", "text/event-stream")

	// swagger:operation GET /api/functions/{serviceName}/revisions watchGlobalServiceRevisionList
	// ---
	// description: |
	//   Watch the revision list of a global scoped knative service.
	//   Note: This is a Server-Sent-Event endpoint, and will not work with the default swagger client.
	// summary: Watch Global Service Revision List
	// tags:
	// - "Global Services"
	// parameters:
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
	r.HandleFunc("/{svn}/revisions", h.watchGlobalRevisions).Name(RN_WatchRevisions).Methods(http.MethodGet).Headers("Accept", "text/event-stream")

	// swagger:operation GET /api/functions/{serviceName}/revisions/{revisionGeneration} watchGlobalServiceRevision
	// ---
	// description: |
	//   Watch a global scoped knative service revision.
	//   The target revision generation is the number suffix on a revision.
	//   Example: A revisions named 'global-fast-request-00003' would have the revisionGeneration '00003'.
	//   Note: This is a Server-Sent-Event endpoint, and will not work with the default swagger client.
	// summary: Watch Global Service Revision
	// tags:
	// - "Global Services"
	// parameters:
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
	r.HandleFunc("/{svn}/revisions/{rev}", h.watchGlobalRevision).Name(RN_WatchRevisions).Methods(http.MethodGet).Headers("Accept", "text/event-stream")

	// TODO: SWAGGER-SPEC
	r.HandleFunc("/logs/pod/{pod}", h.watchLogs).Methods(http.MethodGet).Name(RN_WatchLogs)

	// swagger:operation POST /api/functions createGlobalService
	// ---
	// description: |
	//   Creates global scoped knative service.
	//   Service Names are unique on a scope level.
	//   These services can be used as functions in workflows, more about this can be read here:
	//   https://docs.direktiv.io/docs/walkthrough/using-functions.html
	// summary: Create Global Service
	// tags:
	// - "Global Services"
	// parameters:
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
	//       minScale: "1"
	//       size: "small"
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
	//         type: string
	//         description: Size of created service pods
	//         enum:
	//           - small
	//           - medium
	//           - large
	// responses:
	//   '200':
	//     "description": "successfully created service"
	r.HandleFunc("", h.createGlobalService).Methods(http.MethodPost).Name(RN_CreateService)

	// swagger:operation DELETE /api/functions/{serviceName} deleteGlobalService
	// ---
	// description: |
	//   Deletes global scoped knative service and all its revisions.
	// summary: Delete Global Service
	// tags:
	// - "Global Services"
	// parameters:
	// - in: path
	//   name: serviceName
	//   type: string
	//   required: true
	//   description: 'target service name'
	// responses:
	//   '200':
	//     "description": "successfully deleted service"
	r.HandleFunc("/{svn}", h.deleteGlobalService).Methods(http.MethodDelete).Name(RN_DeleteServices)

	// swagger:operation GET /api/functions/{serviceName} getGlobalService
	// ---
	// description: |
	//   Get details of a global scoped knative service.
	// summary: Get Global Service Details
	// tags:
	// - "Global Services"
	// parameters:
	// - in: path
	//   name: serviceName
	//   type: string
	//   required: true
	//   description: 'target service name'
	// responses:
	//   '200':
	//     "description": "successfully got service details"
	r.HandleFunc("/{svn}", h.getGlobalService).Methods(http.MethodGet).Name(RN_GetService)

	// swagger:operation POST /api/functions/{serviceName} updateGlobalService
	// ---
	// description: |
	//   Creates a new global scoped knative service revision
	//   Revisions are created with a traffic percentage. This percentage controls how much traffic will be directed to this revision.
	//   Traffic can be set to 100 to direct all traffic.
	// summary: Create Global Service Revision
	// tags:
	// - "Global Services"
	// parameters:
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
	//       trafficPercent: 50
	//       image: "direktiv/request:v10"
	//       cmd: ""
	//       minScale: "1"
	//       size: "small"
	//     required:
	//       - image
	//       - cmd
	//       - minScale
	//       - size
	//       - trafficPercent
	//     properties:
	//       trafficPercent:
	//         type: integer
	//         description: Traffic percentage new revision will use
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
	r.HandleFunc("/{svn}", h.updateGlobalService).Methods(http.MethodPost).Name(RN_UpdateService)

	// swagger:operation PATCH /api/functions/{serviceName} updateGlobalServiceTraffic
	// ---
	// description: |
	//   Update Global Service traffic directed to each revision, traffic can only be configured between two revisions.
	//   All other revisions will bet set to 0 traffic.
	// summary: Update Global Service Traffic
	// tags:
	// - "Global Services"
	// parameters:
	// - in: path
	//   name: serviceName
	//   type: string
	//   required: true
	//   description: 'target service name'
	// - in: body
	//   name: Service Traffic
	//   description: Payload that contains information on service traffic
	//   required: true
	//   schema:
	//     type: object
	//     example:
	//       values:
	//         - percent: 60
	//           revision: global-fast-request-00002
	//         - percent: 40
	//           revision: global-fast-request-00001
	//     required:
	//       - values
	//     properties:
	//       values:
	//         description: List of revision traffic targets
	//         type: array
	//         items:
	//           type: object
	//           properties:
	//             percent:
	//               description: Target traffice percentage
	//               type: integer
	//             revision:
	//               description: Target service revision
	//               type: string
	//
	// responses:
	//   '200':
	//     "description": "successfully updated service traffic"
	r.HandleFunc("/{svn}", h.updateGlobalServiceTraffic).Methods(http.MethodPatch).Name(RN_UpdateServiceTraffic)

	// swagger:operation DELETE /api/functions/{serviceName}/revisions/{revisionGeneration} deleteGlobalRevision
	// ---
	// description: |
	//   Delete a global scoped knative service revision.
	//   The target revision generation is the number suffix on a revision.
	//   Example: A revisions named 'global-fast-request-00003' would have the revisionGeneration '00003'.
	//   Note: Revisions with traffic cannot be deleted.
	// summary: Delete Global Service Revision
	// tags:
	// - "Global Services"
	// parameters:
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
	r.HandleFunc("/{svn}/revisions/{rev}", h.deleteGlobalRevision).Methods(http.MethodDelete).Name(RN_DeleteRevision)

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
	//   Example: A revisions named 'namespace-direktiv-fast-request-00003' would have the revisionGeneration '00003'.
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
	//   Example: A revisions named 'namespace-direktiv-fast-request-00003' would have the revisionGeneration '00003'.
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
	//   https://docs.direktiv.io/docs/walkthrough/using-functions.html
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
	//       minScale: "1"
	//       size: "small"
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
	//         type: string
	//         description: Size of created service pods
	//         enum:
	//           - small
	//           - medium
	//           - large
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
	//   Revisions are created with a traffic percentage. This percentage controls
	//   how much traffic will be directed to this revision. Traffic can be set to 100
	//   to direct all traffic.
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
	//       trafficPercent: 50
	//       image: "direktiv/request:v10"
	//       cmd: ""
	//       minScale: "1"
	//       size: "small"
	//     required:
	//       - image
	//       - cmd
	//       - minScale
	//       - size
	//       - trafficPercent
	//     properties:
	//       trafficPercent:
	//         type: integer
	//         description: Traffic percentage new revision will use
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

	// swagger:operation PATCH /api/functions/namespaces/{namespace}/function/{serviceName} updateNamespaceServiceTraffic
	// ---
	// description: |
	//   Update Namespace Service traffic directed to each revision,
	//   traffic can only be configured between two revisions. All other revisions
	//   will bet set to 0 traffic.
	// summary: Update Namespace Service Traffic
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
	//   name: Service Traffic
	//   description: Payload that contains information on service traffic
	//   required: true
	//   schema:
	//     type: object
	//     example:
	//       values:
	//         - percent: 60
	//           revision: namespace-direktiv-fast-request-00002
	//         - percent: 40
	//           revision: namespace-direktiv-fast-request-00001
	//     required:
	//       - values
	//     properties:
	//       values:
	//         description: List of revision traffic targets
	//         type: array
	//         items:
	//           type: object
	//           properties:
	//             percent:
	//               description: Target traffice percentage
	//               type: integer
	//             revision:
	//               description: Target service revision
	//               type: string
	//
	// responses:
	//   '200':
	//     "description": "successfully updated service traffic"
	r.HandleFunc("/namespaces/{ns}/function/{svn}", h.updateNamespaceServiceTraffic).Methods(http.MethodPatch).Name(RN_UpdateNamespaceServiceTraffic)

	// swagger:operation DELETE /api/functions/namespaces/{namespace}/function/{serviceName}/revisions/{revisionGeneration} deleteNamespaceRevision
	// ---
	// description: |
	//   Delete a namespace scoped knative service revision.
	//   The target revision generation is the number suffix on a revision.
	//   Example: A revisions named 'namespace-direktiv-fast-request-00003' would have the revisionGeneration '00003'.
	//   Note: Revisions with traffic cannot be deleted.
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
	r.HandleFunc("/namespaces/{ns}/function/{svn}/revisions/{rev}", h.deleteNamespaceRevision).Methods(http.MethodDelete).Name(RN_DeleteNamespaceRevision)

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

	// swagger:operation GET /api/functions/namespaces/{namespace}/tree/{workflow}?op=pods listWorkflowServiceRevisionPods
	// ---
	// description: |
	//   List a revisions pods of a workflow scoped knative service.
	//   The target revision generation (rev query) is the number suffix on a revision.
	//   Example: A revisions named 'workflow-10640097968065193909-get-00001' would have the revisionGeneration '00001'.
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
	//   Example: A revisions named 'workflow-10640097968065193909-get-00001' would have the revisionGeneration '00001'.
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

	// TODO: direct control?
	// r.HandleFunc("/namespaces/{ns}/workflow/{wf}", h.createWorkflowService).Methods(http.MethodPost).Name(RN_CreateNamespaceService)
	// r.HandleFunc("/namespaces/{ns}/function/{svn}", h.deleteNamespaceService).Methods(http.MethodDelete).Name(RN_DeleteNamespaceServices)
	// r.HandleFunc("/namespaces/{ns}/function/{svn}", h.getNamespaceService).Methods(http.MethodGet).Name(RN_GetNamespaceService)
	// r.HandleFunc("/namespaces/{ns}/function/{svn}", h.updateNamespaceService).Methods(http.MethodPost).Name(RN_UpdateNamespaceService)
	// r.HandleFunc("/namespaces/{ns}/function/{svn}", h.updateNamespaceServiceTraffic).Methods(http.MethodPatch).Name(RN_UpdateNamespaceServiceTraffic)
	// r.HandleFunc("/namespaces/{ns}/function/{svn}/revisions/{rev}", h.deleteNamespaceRevision).Methods(http.MethodDelete).Name(RN_DeleteNamespaceRevision)

	// Registry ..

	// swagger:operation GET /api/namespaces/{namespace}/registries Registries getRegistries
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
	r.HandleFunc("/namespaces/{ns}/registries", h.getRegistries).Methods(http.MethodGet).Name(RN_ListRegistries)

	// swagger:operation POST /api/namespaces/{namespace}/registries Registries createRegistry
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
	r.HandleFunc("/namespaces/{ns}/registries", h.createRegistry).Methods(http.MethodPost).Name(RN_CreateRegistry)

	// swagger:operation POST /api/namespaces/{namespace}/registries Registries deleteRegistry
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
	r.HandleFunc("/namespaces/{ns}/registries", h.deleteRegistry).Methods(http.MethodDelete).Name(RN_DeleteRegistry)

}

func (h *functionHandler) deleteRegistry(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	n := mux.Vars(r)["ns"]

	d := make(map[string]string)

	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		respond(w, nil, err)
	}
	reg := d["reg"]

	resp, err := h.client.DeleteRegistry(r.Context(), &grpc.DeleteRegistryRequest{
		Namespace: &n,
		Name:      &reg,
	})

	respond(w, resp, err)
}

func (h *functionHandler) createRegistry(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	n := mux.Vars(r)["ns"]
	d := make(map[string]string)

	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		respond(w, nil, err)
	}
	reg := d["reg"]

	resp, err := h.client.StoreRegistry(r.Context(), &grpc.StoreRegistryRequest{
		Namespace: &n,
		Name:      &reg,
		Data:      []byte(d["data"]),
	})

	respond(w, resp, err)

}

func (h *functionHandler) getRegistries(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	n := mux.Vars(r)["ns"]

	var resp *grpc.GetRegistriesResponse
	resp, err := h.client.GetRegistries(r.Context(), &grpc.GetRegistriesRequest{
		Namespace: &n,
	})

	respond(w, resp, err)
}

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"strings"
// 	"time"
//
// 	"github.com/direktiv/direktiv/pkg/functions"
// 	"github.com/direktiv/direktiv/pkg/model"
//
// 	"github.com/gorilla/mux"
// 	"github.com/direktiv/direktiv/pkg/functions/grpc"
// 	"github.com/direktiv/direktiv/pkg/ingress"
// )
//
// type functionAnnotationsRequest struct {
// 	Scope     string `json:"scope"`
// 	Name      string `json:"name"`
// 	Namespace string `json:"namespace"`
// 	Workflow  string `json:"workflow"`
// }
//
// type functionResponseList struct {
// 	Config   *grpc.FunctionsConfig     `json:"config,omitempty"`
// 	Services []*functionResponseObject `json:"services"`
// }
//
// type functionResponseObject struct {
// 	Info struct {
// 		Workflow  string `json:"workflow"`
// 		Name      string `json:"name"`
// 		Namespace string `json:"namespace"`
// 		Image     string `json:"image"`
// 		Cmd       string `json:"cmd"`
// 	} `json:"info"`
// 	ServiceName string         funcClient   `json:"serviceName"`
// 	Status      string            `json:"status"`
// 	Conditions  []*grpc.Condition `json:"conditions"`
// }
//
// var functionsQueryLabelMapping = map[string]string{
// 	"scope":     functions.ServiceHeaderScope,
// 	"name":      functions.ServiceHeaderName,
// 	"namespace": functions.ServiceHeaderNamespace,
// 	"workflow":  functions.ServiceHeaderWorkflow,
// }
//
// func accepted(w http.ResponseWriter) {
// 	w.WriteHeader(http.StatusAccepted)
// }
//
// func getFunctionAnnotations(r *http.Request) (map[string]string, error) {
//
// 	annotations := make(map[string]string)
//
// 	// Get function labels from url queries
// 	for k, v := range r.URL.Query() {
// 		if aLabel, ok := functionsQueryLabelMapping[k]; ok && len(v) > 0 {
// 			annotations[aLabel] = v[0]
// 		}
// 	}
//
// 	// Get functions from body
// 	rb := new(functionAnnotationsRequest)
// 	err := json.NewDecoder(r.Body).Decode(rb)
// 	if err != nil && err != io.EOF {
// 		return nil, err
// 	} else if err == nil {
// 		annotations[functions.ServiceHeaderName] = rb.Name
// 		annotations[functions.ServiceHeaderNamespace] = rb.Namespace
// 		annotations[functions.ServiceHeaderWorkflow] = rb.Workflow
// 		annotations[functions.ServiceHeaderScope] = rb.Scope
// 	}
//
// 	// Split serviceName
// 	svc := mux.Vars(r)["serviceName"]
// 	if svc != "" {
// 		// Split namespaced service name
// 		if strings.HasPrefix(svc, functions.PrefixNamespace) {
// 			if strings.Count(svc,
// 	resp, err = h.s.functions.StoreRegistry(ctx, &grpc.StoreRegistryRequest{
// 		Namespace: &n,
// 		Name:      &st.Name,
// 		Data:      []byte(st.Data),
// 	})
// "-") < 2 {
// 				return nil, fmt.Errorf("service name is incorrect format, does not include scope and name")
// 			}
//
// 			annotations[functions.ServiceHeaderName] = rb.Name
// 			annotations[functions.ServiceHeaderNamespace] = rb.Namespace
// 			annotations[functions.ServiceHeaderWorkflow] = rb.Workflow
// 			annotations[functions.ServiceHeaderScope] = rb.Scope
//
// 			firstInd := strings.Index(svc, "-")
// 			lastInd := strings.LastIndex(svc, "-")
// 			annotations[functions.ServiceHeaderNamespace] = svc[firstInd+1 : lastInd]
// 			annotations[functions.ServiceHeaderName] = svc[lastInd+1:]
// 			annotations[functions.ServiceHeaderScope] = svc[:firstInd]
// 		} else {
// 			if strings.Count(svc, "-") < 1 {
// 				return nil, fmt.Errorf("service name is incorrect format, does not include scope")
// 			}
//
// 			firstInd := strings.Index(svc, "-")
// 			annotations[functions.ServiceHeaderName] = svc[firstInd+1:]
// 			annotations[functions.ServiceHeaderScope] = svc[:firstInd]
// 		}
// 	}
//
// 	// Handle if this was reached via the workflow route
// 	wf := mux.Vars(r)["workflowTarget"]
// 	if wf != "" {
// 		if annotations[functions.ServiceHeaderScope] != "" && annotations[functions.ServiceHeaderScope] != functions.PrefixWorkflow {
// 			return nil, fmt.Errorf("this route is for workflow-scoped requests")
// 		}
//
// 		annotations[functions.ServiceHeaderWorkflow] = wf
// 		annotations[functions.ServiceHeaderScope] = functions.PrefixWorkflow
// 	}
//
// 	// Handle if this was reached via the namespaced route
// 	ns := mux.Vars(r)["namespace"]
// 	if ns != "" {
// 		if annotations[functions.ServiceHeaderScope] == functions.PrefixGlobal {
// 			return nil, fmt.Errorf("this route is for namespace-scoped requests or lower, not global")
// 		}
//
// 		annotations[functions.ServiceHeaderNamespace] = ns
//
// 		if annotations[functions.ServiceHeaderScope] == "" {
// 			annotations[functions.ServiceHeaderScope] = functions.PrefixNamespace
// 		}
// 	}
//
// 	del := make([]string, 0)
// 	for k, v := range annotations {
// 		if v == "" {
// 			del = append(del, k)
// 		}
// 	}
//
// 	for _, v := range del {
// 		delete(annotations, v)
// 	}
//
// 	return annotations, nil
// }
//
// func prepareFunctionsForResponse(functions []*grpc.FunctionsInfo) []*functionResponseObject {
// 	out := make([]*functionResponseObject, 0)
//
// 	for _, function := range functions {
//
// 		obj := new(functionResponseObject)
// 		iinf := function.GetInfo()
// 		if iinf != nil {
// 			if iinf.Workflow != nil {
// 				obj.Info.Workflow = *iinf.Workflow
// 			}
// 			if iinf.Name != nil {
// 				obj.Info.Name = *iinf.Name
// 			}
// 			if iinf.Namespace != nil {
// 				obj.Info.Namespace = *iinf.Namespace
// 			}
// 			if iinf.Image != nil {
// 				obj.Info.Image = *iinf.Image
// 			}
// 			if iinf.Cmd != nil {
// 				obj.Info.Cmd = *iinf.Cmd
// 			}
// 		}
//
// 		obj.ServiceName = function.GetServiceName()
// 		obj.Status = function.GetStatus()
// 		obj.Conditions = function.GetConditions()
//
// 		out = append(out, obj)
// 	}
//
// 	return out
// }
//
// func (h *functionHandler) listFunctions(w http.ResponseWriter, r *http.Request) {
// 	h.logger.Infof("LIST FUNCTIONS")
// 	w.Write([]byte("LIST FUNCTIONS"))
// }

// var functionsQueryLabelMapping = map[string]string{
// 	"scope":     functions.ServiceHeaderScope,
// 	"name":      functions.ServiceHeaderName,
// 	"namespace": functions.ServiceHeaderNamespace,
// 	"workflow":  functions.ServiceHeaderWorkflow,
// }

func (h *functionHandler) listGlobalServices(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	annotations := make(map[string]string)
	annotations[functions.ServiceHeaderScope] = functions.PrefixGlobal
	h.listServices(annotations, w, r)

}

func (h *functionHandler) listNamespaceServices(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()

	resp, err := h.srv.flowClient.Namespace(ctx, &igrpc.NamespaceRequest{
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

	resp, err := h.srv.flowClient.Namespace(ctx, &igrpc.NamespaceRequest{
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

	ctx := r.Context()

	resp, err := h.srv.flowClient.Workflow(ctx, &igrpc.WorkflowRequest{
		Namespace: mux.Vars(r)["ns"],
		Path:      mux.Vars(r)["path"],
	})
	if err != nil {
		respond(w, nil, err)
		return
	}

	annotations := make(map[string]string)
	annotations[functions.ServiceHeaderWorkflowID] = resp.GetOid()

	h.listServices(annotations, w, r)

}

func (h *functionHandler) listWorkflowServicesSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()

	resp, err := h.srv.flowClient.Workflow(ctx, &igrpc.WorkflowRequest{
		Namespace: mux.Vars(r)["ns"],
		Path:      mux.Vars(r)["path"],
	})
	if err != nil {
		respond(w, nil, err)
		return
	}

	annotations := make(map[string]string)
	annotations[functions.ServiceHeaderWorkflowID] = resp.GetOid()

	h.listServicesSSE(annotations, w, r)

}

func (h *functionHandler) singleGlobalServiceSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	annotations := make(map[string]string)
	annotations[functions.ServiceHeaderScope] = functions.PrefixGlobal
	annotations[functions.ServiceHeaderName] = mux.Vars(r)["svn"]
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

	ctx := r.Context()

	vers := r.URL.Query().Get("version")

	resp, err := h.srv.flowClient.Workflow(ctx, &igrpc.WorkflowRequest{
		Namespace: mux.Vars(r)["ns"],
		Path:      mux.Vars(r)["path"],
	})
	if err != nil {
		respond(w, nil, err)
		return
	}

	hash, err := strconv.ParseUint(vers, 10, 64)
	if err != nil {
		respond(w, nil, err)
		return
	}

	svn := r.URL.Query().Get("svn")
	svc := functions.AssembleWorkflowServiceName(resp.Oid, svn, hash)

	annotations := make(map[string]string)

	annotations[functions.ServiceKnativeHeaderName] = svc

	h.listServicesSSE(annotations, w, r)

}

func (h *functionHandler) listGlobalServicesSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	annotations := make(map[string]string)
	annotations[functions.ServiceHeaderScope] = functions.PrefixGlobal
	h.listServicesSSE(annotations, w, r)

}

func (h *functionHandler) listServicesSSE(
	annotations map[string]string, w http.ResponseWriter, r *http.Request) {

	grpcReq := grpcfunc.WatchFunctionsRequest{
		Annotations: annotations,
	}

	client, err := h.client.WatchFunctions(r.Context(), &grpcReq)
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
	annotations map[string]string, w http.ResponseWriter, r *http.Request) {

	grpcReq := grpcfunc.ListFunctionsRequest{
		Annotations: annotations,
	}

	resp, err := h.client.ListFunctions(r.Context(), &grpcReq)
	respond(w, resp, err)
}

// sse

func (h *functionHandler) deleteGlobalService(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	annotations := make(map[string]string)
	annotations[functions.ServiceHeaderScope] = functions.PrefixGlobal
	annotations[functions.ServiceHeaderName] = mux.Vars(r)["svn"]
	h.deleteService(annotations, w, r)
}

func (h *functionHandler) deleteNamespaceService(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	annotations := make(map[string]string)
	annotations[functions.ServiceHeaderScope] = functions.PrefixNamespace
	annotations[functions.ServiceHeaderName] = mux.Vars(r)["svn"]
	h.deleteService(annotations, w, r)
}

func (h *functionHandler) deleteService(annotations map[string]string,
	w http.ResponseWriter, r *http.Request) {

	grpcReq := grpcfunc.ListFunctionsRequest{
		Annotations: annotations,
	}

	resp, err := h.client.DeleteFunctions(r.Context(), &grpcReq)
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
	Name       string            `json:"name,omitempty"`
	Image      string            `json:"image,omitempty"`
	Cmd        string            `json:"cmd,omitempty"`
	Size       int32             `json:"size,omitempty"`
	MinScale   int32             `json:"minScale,omitempty"`
	Generation int64             `json:"generation,omitempty"`
	Created    int64             `json:"created,omitempty"`
	Status     string            `json:"status,omitempty"`
	Conditions []*grpc.Condition `json:"conditions,omitempty"`
	Traffic    int64             `json:"traffic,omitempty"`
	Revision   string            `json:"revision,omitempty"`
}

func (h *functionHandler) getGlobalService(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	h.getService(fmt.Sprintf("%s-%s", functions.PrefixGlobal,
		mux.Vars(r)["svn"]), w, r)
}

func (h *functionHandler) getGlobalServiceSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	annotations := make(map[string]string)
	annotations[functions.ServiceHeaderScope] = functions.PrefixGlobal
	annotations[functions.ServiceHeaderName] = fmt.Sprintf("%s-%s", functions.PrefixGlobal, mux.Vars(r)["svn"])
	h.getServiceSSE(annotations, w, r)
}

func (h *functionHandler) getNamespaceService(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	h.getService(fmt.Sprintf("%s-%s-%s", functions.PrefixNamespace, mux.Vars(r)["ns"],
		mux.Vars(r)["svn"]), w, r)
}

func (h *functionHandler) getServiceSSE(annotations map[string]string,
	w http.ResponseWriter, r *http.Request) {

	grpcReq := &grpcfunc.WatchFunctionsRequest{
		Annotations: annotations,
	}

	client, err := h.client.WatchFunctions(r.Context(), grpcReq)
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

func (h *functionHandler) getService(svn string, w http.ResponseWriter, r *http.Request) {

	grpcReq := new(grpc.GetFunctionRequest)
	grpcReq.ServiceName = &svn

	resp, err := h.client.GetFunction(r.Context(), grpcReq)

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
			Traffic:    rev.GetTraffic(),
			Revision:   rev.GetRev(),
		})
	}

	respondStruct(w, out, http.StatusOK, nil)

}

type createFunctionRequest struct {
	Name     *string `json:"name,omitempty"`
	Image    *string `json:"image,omitempty"`
	Cmd      *string `json:"cmd,omitempty"`
	Size     *int32  `json:"size,omitempty"`
	MinScale *int32  `json:"minScale,omitempty"`
}

func (h *functionHandler) createGlobalService(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	h.createService("", "", "", "", "", w, r)
}

func (h *functionHandler) createNamespaceService(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()

	nsName := mux.Vars(r)["ns"]

	resp, err := h.srv.flowClient.Namespace(ctx, &igrpc.NamespaceRequest{
		Name: nsName,
	})
	if err != nil {
		respond(w, nil, err)
		return
	}

	h.createService(resp.Namespace.GetOid(), nsName, "", "", "", w, r)

}

func (h *functionHandler) createService(ns, nsName, wf, path, rev string,
	w http.ResponseWriter, r *http.Request) {

	obj := new(createFunctionRequest)
	err := json.NewDecoder(r.Body).Decode(obj)
	if err != nil {
		respond(w, nil, err)
		return
	}

	grpcReq := new(grpcfunc.CreateFunctionRequest)
	grpcReq.Info = &grpc.BaseInfo{
		Name:          obj.Name,
		Namespace:     &ns,
		Workflow:      &wf,
		Image:         obj.Image,
		Cmd:           obj.Cmd,
		Size:          obj.Size,
		MinScale:      obj.MinScale,
		NamespaceName: &nsName,
		Path:          &path,
		Revision:      &rev,
	}

	// returns an empty body
	resp, err := h.client.CreateFunction(r.Context(), grpcReq)
	respond(w, resp, err)

}

// UpdateServiceRequest UpdateServiceRequest update service request
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
	// trafficPercent
	// Required: true
	TrafficPercent int64 `json:"trafficPercent"`
}

func (h *functionHandler) updateGlobalService(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	h.updateService(fmt.Sprintf("%s-%s",
		functions.PrefixGlobal, mux.Vars(r)["svn"]), w, r)
}

func (h *functionHandler) updateNamespaceService(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	h.updateService(fmt.Sprintf("%s-%s-%s",
		functions.PrefixNamespace, mux.Vars(r)["ns"], mux.Vars(r)["svn"]), w, r)
}

func (h *functionHandler) updateService(svc string, w http.ResponseWriter, r *http.Request) {

	obj := new(updateServiceRequest)
	err := json.NewDecoder(r.Body).Decode(obj)
	if err != nil {
		respond(w, nil, err)
		return
	}

	grpcReq := new(grpcfunc.UpdateFunctionRequest)
	grpcReq.ServiceName = &svc
	grpcReq.Info = &grpc.BaseInfo{
		Image:    obj.Image,
		Cmd:      obj.Cmd,
		Size:     obj.Size,
		MinScale: obj.MinScale,
	}

	grpcReq.TrafficPercent = &obj.TrafficPercent

	// returns an empty body
	resp, err := h.client.UpdateFunction(r.Context(), grpcReq)
	respond(w, resp, err)

}

type updateServiceTrafficRequest struct {
	Values []struct {
		Revision string `json:"revision"`
		Percent  int64  `json:"percent"`
	} `json:"values"`
}

func (h *functionHandler) updateGlobalServiceTraffic(w http.ResponseWriter,
	r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	h.updateServiceTraffic(fmt.Sprintf("%s-%s",
		functions.PrefixGlobal, mux.Vars(r)["svn"]), w, r)

}

func (h *functionHandler) updateNamespaceServiceTraffic(w http.ResponseWriter,
	r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	h.updateServiceTraffic(fmt.Sprintf("%s-%s-%s", functions.PrefixNamespace, mux.Vars(r)["ns"],
		mux.Vars(r)["svn"]), w, r)
}

func (h *functionHandler) updateServiceTraffic(svc string,
	w http.ResponseWriter, r *http.Request) {

	obj := new(updateServiceTrafficRequest)
	err := json.NewDecoder(r.Body).Decode(obj)
	if err != nil {
		respond(w, nil, err)
		return
	}

	if obj.Values == nil {
		respond(w, nil, fmt.Errorf("no traffic values"))
		return
	}

	grpcReq := &grpc.SetTrafficRequest{
		Name:    &svc,
		Traffic: make([]*grpc.TrafficValue, 0),
	}

	for _, v := range obj.Values {
		x := v
		grpcReq.Traffic = append(grpcReq.Traffic, &grpc.TrafficValue{
			Revision: &x.Revision,
			Percent:  &x.Percent,
		})
	}

	resp, err := h.client.SetFunctionsTraffic(r.Context(), grpcReq)
	respond(w, resp, err)

}

func (h *functionHandler) deleteGlobalRevision(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	h.deleteRevision(fmt.Sprintf("%s-%s-%s",
		functions.PrefixGlobal, mux.Vars(r)["svn"], mux.Vars(r)["rev"]), w, r)
}

func (h *functionHandler) deleteNamespaceRevision(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	h.deleteRevision(fmt.Sprintf("%s-%s-%s-%s",
		functions.PrefixNamespace, mux.Vars(r)["ns"],
		mux.Vars(r)["svn"], mux.Vars(r)["rev"]), w, r)
}

func (h *functionHandler) deleteRevision(rev string,
	w http.ResponseWriter, r *http.Request) {

	grpcReq := &grpcfunc.DeleteRevisionRequest{
		Revision: &rev,
	}

	resp, err := h.client.DeleteRevision(r.Context(), grpcReq)
	respond(w, resp, err)

}

// type serviceItem struct {
// 	name, service string
// }
//
// func calculateList(client grpc.FunctionsServiceClient,
// 	items []serviceItem, annotations map[string]string, ns string) ([]*grpc.FunctionsInfo, error) {
//
// 	resp, err := client.ListFunctions(context.Background(),
// 		&grpc.ListFunctionsRequest{
// 			Annotations: annotations,
// 		})
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	gisos := make(map[string]*grpc.FunctionsInfo)
//
// 	imgStatus := "False"
// 	imgErr := "not found"
// 	imgNS := ""
//
// 	condName := "Ready"
// 	condStatus := "False"
//
// 	condMessage := "Global service does not exist"
//
// 	if len(annotations) > 1 {
// 		condMessage = "Namespace service does not exist"
// 		imgNS = ns
// 	}
//
// 	cond := &grpc.Condition{
// 		Name:    &condName,
// 		Status:  &condStatus,
// 		Message: &condMessage,
// 	}
//
// 	// populate the map with "error items"
// 	for i := range items {
// 		li := items[i]
//
// 		ns := ""
// 		if annons, ok := annotations[functions.ServiceHeaderNamespace]; ok {
// 			ns = annons
// 		}
//
// 		svcName, _, err := functions.GenerateServiceName(ns, "", li.service)
// 		if err != nil {
// 			logger.Errorf("can not generate service name: %v", err)
// 			continue
// 		}
//
// 		info := &grpc.FunctionsInfo{
// 			Status:      &imgStatus,
// 			ServiceName: &li.service,
// 			Info: &grpc.BaseInfo{
// 				Image:     &imgErr,
// 				Namespace: &imgNS,
// 			},
// 			Conditions: []*grpc.Condition{
// 				cond,
// 			},
// 		}
// 		gisos[svcName] = info
//
// 	}
//
// 	isos := resp.GetFunctions()
//
// 	for i := range isos {
// 		// that item exists, we replace
// 		logger.Debugf("checking %v", isos[i].GetServiceName())
// 		if _, ok := gisos[isos[i].GetServiceName()]; ok {
// 			gisos[isos[i].GetServiceName()] = isos[i]
// 		}
// 	}
//
// 	var retIsos []*grpc.FunctionsInfo
//
// 	for _, v := range gisos {
// 		retIsos = append(retIsos, v)
// 	}
// 	return retIsos, nil
//
// }
//
// // /api/namespaces/{namespace}/workflows/{workflowTarget}/functions
// func (h *Handler) getWorkflowFunctions(w http.ResponseWriter, r *http.Request) {
//
// 	ns := mux.Vars(r)["namespace"]
// 	wf := mux.Vars(r)["workflowTarget"]
//
// 	grpcReq1 := &ingress.GetWorkflowByNameRequest{
// 		Namespace: &ns,
// 		Name:      &wf,
// 	}
//
// 	resp, err := h.s.direktiv.GetWorkflowByName(r.Context(), grpcReq1)
// 	if err != nil {
// 		ErrResponse(w, err)
// 		return
// 	}
//
// 	workflow := new(model.Workflow)
// 	err = workflow.Load(resp.Workflow)
// 	if err != nil {
// 		ErrResponse(w, err)
// 		return
// 	}
//
// 	var fnNS, fnGlobal []serviceItem
//
// 	allFunctions := make([]*grpc.FunctionsInfo, 0)
// 	wfFns := false
//
// 	for _, fn := range workflow.Functions {
// 		switch fn.GetType() {
// 		case model.ReusableContainerFunctionType:
// 			wfFns = true
// 		case model.NamespacedKnativeFunctionType:
// 			fnNS = append(fnNS, serviceItem{
// 				name:    fn.GetID(),
// 				service: fn.(*model.NamespacedFunctionDefinition).KnativeService,
// 			})
// 		case model.GlobalKnativeFunctionType:
// 			fnGlobal = append(fnGlobal, serviceItem{
// 				name:    fn.GetID(),
// 				service: fn.(*model.GlobalFunctionDefinition).KnativeService,
// 			})
// 		}
// 	}
//
// 	// we add all workflow functions
// 	if wfFns {
// 		wfResp, err := h.s.functions.ListFunctions(r.Context(), &grpc.ListFunctionsRequest{
// 			Annotations: map[string]string{
// 				functions.ServiceHeaderWorkflow:  wf,
// 				functions.ServiceHeaderNamespace: ns,
// 				functions.ServiceHeaderScope:     functions.PrefixWorkflow,
// 			},
// 		})
// 		if err != nil {
// 			ErrResponse(w, err)
// 			return
// 		}
// 		allFunctions = append(allFunctions, wfResp.GetFunctions()...)
// 	}
//
// 	if len(fnNS) > 0 {
//
// 		i, err := calculateList(h.s.functions, fnNS,
// 			map[string]string{
// 				functions.ServiceHeaderNamespace: ns,
// 				functions.ServiceHeaderScope:     functions.PrefixNamespace,
// 			}, ns)
//
// 		if err != nil {
// 			ErrResponse(w, err)
// 			return
// 		}
// 		allFunctions = append(allFunctions, i...)
//
// 	}
//
// 	if len(fnGlobal) > 0 {
//
// 		i, err := calculateList(h.s.functions, fnGlobal,
// 			map[string]string{
// 				functions.ServiceHeaderScope: functions.PrefixGlobal,
// 			}, ns)
//
// 		if err != nil {
// 			ErrResponse(w, err)
// 			return
// 		}
// 		allFunctions = append(allFunctions, i...)
//
// 	}
//
// 	out := prepareFunctionsForResponse(allFunctions)
// 	if err := json.NewEncoder(w).Encode(out); err != nil {
// 		ErrResponse(w, err)
// 		return
// 	}
//
// }
//
// func (h *Handler) watchFunctions(w http.ResponseWriter, r *http.Request) {
//
// 	a, err := getFunctionAnnotations(r)
// 	if err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		w.Write([]byte(err.Error()))
// 		return
// 	}
//
// 	grpcReq := grpc.WatchFunctionsRequest{
// 		Annotations: a,
// 	}
//
// 	client, err := h.s.functions.WatchFunctions(r.Context(), &grpcReq)
// 	if err != nil {
// 		ErrResponse(w, err)
// 		return
// 	}
//
// 	flusher, err := SetupSEEWriter(w)
// 	if err != nil {
// 		ErrResponse(w, err)
// 		return
// 	}
//
// 	// Create Heartbeat Ticker
// 	heartbeat := time.NewTicker(10 * time.Second)
// 	defer heartbeat.Stop()
//
// 	// Start watcher client stream channels
// 	dataCh := make(chan interface{})
// 	errorCh := make(chan error)
// 	go func() {
// 		for {
// 			data, err := client.Recv()
// 			if err != nil {
// 				errorCh <- err
// 				break
// 			} else {
// 				dataCh <- data
// 			}
// 		}
// 	}()
//
// 	for {
// 		select {
// 		case data := <-dataCh:
// 			err = WriteSSEJSONData(w, flusher, data)
// 		case err = <-errorCh:
// 		case <-client.Context().Done():
// 			err = fmt.Errorf("requested stream has timed out")
// 		case <-heartbeat.C:
// 			SendSSEHeartbeat(w, flusher)
// 		}
//
// 		// Check for errors
// 		if err != nil {
// 			ErrSSEResponse(w, flusher, err)
// 			heartbeat.Stop()
// 			return
// 		}
// 	}
// }

func (h *functionHandler) watchGlobalRevision(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())
	svn := fmt.Sprintf("%s-%s", functions.PrefixGlobal, mux.Vars(r)["svn"])
	h.watchRevisions(svn, mux.Vars(r)["rev"] /*functions.PrefixGlobal,*/, w, r)
}

func (h *functionHandler) watchNamespaceRevision(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())
	svn := fmt.Sprintf("%s-%s-%s", functions.PrefixNamespace, mux.Vars(r)["ns"], mux.Vars(r)["svn"])
	h.watchRevisions(svn, mux.Vars(r)["rev"] /*functions.PrefixNamespace,*/, w, r)
}

func (h *functionHandler) singleWorkflowServiceRevision(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	http.Error(w, "text/event-stream only", http.StatusBadRequest)

}

func (h *functionHandler) singleWorkflowServiceRevisionSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()

	vers := r.URL.Query().Get("version")

	resp, err := h.srv.flowClient.Workflow(ctx, &igrpc.WorkflowRequest{
		Namespace: mux.Vars(r)["ns"],
		Path:      mux.Vars(r)["path"],
	})
	if err != nil {
		respond(w, nil, err)
		return
	}

	svn := r.URL.Query().Get("svn")
	rev := r.URL.Query().Get("rev")
	if rev == "" {
		rev = "00001"
	}

	hash, err := strconv.ParseUint(vers, 10, 64)
	if err != nil {
		respond(w, nil, err)
		return
	}

	svc := functions.AssembleWorkflowServiceName(resp.Oid, svn, hash)

	h.watchRevisions(svc, rev /*functions.PrefixWorkflow,*/, w, r)

}

func (h *functionHandler) watchGlobalRevisions(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	svn := fmt.Sprintf("%s-%s", functions.PrefixGlobal, mux.Vars(r)["svn"])
	h.watchRevisions(svn, "" /*functions.PrefixGlobal,*/, w, r)
}

func (h *functionHandler) watchNamespaceRevisions(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	svn := fmt.Sprintf("%s-%s-%s", functions.PrefixNamespace, mux.Vars(r)["ns"], mux.Vars(r)["svn"])
	h.watchRevisions(svn, "" /*functions.PrefixNamespace,*/, w, r)
}

func (h *functionHandler) singleWorkflowServiceRevisions(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	http.Error(w, "text/event-stream only", http.StatusBadRequest)

}

func (h *functionHandler) singleWorkflowServiceRevisionsSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	vers := r.URL.Query().Get("version")

	resp, err := h.srv.flowClient.Workflow(ctx, &igrpc.WorkflowRequest{
		Namespace: mux.Vars(r)["ns"],
		Path:      mux.Vars(r)["path"],
	})
	if err != nil {
		respond(w, nil, err)
		return
	}

	hash, err := strconv.ParseUint(vers, 10, 64)
	if err != nil {
		respond(w, nil, err)
		return
	}

	svn := r.URL.Query().Get("svn")
	svc := functions.AssembleWorkflowServiceName(resp.Oid, svn, hash)

	h.watchRevisions(svc, "" /*functions.PrefixWorkflow,*/, w, r)

}

func (h *functionHandler) watchRevisions(svc, rev /*, scope*/ string,
	w http.ResponseWriter, r *http.Request) {

	if rev != "" {
		rev = fmt.Sprintf("%s-%s", svc, rev)
	}

	grpcReq := &grpc.WatchRevisionsRequest{
		ServiceName:  &svc,
		RevisionName: &rev,
		// Scope:        &scope,
	}

	client, err := h.client.WatchRevisions(r.Context(), grpcReq)
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

// 	sn := mux.Vars(r)["serviceName"]
// 	rn := mux.Vars(r)["revisionName"]
//
// 	// Append prefixNamespace if in namespace route and not
// 	ns := mux.Vars(r)["namespace"]
// 	if ns != "" && !strings.HasPrefix(sn, functions.PrefixNamespace+"-") {
// 		sn = fmt.Sprintf("%s-%s-%s", functions.PrefixNamespace, ns, sn)
// 	}
//
// 	grpcReq := new(grpc.WatchRevisionsRequest)
// 	grpcReq.ServiceName = &sn
// 	grpcReq.RevisionName = &rn
//
// 	client, err := h.s.functions.WatchRevisions(r.Context(), grpcReq)
// 	if err != nil {
// 		ErrResponse(w, err)
// 		return
// 	}
//
// 	defer client.CloseSend()
// 	flusher, err := SetupSEEWriter(w)
// 	if err != nil {
// 		ErrResponse(w, err)
// 		return
// 	}
//
// 	// Create Heartbeat Ticker
// 	heartbeat := time.NewTicker(10 * time.Second)
// 	defer heartbeat.Stop()
//
// 	// Start watcher client stream channels
// 	dataCh := make(chan interface{})
// 	errorCh := make(chan error)
// 	go func() {
// 		for {
// 			data, err := client.Recv()
// 			if err != nil {
// 				errorCh <- err
// 				break
// 			} else {
// 				dataCh <- data
// 			}
// 		}
// 	}()
//
// 	for {
// 		select {
// 		case data := <-dataCh:
// 			err = WriteSSEJSONData(w, flusher, data)
// 		case err = <-errorCh:
// 		case <-client.Context().Done():
// 			err = fmt.Errorf("requested stream has timed out")
// 		case <-heartbeat.C:
// 			SendSSEHeartbeat(w, flusher)
// 		}
//
// 		// Check for errors
// 		if err != nil {
// 			ErrSSEResponse(w, flusher, err)
// 			heartbeat.Stop()
// 			return
// 		}
// 	}
// }
//
func (h *functionHandler) watchLogs(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	sn := mux.Vars(r)["pod"]
	grpcReq := new(grpc.WatchLogsRequest)
	grpcReq.PodName = &sn

	client, err := h.client.WatchLogs(r.Context(), grpcReq)
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

//
// 	sn := mux.Vars(r)["podName"]
// 	grpcReq := new(grpc.WatchLogsRequest)
// 	grpcReq.PodName = &sn
//
// 	client, err := h.s.functions.WatchLogs(r.Context(), grpcReq)
// 	if err != nil {
// 		ErrResponse(w, err)
// 		return
// 	}
//
// 	defer client.CloseSend()
// 	flusher, err := SetupSEEWriter(w)
// 	if err != nil {
// 		ErrResponse(w, err)
// 		return
// 	}
//
// 	// Create Heartbeat Ticker
// 	heartbeat := time.NewTicker(10 * time.Second)
// 	defer heartbeat.Stop()
//
// 	// Start watcher client stream channels
// 	dataCh := make(chan string)
// 	errorCh := make(chan error)
// 	go func() {
// 		for {
// 			data, err := client.Recv()
// 			if err != nil {
// 				errorCh <- err
// 				break
// 			} else {
//
// 				dataCh <- *data.Data
// 			}
// 		}
// 	}()
//
// 	for {
// 		select {
// 		case data := <-dataCh:
// 			err = WriteSSEData(w, flusher, []byte(data))
// 		case err = <-errorCh:
// 		case <-client.Context().Done():
// 			err = fmt.Errorf("requested stream has timed out")
// 		case <-heartbeat.C:
// 			SendSSEHeartbeat(w, flusher)
// 		}
//
// 		// Check for errors
// 		if err != nil {
// 			ErrSSEResponse(w, flusher, err)
// 			heartbeat.Stop()
// 			return
// 		}
// 	}
//
// }
//
func (h *functionHandler) listGlobalPods(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	annotations := make(map[string]string)
	annotations[functions.ServiceKnativeHeaderRevision] = fmt.Sprintf("%s-%s-%s",
		functions.PrefixGlobal, mux.Vars(r)["svn"], mux.Vars(r)["rev"])
	annotations[functions.ServiceHeaderScope] = functions.PrefixGlobal
	h.listPods(annotations, w, r)
}

func (h *functionHandler) listNamespacePods(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()

	resp, err := h.srv.flowClient.Namespace(ctx, &igrpc.NamespaceRequest{
		Name: mux.Vars(r)["ns"],
	})
	if err != nil {
		respond(w, nil, err)
		return
	}

	ns := resp.Namespace.GetOid()
	svn := mux.Vars(r)["svn"]

	annotations := make(map[string]string)

	svn, _, _ = functions.GenerateServiceName(&grpcfunc.BaseInfo{
		Namespace: &ns,
		Name:      &svn,
	})

	annotations[functions.ServiceKnativeHeaderRevision] = fmt.Sprintf("%s-%s", svn, mux.Vars(r)["rev"])

	h.listPods(annotations, w, r)

}

func (h *functionHandler) listWorkflowPods(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()

	resp, err := h.srv.flowClient.Workflow(ctx, &igrpc.WorkflowRequest{
		Namespace: mux.Vars(r)["ns"],
		Path:      mux.Vars(r)["path"],
	})
	if err != nil {
		respond(w, nil, err)
		return
	}

	svn := r.URL.Query().Get("svn")
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

	svc := functions.AssembleWorkflowServiceName(resp.Oid, svn, hash)

	knrev := fmt.Sprintf("%s-%s", svc, rev)

	annotations := make(map[string]string)

	annotations[functions.ServiceKnativeHeaderRevision] = knrev
	annotations[functions.ServiceHeaderWorkflowID] = resp.Oid

	h.listPods(annotations, w, r)

}

func (h *functionHandler) listGlobalPodsSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	svc := fmt.Sprintf("%s-%s", functions.PrefixGlobal, mux.Vars(r)["svn"])
	rev := fmt.Sprintf("%s-%s", svc, mux.Vars(r)["rev"])
	h.listPodsSSE(svc, rev, w, r)
}

func (h *functionHandler) listNamespacePodsSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	svc := fmt.Sprintf("%s-%s-%s", functions.PrefixNamespace, mux.Vars(r)["ns"], mux.Vars(r)["svn"])
	rev := fmt.Sprintf("%s-%s", svc, mux.Vars(r)["rev"])
	h.listPodsSSE(svc, rev, w, r)
}

func (h *functionHandler) listWorkflowPodsSSE(w http.ResponseWriter, r *http.Request) {

	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()

	resp, err := h.srv.flowClient.Workflow(ctx, &igrpc.WorkflowRequest{
		Namespace: mux.Vars(r)["ns"],
		Path:      mux.Vars(r)["path"],
	})
	if err != nil {
		respond(w, nil, err)
		return
	}

	svn := r.URL.Query().Get("svn")
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

	svc := functions.AssembleWorkflowServiceName(resp.Oid, svn, hash)

	knrev := fmt.Sprintf("%s-%s", svc, rev)

	h.listPodsSSE(svc, knrev, w, r)

}

func (h *functionHandler) listPodsSSE(svc, rev string,
	w http.ResponseWriter, r *http.Request) {

	grpcReq := &grpc.WatchPodsRequest{
		ServiceName:  &svc,
		RevisionName: &rev,
	}

	client, err := h.client.WatchPods(r.Context(), grpcReq)
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
	w http.ResponseWriter, r *http.Request) {

	grpcReq := grpc.ListPodsRequest{
		Annotations: annotations,
	}

	resp, err := h.client.ListPods(r.Context(), &grpcReq)
	respond(w, resp, err)
}

// func (h *functionHandler) watchPods(w http.ResponseWriter, r *http.Request) {
//
// }
//
// 	sn := mux.Vars(r)["serviceName"]
// 	rn := mux.Vars(r)["revisionName"]
//
// 	// Append prefixNamespace if in namespace route and not found
// 	ns := mux.Vars(r)["namespace"]
// 	if ns != "" && !strings.HasPrefix(sn, functions.PrefixNamespace+"-") {
// 		sn = fmt.Sprintf("%s-%s-%s", functions.PrefixNamespace, ns, sn)
// 	}
//
// 	grpcReq := new(grpc.WatchPodsRequest)
// 	grpcReq.ServiceName = &sn
// 	grpcReq.RevisionName = &rn
//
// 	client, err := h.s.functions.WatchPods(r.Context(), grpcReq)
// 	if err != nil {
// 		ErrResponse(w, err)
// 		return
// 	}
//
// 	defer client.CloseSend()
// 	flusher, err := SetupSEEWriter(w)
// 	if err != nil {
// 		ErrResponse(w, err)
// 		return
// 	}
//
// 	// Create Heartbeat Ticker
// 	heartbeat := time.NewTicker(10 * time.Second)
// 	defer heartbeat.Stop()
//
// 	// Start watcher client stream channels
// 	dataCh := make(chan interface{})
// 	errorCh := make(chan error)
// 	go func() {
// 		for {
// 			data, err := client.Recv()
// 			if err != nil {
// 				errorCh <- err
// 				break
// 			} else {
// 				dataCh <- data
// 			}
// 		}
// 	}()
//
// 	for {
// 		select {
// 		case data := <-dataCh:
// 			err = WriteSSEJSONData(w, flusher, data)
// 		case err = <-errorCh:
// 		case <-client.Context().Done():
// 			err = fmt.Errorf("requested stream has timed out")
// 		case <-heartbeat.C:
// 			SendSSEHeartbeat(w, flusher)
// 		}
//
// 		// Check for errors
// 		if err != nil {
// 			ErrSSEResponse(w, flusher, err)
// 			heartbeat.Stop()
// 			return
// 		}
// 	}
//
// }
//
// func (h *Handler) watchInstanceLogs(w http.ResponseWriter, r *http.Request) {
//
// 	ns := mux.Vars(r)["namespace"]
// 	wf := mux.Vars(r)["workflowTarget"]
// 	id := mux.Vars(r)["id"]
// 	iid := fmt.Sprintf("%s/%s/%s", ns, wf, id)
//
// 	flusher, err := SetupSEEWriter(w)
// 	if err != nil {
// 		ErrResponse(w, err)
// 		return
// 	}
//
// 	grpcReq := new(ingress.WatchWorkflowInstanceLogsRequest)
// 	grpcReq.InstanceId = &iid
//
// 	client, err := h.s.direktiv.WatchWorkflowInstanceLogs(r.Context(), grpcReq)
// 	defer client.CloseSend()
// 	if err != nil {
// 		ErrResponse(w, err)
// 		return
// 	}
//
// 	// Create Heartbeat Ticker
// 	heartbeat := time.NewTicker(10 * time.Second)
// 	defer heartbeat.Stop()
//
// 	// Start watcher client stream channels
// 	dataCh := make(chan interface{})
// 	errorCh := make(chan error)
// 	go func() {
// 		for {
// 			data, err := client.Recv()
// 			if err != nil {
// 				errorCh <- err
// 				break
// 			} else {
// 				dataCh <- data
// 			}
// 		}
// 	}()
//
// 	for {
// 		select {
// 		case data := <-dataCh:
// 			err = WriteSSEJSONData(w, flusher, data)
// 		case err = <-errorCh:
// 		case <-client.Context().Done():
// 			err = fmt.Errorf("requested stream has timed out")
// 		case <-heartbeat.C:
// 			SendSSEHeartbeat(w, flusher)
// 		}
//
// 		// Check for errors
// 		if err != nil {
// 			ErrSSEResponse(w, flusher, err)
// 			heartbeat.Stop()
// 			return
// 		}
// 	}
// }
