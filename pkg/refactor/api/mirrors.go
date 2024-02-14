// nolint
package api

import (
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/go-chi/chi/v5"
)

type mirrorsController struct {
	db *database.DB
}

func (e *mirrorsController) mountRouter(r chi.Router) {
	r.Get("/", e.list)
	r.Post("/", e.create)
}

func (e *mirrorsController) create(w http.ResponseWriter, r *http.Request) {
}

func (e *mirrorsController) list(w http.ResponseWriter, r *http.Request) {
}
