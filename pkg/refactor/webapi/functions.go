// nolint
package webapi

import (
	"encoding/json"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/function"
	"github.com/go-chi/chi/v5"
)

type functionsController struct {
	manager *function.Manager
}

func (e *functionsController) mountRouter(r chi.Router) {
	r.Post("/", e.insert)
	r.Get("/", e.all)
}

func (e *functionsController) insert(w http.ResponseWriter, r *http.Request) {
	req := &function.Config{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, &Error{
			Code:    "body_not_json",
			Message: "couldn't parse request payload in json format",
		})

		return
	}

	e.manager.SetOneService(req)
}

func (e *functionsController) all(w http.ResponseWriter, r *http.Request) {
	list, err := e.manager.GetList()
	if err != nil {
		writeError(w, &Error{
			Code:    "internal",
			Message: "internal error: %s" + err.Error(),
		})

		return
	}

	writeData(w, list)
}
