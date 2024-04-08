// nolint
package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/direktiv/direktiv/pkg/refactor/datastore"

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"github.com/go-chi/chi/v5"
)

type nsController struct {
	db  *database.DB
	bus *pubsub.Bus
}

func (e *nsController) mountRouter(r chi.Router) {
	r.Get("/{name}", e.get)
	r.Delete("/{name}", e.delete)
	r.Put("/{name}", e.update)

	r.Get("/", e.list)
	r.Post("/", e.create)
}

func (e *nsController) get(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

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
	if err != nil && !errors.Is(err, datastore.ErrNotFound) {
		writeDataStoreError(w, err)
		return
	}

	writeJSON(w, namespaceApiObject(ns, settings))
}

func (e *nsController) delete(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

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

	err = e.bus.DebouncedPublish(pubsub.NamespaceDelete, name)
	// nolint
	if err != nil {
		// TODO: log error here.
	}

	writeOk(w)
}

func (e *nsController) update(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

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

	// Parse request.
	req := struct {
		Mirror *datastore.MirrorConfig `json:"mirror"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeNotJSONError(w, err)
		return
	}
	if req.Mirror == nil {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "field mirror must be provided",
		})

		return
	}

	// Update mirroring config.
	req.Mirror.Namespace = name
	settings, err := dStore.Mirror().UpdateConfig(r.Context(), req.Mirror)
	// If no mirroring config already set, create one.
	if errors.Is(err, datastore.ErrNotFound) {
		settings, err = dStore.Mirror().CreateConfig(r.Context(), req.Mirror)
	}
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

	writeJSON(w, namespaceApiObject(ns, settings))
}

func (e *nsController) create(w http.ResponseWriter, r *http.Request) {
	// Parse request.
	req := struct {
		Name   string                  `json:"name"`
		Mirror *datastore.MirrorConfig `json:"mirror"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeNotJSONError(w, err)
		return
	}

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	ns, err := db.DataStore().Namespaces().Create(r.Context(), &datastore.Namespace{
		Name: req.Name,
	})
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	var mConfig *datastore.MirrorConfig
	if req.Mirror != nil {
		req.Mirror.Namespace = req.Name
		mConfig, err = db.DataStore().Mirror().CreateConfig(r.Context(), req.Mirror)
		if err != nil {
			writeDataStoreError(w, err)
			return
		}
	}

	_, err = db.FileStore().CreateRoot(r.Context(), uuid.New(), ns.Name)
	if err != nil {
		writeFileStoreError(w, err)
		return
	}

	err = db.Commit(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}

	// TODO: Alan, check if here we need to fire some pubsub events.

	err = e.bus.DebouncedPublish(pubsub.NamespaceCreate, req.Name)
	// nolint
	if err != nil {
		// TODO: log error here.
	}

	writeJSON(w, namespaceApiObject(ns, mConfig))
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
	indexedMirrors := map[string]*datastore.MirrorConfig{}
	for _, m := range mirrors {
		indexedMirrors[m.Namespace] = m
	}

	var result []any
	for _, ns := range namespaces {
		settings, _ := indexedMirrors[ns.Name]
		result = append(result, namespaceApiObject(ns, settings))
	}

	writeJSON(w, result)
}

func namespaceApiObject(ns *datastore.Namespace, mConfig *datastore.MirrorConfig) any {
	type apiObject struct {
		*datastore.Namespace
		Mirror *datastore.MirrorConfig `json:"mirror"`
	}

	return &apiObject{
		Namespace: ns,
		Mirror:    mConfig,
	}
}
