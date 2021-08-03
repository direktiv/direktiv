package api

import (
	"encoding/json"
	"net/http"

	"github.com/vorteil/direktiv/pkg/isolates/grpc"
)

type listFunctionsRequest struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Workflow  string `json:"workflow"`
}

type functionResponseObject struct {
	Info struct {
		Size     int32  `json:"size"`
		Workflow string `json:"workflow"`
	} `json:"info"`
	ServiceName   string `json:"serviceName"`
	Status        string `json:"status"`
	StatusMessage string `json:"statusMessage"`
}

func (h *Handler) listFunctions(w http.ResponseWriter, r *http.Request) {

	rb := new(listFunctionsRequest)
	err := json.NewDecoder(r.Body).Decode(rb)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	grpcReq := new(grpc.ListIsolatesRequest)
	grpcReq.Annotations = make(map[string]string)

	grpcReq.Annotations["direktiv.io/name"] = rb.Name
	grpcReq.Annotations["direktiv.io/namespace"] = rb.Namespace
	grpcReq.Annotations["direktiv.io/workflow"] = rb.Workflow

	del := make([]string, 0)
	for k, v := range grpcReq.Annotations {
		if v == "" {
			del = append(del, k)
		}
	}

	for _, v := range del {
		delete(grpcReq.Annotations, v)
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
