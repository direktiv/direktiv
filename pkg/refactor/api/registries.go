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
	r.Delete("/{id}", e.delete)
	r.Post("/", e.create)
	r.Post("/test", e.test)
}

func (e *registryController) all(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)

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
	ns := extractContextNamespace(r)
	id := chi.URLParam(r, "id")

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

	writeOk(w)
}

func (e *registryController) create(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)

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

	writeJSON(w, newReg)
}

func (e *registryController) test(w http.ResponseWriter, r *http.Request) {
	reg := &core.Registry{}

	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		writeNotJsonError(w, err)

		return
	}

	err := e.manager.TestLogin(reg)
	if err != nil {
		writeInternalError(w, err)

		return
	}

	writeOk(w)
}
