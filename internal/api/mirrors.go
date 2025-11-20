package api

import (
	"net/http"

	"github.com/direktiv/direktiv/internal/cluster/pubsub"
	"github.com/direktiv/direktiv/internal/datastore"
	"github.com/direktiv/direktiv/internal/datastore/datasql"
	"github.com/direktiv/direktiv/internal/mirroring"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type mirrorsController struct {
	db  *gorm.DB
	bus pubsub.EventBus
}

func (e *mirrorsController) mountRouter(r chi.Router) {
	r.Get("/", e.list)
	r.Post("/", e.create)
}

func (e *mirrorsController) create(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")

	db := e.db.WithContext(r.Context()).Begin()
	if db.Error != nil {
		writeInternalError(w, db.Error)
		return
	}
	defer db.Rollback()

	mirConfig, err := datasql.NewStore(db).Mirror().GetConfig(r.Context(), namespace)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	proc, err := mirroring.MirrorExec(r.Context(), e.bus, e.db, mirConfig, datastore.ProcessTypeInit)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	writeJSON(w, proc)
}

func (e *mirrorsController) list(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")

	db := e.db.WithContext(r.Context()).Begin()
	if db.Error != nil {
		writeInternalError(w, db.Error)
		return
	}
	defer db.Rollback()

	processes, err := datasql.NewStore(db).Mirror().GetProcessesByNamespace(r.Context(), namespace)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	writeJSON(w, processes)
}
