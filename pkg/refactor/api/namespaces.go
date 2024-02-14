package api

import (
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/go-chi/chi/v5"
)

type nsController struct {
	db *database.DB
}

func (e *nsController) mountRouter(r chi.Router) {
	r.Get("/{nsName}", e.get)
	r.Delete("/{nsName}", e.delete)
	r.Patch("/{nsName}", e.update)

	r.Get("/", e.list)
	r.Post("/", e.create)
}

func (e *nsController) get(w http.ResponseWriter, r *http.Request) {
}

func (e *nsController) delete(w http.ResponseWriter, r *http.Request) {
}

func (e *nsController) update(w http.ResponseWriter, r *http.Request) {
}

func (e *nsController) create(w http.ResponseWriter, r *http.Request) {

}

func (e *nsController) list(w http.ResponseWriter, r *http.Request) {
}
