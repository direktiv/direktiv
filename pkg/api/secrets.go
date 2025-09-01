package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/go-chi/chi/v5"
)

type secretRequest struct {
	Name string `json:"name"`
	Data []byte `json:"data"`
}

type secretsController struct {
	secretsManager core.SecretsManager
}

func (e *secretsController) mountRouter(r chi.Router) {
	r.Get("/{secretName}", e.get)
	r.Delete("/{secretName}", e.delete)
	r.Patch("/{secretName}", e.update)

	r.Get("/", e.list)
	r.Post("/", e.create)
}

func (e *secretsController) get(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	secretName := chi.URLParam(r, "secretName")

	sh, err := e.secretsManager.SecretsForNamespace(r.Context(), namespace)
	if err != nil {
		writeSecretsError(w, err)
		return
	}

	s, err := sh.Get(r.Context(), secretName)
	if err != nil {
		writeSecretsError(w, err)
		return
	}

	writeJSON(w, convert(s))
}

func (e *secretsController) delete(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	secretName := chi.URLParam(r, "secretName")

	sh, err := e.secretsManager.SecretsForNamespace(r.Context(), namespace)
	if err != nil {
		writeSecretsError(w, err)
		return
	}

	err = sh.Delete(r.Context(), secretName)
	if err != nil {
		writeSecretsError(w, err)
		return
	}

	writeOk(w)
}

func (e *secretsController) update(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	secretName := chi.URLParam(r, "secretName")

	var req secretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeNotJSONError(w, err)
		return
	}

	sh, err := e.secretsManager.SecretsForNamespace(r.Context(), namespace)
	if err != nil {
		writeSecretsError(w, err)
		return
	}

	s, err := sh.Update(r.Context(), &core.Secret{
		Name: secretName,
		Data: req.Data,
	})
	if err != nil {
		writeSecretsError(w, err)
		return
	}

	writeJSON(w, convert(s))
}

func (e *secretsController) create(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")

	// parse request
	var req secretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeNotJSONError(w, err)
		return
	}

	sh, err := e.secretsManager.SecretsForNamespace(r.Context(), namespace)
	if err != nil {
		writeSecretsError(w, err)
		return
	}

	s, err := sh.Set(r.Context(), &core.Secret{
		Name: req.Name,
		Data: req.Data,
	})
	if err != nil {
		writeSecretsError(w, err)
		return
	}

	writeJSON(w, convert(s))
}

func (e *secretsController) list(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")

	sh, err := e.secretsManager.SecretsForNamespace(r.Context(), namespace)
	if err != nil {
		writeSecretsError(w, err)
		return
	}

	secretsList, err := sh.GetAll(r.Context())
	if err != nil {
		writeSecretsError(w, err)
		return
	}

	res := make([]any, len(secretsList))
	for i := range secretsList {
		res[i] = convert(secretsList[i])
	}

	writeJSON(w, res)
}

func convert(v *core.Secret) any {
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
