package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vorteil/direktiv/pkg/isolates/grpc"
)

type listFunctionsRequest struct {
	Scope     string `json:"scope"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Workflow  string `json:"workflow"`
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
	ServiceName   string `json:"serviceName"`
	Status        string `json:"status"`
	StatusMessage string `json:"statusMessage"`
}

func listRequestObjectFromHTTPRequest(r *http.Request) (*grpc.ListIsolatesRequest, error) {

	rb := new(listFunctionsRequest)
	err := json.NewDecoder(r.Body).Decode(rb)
	if err != nil {
		return nil, err
	}

	grpcReq := new(grpc.ListIsolatesRequest)
	grpcReq.Annotations = make(map[string]string)

	grpcReq.Annotations["direktiv.io/name"] = rb.Name
	grpcReq.Annotations["direktiv.io/namespace"] = rb.Namespace
	grpcReq.Annotations["direktiv.io/workflow"] = rb.Workflow
	grpcReq.Annotations["direktiv.io/scope"] = rb.Scope

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
		obj.StatusMessage = isolate.GetStatusMessage()

		out = append(out, obj)
	}

	if err := json.NewEncoder(w).Encode(out); err != nil {
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
	Revisions []getFunctionResponse_Revision `json:"revisions,omitempty"`
}

type getFunctionResponse_Revision struct {
	Name          string `json:"name,omitempty"`
	Image         string `json:"image,omitempty"`
	Cmd           string `json:"cmd,omitempty"`
	Size          int32  `json:"size,omitempty"`
	MinScale      int32  `json:"minScale,omitempty"`
	Generation    int64  `json:"generation,omitempty"`
	Created       int64  `json:"created,omitempty"`
	Status        string `json:"status,omitempty"`
	StatusMessage string `json:"statusMessage,omitempty"`
	Traffic       int64  `json:"traffic,omitempty"`
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
	}

	for _, rev := range resp.GetRevisions() {
		out.Revisions = append(out.Revisions, getFunctionResponse_Revision{
			Name:          rev.GetName(),
			Image:         rev.GetImage(),
			Cmd:           rev.GetCmd(),
			Size:          rev.GetSize(),
			MinScale:      rev.GetMinScale(),
			Generation:    rev.GetGeneration(),
			Created:       rev.GetCreated(),
			Status:        rev.GetStatus(),
			StatusMessage: rev.GetStatusMessage(),
			Traffic:       rev.GetTraffic(),
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
	Image    *string `json:"image,omitempty"`
	Cmd      *string `json:"cmd,omitempty"`
	Size     *int32  `json:"size,omitempty"`
	MinScale *int32  `json:"minScale,omitempty"`
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
		grpcReq.Traffic = append(grpcReq.Traffic, &grpc.TrafficValue{
			Revision: &v.Revision,
			Percent:  &v.Percent,
		})
	}

	_, err = h.s.isolates.SetIsolateTraffic(r.Context(), grpcReq)
	if err != nil {
		ErrResponse(w, err)
		return
	}

}
