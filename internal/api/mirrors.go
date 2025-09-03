package api

import (
	"fmt"
	"net/http"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/database"
	"github.com/direktiv/direktiv/internal/datastore/datasql"
	"github.com/go-chi/chi/v5"
)

type mirrorsController struct {
	db  *database.DB
	bus core.PubSub
}

func (e *mirrorsController) mountRouter(r chi.Router) {
	r.Get("/", e.list)
	r.Post("/", e.create)
}

func (e *mirrorsController) create(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Conn().Rollback()

	mirConfig, err := datasql.NewStore(db.Conn()).Mirror().GetConfig(r.Context(), namespace)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	// TODO: sync
	fmt.Println(namespace)
	fmt.Println(mirConfig)

	fmt.Println("------")

	// proc, err := e.syncNamespace(nil, mirConfig)
	// if err != nil {
	// 	writeDataStoreError(w, err)
	// 	return
	// }

	// writeJSON(w, proc)
}

func (e *mirrorsController) list(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Conn().Rollback()

	processes, err := datasql.NewStore(db.Conn()).Mirror().GetProcessesByNamespace(r.Context(), namespace)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	writeJSON(w, processes)
}
