package api

import (
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
