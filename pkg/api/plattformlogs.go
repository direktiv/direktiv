package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// nolint:canonicalheader
func extractLogRequestParams(r *http.Request) map[string]string {
	params := map[string]string{}
	if v := r.Header.Get("Last-Event-ID"); v != "" {
		params["lastID"] = v
	}
	if v := chi.URLParam(r, "namespace"); v != "" {
		params["namespace"] = v
	}
	if v := r.URL.Query().Get("route"); v != "" {
		params["route"] = v
	}
	if v := r.URL.Query().Get("instance"); v != "" {
		params["instance"] = v
	}
	if v := r.URL.Query().Get("branch"); v != "" {
		params["branch"] = v
	}
	if v := r.URL.Query().Get("level"); v != "" {
		params["level"] = v
	}
	if v := r.URL.Query().Get("before"); v != "" {
		params["before"] = v
	}
	if v := r.URL.Query().Get("after"); v != "" {
		params["after"] = v
	}
	if v := r.URL.Query().Get("trace"); v != "" {
		params["trace"] = v
	}
	if v := r.URL.Query().Get("span"); v != "" {
		params["span"] = v
	}
	if v := r.URL.Query().Get("activity"); v != "" {
		params["activity"] = v
	}

	return params
}

type logEntry struct {
	ID        int64                 `json:"id"`
	Time      time.Time             `json:"time"`
	Msg       interface{}           `json:"msg"`
	Level     interface{}           `json:"level"`
	Namespace interface{}           `json:"namespace"`
	Trace     interface{}           `json:"trace"`
	Span      interface{}           `json:"span"`
	Workflow  *WorkflowEntryContext `json:"workflow,omitempty"`
	Activity  *ActivityEntryContext `json:"activity,omitempty"`
	Route     *RouteEntryContext    `json:"route,omitempty"`
	Error     interface{}           `json:"error"`
}

type WorkflowEntryContext struct {
	Status interface{} `json:"status"`

	State    interface{} `json:"state"`
	Branch   interface{} `json:"branch"`
	Path     interface{} `json:"path"`
	Workflow interface{} `json:"workflow"`
	CalledAs interface{} `json:"calledAs"`
	Instance interface{} `json:"instance"`
}

type ActivityEntryContext struct {
	ID interface{} `json:"id,omitempty"`
}
type RouteEntryContext struct {
	Path interface{} `json:"path,omitempty"`
}
