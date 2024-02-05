// nolint
package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/go-chi/chi/v5"
)

type registryController struct {
	manager core.RegistryManager
}

func (e *registryController) mountRouter(r chi.Router) {
	r.Get("/", e.all)
	r.Delete("/{registry}", e.delete)
	r.Post("/", e.create)
}

func (e *registryController) all(w http.ResponseWriter, r *http.Request) {
	ns := r.Context().Value(ctxKeyNamespace{}).(*core.Namespace)

	list, err := e.manager.ListRegistries(ns.Name)
	if err != nil {
		writeError(w, &Error{
			Code:    "internal",
			Message: "internal error: %s" + err.Error(),
		})

		return
	}

	writeJSON(w, list)
}

func (e *registryController) delete(w http.ResponseWriter, r *http.Request) {
	ns := r.Context().Value(ctxKeyNamespace{}).(*core.Namespace)
	id := chi.URLParam(r, "registry")

	err := e.manager.DeleteRegistry(ns.Name, id)
	if errors.Is(err, core.ErrNotFound) {
		writeError(w, &Error{
			Code:    "resource_not_found",
			Message: "resource(registry) is not found",
		})

		return
	}
	if err != nil {
		writeInternalError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type Registry struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	User string `json:"user"`
}

func (e *registryController) create(w http.ResponseWriter, r *http.Request) {
	ns := r.Context().Value(ctxKeyNamespace{}).(*core.Namespace)

	reg := &core.Registry{}

	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		writeNotJsonError(w, err)

		return
	}

	reg.Namespace = ns.Name
	newReg, err := e.manager.StoreRegistry(reg)
	if err != nil {
		writeInternalError(w, err)

		return
	}

	resp := new(Registry)
	remarshal(newReg, &resp)
	writeJSON(w, resp)
}
