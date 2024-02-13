package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type varController struct {
	db *database.DB
}

func (e *varController) mountRouter(r chi.Router) {
	r.Get("/{variableID}", e.get)
	r.Delete("/{variableID}", e.delete)
	r.Patch("/{variableID}", e.update)

	r.Get("/", e.list)
	r.Post("/", e.create)
}

func (e *varController) get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "variableID"))
	if err != nil {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "variable id is invalid uuid string",
		})

		return
	}

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()
	dStore := db.DataStore()

	// Fetch one
	variable, err := dStore.RuntimeVariables().GetByID(r.Context(), id)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	writeJSON(w, variable)
}

func (e *varController) delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "variableID"))
	if err != nil {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "variable id is invalid uuid string",
		})

		return
	}

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()
	dStore := db.DataStore()

	// Fetch one
	err = dStore.RuntimeVariables().Delete(r.Context(), id)
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

func (e *varController) update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "variableID"))
	if err != nil {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "variable id is invalid uuid string",
		})

		return
	}

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()
	dStore := db.DataStore()

	// Fetch one
	variable, err := dStore.RuntimeVariables().GetByID(r.Context(), id)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	req := struct {
		Name     *string `json:"name"`
		Data     string  `json:"data"`
		MimeType *string `json:"mimeType"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeNotJsonError(w, err)
		return
	}

	// Check if data is valid base64 encoded string.
	decodedBytes, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil && req.Data != "" {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "updated variable data has invalid base64 string",
		})

		return
	}
	if req.MimeType != nil {
		variable.MimeType = *req.MimeType
	}
	if req.Name != nil {
		variable.MimeType = *req.Name
	}
	if req.Data != "" {
		variable.Data = decodedBytes
	}

	updatedVar, err := dStore.RuntimeVariables().Set(r.Context(), variable)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	err = db.Commit(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, updatedVar)
}

func (e *varController) create(w http.ResponseWriter, r *http.Request) {
	//nolint:forcetypeassert
	ns := r.Context().Value(ctxKeyNamespace{}).(*core.Namespace)

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()
	dStore := db.DataStore()

	// Parse request.
	req := struct {
		Name             string `json:"name"`
		MimeType         string `json:"mimeType"`
		Data             string `json:"data"`
		InstanceIDString string `json:"instanceId"`
		WorkflowPath     string `json:"workflowPath"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeNotJsonError(w, err)
		return
	}
	instanceID, err := uuid.Parse(req.InstanceIDString)
	if err != nil && req.InstanceIDString != "" {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "field instanceId has uuid string",
		})

		return
	}

	// Check if data is valid base64 encoded string.
	decodedBytes, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "field data has invalid base64 string",
		})

		return
	}

	// Create variable.
	newVar, err := dStore.RuntimeVariables().Set(r.Context(), &core.RuntimeVariable{
		Namespace:    ns.Name,
		Name:         req.Name,
		Data:         decodedBytes,
		MimeType:     req.MimeType,
		InstanceID:   instanceID,
		WorkflowPath: req.WorkflowPath,
	})
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	err = db.Commit(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, newVar)
}

func (e *varController) list(w http.ResponseWriter, r *http.Request) {
	//nolint:forcetypeassert
	ns := r.Context().Value(ctxKeyNamespace{}).(*core.Namespace)

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	writeJSON(w, ns)
}
