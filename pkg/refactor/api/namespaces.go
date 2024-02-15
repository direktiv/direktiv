// nolint
package api

import (
	"errors"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"github.com/go-chi/chi/v5"
)

type nsController struct {
	db  *database.DB
	bus *pubsub.Bus
}

type namespaceWithSettings struct {
	*core.Namespace
	MirrorSettings *mirror.Config `json:"mirrorSettings"`
}

func (e *nsController) mountRouter(r chi.Router) {
	r.Get("/{nsName}", e.get)
	r.Delete("/{nsName}", e.delete)
	r.Patch("/{nsName}", e.update)

	r.Get("/", e.list)
	r.Post("/", e.create)
}

func (e *nsController) get(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "nsName")

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()
	dStore := db.DataStore()

	ns, err := dStore.Namespaces().GetByName(r.Context(), name)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}
	settings, err := dStore.Mirror().GetConfig(r.Context(), name)
	if err != nil && !errors.Is(err, core.ErrNotFound) {
		writeDataStoreError(w, err)
		return
	}
	result := namespaceWithSettings{
		Namespace:      ns,
		MirrorSettings: settings,
	}

	writeJSON(w, result)
}

func (e *nsController) delete(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "nsName")

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()
	dStore := db.DataStore()

	err = dStore.Namespaces().Delete(r.Context(), name)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	err = db.Commit(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}

	// TODO: Alan, check if here we need to fire some pubsub events.

	err = e.bus.Publish(pubsub.NamespaceDelete, name)
	// nolint
	if err != nil {
		// TODO: log error here.
	}

	writeOk(w)
}

func (e *nsController) update(w http.ResponseWriter, r *http.Request) {
}

func (e *nsController) create(w http.ResponseWriter, r *http.Request) {
}

func (e *nsController) list(w http.ResponseWriter, r *http.Request) {
	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()
	dStore := db.DataStore()

	namespaces, err := dStore.Namespaces().GetAll(r.Context())
	if err != nil {
		writeDataStoreError(w, err)
		return
	}
	mirrors, err := dStore.Mirror().GetAllConfigs(r.Context())
	if err != nil {
		writeDataStoreError(w, err)
		return
	}
	indexedMirrors := map[string]*mirror.Config{}
	for _, m := range mirrors {
		indexedMirrors[m.Namespace] = m
	}

	var result []namespaceWithSettings
	for _, ns := range namespaces {
		settings, _ := indexedMirrors[ns.Name]
		result = append(result, namespaceWithSettings{
			Namespace:      ns,
			MirrorSettings: settings,
		})
	}

	writeJSON(w, result)
}
