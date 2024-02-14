package api

import (
	"encoding/json"
	"net/http"
	"path/filepath"

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

	// Parse request body.
	req := &core.RuntimeVariablePatch{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeNotJsonError(w, err)
		return
	}

	updatedVar, err := dStore.RuntimeVariables().Patch(r.Context(), id, req)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}
	updatedVar.Data, err = dStore.RuntimeVariables().LoadData(r.Context(), updatedVar.ID)
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
		Data             []byte `json:"data"`
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
			Message: "field instanceId has invalid uuid string",
		})
		return
	}

	// Create variable.
	newVar, err := dStore.RuntimeVariables().Create(r.Context(), &core.RuntimeVariable{
		Namespace:    ns.Name,
		Name:         req.Name,
		Data:         req.Data,
		MimeType:     req.MimeType,
		InstanceID:   instanceID,
		WorkflowPath: req.WorkflowPath,
	})
	if err != nil {
		writeDataStoreError(w, err)
		return
	}
	newVar.Data, err = dStore.RuntimeVariables().LoadData(r.Context(), newVar.ID)
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
	dStore := db.DataStore()

	forInstanceID := chi.URLParam(r, "instanceId")
	_, err = uuid.Parse(forInstanceID)
	if err != nil && forInstanceID != "" {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "query param instanceId invalid uuid string",
		})
		return
	}
	forWorkflowPath := chi.URLParam(r, "workflowPath")
	if forWorkflowPath != "" && forWorkflowPath != filepath.Clean(forWorkflowPath) {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "query param workflowPath invalid file path",
		})
		return
	}

	var list []*core.RuntimeVariable
	if forInstanceID != "" {
		list, err = dStore.RuntimeVariables().ListForInstance(r.Context(), uuid.MustParse(forInstanceID))
	} else if forWorkflowPath != "" {
		list, err = dStore.RuntimeVariables().ListForWorkflow(r.Context(), ns.Name, forWorkflowPath)
	} else {
		list, err = dStore.RuntimeVariables().ListForNamespace(r.Context(), ns.Name)
	}

	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	writeJSON(w, list)
}
