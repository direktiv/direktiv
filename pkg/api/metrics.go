package api

import (
	"net/http"

	"github.com/direktiv/direktiv/pkg/database"
	"github.com/go-chi/chi/v5"
)

type metricsController struct {
	db *database.DB
}

func (e *metricsController) mountRouter(r chi.Router) {
	r.Get("/instances", e.dummy)
}

func (e *metricsController) dummy(w http.ResponseWriter, r *http.Request) {
}
