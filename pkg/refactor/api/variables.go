package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type varController struct {
	db *database.SQLStore
}

func (e *varController) mountRouter(r chi.Router) {
	r.Get("/{variableID}", e.get)
	r.Delete("/{variableID}", e.delete)
	r.Patch("/{variableID}", e.update)

	r.Get("/", e.list)
	r.Post("/", e.create)
}

func (e *varController) get(w http.ResponseWriter, r *http.Request) {
	// handle raw var read.
	if r.URL.Query().Get("raw") == "true" {
		e.getRaw(w, r)
		return
	}
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
	variable.Data, err = dStore.RuntimeVariables().LoadData(r.Context(), variable.ID)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}

	writeJSON(w, convertVariable(variable))
}

func (e *varController) getRaw(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "variableID"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Rollback()
	dStore := db.DataStore()

	// Fetch one
	variable, err := dStore.RuntimeVariables().GetByID(r.Context(), id)
	if errors.Is(err, datastore.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	variable.Data, err = dStore.RuntimeVariables().LoadData(r.Context(), variable.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", variable.MimeType)
	_, err = w.Write(variable.Data)
	if err != nil {
		slog.Error("write raw variable response", "err", err)
	}
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
	req := &datastore.RuntimeVariablePatch{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeNotJSONError(w, err)
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

	writeJSON(w, convertVariable(updatedVar))
}

func (e *varController) create(w http.ResponseWriter, r *http.Request) {
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
		Name             string `json:"name"`
		MimeType         string `json:"mimeType"`
		Data             []byte `json:"data"`
		InstanceIDString string `json:"instanceId"`
		WorkflowPath     string `json:"workflowPath"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeNotJSONError(w, err)
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
	newVar, err := dStore.RuntimeVariables().Create(r.Context(), &datastore.RuntimeVariable{
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

	writeJSON(w, convertVariable(newVar))
}

func (e *varController) list(w http.ResponseWriter, r *http.Request) {
	// handle raw var read.
	if r.URL.Query().Get("raw") == "true" {
		e.listRaw(w, r)
		return
	}
	ns := extractContextNamespace(r)

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()
	dStore := db.DataStore()

	forInstanceID := r.URL.Query().Get("instanceId")
	_, err = uuid.Parse(forInstanceID)
	if err != nil && forInstanceID != "" {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "query param instanceId invalid uuid string",
		})

		return
	}
	forWorkflowPath := r.URL.Query().Get("workflowPath")
	if forWorkflowPath != "" && forWorkflowPath != filepath.Clean(forWorkflowPath) {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "query param workflowPath invalid file path",
		})

		return
	}

	var list []*datastore.RuntimeVariable
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

	filterByName := r.URL.Query().Get("name")
	if filterByName != "" {
		var filteredList []*datastore.RuntimeVariable
		for _, item := range list {
			if item.Name == filterByName {
				filteredList = append(filteredList, item)
			}
		}
		list = filteredList
	}

	res := make([]any, len(list))
	for i := range list {
		res[i] = convertVariable(list[i])
	}

	writeJSON(w, res)
}

func (e *varController) listRaw(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Rollback()
	dStore := db.DataStore()

	forInstanceID := r.URL.Query().Get("instanceId")
	_, err = uuid.Parse(forInstanceID)
	if err != nil && forInstanceID != "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	forWorkflowPath := r.URL.Query().Get("workflowPath")
	if forWorkflowPath != "" && forWorkflowPath != filepath.Clean(forWorkflowPath) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var list []*datastore.RuntimeVariable
	if forInstanceID != "" {
		list, err = dStore.RuntimeVariables().ListForInstance(r.Context(), uuid.MustParse(forInstanceID))
	} else if forWorkflowPath != "" {
		list, err = dStore.RuntimeVariables().ListForWorkflow(r.Context(), ns.Name, forWorkflowPath)
	} else {
		list, err = dStore.RuntimeVariables().ListForNamespace(r.Context(), ns.Name)
	}
	if errors.Is(err, datastore.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	filterByName := r.URL.Query().Get("name")
	if filterByName != "" {
		var filteredList []*datastore.RuntimeVariable
		for _, item := range list {
			if item.Name == filterByName {
				filteredList = append(filteredList, item)
			}
		}
		list = filteredList
	}
	if len(list) != 1 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	variable := list[0]
	variable.Data, err = dStore.RuntimeVariables().LoadData(r.Context(), variable.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", variable.MimeType)
	_, err = w.Write(variable.Data)
	if err != nil {
		slog.Error("write raw variable response", "err", err)
	}
}

func convertVariable(v *datastore.RuntimeVariable) any {
	type variableForAPI struct {
		ID        uuid.UUID `json:"id"`
		Typ       string    `json:"type"`
		Reference string    `json:"reference"`
		Name      string    `json:"name"`

		Size      int       `json:"size"`
		MimeType  string    `json:"mimeType"`
		Data      []byte    `json:"data,omitempty"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}

	res := &variableForAPI{
		ID:        v.ID,
		Name:      v.Name,
		Size:      v.Size,
		MimeType:  v.MimeType,
		Data:      v.Data,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}

	res.Typ = "namespace-variable"
	res.Reference = v.Namespace
	if v.InstanceID.String() != (uuid.UUID{}).String() {
		res.Reference = v.InstanceID.String()
		res.Typ = "instance-variable"
	}
	if v.WorkflowPath != "" {
		res.Reference = v.WorkflowPath
		res.Typ = "workflow-variable"
	}

	return res
}
