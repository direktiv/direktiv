package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vorteil/direktiv/pkg/ingress"
	"google.golang.org/grpc/status"
)

func (h *Handler) Namespaces(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := CtxDeadline()
	defer cancel()

	resp, err := h.s.direktiv.GetNamespaces(ctx, &ingress.GetNamespacesRequest{})
	if err != nil {
		ErrResponse(w, 0, err)
		return
	}

	h.s.json.Marshal(w, resp)
}

func (h *Handler) AddNamespace(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]

	ctx, cancel := CtxDeadline()
	defer cancel()

	resp, err := h.s.direktiv.AddNamespace(ctx, &ingress.AddNamespaceRequest{
		Name: &n,
	})
	if err != nil {
		ErrResponse(w, 0, err)
		return
	}

	h.s.json.Marshal(w, resp)
}

func (h *Handler) DeleteNamespace(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]

	ctx, cancel := CtxDeadline()
	defer cancel()

	resp, err := h.s.direktiv.DeleteNamespace(ctx, &ingress.DeleteNamespaceRequest{
		Name: &n,
	})
	if err != nil {
		ErrResponse(w, 0, err)
		return
	}

	h.s.json.Marshal(w, resp)
}

func (h *Handler) NamespaceEvent(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ErrResponse(w, 0, err)
		return
	}

	var contentType string
	if typeMap, ok := r.Header["Content-Type"]; ok {
		contentType = typeMap[0]
	}

	switch contentType {
	case "application/cloudevents+json; charset=utf-8":
	case "application/cloudevents+json":
	case "application/json":
		break
	default:
		ErrResponse(w, http.StatusUnsupportedMediaType, fmt.Errorf("content type '%s' is not supported. supported media types: 'application/json' ", contentType))
		return
	}

	req := ingress.BroadcastEventRequest{
		Namespace:  &n,
		Cloudevent: b,
	}

	ctx, cancel := CtxDeadline()
	defer cancel()

	resp, err := h.s.direktiv.BroadcastEvent(ctx, &req)
	if err != nil {
		s := status.Convert(err)
		ErrResponse(w, convertGRPCStatusCodeToHTTPCode(s.Code()), fmt.Errorf(s.Message()))
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := h.s.json.Marshal(w, resp); err != nil {
		ErrResponse(w, 0, err)
		return
	}

}
