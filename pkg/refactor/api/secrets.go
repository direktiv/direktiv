// nolint
package api

import (
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/go-chi/chi/v5"
)

type secretsController struct {
	db *database.DB
}

func (e *secretsController) mountRouter(r chi.Router) {
	r.Get("/{secretName}", e.get)
	r.Delete("/{secretName}", e.delete)
	r.Patch("/{secretName}", e.update)

	r.Get("/", e.list)
	r.Post("/", e.create)
}

func (e *secretsController) get(w http.ResponseWriter, r *http.Request) {
}

func (e *secretsController) delete(w http.ResponseWriter, r *http.Request) {
}

func (e *secretsController) update(w http.ResponseWriter, r *http.Request) {
}

func (e *secretsController) create(w http.ResponseWriter, r *http.Request) {
}

func (e *secretsController) list(w http.ResponseWriter, r *http.Request) {
}
