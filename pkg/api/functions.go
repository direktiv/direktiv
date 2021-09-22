// Package classification Direktiv API.
//
// direktiv api
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

	"github.com/gorilla/mux"
	"github.com/vorteil/direktiv/pkg/functions"
	"github.com/vorteil/direktiv/pkg/functions/grpc"
	grpcfunc "github.com/vorteil/direktiv/pkg/functions/grpc"
	"github.com/vorteil/direktiv/pkg/util"
	"go.uber.org/zap"
)

type functionHandler struct {
	logger *zap.SugaredLogger
	client grpcfunc.FunctionsServiceClient
}

func newFunctionHandler(logger *zap.SugaredLogger,
	router *mux.Router, addr string) (*functionHandler, error) {

	funcAddr := fmt.Sprintf("%s:5555", addr)
	logger.Infof("connecting to functions %s", funcAddr)

	conn, err := util.GetEndpointTLS(funcAddr)
	if err != nil {
		logger.Errorf("can not connect to direktiv function: %v", err)
		return nil, err
	}

	fh := &functionHandler{
		logger: logger,
		client: grpcfunc.NewFunctionsServiceClient(conn),
	}

	fh.initRoutes(router)

	return fh, err

}

func (h *functionHandler) initRoutes(r *mux.Router) {

	handlerPair(r, RN_ListServices, "", h.listGlobalServices, h.listGlobalServicesSSE)
	handlerPair(r, RN_ListPods, "/{svn}/revision/{rev}/pods", h.listGlobalPods, h.listGlobalPodsSSE)

	r.HandleFunc("/{svn}", h.listGlobalServiceSSE).Name(RN_WatchServices).Methods(http.MethodGet).Headers("Accept", "text/event-stream")

	r.HandleFunc("/{svn}/revisions", h.watchGlobalRevisions).Name(RN_WatchRevisions).Methods(http.MethodGet).Headers("Accept", "text/event-stream")
	// r.HandleFunc("/{svn}", h.listGlobalServiceSSE).Name(RN_WatchRevisions).Methods(http.MethodGet).Headers("Accept", "text/event-stream")

	// r.HandleFunc("/{svn}", h.listGlobalServiceSSE).Name(RN_WatchRevisions).Methods(http.MethodGet).Headers("Accept", "text/event-stream")
	// r.HandleFunc("/{svn}", h.listGlobalServiceSSE).Name(RN_WatchRevisions).Methods(http.MethodGet).Headers("Accept", "text/event-stream")

	// s.Router().HandleFunc("/api/watch/functions/{serviceName}/revisions/", s.handler.watchRevisions).Methods(http.MethodGet).Name(RN_WatchRevisions)
	// s.Router().HandleFunc("/api/watch/functions/{serviceName}/revisions/{revisionName}", s.handler.watchRevisions).Methods(http.MethodGet).Name(RN_WatchRevisions)

	// handlerPair(r, RN_ListPods, "/{svn}/revision/{rev}", h.listGlobalPods, h.listGlobalPodsSSE)

	// swagger:operation GET /api/functions getFunctions
	// ---
	// summary: Returns list of global functions.
	// description: Returns list of global Knative functions with 'global-' prefix.
	// responses:
	//   "201":
	//     "description": "service created"
	r.HandleFunc("", h.createGlobalService).Methods(http.MethodPost).Name(RN_CreateService)
	r.HandleFunc("/{svn}", h.deleteGlobalService).Methods(http.MethodDelete).Name(RN_DeleteServices)
	r.HandleFunc("/{svn}", h.getGlobalService).Methods(http.MethodGet).Name(RN_GetService)
	r.HandleFunc("/{svn}", h.updateGlobalService).Methods(http.MethodPost).Name(RN_UpdateService)
	r.HandleFunc("/{svn}", h.updateGlobalServiceTraffic).Methods(http.MethodPatch).Name(RN_UpdateServiceTraffic)
	r.HandleFunc("/{svn}/revision/{rev}", h.deleteGlobalRevision).Methods(http.MethodDelete).Name(RN_DeleteRevision)

	// namespace
	handlerPair(r, RN_ListNamespaceServices, "/namespaces/{ns}", h.listNamespaceServices, h.listNamespaceServicesSSE)
	handlerPair(r, RN_ListNamespacePods, "/namespaces/{ns}/function/{svn}/revision/{rev}/pods", h.listNamespacePods, h.listNamespacePodsSSE)

	r.HandleFunc("/namespaces/{ns}", h.createNamespaceService).Methods(http.MethodPost).Name(RN_CreateNamespaceService)
	r.HandleFunc("/namespaces/{ns}/function/{svn}", h.deleteNamespaceService).Methods(http.MethodDelete).Name(RN_DeleteNamespaceServices)
	r.HandleFunc("/namespaces/{ns}/function/{svn}", h.getNamespaceService).Methods(http.MethodGet).Name(RN_GetNamespaceService)
	r.HandleFunc("/namespaces/{ns}/function/{svn}", h.updateNamespaceService).Methods(http.MethodPost).Name(RN_UpdateNamespaceService)
	r.HandleFunc("/namespaces/{ns}/function/{svn}", h.updateNamespaceServiceTraffic).Methods(http.MethodPatch).Name(RN_UpdateNamespaceServiceTraffic)
	r.HandleFunc("/namespaces/{ns}/function/{svn}/revision/{rev}", h.deleteNamespaceRevision).Methods(http.MethodDelete).Name(RN_DeleteNamespaceRevision)

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
// 	"github.com/vorteil/direktiv/pkg/functions"
// 	"github.com/vorteil/direktiv/pkg/model"
//
// 	"github.com/gorilla/mux"
// 	"github.com/vorteil/direktiv/pkg/functions/grpc"
// 	"github.com/vorteil/direktiv/pkg/ingress"
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
// 			if strings.Count(svc, "-") < 2 {
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

	annotations := make(map[string]string)
	annotations[functions.ServiceHeaderScope] = functions.PrefixGlobal
	h.listServices(annotations, w, r)

}

