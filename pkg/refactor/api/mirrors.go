// nolint
package api

import (
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"github.com/go-chi/chi/v5"
)

type mirrorsController struct {
	db            *database.DB
	syncNamespace core.SyncNamespace
	bus           *pubsub.Bus
}

func (e *mirrorsController) mountRouter(r chi.Router) {
	r.Get("/", e.list)
	r.Post("/", e.create)
}

func (e *mirrorsController) create(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	mirConfig, err := db.DataStore().Mirror().GetConfig(r.Context(), ns.Name)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	proc, err := e.syncNamespace(ns, mirConfig)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	writeJSON(w, proc)
}

func (e *mirrorsController) list(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	processes, err := db.DataStore().Mirror().GetProcessesByNamespace(r.Context(), ns.Name)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	writeJSON(w, processes)
}
