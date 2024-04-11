package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type nsController struct {
	db  *database.SQLStore
	bus *pubsub.Bus
}

func (e *nsController) mountRouter(r chi.Router) {
	r.Get("/{name}", e.get)
	r.Delete("/{name}", e.delete)
	r.Patch("/{name}", e.update)

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

	writeJSON(w, namespaceAPIObject(ns, settings))
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

//nolint:gocognit
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
		Mirror *struct {
			URL                  *string `json:"url"`
			GitRef               *string `json:"gitRef"`
			AuthToken            *string `json:"authToken"`
			PublicKey            *string `json:"publicKey"`
			PrivateKey           *string `json:"privateKey"`
			PrivateKeyPassphrase *string `json:"privateKeyPassphrase"`
			Insecure             *bool   `json:"insecure"`
		} `json:"mirror"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeNotJSONError(w, err)

		return
	}
	if req.Mirror == nil {
		err := dStore.Mirror().DeleteConfig(r.Context(), ns.Name)
		// if no mirror stored, then nothing to do
		if errors.Is(err, datastore.ErrNotFound) {
			writeOk(w)
			return
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
		writeOk(w)

		return
	}

	settings, err := dStore.Mirror().GetConfig(r.Context(), ns.Name)
	if err != nil && !errors.Is(err, datastore.ErrNotFound) {
		writeDataStoreError(w, err)

		return
	}
	// old setting was not set
	if errors.Is(err, datastore.ErrNotFound) {
		if req.Mirror.URL == nil || *req.Mirror.URL == "" {
			writeError(w, &Error{
				Code:    "request_data_invalid",
				Message: "mirror was not initialized, field 'url' must be provided and not empty",
			})

			return
		}
		if req.Mirror.GitRef == nil || *req.Mirror.GitRef == "" {
			writeError(w, &Error{
				Code:    "request_data_invalid",
				Message: "mirror was not initialized, field 'gitRef' must be provided and not empty",
			})

			return
		}
		settings, err = dStore.Mirror().CreateConfig(r.Context(), &datastore.MirrorConfig{
			Namespace: ns.Name,
			URL:       *req.Mirror.URL,
			GitRef:    *req.Mirror.GitRef,
		})
		if err != nil {
			writeDataStoreError(w, err)
			return
		}
	}

	if req.Mirror.URL != nil && *req.Mirror.URL != "" {
		settings.URL = *req.Mirror.URL
	}
	if req.Mirror.GitRef != nil && *req.Mirror.GitRef != "" {
		settings.GitRef = *req.Mirror.GitRef
	}
	if req.Mirror.AuthToken != nil {
		settings.AuthToken = *req.Mirror.AuthToken
	}
	if req.Mirror.PublicKey != nil {
		settings.PublicKey = *req.Mirror.PublicKey
	}
	if req.Mirror.PrivateKey != nil {
		settings.PrivateKey = *req.Mirror.PrivateKey
	}
	if req.Mirror.PrivateKeyPassphrase != nil {
		settings.PrivateKeyPassphrase = *req.Mirror.PrivateKeyPassphrase
	}
	if req.Mirror.Insecure != nil {
		settings.Insecure = *req.Mirror.Insecure
	}

	// Update mirroring config.
	settings, err = dStore.Mirror().UpdateConfig(r.Context(), settings)
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

	writeJSON(w, namespaceAPIObject(ns, settings))
}

func (e *nsController) create(w http.ResponseWriter, r *http.Request) {
	// Parse request.

	req := struct {
		Name   string `json:"name"`
		Mirror *struct {
			URL                  string `json:"url"`
			GitRef               string `json:"gitRef"`
			AuthToken            string `json:"authToken"`
			PublicKey            string `json:"publicKey"`
			PrivateKey           string `json:"privateKey"`
			PrivateKeyPassphrase string `json:"privateKeyPassphrase"`
			Insecure             bool   `json:"insecure"`
		} `json:"mirror"`
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
		if req.Mirror.URL == "" {
			writeError(w, &Error{
				Code:    "request_data_invalid",
				Message: "field 'url' must be provided",
			})

			return
		}
		if req.Mirror.GitRef == "" {
			writeError(w, &Error{
				Code:    "request_data_invalid",
				Message: "field 'gitRef' must be provided",
			})

			return
		}
		mirrorConfig := &datastore.MirrorConfig{
			Namespace:            req.Name,
			URL:                  req.Mirror.URL,
			GitRef:               req.Mirror.GitRef,
			AuthToken:            req.Mirror.AuthToken,
			PublicKey:            req.Mirror.PublicKey,
			PrivateKey:           req.Mirror.PrivateKey,
			PrivateKeyPassphrase: req.Mirror.PrivateKeyPassphrase,
			Insecure:             req.Mirror.Insecure,
		}
		mConfig, err = db.DataStore().Mirror().CreateConfig(r.Context(), mirrorConfig)
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

	writeJSON(w, namespaceAPIObject(ns, mConfig))
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

	//nolint:prealloc
	var result []any
	for _, ns := range namespaces {
		settings := indexedMirrors[ns.Name]
		result = append(result, namespaceAPIObject(ns, settings))
	}

	writeJSON(w, result)
}

func namespaceAPIObject(ns *datastore.Namespace, mConfig *datastore.MirrorConfig) any {
	type apiObject struct {
		*datastore.Namespace
		Mirror *datastore.MirrorConfig `json:"mirror"`
	}

	return &apiObject{
		Namespace: ns,
		Mirror:    mConfig,
	}
}
