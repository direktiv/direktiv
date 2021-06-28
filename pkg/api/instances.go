package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rung/go-safecast"

	"github.com/vorteil/direktiv/pkg/ingress"
)

func (h *Handler) instances(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]
	o, l := paginationParams(r)

	if l < 1 {
		l = 10
	}

	if o < 0 {
		o = 0
	}

	limit, err := safecast.Int32(l)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	offset, err := safecast.Int32(o)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.GetWorkflowInstances(ctx, &ingress.GetWorkflowInstancesRequest{
		Namespace: &n,
		Offset:    &offset,
		Limit:     &limit,
	})

	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeData(resp, w)

}

func (h *Handler) getInstance(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]
	name := mux.Vars(r)["workflowTarget"]
	id := mux.Vars(r)["id"]

	iid := fmt.Sprintf("%s/%s/%s", n, name, id)

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.GetWorkflowInstance(ctx, &ingress.GetWorkflowInstanceRequest{
		Id: &iid,
	})
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeData(resp, w)

}

func (h *Handler) cancelInstance(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]
	name := mux.Vars(r)["workflowTarget"]
	id := mux.Vars(r)["id"]

	iid := fmt.Sprintf("%s/%s/%s", n, name, id)

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.CancelWorkflowInstance(ctx, &ingress.CancelWorkflowInstanceRequest{
		Id: &iid,
	})
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeData(resp, w)

}

func (h *Handler) instanceLogs(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]
	name := mux.Vars(r)["workflowTarget"]
	id := mux.Vars(r)["id"]

	iid := fmt.Sprintf("%s/%s/%s", n, name, id)

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	o, l := paginationParams(r)
	if l < 1 {
		l = 10
	}

	if o < 0 {
		o = 0
	}

	limit, err := safecast.Int32(l)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	offset, err := safecast.Int32(o)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	resp, err := h.s.direktiv.GetWorkflowInstanceLogs(ctx, &ingress.GetWorkflowInstanceLogsRequest{
		InstanceId: &iid,
		Limit:      &limit,
		Offset:     &offset,
	})
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeData(resp, w)

}
