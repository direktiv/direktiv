package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/internal/cluster/cache"
	"github.com/direktiv/direktiv/internal/cluster/pubsub"
	"github.com/direktiv/direktiv/internal/compiler"
	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/datastore/datasql"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/filestore/filesql"
	"github.com/go-chi/chi/v5"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type fsController struct {
	db  *gorm.DB
	bus pubsub.EventBus

	cache cache.Cache[core.TypescriptFlow]
}

func (e *fsController) mountRouter(r chi.Router) {
	r.Get("/*", e.read)
	r.Delete("/*", e.delete)
	r.Post("/*", e.createFile)
	r.Post("/", e.createFile)
	r.Patch("/*", e.updateFile)
}

func (e *fsController) read(w http.ResponseWriter, r *http.Request) {
	// handle raw file read.
	if r.URL.Query().Get("raw") == "true" {
		e.readRaw(w, r)
		return
	}

	namespace := chi.URLParam(r, "namespace")

	db := e.db.WithContext(r.Context()).Begin()
	if db.Error != nil {
		writeInternalError(w, db.Error)
		return
	}
	defer db.Rollback()

	fStore := filesql.NewStore(db)

	path := strings.SplitN(r.URL.Path, "/files", 2)[1]
	path = filepath.Join("/", path)
	path = filepath.Clean(path)

	// Fetch file
	file, err := fStore.ForRoot(namespace).GetFile(r.Context(), path)
	if err != nil {
		writeFileStoreError(w, err)
		return
	}

	var children []*filestore.File
	var data []byte
	if file.Typ == filestore.FileTypeDirectory {
		children, err = fStore.ForRoot(namespace).ReadDirectory(r.Context(), path)
		if err != nil {
			writeInternalError(w, err)
			return
		}
	} else {
		data, err = fStore.ForFile(file).GetData(r.Context())
		if err != nil {
			writeInternalError(w, err)
			return
		}
	}

	res := struct {
		*filestore.File

		Data []byte `json:"data,omitempty"`

		Children []*filestore.File `json:"children"`
	}{
		File:     file,
		Data:     data,
		Children: children,
	}

	writeJSON(w, res)
}