func (h *functionHandler) listNamespaceServices(w http.ResponseWriter, r *http.Request) {

	annotations := make(map[string]string)
	annotations[functions.ServiceHeaderScope] = functions.PrefixNamespace
	annotations[functions.ServiceHeaderNamespace] = mux.Vars(r)["ns"]
	h.listServices(annotations, w, r)

}

func (h *functionHandler) listNamespaceServicesSSE(w http.ResponseWriter, r *http.Request) {

	annotations := make(map[string]string)
	annotations[functions.ServiceHeaderScope] = functions.PrefixNamespace
	annotations[functions.ServiceHeaderNamespace] = mux.Vars(r)["ns"]
	h.listServicesSSE(annotations, w, r)

}

func (h *functionHandler) listGlobalServiceSSE(w http.ResponseWriter, r *http.Request) {

	annotations := make(map[string]string)
	annotations[functions.ServiceHeaderScope] = functions.PrefixGlobal
	annotations[functions.ServiceHeaderName] = mux.Vars(r)["svn"]
	h.listServicesSSE(annotations, w, r)

}

func (h *functionHandler) listGlobalServicesSSE(w http.ResponseWriter, r *http.Request) {

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
	annotations := make(map[string]string)
	annotations[functions.ServiceHeaderScope] = functions.PrefixGlobal
	annotations[functions.ServiceHeaderName] = mux.Vars(r)["svn"]
	h.deleteService(annotations, w, r)
}

func (h *functionHandler) deleteNamespaceService(w http.ResponseWriter, r *http.Request) {
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
	h.getService(fmt.Sprintf("%s-%s", functions.PrefixGlobal,
		mux.Vars(r)["svn"]), w, r)
}

func (h *functionHandler) getGlobalServiceSSE(w http.ResponseWriter, r *http.Request) {
	annotations := make(map[string]string)
	annotations[functions.ServiceHeaderScope] = functions.PrefixGlobal
	annotations[functions.ServiceHeaderName] = fmt.Sprintf("%s-%s", functions.PrefixGlobal, mux.Vars(r)["svn"])
	h.getServiceSSE(annotations, w, r)
}

func (h *functionHandler) getNamespaceService(w http.ResponseWriter, r *http.Request) {
	h.getService(fmt.Sprintf("%s-%s-%s", functions.PrefixNamespace, mux.Vars(r)["ns"],
		mux.Vars(r)["svn"]), w, r)
}

func (h *functionHandler) getServiceSSE(annotations map[string]string,
	w http.ResponseWriter, r *http.Request) {

	grpcReq := &grpcfunc.WatchFunctionsRequest{
		Annotations: annotations,
	}

	fmt.Printf("STARTWATCHING API")
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
			fmt.Printf("STARTWATCHING API2")
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
	h.createService("", "", w, r)
}

