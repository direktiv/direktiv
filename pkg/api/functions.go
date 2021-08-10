package api

import (
	"encoding/json"
	"net/http"

	"github.com/vorteil/direktiv/pkg/model"

	"github.com/gorilla/mux"
	"github.com/vorteil/direktiv/pkg/ingress"
	"github.com/vorteil/direktiv/pkg/isolates/grpc"
)

type listFunctionsRequest struct {
	Scope     string `json:"scope"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Workflow  string `json:"workflow"`
}

type functionResponseList struct {
	Config   *grpc.IsolateConfig       `json:"config,omitempty"`
	Services []*functionResponseObject `json:"services"`
}

type functionResponseObject struct {
	Info struct {
		Size      int32  `json:"size"`
		Workflow  string `json:"workflow"`
		MinScale  int32  `json:"minScale"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Image     string `json:"image"`
		Cmd       string `json:"cmd"`
	} `json:"info"`
	ServiceName string            `json:"serviceName"`
	Status      string            `json:"status"`
	Conditions  []*grpc.Condition `json:"conditions"`
}

const (
	isolateServiceNameAnnotation      = "direktiv.io/name"
	isolateServiceNamespaceAnnotation = "direktiv.io/namespace"
	isolateServiceWorkflowAnnotation  = "direktiv.io/workflow"
	isolateServiceScopeAnnotation     = "direktiv.io/scope"

	prefixWorkflow  = "w"
	prefixNamespace = "ns"
	prefixGlobal    = "g"
	prefixService   = "s"
)

func accepted(w http.ResponseWriter) {
	w.WriteHeader(http.StatusAccepted)
}

func listRequestObjectFromHTTPRequest(r *http.Request) (*grpc.ListIsolatesRequest, error) {

	rb := new(listFunctionsRequest)
	err := json.NewDecoder(r.Body).Decode(rb)
	if err != nil {
		return nil, err
	}

	grpcReq := new(grpc.ListIsolatesRequest)
	grpcReq.Annotations = make(map[string]string)

	grpcReq.Annotations[isolateServiceNameAnnotation] = rb.Name
	grpcReq.Annotations[isolateServiceNamespaceAnnotation] = rb.Namespace
	grpcReq.Annotations[isolateServiceWorkflowAnnotation] = rb.Workflow
	grpcReq.Annotations[isolateServiceScopeAnnotation] = rb.Scope

	del := make([]string, 0)
	for k, v := range grpcReq.Annotations {
		if v == "" {
			del = append(del, k)
		}
	}

	for _, v := range del {
		delete(grpcReq.Annotations, v)
	}

	return grpcReq, nil
}

func prepareIsolatesForResponse(isolates []*grpc.IsolateInfo) []*functionResponseObject {
	out := make([]*functionResponseObject, 0)

	for _, isolate := range isolates {

		obj := new(functionResponseObject)
		iinf := isolate.GetInfo()
		if iinf != nil {
			if iinf.Size != nil {
				obj.Info.Size = *iinf.Size
			}
			if iinf.Workflow != nil {
				obj.Info.Workflow = *iinf.Workflow
			}
			if iinf.MinScale != nil {
				obj.Info.MinScale = *iinf.MinScale
			}
			if iinf.Name != nil {
				obj.Info.Name = *iinf.Name
			}
			if iinf.Namespace != nil {
				obj.Info.Namespace = *iinf.Namespace
			}
			if iinf.Image != nil {
				obj.Info.Image = *iinf.Image
			}
			if iinf.Cmd != nil {
				obj.Info.Cmd = *iinf.Cmd
			}
		}

		obj.ServiceName = isolate.GetServiceName()
		obj.Status = isolate.GetStatus()
		obj.Conditions = isolate.GetConditions()

		out = append(out, obj)
	}

	return out
}

func (h *Handler) listServices(w http.ResponseWriter, r *http.Request) {

	grpcReq, err := listRequestObjectFromHTTPRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	resp, err := h.s.isolates.ListIsolates(r.Context(), grpcReq)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	isolates := resp.GetIsolates()
	out := prepareIsolatesForResponse(isolates)

	l := &functionResponseList{
		Config:   resp.GetConfig(),
		Services: out,
	}

	if err := json.NewEncoder(w).Encode(l); err != nil {
		ErrResponse(w, err)
		return
	}
}

