package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/functions"
	"github.com/vorteil/direktiv/pkg/model"

	"github.com/gorilla/mux"
	"github.com/vorteil/direktiv/pkg/functions/grpc"
	"github.com/vorteil/direktiv/pkg/ingress"
)

type functionAnnotationsRequest struct {
	Scope     string `json:"scope"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Workflow  string `json:"workflow"`
}

type functionResponseList struct {
	Config   *grpc.FunctionsConfig     `json:"config,omitempty"`
	Services []*functionResponseObject `json:"services"`
}

type functionResponseObject struct {
	Info struct {
		Workflow  string `json:"workflow"`
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
	functionsServiceNameAnnotation      = "direktiv.io/name"
	functionsServiceNamespaceAnnotation = "direktiv.io/namespace"
	functionsServiceWorkflowAnnotation  = "direktiv.io/workflow"
	functionsServiceScopeAnnotation     = "direktiv.io/scope"

	prefixWorkflow  = "w"
	prefixNamespace = "ns"
	prefixGlobal    = "g"
	prefixService   = "s"
)

var functionsQueryLabelMapping = map[string]string{
	"scope":     functionsServiceScopeAnnotation,
	"name":      functionsServiceNameAnnotation,
	"namespace": functionsServiceNamespaceAnnotation,
	"workflow":  functionsServiceWorkflowAnnotation,
}

func accepted(w http.ResponseWriter) {
	w.WriteHeader(http.StatusAccepted)
}

func getFunctionAnnotations(r *http.Request) (map[string]string, error) {

	annotations := make(map[string]string)

	// Get function labels from url queries
	for k, v := range r.URL.Query() {
		if aLabel, ok := functionsQueryLabelMapping[k]; ok && len(v) > 0 {
			annotations[aLabel] = v[0]
		}
	}

	// Get functions from body
	rb := new(functionAnnotationsRequest)
	err := json.NewDecoder(r.Body).Decode(rb)
	if err != nil && err != io.EOF {
		return nil, err
	} else if err == nil {
		annotations[functionsServiceNameAnnotation] = rb.Name
		annotations[functionsServiceNamespaceAnnotation] = rb.Namespace
		annotations[functionsServiceWorkflowAnnotation] = rb.Workflow
		annotations[functionsServiceScopeAnnotation] = rb.Scope
	}

	// Split serviceName
	svc := mux.Vars(r)["serviceName"]
	if svc != "" {
		fmt.Printf("svc ======= %v\n", svc)

		// Split namespaced service name
		if strings.HasPrefix(svc, prefixNamespace) {
			if strings.Count(svc, "-") < 2 {
				return nil, fmt.Errorf("service name is incorrect format, does not include scope and name")
			}

			firstInd := strings.Index(svc, "-")
			lastInd := strings.LastIndex(svc, "-")
			annotations[functionsServiceNamespaceAnnotation] = svc[firstInd+1 : lastInd]
			annotations[functionsServiceNameAnnotation] = svc[lastInd+1:]
			annotations[functionsServiceScopeAnnotation] = svc[:firstInd]
		} else {
			if strings.Count(svc, "-") < 1 {
				return nil, fmt.Errorf("service name is incorrect format, does not include scope")
			}

			firstInd := strings.Index(svc, "-")
			annotations[functionsServiceNameAnnotation] = svc[firstInd+1:]
			annotations[functionsServiceScopeAnnotation] = svc[:firstInd]
		}
	}

	// Handle if this was reached via the workflow route
	wf := mux.Vars(r)["workflowTarget"]
	if wf != "" {
		if annotations[functionsServiceScopeAnnotation] != "" && annotations[functionsServiceScopeAnnotation] != prefixWorkflow {
			return nil, fmt.Errorf("this route is for workflow-scoped requests")
		}

		annotations[functionsServiceWorkflowAnnotation] = wf
		annotations[functionsServiceScopeAnnotation] = prefixWorkflow
	}

	// Handle if this was reached via the namespaced route
	ns := mux.Vars(r)["namespace"]
	if ns != "" {
		if annotations[functionsServiceScopeAnnotation] == prefixGlobal {
			return nil, fmt.Errorf("this route is for namespace-scoped requests or lower, not global")
		}

		annotations[functionsServiceNamespaceAnnotation] = ns

		if annotations[functionsServiceScopeAnnotation] == "" {
			annotations[functionsServiceScopeAnnotation] = prefixNamespace
		}
	}

	del := make([]string, 0)
	for k, v := range annotations {
		if v == "" {
			del = append(del, k)
		}
	}

	for _, v := range del {
		delete(annotations, v)
	}

	fmt.Printf("annotations !!!!!!!!!!!! = %v", annotations)

	return annotations, nil
}

func prepareFunctionsForResponse(functions []*grpc.FunctionsInfo) []*functionResponseObject {
	out := make([]*functionResponseObject, 0)

	for _, function := range functions {

		obj := new(functionResponseObject)
		iinf := function.GetInfo()
		if iinf != nil {
			if iinf.Workflow != nil {
				obj.Info.Workflow = *iinf.Workflow
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

		obj.ServiceName = function.GetServiceName()
		obj.Status = function.GetStatus()
		obj.Conditions = function.GetConditions()

		out = append(out, obj)
	}

	return out
}

func (h *Handler) listServices(w http.ResponseWriter, r *http.Request) {

	a, err := getFunctionAnnotations(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	grpcReq := grpc.ListFunctionsRequest{
		Annotations: a,
	}

	resp, err := h.s.functions.ListFunctions(r.Context(), &grpcReq)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	functions := resp.GetFunctions()
	out := prepareFunctionsForResponse(functions)

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

	a, err := getFunctionAnnotations(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	grpcReq := grpc.ListFunctionsRequest{
		Annotations: a,
	}

	// returns an empty response
	_, err = h.s.functions.DeleteFunctions(r.Context(), &grpcReq)
	if err != nil {
		ErrResponse(w, err)
		return
	}

}

func (h *Handler) deleteService(w http.ResponseWriter, r *http.Request) {

	sn := mux.Vars(r)["serviceName"]
	grpcReq := new(grpc.GetFunctionRequest)
	grpcReq.ServiceName = &sn

	_, err := h.s.functions.DeleteFunction(r.Context(), grpcReq)
	if err != nil {
		ErrResponse(w, err)
		return
	}

}

type getFunctionResponse struct {
	Name      string                         `json:"name,omitempty"`
	Namespace string                         `json:"namespace,omitempty"`
	Workflow  string                         `json:"workflow,omitempty"`
	Config    *grpc.FunctionsConfig          `json:"config,omitempty"`
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
	grpcReq := new(grpc.GetFunctionRequest)
	grpcReq.ServiceName = &sn

	resp, err := h.s.functions.GetFunction(r.Context(), grpcReq)
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

	grpcReq := new(grpc.CreateFunctionRequest)
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
	_, err = h.s.functions.CreateFunction(r.Context(), grpcReq)
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

	grpcReq := new(grpc.UpdateFunctionRequest)
	grpcReq.ServiceName = &sn
	grpcReq.Info = &grpc.BaseInfo{
		Image:    obj.Image,
		Cmd:      obj.Cmd,
		Size:     obj.Size,
		MinScale: obj.MinScale,
	}
	grpcReq.TrafficPercent = &obj.TrafficPercent

	// returns an empty body
	_, err = h.s.functions.UpdateFunction(r.Context(), grpcReq)
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

	_, err = h.s.functions.SetFunctionsTraffic(r.Context(), grpcReq)
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

	_, err := h.s.functions.DeleteRevision(r.Context(), grpcReq)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	accepted(w)
}

type serviceItem struct {
	name, service string
}

func calculateList(client grpc.FunctionsServiceClient,
	items []serviceItem, annotations map[string]string, ns string) ([]*grpc.FunctionsInfo, error) {

	resp, err := client.ListFunctions(context.Background(),
		&grpc.ListFunctionsRequest{
			Annotations: annotations,
		})
	if err != nil {
		return nil, err
	}

	gisos := make(map[string]*grpc.FunctionsInfo)

	imgStatus := "False"
	imgErr := "not found"
	imgNS := ""

	condName := "Ready"
	condStatus := "False"

	condMessage := "Global service does not exist"

	if len(annotations) > 1 {
		condMessage = "Namespace service does not exist"
		imgNS = ns
	}

	cond := &grpc.Condition{
		Name:    &condName,
		Status:  &condStatus,
		Message: &condMessage,
	}

	// populate the map with "error items"
	for i := range items {
		li := items[i]

		ns := ""
		if annons, ok := annotations[functionsServiceNamespaceAnnotation]; ok {
			ns = annons
		}

		svcName, _, err := functions.GenerateServiceName(ns, "", li.service)
		if err != nil {
			log.Errorf("can not generate service name: %v", err)
			continue
		}

		info := &grpc.FunctionsInfo{
			Status:      &imgStatus,
			ServiceName: &li.service,
			Info: &grpc.BaseInfo{
				Image:     &imgErr,
				Namespace: &imgNS,
			},
			Conditions: []*grpc.Condition{
				cond,
			},
		}
		gisos[svcName] = info

	}

	isos := resp.GetFunctions()

	for i := range isos {
		// that item exists, we replace
		log.Debugf("checking %v", isos[i].GetServiceName())
		if _, ok := gisos[isos[i].GetServiceName()]; ok {
			gisos[isos[i].GetServiceName()] = isos[i]
		}
	}

	var retIsos []*grpc.FunctionsInfo

	for _, v := range gisos {
		retIsos = append(retIsos, v)
	}
	return retIsos, nil

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

	var fnNS, fnGlobal []serviceItem

	allFunctions := make([]*grpc.FunctionsInfo, 0)
	wfFns := false

	for _, fn := range workflow.Functions {
		switch fn.GetType() {
		case model.ReusableContainerFunctionType:
			wfFns = true
		case model.NamespacedKnativeFunctionType:
			fnNS = append(fnNS, serviceItem{
				name:    fn.GetID(),
				service: fn.(*model.NamespacedFunctionDefinition).KnativeService,
			})
		case model.GlobalKnativeFunctionType:
			fnGlobal = append(fnGlobal, serviceItem{
				name:    fn.GetID(),
				service: fn.(*model.GlobalFunctionDefinition).KnativeService,
			})
		}
	}

	// we add all workflow functions
	if wfFns {
		wfResp, err := h.s.functions.ListFunctions(r.Context(), &grpc.ListFunctionsRequest{
			Annotations: map[string]string{
				functionsServiceWorkflowAnnotation:  wf,
				functionsServiceNamespaceAnnotation: ns,
				functionsServiceScopeAnnotation:     prefixWorkflow,
			},
		})
		if err != nil {
			ErrResponse(w, err)
			return
		}
		allFunctions = append(allFunctions, wfResp.GetFunctions()...)
	}

	if len(fnNS) > 0 {

		i, err := calculateList(h.s.functions, fnNS,
			map[string]string{
				functionsServiceNamespaceAnnotation: ns,
				functionsServiceScopeAnnotation:     prefixNamespace,
			}, ns)

		if err != nil {
			ErrResponse(w, err)
			return
		}
		allFunctions = append(allFunctions, i...)

	}

	if len(fnGlobal) > 0 {

		i, err := calculateList(h.s.functions, fnGlobal,
			map[string]string{
				functionsServiceScopeAnnotation: prefixGlobal,
			}, ns)

		if err != nil {
			ErrResponse(w, err)
			return
		}
		allFunctions = append(allFunctions, i...)

	}

	out := prepareFunctionsForResponse(allFunctions)
	if err := json.NewEncoder(w).Encode(out); err != nil {
		ErrResponse(w, err)
		return
	}

}

func (h *Handler) watchFunctions(w http.ResponseWriter, r *http.Request) {

	a, err := getFunctionAnnotations(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	grpcReq := grpc.WatchFunctionsRequest{
		Annotations: a,
	}

	client, err := h.s.functions.WatchFunctions(r.Context(), &grpcReq)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	go func() {
		<-client.Context().Done()
		//TODO: done
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		ErrResponse(w, fmt.Errorf("streaming unsupported"))
		return
	}

	for {
		resp, err := client.Recv()
		if err != nil {
			fmt.Println("superfluos breaks")
			ErrResponse(w, err)
			return
		}

		b, err := json.Marshal(resp.Event)
		if err != nil {
			ErrResponse(w, fmt.Errorf("got bad data: %w", err))
			return
		}

		w.Write([]byte(fmt.Sprintf("event: %s", *resp.Event)))
		w.Write([]byte(fmt.Sprintf("data: %s", string(b))))
		fmt.Println("Writing event")

		flusher.Flush()
	}
}

func (h *Handler) watchFunctionsV2(w http.ResponseWriter, r *http.Request) {

	annotations := make(map[string]string)
	ns := mux.Vars(r)["namespace"]

	annotations[functionsServiceNamespaceAnnotation] = ns
	annotations[functionsServiceScopeAnnotation] = "ns"

	grpcReq := grpc.WatchFunctionsRequest{
		Annotations: annotations,
	}

	client, err := h.s.functions.WatchFunctions(r.Context(), &grpcReq)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	go func() {
		<-client.Context().Done()
		//TODO: done
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		fmt.Println("HELLO1")
		ErrResponse(w, fmt.Errorf("streaming unsupported"))
		return
	}

	for {
		resp, err := client.Recv()
		if err != nil {
			fmt.Println("HELLO2")
			ErrResponse(w, err)
			return
		}

		b, err := json.Marshal(resp)
		if err != nil {
			ErrResponse(w, fmt.Errorf("got bad data: %w", err))
			return
		}

		// w.Write([]byte(fmt.Sprintf("event: %s", *resp.Event)))
		fmt.Printf("writing: %s", fmt.Sprintf("event: %s\ndata: %s\n\n", *resp.Event, string(b)))
		w.Write([]byte(fmt.Sprintf("data: %s\n\n", string(b))))
		fmt.Println("Writing event")

		flusher.Flush()
	}
}

func (h *Handler) watchFunctionsV3(w http.ResponseWriter, r *http.Request) {

	fmt.Println("jon 1")
	a, err := getFunctionAnnotations(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	fmt.Printf("jon 2 === %v\n", a)
	grpcReq := grpc.WatchFunctionsRequest{
		Annotations: a,
	}

	fmt.Println("jon 3")

	client, err := h.s.functions.WatchFunctions(r.Context(), &grpcReq)
	if err != nil {
		ErrResponse(w, err)
		return
	}
	fmt.Println("jon 4")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	go func() {
		<-client.Context().Done()
		//TODO: done
	}()
	fmt.Println("jon 5")
	flusher, ok := w.(http.Flusher)
	if !ok {
		ErrResponse(w, fmt.Errorf("streaming unsupported"))
		return
	}

	// var httpOk bool
	// errch := make(chan error)

	for {
		fmt.Println("jon 6")
		resp, err := client.Recv()
		if err != nil {
			fmt.Printf("Im blocked1, err = %v", err)

			// errch <- fmt.Errorf("client failed: %w", err)
			ErrResponse(w, err)
			fmt.Printf("Im blocked 1!!!, err = %v", err)

			return
		}

		fmt.Println("jon 7")
		b, err := json.Marshal(resp)
		if err != nil {
			fmt.Printf("Im blocked2, err = %v", err)

			// errch <- fmt.Errorf("got bad data: %w", err)
			ErrResponse(w, err)
			fmt.Printf("Im blocked2 !!!, err = %v", err)

			return
		}

		fmt.Println("jon 8")
		// w.Write([]byte(fmt.Sprintf("event: +%v", resp)))
		fmt.Printf("WRITING ======= %s", fmt.Sprintf("data: %s", string(b)))
		_, err = w.Write([]byte(fmt.Sprintf("data: %s\n\n", string(b))))
		if err != nil {
			fmt.Printf("Im blocked3, err = %v", err)
			// errch <- fmt.Errorf("failed to write data: %w", err)
			ErrResponse(w, err)
			fmt.Printf("Im blocked 3!!!, err = %v", err)
			return
		}

		fmt.Println("jon 9")
		flusher.Flush()
		// httpOk = true
	}

	// fmt.Println("jon 10")
	// err = <-errch
	// fmt.Printf("err = +%v", err)
	// if !httpOk {
	// 	ErrResponse(w, err)
	// }

}

func (h *Handler) WatchRevisions(w http.ResponseWriter, r *http.Request) {

	sn := mux.Vars(r)["serviceName"]
	rn := mux.Vars(r)["revisionName"]

	grpcReq := new(grpc.WatchRevisionsRequest)
	grpcReq.ServiceName = &sn
	grpcReq.RevisionName = &rn

	client, err := h.s.functions.WatchRevisions(r.Context(), grpcReq)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	go func() {
		<-client.Context().Done()
		//TODO: done
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		ErrResponse(w, fmt.Errorf("streaming unsupported"))
		return
	}

	// var httpOk bool
	// errch := make(chan error)

	// go func() {
	for {
		resp, err := client.Recv()
		if err != nil {
			// errch <- fmt.Errorf("client failed: %w", err)
			ErrResponse(w, err)
			return
		}

		b, err := json.Marshal(resp)
		if err != nil {
			// errch <- fmt.Errorf("got bad data: %w", err)
			ErrResponse(w, err)
			return
		}

		// w.Write([]byte(fmt.Sprintf("event: %s", *resp.Event)))
		_, err = w.Write([]byte(fmt.Sprintf("data: %s\n\n", string(b))))
		if err != nil {
			// errch <- fmt.Errorf("failed to write data: %w", err)
			ErrResponse(w, err)
			return
		}

		flusher.Flush()
		// httpOk = true
	}
	// }()

	// err = <-errch
	// if !httpOk {
	// ErrResponse(w, err)
	// }

}

func (h *Handler) watchLogs(w http.ResponseWriter, r *http.Request) {

	sn := mux.Vars(r)["podName"]
	grpcReq := new(grpc.WatchLogsRequest)
	grpcReq.PodName = &sn

	client, err := h.s.functions.WatchLogs(r.Context(), grpcReq)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	go func() {
		<-client.Context().Done()
		//TODO: done
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		ErrResponse(w, fmt.Errorf("streaming unsupported"))
		return
	}

	// var httpOk bool
	// errch := make(chan error)

	// go func() {
	for {
		resp, err := client.Recv()
		if err != nil {
			// errch <- fmt.Errorf("client failed: %w", err)
			ErrResponse(w, err)
			return
		}

		// w.Write([]byte(fmt.Sprintf("event: %s", *resp.Event)))
		_, err = w.Write([]byte(fmt.Sprintf("data: %s\n\n", string(*resp.Data))))
		if err != nil {
			// errch <- fmt.Errorf("failed to write data: %w", err)
			ErrResponse(w, err)
			return
		}

		flusher.Flush()
		// httpOk = true
	}
	// }()

	// err = <-errch
	// if !httpOk {
	// ErrResponse(w, err)
	// }

}

func (h *Handler) listPods(w http.ResponseWriter, r *http.Request) {

	a, err := getFunctionAnnotations(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	grpcReq := grpc.ListPodsRequest{
		Annotations: a,
	}

	resp, err := h.s.functions.ListPods(r.Context(), &grpcReq)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		ErrResponse(w, err)
		return
	}
}

func (h *Handler) watchPods(w http.ResponseWriter, r *http.Request) {
	sn := mux.Vars(r)["serviceName"]
	rn := mux.Vars(r)["revisionName"]

	grpcReq := new(grpc.WatchPodsRequest)
	grpcReq.ServiceName = &sn
	grpcReq.RevisionName = &rn

	client, err := h.s.functions.WatchPods(r.Context(), grpcReq)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	go func() {
		<-client.Context().Done()
		//TODO: done
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		ErrResponse(w, fmt.Errorf("streaming unsupported"))
		return
	}

	for {
		resp, err := client.Recv()
		if err != nil {
			ErrResponse(w, err)
			return
		}

		fmt.Println("Writing event")

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			ErrResponse(w, err)
			return
		}

		flusher.Flush()
	}
}

func (h *Handler) watchPodsV2(w http.ResponseWriter, r *http.Request) {

	sn := mux.Vars(r)["serviceName"]
	rn := mux.Vars(r)["revisionName"]

	grpcReq := new(grpc.WatchPodsRequest)
	grpcReq.ServiceName = &sn
	grpcReq.RevisionName = &rn

	client, err := h.s.functions.WatchPods(r.Context(), grpcReq)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	go func() {
		<-client.Context().Done()
		//TODO: done
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		ErrResponse(w, fmt.Errorf("streaming unsupported"))
		return
	}

	var httpOk bool
	errch := make(chan error)

	go func() {
		for {
			resp, err := client.Recv()
			if err != nil {
				errch <- fmt.Errorf("client failed: %w", err)
				return
			}

			b, err := json.Marshal(resp)
			if err != nil {
				errch <- fmt.Errorf("got bad data: %w", err)
				return
			}

			// w.Write([]byte(fmt.Sprintf("event: %s", *resp.Event)))
			_, err = w.Write([]byte(fmt.Sprintf("data: %s\n\n", string(b))))
			if err != nil {
				errch <- fmt.Errorf("failed to write data: %w", err)
				return
			}

			flusher.Flush()
			httpOk = true
		}
	}()

	err = <-errch
	if !httpOk {
		ErrResponse(w, err)
	}

}