func (h *functionHandler) createNamespaceService(w http.ResponseWriter, r *http.Request) {
	h.createService(mux.Vars(r)["ns"], "", w, r)
}

func (h *functionHandler) createService(ns, wf string,
	w http.ResponseWriter, r *http.Request) {

	obj := new(createFunctionRequest)
	err := json.NewDecoder(r.Body).Decode(obj)
	if err != nil {
		respond(w, nil, err)
		return
	}

	grpcReq := new(grpcfunc.CreateFunctionRequest)
	grpcReq.Info = &grpc.BaseInfo{
		Name:      obj.Name,
		Namespace: &ns,
		Workflow:  &wf,
		Image:     obj.Image,
		Cmd:       obj.Cmd,
		Size:      obj.Size,
		MinScale:  obj.MinScale,
	}

	// returns an empty body
	resp, err := h.client.CreateFunction(r.Context(), grpcReq)
	respond(w, resp, err)

}

type updateServiceRequest struct {
	Image          *string `json:"image,omitempty"`
	Cmd            *string `json:"cmd,omitempty"`
	Size           *int32  `json:"size,omitempty"`
	MinScale       *int32  `json:"minScale,omitempty"`
	TrafficPercent int64   `json:"trafficPercent"`
}

func (h *functionHandler) updateGlobalService(w http.ResponseWriter, r *http.Request) {
	h.updateService(fmt.Sprintf("%s-%s",
		functions.PrefixGlobal, mux.Vars(r)["svn"]), w, r)
}

func (h *functionHandler) updateNamespaceService(w http.ResponseWriter, r *http.Request) {
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

	h.updateServiceTraffic(fmt.Sprintf("%s-%s",
		functions.PrefixGlobal, mux.Vars(r)["svn"]), w, r)

}

func (h *functionHandler) updateNamespaceServiceTraffic(w http.ResponseWriter,
	r *http.Request) {
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
	h.deleteRevision(fmt.Sprintf("%s-%s-%s",
		functions.PrefixGlobal, mux.Vars(r)["svn"], mux.Vars(r)["rev"]), w, r)
}

func (h *functionHandler) deleteNamespaceRevision(w http.ResponseWriter, r *http.Request) {
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

func (h *functionHandler) watchGlobalRevisions(w http.ResponseWriter, r *http.Request) {
	// jens
}

func (h *functionHandler) watchRevisions(annotations map[string]string,
	w http.ResponseWriter, r *http.Request) {
	// jens
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
// func (h *Handler) watchLogs(w http.ResponseWriter, r *http.Request) {
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
	annotations := make(map[string]string)
	annotations[functions.ServiceKnativeHeaderRevision] = fmt.Sprintf("%s-%s-%s",
		functions.PrefixGlobal, mux.Vars(r)["svn"], mux.Vars(r)["rev"])
	annotations[functions.ServiceHeaderScope] = functions.PrefixGlobal
	h.listPods(annotations, w, r)
}

func (h *functionHandler) listNamespacePods(w http.ResponseWriter, r *http.Request) {
	annotations := make(map[string]string)
	annotations[functions.ServiceKnativeHeaderRevision] = fmt.Sprintf("%s-%s-%s",
		functions.PrefixNamespace, mux.Vars(r)["svn"], mux.Vars(r)["rev"])
	annotations[functions.ServiceHeaderScope] = functions.PrefixNamespace
	annotations[functions.ServiceHeaderNamespace] = mux.Vars(r)["ns"]
	h.listPods(annotations, w, r)
}

func (h *functionHandler) listGlobalPodsSSE(w http.ResponseWriter, r *http.Request) {
	svc := fmt.Sprintf("%s-%s", functions.PrefixGlobal, mux.Vars(r)["svn"])
	rev := fmt.Sprintf("%s-%s", svc, mux.Vars(r)["rev"])
	h.listPodsSSE(svc, rev, w, r)
}

func (h *functionHandler) listNamespacePodsSSE(w http.ResponseWriter, r *http.Request) {
	svc := fmt.Sprintf("%s-%s-%s", functions.PrefixNamespace, mux.Vars(r)["ns"], mux.Vars(r)["svn"])
	rev := fmt.Sprintf("%s-%s", svc, mux.Vars(r)["rev"])
	h.listPodsSSE(svc, rev, w, r)
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

//
// func (h *Handler) watchPods(w http.ResponseWriter, r *http.Request) {
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