func (h *Handler) deleteServices(w http.ResponseWriter, r *http.Request) {

	grpcReq, err := listRequestObjectFromHTTPRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// returns an empty response
	_, err = h.s.isolates.DeleteIsolates(r.Context(), grpcReq)
	if err != nil {
		ErrResponse(w, err)
		return
	}

}

func (h *Handler) deleteService(w http.ResponseWriter, r *http.Request) {

	sn := mux.Vars(r)["serviceName"]
	grpcReq := new(grpc.GetIsolateRequest)
	grpcReq.ServiceName = &sn

	_, err := h.s.isolates.DeleteIsolate(r.Context(), grpcReq)
	if err != nil {
		ErrResponse(w, err)
		return
	}

}

type getFunctionResponse struct {
	Name      string                         `json:"name,omitempty"`
	Namespace string                         `json:"namespace,omitempty"`
	Workflow  string                         `json:"workflow,omitempty"`
	Config    *grpc.IsolateConfig            `json:"config,omitempty"`
	Revisions []getFunctionResponse_Revision `json:"revisions,omitempty"`
}

type getFunctionResponse_Revision struct {
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
}

func (h *Handler) getService(w http.ResponseWriter, r *http.Request) {

	sn := mux.Vars(r)["serviceName"]
	grpcReq := new(grpc.GetIsolateRequest)
	grpcReq.ServiceName = &sn

	resp, err := h.s.isolates.GetIsolate(r.Context(), grpcReq)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	out := &getFunctionResponse{
		Name:      resp.GetName(),
		Namespace: resp.GetNamespace(),
		Workflow:  resp.GetWorkflow(),
		Revisions: make([]getFunctionResponse_Revision, 0),
		Config:    resp.GetConfig(),
	}

	for _, rev := range resp.GetRevisions() {
		out.Revisions = append(out.Revisions, getFunctionResponse_Revision{
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
		})
	}

	if err := json.NewEncoder(w).Encode(out); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

type createFunctionRequest struct {
	Name      *string `json:"name,omitempty"`
	Namespace *string `json:"namespace,omitempty"`
	Workflow  *string `json:"workflow,omitempty"`
	Image     *string `json:"image,omitempty"`
	Cmd       *string `json:"cmd,omitempty"`
	Size      *int32  `json:"size,omitempty"`
	MinScale  *int32  `json:"minScale,omitempty"`
}

func (h *Handler) createService(w http.ResponseWriter, r *http.Request) {

	obj := new(createFunctionRequest)
	err := json.NewDecoder(r.Body).Decode(obj)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	grpcReq := new(grpc.CreateIsolateRequest)
	grpcReq.Info = &grpc.BaseInfo{
		Name:      obj.Name,
		Namespace: obj.Namespace,
		Workflow:  obj.Workflow,
		Image:     obj.Image,
		Cmd:       obj.Cmd,
		Size:      obj.Size,
		MinScale:  obj.MinScale,
	}

	// returns an empty body
	_, err = h.s.isolates.CreateIsolate(r.Context(), grpcReq)
	if err != nil {
		ErrResponse(w, err)
		return
	}

}

type updateServiceRequest struct {
	Image          *string `json:"image,omitempty"`
	Cmd            *string `json:"cmd,omitempty"`
	Size           *int32  `json:"size,omitempty"`
	MinScale       *int32  `json:"minScale,omitempty"`
	TrafficPercent int64   `json:"trafficPercent"`
}

func (h *Handler) updateService(w http.ResponseWriter, r *http.Request) {

	obj := new(updateServiceRequest)
	err := json.NewDecoder(r.Body).Decode(obj)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	sn := mux.Vars(r)["serviceName"]

	grpcReq := new(grpc.UpdateIsolateRequest)
	grpcReq.ServiceName = &sn
	grpcReq.Info = &grpc.BaseInfo{
		Image:    obj.Image,
		Cmd:      obj.Cmd,
		Size:     obj.Size,
		MinScale: obj.MinScale,
	}
	grpcReq.TrafficPercent = &obj.TrafficPercent

	// returns an empty body
	_, err = h.s.isolates.UpdateIsolate(r.Context(), grpcReq)
	if err != nil {
		ErrResponse(w, err)
		return
	}

}

type updateServiceTrafficRequest struct {
	Values []struct {
		Revision string `json:"revision"`
		Percent  int64  `json:"percent"`
	} `json:"values"`
}

func (h *Handler) updateServiceTraffic(w http.ResponseWriter, r *http.Request) {

	obj := new(updateServiceTrafficRequest)
	err := json.NewDecoder(r.Body).Decode(obj)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if obj.Values == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sn := mux.Vars(r)["serviceName"]
	grpcReq := &grpc.SetTrafficRequest{
		Name:    &sn,
		Traffic: make([]*grpc.TrafficValue, 0),
	}

	for _, v := range obj.Values {
		x := v
		grpcReq.Traffic = append(grpcReq.Traffic, &grpc.TrafficValue{
			Revision: &x.Revision,
			Percent:  &x.Percent,
		})
	}

	_, err = h.s.isolates.SetIsolateTraffic(r.Context(), grpcReq)
	if err != nil {
		ErrResponse(w, err)
		return
	}

}

func (h *Handler) deleteRevision(w http.ResponseWriter, r *http.Request) {

	rev := mux.Vars(r)["revision"]
	grpcReq := &grpc.DeleteRevisionRequest{
		Revision: &rev,
	}

	_, err := h.s.isolates.DeleteRevision(r.Context(), grpcReq)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	accepted(w)
}

// /api/namespaces/{namespace}/workflows/{workflowTarget}/functions
func (h *Handler) getWorkflowFunctions(w http.ResponseWriter, r *http.Request) {

	ns := mux.Vars(r)["namespace"]
	wf := mux.Vars(r)["workflowTarget"]

	grpcReq1 := &ingress.GetWorkflowByNameRequest{
		Namespace: &ns,
		Name:      &wf,
	}

	resp, err := h.s.direktiv.GetWorkflowByName(r.Context(), grpcReq1)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	workflow := new(model.Workflow)
	err = workflow.Load(resp.Workflow)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	var hasWorkflowFunctions, hasNamespaceFunctions, hasGlobalFunctions bool
	nsFunctions := make(map[string]interface{})
	globalFunctions := make(map[string]interface{})

	allIsolates := make([]*grpc.IsolateInfo, 0)

	for _, fn := range workflow.Functions {
		switch fn.GetType() {
		case model.DefaultFunctionType:
			fallthrough
		case model.IsolatedContainerFunctionType:
			fallthrough
		case model.SubflowFunctionType:
			continue
		case model.ReusableContainerFunctionType:
			// Workflow scope
			hasWorkflowFunctions = true
		case model.NamespacedKnativeFunctionType:
			// Namespace scope
			hasNamespaceFunctions = true
			nsFunctions[fn.GetID()] = nil
		case model.GlobalKnativeFunctionType:
			// Global scope
			hasGlobalFunctions = true
			globalFunctions[fn.GetID()] = nil
		}
	}

	if hasWorkflowFunctions {
		resp, err := h.s.isolates.ListIsolates(r.Context(), &grpc.ListIsolatesRequest{
			Annotations: map[string]string{
				isolateServiceWorkflowAnnotation:  wf,
				isolateServiceNamespaceAnnotation: ns,
				isolateServiceScopeAnnotation:     prefixWorkflow,
			},
		})
		if err != nil {
			ErrResponse(w, err)
			return
		}

		allIsolates = append(allIsolates, resp.GetIsolates()...)
	}

	if hasNamespaceFunctions {
		resp, err := h.s.isolates.ListIsolates(r.Context(), &grpc.ListIsolatesRequest{
			Annotations: map[string]string{
				isolateServiceNamespaceAnnotation: ns,
				isolateServiceScopeAnnotation:     prefixNamespace,
			},
		})
		if err != nil {
			ErrResponse(w, err)
			return
		}

		for _, fn := range resp.GetIsolates() {
			if _, ok := nsFunctions[fn.GetInfo().GetName()]; ok {
				allIsolates = append(allIsolates, fn)
			}
		}
	}

	if hasGlobalFunctions {
		resp, err := h.s.isolates.ListIsolates(r.Context(), &grpc.ListIsolatesRequest{
			Annotations: map[string]string{
				isolateServiceScopeAnnotation: prefixGlobal,
			},
		})
		if err != nil {
			ErrResponse(w, err)
			return
		}

		for _, fn := range resp.GetIsolates() {
			if _, ok := globalFunctions[fn.GetInfo().GetName()]; ok {
				allIsolates = append(allIsolates, fn)
			}
		}
	}

	out := prepareIsolatesForResponse(allIsolates)
	if err := json.NewEncoder(w).Encode(out); err != nil {
		ErrResponse(w, err)
		return
	}

}
