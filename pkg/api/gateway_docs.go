package api

import (
	"net/http"

	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/go-chi/chi/v5"
)

type gatewayDocsController struct {
	store datastore.Store
}

func (c *gatewayDocsController) mountRouter(r chi.Router) {
	r.Get("/", c.get) // list the openapi docs for the gateway routes of the current namepsace
}

func (c *gatewayDocsController) get(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, "TODO")
	return
}
