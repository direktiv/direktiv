package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/database"
	"github.com/direktiv/direktiv/internal/transpiler"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/go-chi/chi/v5"
	"gopkg.in/yaml.v3"
)

type fsController struct {
	db  *database.DB
	bus core.PubSub
}

func (e *fsController) mountRouter(r chi.Router) {
	r.Get("/*", e.read)
	r.Delete("/*", e.delete)
	r.Post("/*", e.createFile)
	r.Patch("/*", e.updateFile)
}

func (e *fsController) read(w http.ResponseWriter, r *http.Request) {
	// handle raw file read.
	if r.URL.Query().Get("raw") == "true" {
		e.readRaw(w, r)
		return
	}

	namespace := chi.URLParam(r, "namespace")

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	fStore := db.FileStore()

	path := strings.SplitN(r.URL.Path, "/files", 2)[1]
	path = filepath.Clean("/" + path)

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

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	fStore := db.FileStore()

	path := strings.SplitN(r.URL.Path, "/files", 2)[1]
	path = filepath.Clean("/" + path)

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

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	fStore := db.FileStore()

	path := strings.SplitN(r.URL.Path, "/files", 2)[1]
	path = filepath.Clean("/" + path)

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
	dStore := db.DataStore()
	err = dStore.RuntimeVariables().DeleteForWorkflow(r.Context(), namespace, path)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	err = db.Commit(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}

	// TODO: yassir, check the logic of sending events on fs change in all actions.
	// Publish pubsub event.
	if file.Typ.IsDirektivSpecFile() {
		err = e.bus.Publish(core.FileSystemChangeEvent, nil)
		if err != nil {
			slog.Error("pubsub publish", "err", err)
		}
	}

	writeOk(w)
}

func (e *fsController) createFile(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	fStore := db.FileStore()

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
	// Validate if data is valid typescript with direktiv files.
	isDirektivFile := req.Typ != filestore.FileTypeDirectory && req.Typ != filestore.FileTypeFile
	if err = transpiler.TestCompile(string(decodedBytes)); err != nil && isDirektivFile {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "file data has invalid typescript",
		})

		return
	}

	path := strings.SplitN(r.URL.Path, "/files", 2)[1]
	path = filepath.Clean("/" + path)

	// Create file.
	newFile, err := fStore.ForRoot(namespace).CreateFile(r.Context(),
		"/"+path+"/"+req.Name,
		req.Typ,
		req.MIMEType,
		decodedBytes)
	if err != nil {
		writeFileStoreError(w, err)
		return
	}

	err = db.Commit(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}

	// Publish pubsub event.
	if newFile.Typ.IsDirektivSpecFile() {
		err = e.bus.Publish(core.FileSystemChangeEvent, nil)
		// nolint:staticcheck
		if err != nil {
			slog.With("component", "api").
				Error("publish filesystem event", "err", err)
		}
	}

	res := struct {
		*filestore.File

		Data []byte `json:"data,omitempty"`
	}{
		File: newFile,
		Data: decodedBytes,
	}

	writeJSON(w, res)
}

func (e *fsController) updateFile(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	fStore := db.FileStore()

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
	path = filepath.Clean("/" + path)

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

	dStore := db.DataStore()

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

	err = db.Commit(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}

	// Publish pubsub event (rename).
	if req.Path != "" && updatedFile.Typ.IsDirektivSpecFile() {
		err = e.bus.Publish(core.FileSystemChangeEvent, nil)
		if err != nil {
			slog.Error("pubsub publish", "err", err)
		}
	}

	// Publish pubsub event (update).
	if req.Data != "" && updatedFile.Typ.IsDirektivSpecFile() {
		err = e.bus.Publish(core.FileSystemChangeEvent, nil)
		// nolint:staticcheck
		if err != nil {
			slog.With("component", "api").
				Error("publish filesystem event", "err", err)
		}
	}

	res := struct {
		*filestore.File

		Data []byte `json:"data,omitempty"`
	}{
		File: updatedFile,
		Data: decodedBytes,
	}

	writeJSON(w, res)
}
