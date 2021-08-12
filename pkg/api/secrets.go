package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/gorilla/mux"

	"github.com/vorteil/direktiv/pkg/functions/grpc"
	"github.com/vorteil/direktiv/pkg/ingress"
	"github.com/vorteil/direktiv/pkg/util"
)

func (h *Handler) getSecretsOrRegistries(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]
	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	var err error
	var data interface{}

	switch mux.CurrentRoute(r).GetName() {
	case RN_ListSecrets:

		var resp *ingress.GetSecretsResponse
		resp, err = h.s.direktiv.GetSecrets(ctx, &ingress.GetSecretsRequest{
			Namespace: &n,
		})
		data = resp

	case RN_ListRegistries:

		var resp *grpc.GetRegistriesResponse
		resp, err = h.s.isolates.GetRegistries(ctx, &grpc.GetRegistriesRequest{
			Namespace: &n,
		})
		data = resp

	default:
		ErrResponse(w, fmt.Errorf(http.StatusText(http.StatusBadRequest)))
		return

	}

	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeData(data, w)

}

func (h *Handler) createSecretOrRegistry(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]
	st := new(NameDataTuple)

	err := json.NewDecoder(r.Body).Decode(st)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	var resp *empty.Empty

	switch mux.CurrentRoute(r).GetName() {

	case RN_CreateSecret:

		if ok := util.MatchesVarRegex(st.Name); !ok {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(errSecretRegex.Error()))
			return
		}

		resp, err = h.s.direktiv.StoreSecret(ctx, &ingress.StoreSecretRequest{
			Namespace: &n,
			Name:      &st.Name,
			Data:      []byte(st.Data),
		})

	case RN_CreateRegistry:

		resp, err = h.s.isolates.StoreRegistry(ctx, &grpc.StoreRegistryRequest{
			Namespace: &n,
			Name:      &st.Name,
			Data:      []byte(st.Data),
		})

	default:

		ErrResponse(w, fmt.Errorf(http.StatusText(http.StatusBadRequest)))
		return

	}

	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeData(resp, w)

}

func (h *Handler) deleteSecretOrRegistry(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]
	st := new(NameDataTuple)

	err := json.NewDecoder(r.Body).Decode(st)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	var resp *empty.Empty

	switch mux.CurrentRoute(r).GetName() {
	case RN_DeleteSecret:

		resp, err = h.s.direktiv.DeleteSecret(ctx, &ingress.DeleteSecretRequest{
			Namespace: &n,
			Name:      &st.Name,
		})

	case RN_DeleteRegistry:

		resp, err = h.s.isolates.DeleteRegistry(ctx, &grpc.DeleteRegistryRequest{
			Namespace: &n,
			Name:      &st.Name,
		})

	default:

		ErrResponse(w, fmt.Errorf(http.StatusText(http.StatusBadRequest)))
		return

	}

	if err != nil {
		ErrResponse(w, err)
	}

	writeData(resp, w)

}
