package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/go-chi/chi/v5"
)

type fsController struct {
	db *database.DB
}

func (e *fsController) mountRouter(r chi.Router) {
	r.Get("/*", e.read)
	r.Delete("/*", e.delete)
	r.Post("/create/*", e.createFile)
}

func (e *fsController) read(w http.ResponseWriter, r *http.Request) {
	//nolint:forcetypeassert
	ns := r.Context().Value(ctxKeyNamespace{}).(*core.Namespace)

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	//nolint:errcheck
	defer db.Rollback()

	fStore := db.FileStore()

	// Fetch file
	path := strings.Split(r.URL.Path, "/files-tree")[1]
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
		File  *filestore.File   `json:"file"`
		Paths []*filestore.File `json:"paths"`
	}{
		File:  file,
		Paths: children,
	}

	writeJSON(w, res)
}

func (e *fsController) delete(w http.ResponseWriter, r *http.Request) {
	//nolint:forcetypeassert
	ns := r.Context().Value(ctxKeyNamespace{}).(*core.Namespace)

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	//nolint:errcheck
	defer db.Rollback()

	fStore := db.FileStore()

	// Fetch file
	path := strings.Split(r.URL.Path, "/files-tree")[1]
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

	writeOk(w)
}

func (e *fsController) createFile(w http.ResponseWriter, r *http.Request) {
	//nolint:forcetypeassert
	ns := r.Context().Value(ctxKeyNamespace{}).(*core.Namespace)

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	//nolint:errcheck
	defer db.Rollback()

	fStore := db.FileStore()

	// nolint:tagliatelle
	req := struct {
		Name     string             `json:"name"`
		Typ      filestore.FileType `json:"type"`
		MIMEType string             `json:"mimeType"`
		Content  string             `json:"content"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeNotJsonError(w, err)
		return
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(req.Content)
	if err != nil && req.Typ != filestore.FileTypeDirectory {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "content filed has invalid base64 string",
		})

		return
	}

	// Create file
	path := strings.Split(r.URL.Path, "/files-tree/create")[1]
	newFile, err := fStore.ForNamespace(ns.Name).CreateFile(r.Context(),
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

	writeJSON(w, newFile)
}
