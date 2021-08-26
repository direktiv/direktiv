package api

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	goaway "github.com/TwinProduction/go-away"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/gorilla/mux"
	"github.com/rung/go-safecast"
	"github.com/vorteil/direktiv/pkg/ingress"
	"github.com/vorteil/direktiv/pkg/util"
)

func (h *Handler) namespaces(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.GetNamespaces(ctx, &ingress.GetNamespacesRequest{})
	if err != nil {
		logger.Errorf("error getting namespaces: %v", err)
		ErrResponse(w, err)
		return
	}

	writeData(resp, w)

}

func (h *Handler) addNamespace(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]
	if ok := util.MatchesRegex(n); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errNamespaceRegex.Error()))
		return
	}

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	if goaway.IsProfane(n) {
		ErrResponse(w, errors.New("no profanity allowed in namespace names"))
		return
	}

	// check if namespace is valid
	for _, blockWord := range h.s.blocklist {
		if strings.ToLower(n) == blockWord {
			ErrResponse(w, errors.New(fmt.Sprintf("namespace can not be a reserved word,  reserved words: %v", h.s.blocklist)))
			return
		}
	}

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

	event := new(cloudevents.Event)
	err = event.UnmarshalJSON(b)
	if err != nil {
		ErrResponse(w, err)
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
