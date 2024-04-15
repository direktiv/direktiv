package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/go-chi/chi/v5"
)

type secretsController struct {
	db *database.SQLStore
}

func (e *secretsController) mountRouter(r chi.Router) {
	r.Get("/{secretName}", e.get)
	r.Delete("/{secretName}", e.delete)
	r.Patch("/{secretName}", e.update)

	r.Get("/", e.list)
	r.Post("/", e.create)
}

func (e *secretsController) get(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)
	secretName := chi.URLParam(r, "secretName")

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()
	dStore := db.DataStore()

	// Fetch one
	secret, err := dStore.Secrets().Get(r.Context(), ns.Name, secretName)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	writeJSON(w, convertSecret(secret))
}

func (e *secretsController) delete(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)
	secretName := chi.URLParam(r, "secretName")

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()
	dStore := db.DataStore()

	// Fetch one
	err = dStore.Secrets().Delete(r.Context(), ns.Name, secretName)
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
}

func (e *secretsController) update(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)
	secretName := chi.URLParam(r, "secretName")

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()
	dStore := db.DataStore()

	// Parse request body.
	req := &struct {
		Data []byte `json:"data"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeNotJSONError(w, err)
		return
	}

	err = dStore.Secrets().Update(r.Context(), &datastore.Secret{
		Namespace: ns.Name,
		Name:      secretName,
		Data:      req.Data,
	})
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	// Fetch the updated one
	secret, err := dStore.Secrets().Get(r.Context(), ns.Name, secretName)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	err = db.Commit(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, convertSecret(secret))
}

func (e *secretsController) create(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()
	dStore := db.DataStore()

	// Parse request.
	req := struct {
		Name string `json:"name"`
		Data []byte `json:"data"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeNotJSONError(w, err)
		return
	}
	// Create secret.
	err = dStore.Secrets().Set(r.Context(), &datastore.Secret{
		Namespace: ns.Name,
		Name:      req.Name,
		Data:      req.Data,
	})
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	// Fetch the new one
	secret, err := dStore.Secrets().Get(r.Context(), ns.Name, req.Name)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	err = db.Commit(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, convertSecret(secret))
}

func (e *secretsController) list(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()
	dStore := db.DataStore()

	var list []*datastore.Secret
	list, err = dStore.Secrets().GetAll(r.Context(), ns.Name)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	res := make([]any, len(list))
	for i := range list {
		res[i] = convertSecret(list[i])
	}

	writeJSON(w, res)
}

func convertSecret(v *datastore.Secret) any {
	type secretForAPI struct {
		Name string `json:"name"`

		Initialized bool `json:"initialized"`

		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}

	res := &secretForAPI{
		Name:        v.Name,
		Initialized: v.Data != nil,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}

	return res
}
