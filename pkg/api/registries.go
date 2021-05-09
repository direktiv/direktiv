package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/vorteil/direktiv/pkg/ingress"
)

func (h *Handler) registries(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.GetRegistries(ctx, &ingress.GetRegistriesRequest{
		Namespace: &n,
	})
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeData(resp, w)

}

func (h *Handler) createRegistry(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]

	st := new(NameDataTuple)
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	err = json.Unmarshal(b, st)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.StoreRegistry(ctx, &ingress.StoreRegistryRequest{
		Namespace: &n,
		Name:      &st.Name,
		Data:      []byte(st.Data),
	})
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeData(resp, w)

}

func (h *Handler) deleteRegistry(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]

	st := new(NameDataTuple)
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	err = json.Unmarshal(b, st)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.DeleteRegistry(ctx, &ingress.DeleteRegistryRequest{
		Namespace: &n,
		Name:      &st.Name,
	})
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeData(resp, w)

}
