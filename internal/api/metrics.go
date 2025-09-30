package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type metricsController struct {
	db *gorm.DB
}

func (e *metricsController) mountRouter(r chi.Router) {
	r.Get("/instances", e.dummy)
}

func (e *metricsController) dummy(w http.ResponseWriter, r *http.Request) {
}
