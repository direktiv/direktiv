package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/helpers"
	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"github.com/go-chi/chi/v5"
	"gopkg.in/yaml.v3"
)

type fsController struct {
	db  *database.DB
	bus *pubsub.Bus
}

func (e *fsController) mountRouter(r chi.Router) {
	r.Get("/*", e.read)
	r.Delete("/*", e.delete)
	r.Post("/*", e.createFile)
	r.Patch("/*", e.updateFile)
}

func (e *fsController) read(w http.ResponseWriter, r *http.Request) {
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
		err = helpers.PublishEventDirektivFileChange(e.bus, file.Typ, "delete", &pubsub.FileChangeEvent{
			Namespace:    ns.Name,
			NamespaceID:  ns.ID,
			FilePath:     file.Path,
			DeleteFileID: file.ID,
		})
		// nolint:staticcheck
		if err != nil {
			// TODO: need to log error here.
		}
	}

	writeOk(w)
}

func (e *fsController) createFile(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)

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
	// Validate if data is valid yaml with direktiv files.
	isDirektivFile := req.Typ != filestore.FileTypeDirectory && req.Typ != filestore.FileTypeFile
	var data struct{}
	if err = yaml.Unmarshal(decodedBytes, &data); err != nil && isDirektivFile {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "file data has invalid yaml string",
		})

		return
	}

	path := strings.SplitN(r.URL.Path, "/files", 2)[1]
	path = filepath.Clean("/" + path)

	// Create file.
	newFile, err := fStore.ForNamespace(ns.Name).CreateFile(r.Context(),
		"/"+path+"/"+req.Name,
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
		err = helpers.PublishEventDirektivFileChange(e.bus, newFile.Typ, "create", &pubsub.FileChangeEvent{
			Namespace:   ns.Name,
			NamespaceID: ns.ID,
			FilePath:    newFile.Path,
		})
		// nolint:staticcheck
		if err != nil {
			// TODO: need to log error here.
		}
	}

	writeJSON(w, newFile)
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
	// Validate if data is valid yaml with direktiv files.
	var data struct{}
	if err = yaml.Unmarshal(decodedBytes, &data); err != nil && req.Data != "" {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "updated file data has invalid yaml string",
		})

		return
	}

	path := strings.SplitN(r.URL.Path, "/files", 2)[1]
	path = filepath.Clean("/" + path)

	// Fetch file.
	oldFile, err := fStore.ForNamespace(ns.Name).GetFile(r.Context(), path)
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
		err = helpers.PublishEventDirektivFileChange(e.bus, updatedFile.Typ, "rename", &pubsub.FileChangeEvent{
			Namespace:   ns.Name,
			NamespaceID: ns.ID,
			FilePath:    updatedFile.Path,
			OldPath:     oldFile.Path,
		})
		// nolint:staticcheck
		if err != nil {
			// TODO: need to log error here.
		}
	}

	// Publish pubsub event (update).
	if req.Data != "" && updatedFile.Typ.IsDirektivSpecFile() {
		err = helpers.PublishEventDirektivFileChange(e.bus, updatedFile.Typ, "update", &pubsub.FileChangeEvent{
			Namespace:   ns.Name,
			NamespaceID: ns.ID,
			FilePath:    updatedFile.Path,
		})
		// nolint:staticcheck
		if err != nil {
			// TODO: need to log error here.
		}
	}

	writeJSON(w, updatedFile)
}