func (e *fsController) readRaw(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")

	db := e.db.WithContext(r.Context()).Begin()
	if db.Error != nil {
		writeInternalError(w, db.Error)
		return
	}
	defer db.Rollback()

	fStore := filesql.NewStore(db)

	path := strings.SplitN(r.URL.Path, "/files", 2)[1]
	path = filepath.Join("/", path)
	path = filepath.Clean(path)

	// fetch file.
	file, err := fStore.ForRoot(namespace).GetFile(r.Context(), path)
	if errors.Is(err, filestore.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if file.Typ == filestore.FileTypeDirectory {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", file.MIMEType)

	data, err := fStore.ForFile(file).GetData(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		slog.Error("write response", "err", err)
	}
}

func (e *fsController) delete(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")

	db := e.db.WithContext(r.Context()).Begin()
	if db.Error != nil {
		writeInternalError(w, db.Error)
		return
	}
	defer db.Rollback()

	fStore := filesql.NewStore(db)

	path := strings.SplitN(r.URL.Path, "/files", 2)[1]
	path = filepath.Join("/", path)
	path = filepath.Clean(path)

	// Fetch file
	file, err := fStore.ForRoot(namespace).GetFile(r.Context(), path)
	if err != nil {
		writeFileStoreError(w, err)
		return
	}
	err = fStore.ForFile(file).Delete(r.Context(), true)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	// Remove all associated runtime variables.
	dStore := datasql.NewStore(db)
	err = dStore.RuntimeVariables().DeleteForWorkflow(r.Context(), namespace, path)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	err = db.WithContext(r.Context()).Commit().Error
	if err != nil {
		writeInternalError(w, err)
		return
	}

	// TODO: yassir, check the logic of sending events on fs change in all actions.
	// Publish pubsub event.
	if file.Typ.IsDirektivSpecFile() {
		err = e.bus.Publish(pubsub.SubjFileSystemChange, nil)
		if err != nil {
			slog.Error("pubsub publish", "err", err)
		}
	}

	e.cache.Notify(r.Context(), cache.CacheNotify{
		Key:    fmt.Sprintf("%s-%s-%s", namespace, "script", path),
		Action: cache.CacheUpdate,
	})

	writeOk(w)
}

func (e *fsController) createFile(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")

	db := e.db.WithContext(r.Context()).Begin()
	if db.Error != nil {
		writeInternalError(w, db.Error)
		return
	}
	defer db.Rollback()

	fStore := filesql.NewStore(db)

	req := struct {
		Name     string             `json:"name"`
		Typ      filestore.FileType `json:"type"`
		MIMEType string             `json:"mimeType"`
		Data     string             `json:"data"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeNotJSONError(w, err)
		return
	}

	// Validate if data is valid base64 encoded string.
	decodedBytes, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil && req.Typ != filestore.FileTypeDirectory {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "file data has invalid base64 string",
		})

		return
	}

	path := strings.SplitN(r.URL.Path, "/files", 2)[1]
	path = filepath.Join("/", path, req.Name)
	path = filepath.Clean(path)

	// Create file.
	newFile, err := fStore.ForRoot(namespace).CreateFile(r.Context(),
		path,
		req.Typ,
		req.MIMEType,
		decodedBytes)
	if err != nil {
		writeFileStoreError(w, err)
		return
	}

	err = db.WithContext(r.Context()).Commit().Error
	if err != nil {
		writeInternalError(w, err)
		return
	}

	// publish pubsub event for gateway, consumer, services
	if newFile.Typ.IsDirektivSpecFile() {
		err = e.bus.Publish(pubsub.SubjFileSystemChange, nil)
		// nolint:staticcheck
		if err != nil {
			slog.With("component", "api").
				Error("publish filesystem event", "err", err)
		}
	}

	res := struct {
		*filestore.File

		Data   []byte            `json:"data,omitempty"`
		Errors []json.RawMessage `json:"errors"`
	}{
		File:   newFile,
		Data:   decodedBytes,
		Errors: make([]json.RawMessage, 0),
	}

	// validate flow file. it is stored but we report errors
	if strings.HasSuffix(req.Name, core.FlowFileExtension) {
		ci := compiler.NewCompileItem(decodedBytes, req.Name)
		err = ci.TranspileAndValidate()
		if err != nil {
			jErr, _ := json.Marshal(err)
			res.Errors = append(res.Errors, jErr)
		}

		for i := range ci.ValidationErrors {
			jErr, err := json.Marshal(ci.ValidationErrors[i])
			if err != nil {
				writeInternalError(w, err)
				return
			}
			res.Errors = append(res.Errors, jErr)
		}
	}

	writeJSON(w, res)
}

func (e *fsController) updateFile(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")

	db := e.db.WithContext(r.Context()).Begin()
	if db.Error != nil {
		writeInternalError(w, db.Error)
		return
	}
	defer db.Rollback()

	fStore := filesql.NewStore(db)

	req := struct {
		Path string `json:"path"`
		Data string `json:"data,omitempty"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeNotJSONError(w, err)
		return
	}

	// Validate if data is valid base64 encoded string.
	decodedBytes, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil && req.Data != "" {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "updated file data has invalid base64 string",
		})

		return
	}

	path := strings.SplitN(r.URL.Path, "/files", 2)[1]
	path = filepath.Join("/", path)
	path = filepath.Clean(path)

	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		// Validate if data is valid yaml with direktiv files.
		var data struct{}
		if err = yaml.Unmarshal(decodedBytes, &data); err != nil && req.Data != "" {
			writeError(w, &Error{
				Code:    "request_data_invalid",
				Message: "updated file data has invalid yaml string",
			})

			return
		}
	}

	// Fetch file.
	oldFile, err := fStore.ForRoot(namespace).GetFile(r.Context(), path)
	if err != nil {
		writeFileStoreError(w, err)
		return
	}

	if req.Data != "" {
		_, err = fStore.ForFile(oldFile).SetData(r.Context(), decodedBytes)
		if err != nil {
			writeFileStoreError(w, err)
			return
		}
	}

	dStore := datasql.NewStore(db)

	if req.Path != "" {
		err = fStore.ForFile(oldFile).SetPath(r.Context(), req.Path)
		if err != nil {
			writeFileStoreError(w, err)
			return
		}
		// Update workflow_path of all associated runtime variables.
		err = dStore.RuntimeVariables().SetWorkflowPath(r.Context(), namespace, path, req.Path)
		if err != nil {
			writeInternalError(w, err)
			return
		}
		oldFile.Path = req.Path
	}

	updatedFile, err := fStore.ForRoot(namespace).GetFile(r.Context(), oldFile.Path)
	if err != nil {
		writeFileStoreError(w, err)
		return
	}

	err = db.WithContext(r.Context()).Commit().Error
	if err != nil {
		writeInternalError(w, err)
		return
	}

	// Publish pubsub event (rename).
	// if req.Path != "" && updatedFile.Typ.IsDirektivSpecFile() {
	if updatedFile.Typ.IsDirektivSpecFile() {
		err = e.bus.Publish(pubsub.SubjFileSystemChange, nil)
		if err != nil {
			slog.Error("pubsub publish", "err", err)
		}
	}

	e.cache.Notify(r.Context(), cache.CacheNotify{
		Key:    fmt.Sprintf("%s-%s-%s", namespace, "script", path),
		Action: cache.CacheUpdate,
	})

	res := struct {
		*filestore.File

		Data   []byte            `json:"data,omitempty"`
		Errors []json.RawMessage `json:"errors"`
	}{
		File:   updatedFile,
		Data:   decodedBytes,
		Errors: make([]json.RawMessage, 0),
	}

	if req.Data != "" && strings.HasSuffix(r.URL.Path, core.FlowFileExtension) {
		ci := compiler.NewCompileItem(decodedBytes, r.URL.Path)
		err = ci.TranspileAndValidate()
		if err != nil {
			jErr, _ := json.Marshal(err)
			res.Errors = append(res.Errors, jErr)
		}

		for i := range ci.ValidationErrors {
			jErr, err := json.Marshal(ci.ValidationErrors[i])
			if err != nil {
				writeInternalError(w, err)
				return
			}
			res.Errors = append(res.Errors, jErr)
		}

	}

	writeJSON(w, res)
}
