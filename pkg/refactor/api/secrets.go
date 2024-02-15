// nolint
package api

import (
	"encoding/json"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/go-chi/chi/v5"
)

type secretsController struct {
	db *database.DB
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

	writeJSON(w, secret)
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

	err = dStore.Secrets().Update(r.Context(), &core.Secret{
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

	writeJSON(w, secret)
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
	err = dStore.Secrets().Set(r.Context(), &core.Secret{
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

	writeJSON(w, secret)
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

	var list []*core.Secret
	list, err = dStore.Secrets().GetAll(r.Context(), ns.Name)

	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	writeJSON(w, list)
}
