// nolint
package api

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/service"
	"github.com/go-chi/chi/v5"
)

type serviceController struct {
	manager core.ServiceManager
}

func (e *serviceController) mountRouter(r chi.Router) {
	r.Get("/", e.all)
	r.Get("/{serviceID}/logs/{podNumber}", e.logs)
}

func (e *serviceController) all(w http.ResponseWriter, r *http.Request) {
	ns := r.Context().Value(ctxKeyNamespace{}).(*core.Namespace)

	list, err := e.manager.GetListByNamespace(ns.Name)
	if err != nil {
		writeInternalError(w, err)

		return
	}

	writeJSON(w, list)
}

func (e *serviceController) logs(w http.ResponseWriter, r *http.Request) {
	ns := r.Context().Value(ctxKeyNamespace{}).(*core.Namespace)
	serviceID := chi.URLParam(r, "serviceID")

	readCloser, err := e.manager.StreamLogs(ns.Name, serviceID, 1)
	if errors.Is(err, service.ErrNotFound) {
		writeError(w, &Error{
			Code:    "resource_not_found",
			Message: "resource(service) is not found",
		})

		return
	}
	if err != nil {
		writeInternalError(w, err)

		return
	}
	defer readCloser.Close()

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Accel-Buffering", "no")

	buffer := make([]byte, 4*1024)
	var n int
	for {
		// TODO: this would leak requests as to could block forever.
		n, err = readCloser.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			writeInternalError(w, err)

			break
		}

		_, err := fmt.Fprintf(w, "%X\r\n%s\r\n", n, buffer[:n])
		if err != nil {
			slog.Error("TODO: add log here")
			break
		}
		w.(http.Flusher).Flush()
		time.Sleep(10 * time.Millisecond)
	}
}
