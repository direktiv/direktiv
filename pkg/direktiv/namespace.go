package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/ingress"
)

func (h *Handler) namespaces(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.GetNamespaces(ctx, &ingress.GetNamespacesRequest{})
	if err != nil {
		log.Errorf("error getting namespaces: %v", err)
		ErrResponse(w, err)
		return
	}

	writeData(resp, w)

}

func (h *Handler) addNamespace(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.AddNamespace(ctx, &ingress.AddNamespaceRequest{
		Name: &n,
	})
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeData(resp, w)

}

func (h *Handler) deleteNamespace(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.DeleteNamespace(ctx, &ingress.DeleteNamespaceRequest{
		Name: &n,
	})

	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeData(resp, w)
}

func (h *Handler) namespaceLogs(w http.ResponseWriter, r *http.Request) {
	n := mux.Vars(r)["namespace"]

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	o, l := paginationParams(r)
	if l < 1 {
		l = 10
	}

	if o < 0 {
		o = 0
	}

	limit := int32(l)
	offset := int32(o)

	resp, err := h.s.direktiv.GetNamespaceLogs(ctx, &ingress.GetNamespaceLogsRequest{
		Namespace: &n,
		Limit:     &limit,
		Offset:    &offset,
	})
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeData(resp, w)
}

func (h *Handler) namespaceEvent(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ErrResponse(w, err)
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
		ErrResponse(w, fmt.Errorf("content type '%s' is not supported. supported media types: 'application/json' ", contentType))
		return
	}

	req := ingress.BroadcastEventRequest{
		Namespace:  &n,
		Cloudevent: b,
	}

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.BroadcastEvent(ctx, &req)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeData(resp, w)

}

