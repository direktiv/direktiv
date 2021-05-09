package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/vorteil/direktiv/pkg/ingress"
)

func (h *Handler) secrets(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.GetSecrets(ctx, &ingress.GetSecretsRequest{
		Namespace: &n,
	})
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeData(resp, w)

}

func (h *Handler) createSecret(w http.ResponseWriter, r *http.Request) {

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

	resp, err := h.s.direktiv.StoreSecret(ctx, &ingress.StoreSecretRequest{
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

func (h *Handler) deleteSecret(w http.ResponseWriter, r *http.Request) {

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

	resp, err := h.s.direktiv.DeleteSecret(ctx, &ingress.DeleteSecretRequest{
		Namespace: &n,
		Name:      &st.Name,
	})
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeData(resp, w)

}
