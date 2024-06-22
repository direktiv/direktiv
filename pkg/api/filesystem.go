package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/pkg/compiler"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/pubsub"
	"github.com/direktiv/direktiv/pkg/utils"
	"github.com/go-chi/chi/v5"
	"gopkg.in/yaml.v3"
)

type fsController struct {
	db  *database.SQLStore
	bus *pubsub.Bus
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

	ns := extractContextNamespace(r)

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
	file, err := fStore.ForNamespace(ns.Name).GetFile(r.Context(), path)
	if err != nil {
		writeFileStoreError(w, err)
		return
	}

	var children []*filestore.File
	if file.Typ == filestore.FileTypeDirectory {
		children, err = fStore.ForNamespace(ns.Name).ReadDirectory(r.Context(), path)
		if err != nil {
			writeInternalError(w, err)
			return
		}
	} else {
		data, err := fStore.ForFile(file).GetData(r.Context())
		if err != nil {
			writeInternalError(w, err)
			return
		}
		file.Data = data
	}

	res := struct {
		*filestore.File
		Children []*filestore.File `json:"children"`
	}{
		File:     file,
		Children: children,
	}

	writeJSON(w, res)
}

func (e *fsController) readRaw(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)

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
	file, err := fStore.ForNamespace(ns.Name).GetFile(r.Context(), path)
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
	ns := extractContextNamespace(r)

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
	file, err := fStore.ForNamespace(ns.Name).GetFile(r.Context(), path)
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
	err = dStore.RuntimeVariables().DeleteForWorkflow(r.Context(), ns.Name, path)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	err = db.Commit(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}

	// Publish pubsub event.
	if file.Typ.IsDirektivSpecFile() {
		err = e.bus.DebouncedPublish(&pubsub.FileSystemChangeEvent{
			Action:       "delete",
			FileType:     string(file.Typ),
			Namespace:    ns.Name,
			NamespaceID:  ns.ID,
			FilePath:     file.Path,
			DeleteFileID: file.ID,
		})
		if err != nil {
			slog.Error("pubsub publish", "err", err)
		}
	}

	writeOk(w)
}

type fileRequest struct {
	Name     string             `json:"name"`
	Typ      filestore.FileType `json:"type"`
	MIMEType string             `json:"mimeType"`
	Data     string             `json:"data"`
}

const (
	yamlFlowType = "application/yaml"
)

func (e *fsController) createFile(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	fStore := db.FileStore()

	req := fileRequest{}

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
	path = filepath.Clean("/" + path)

	filePath := filepath.Join("/", path, req.Name)
	dataType := detectFlowContent(req.Typ, req.MIMEType)

	// Validate if data is valid yaml with direktiv files.
	var data struct{}
	if err = yaml.Unmarshal(decodedBytes, &data); err != nil && dataType == yamlFlowType {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "file data has invalid yaml string",
		})
		return
	} else if dataType == utils.TypeScriptMimeType {
		// validate typescript
		compiler, err := compiler.New(filePath, string(decodedBytes))
		if err != nil {
			writeError(w, &Error{
				Code:    "request_data_invalid",
				Message: "file data has invalid typescript string",
			})
			return
		}
		_, err = compiler.CompileFlow()
		if err != nil {
			writeError(w, &Error{
				Code:    "request_data_invalid",
				Message: "file data has invalid typescript string",
			})
			return
		}
	}

	// Create file.
	newFile, err := fStore.ForNamespace(ns.Name).CreateFile(r.Context(),
		filePath,
		req.Typ,
		req.MIMEType,
		decodedBytes)
	if err != nil {
		writeFileStoreError(w, err)
		return
	}
	newFile.Data = decodedBytes

	err = db.Commit(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}

	// Publish pubsub event.
	if newFile.Typ.IsDirektivSpecFile() {
		err = e.bus.DebouncedPublish(&pubsub.FileSystemChangeEvent{
			Action:      "create",
			FileType:    string(newFile.Typ),
			Namespace:   ns.Name,
			NamespaceID: ns.ID,
			FilePath:    newFile.Path,
			MimeType:    dataType,
		})
		// nolint:staticcheck
		if err != nil {
			slog.With("component", "api").
				Error("publish filesystem event", "err", err)
		}
	}

	writeJSON(w, newFile)
}

