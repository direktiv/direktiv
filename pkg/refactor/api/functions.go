// nolint
package api

import (
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/service"
	"github.com/go-chi/chi/v5"
)

type serviceController struct {
	manager *service.Manager
}

func (e *serviceController) mountRouter(r chi.Router) {
	r.Get("/", e.all)
}

func (e *serviceController) all(w http.ResponseWriter, r *http.Request) {
	ns := r.Context().Value(ctxKeyNamespace{}).(*core.Namespace)

	list, err := e.manager.GetListByNamespace(ns.Name)
	if err != nil {
		writeError(w, &Error{
			Code:    "internal",
			Message: "internal error: %s" + err.Error(),
		})

		return
	}

	writeJson(w, list)
}
