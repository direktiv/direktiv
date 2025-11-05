package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/go-chi/chi/v5"
)

type serviceController struct {
	manager core.ServiceManager
}

func (e *serviceController) mountRouter(r chi.Router) {
	r.Get("/", e.all)
	r.Get("/{serviceID}/pods", e.pods)
	r.Get("/{serviceID}/pods/{podID}/logs", e.logs)
	r.Post("/{serviceID}/actions/rebuild", e.rebuild)
}

func (e *serviceController) all(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")

	list, err := e.manager.GeAll(namespace)
	if err != nil {
		writeInternalError(w, err)

		return
	}

	writeJSON(w, list)
}

func (e *serviceController) pods(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	serviceID := chi.URLParam(r, "serviceID")

	svc, err := e.manager.GetPods(namespace, serviceID)
	if errors.Is(err, core.ErrNotFound) {
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

	writeJSON(w, svc)
}

func (e *serviceController) rebuild(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	serviceID := chi.URLParam(r, "serviceID")

	err := e.manager.Rebuild(namespace, serviceID)
	if errors.Is(err, core.ErrNotFound) {
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

	writeOk(w)
}

func (e *serviceController) logs(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	serviceID := chi.URLParam(r, "serviceID")
	podID := chi.URLParam(r, "podID")

	readCloser, err := e.manager.StreamLogs(namespace, serviceID, podID)
	if errors.Is(err, core.ErrNotFound) {
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

	buffer := make([]byte, 1024)
	var n int
	for {
		// TODO: this would leak because read() could block forever.
		n, err = readCloser.Read(buffer)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			writeInternalError(w, err)

			break
		}
		_, _ = fmt.Fprintf(w, "%s", buffer[:n])

		//nolint:forcetypeassert
		w.(http.Flusher).Flush()
		time.Sleep(10 * time.Millisecond)
	}
}