func detectFlowContent(typ filestore.FileType, mimeType string) string {

	// if it is not a standard type return
	if typ == filestore.FileTypeDirectory || typ == filestore.FileTypeFile {
		return ""
	}

	if mimeType == utils.TypeScriptMimeType {
		return utils.TypeScriptMimeType
	}

	return yamlFlowType
}

func (e *fsController) updateFile(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	fStore := db.FileStore()

	req := struct {
		Path string `json:"path"`
		Data string `json:"data"`
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
	oldFile, err := fStore.ForNamespace(ns.Name).GetFile(r.Context(), path)
	if err != nil {
		writeFileStoreError(w, err)
		return
	}

	dataType := detectFlowContent(oldFile.Typ, oldFile.MIMEType)
	filePath := filepath.Join("/", path, oldFile.Path)

	// Validate if data is valid yaml with direktiv files.
	var data struct{}
	if err = yaml.Unmarshal(decodedBytes, &data); err != nil && dataType == yamlFlowType {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "file data has invalid yaml string",
		})
		return
	} else if dataType == utils.TypeScriptMimeType {
		// validate typescript
		compiler, err := compiler.New(filePath, string(decodedBytes))
		if err != nil {
			writeError(w, &Error{
				Code:    "request_data_invalid",
				Message: "file data has invalid typescript string",
			})
			return
		}
		_, err = compiler.CompileFlow()
		if err != nil {
			writeError(w, &Error{
				Code:    "request_data_invalid",
				Message: "file data has invalid typescript string",
			})
			return
		}
	}

	if req.Data != "" {
		_, err = fStore.ForFile(oldFile).SetData(r.Context(), decodedBytes)
		if err != nil {
			writeFileStoreError(w, err)
			return
		}
	}

	if req.Path != "" {
		err = fStore.ForFile(oldFile).SetPath(r.Context(), req.Path)
		if err != nil {
			writeFileStoreError(w, err)
			return
		}
		oldFile.Path = req.Path
	}

	updatedFile, err := fStore.ForNamespace(ns.Name).GetFile(r.Context(), oldFile.Path)
	if err != nil {
		writeFileStoreError(w, err)
		return
	}
	updatedFile.Data = decodedBytes

	// Update workflow_path of all associated runtime variables.
	dStore := db.DataStore()
	err = dStore.RuntimeVariables().SetWorkflowPath(r.Context(), ns.Name, path, req.Path)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	err = db.Commit(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}

	// Publish pubsub event (rename).
	if req.Path != "" && updatedFile.Typ.IsDirektivSpecFile() {
		err = e.bus.DebouncedPublish(&pubsub.FileSystemChangeEvent{
			Action:      "rename",
			FileType:    string(updatedFile.Typ),
			Namespace:   ns.Name,
			NamespaceID: ns.ID,
			FilePath:    updatedFile.Path,
			OldPath:     oldFile.Path,
		})
		if err != nil {
			slog.Error("pubsub publish", "err", err)
		}
	}

	// Publish pubsub event (update).
	if req.Data != "" && updatedFile.Typ.IsDirektivSpecFile() {
		err = e.bus.DebouncedPublish(&pubsub.FileSystemChangeEvent{
			Action:      "update",
			FileType:    string(updatedFile.Typ),
			Namespace:   ns.Name,
			NamespaceID: ns.ID,
			FilePath:    updatedFile.Path,
			MimeType:    dataType,
		})
		// nolint:staticcheck
		if err != nil {
			slog.With("component", "api").
				Error("publish filesystem event", "err", err)
		}
	}

	writeJSON(w, updatedFile)
}
