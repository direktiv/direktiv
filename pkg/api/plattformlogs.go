package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/tracing"
	"github.com/go-chi/chi/v5"
)

type logController struct {
	store datastore.LogStore
}

func (m *logController) mountRouter(r chi.Router) {
	r.Get("/subscribe", m.stream)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		params := extractLogRequestParams(r)

		// Call the Get method with the cursor instead of offset
		data, starting, err := m.getOlder(r.Context(), params)
		if err != nil {
			slog.Error("Fetching logs for request.", "err", err)
			writeInternalError(w, err)

			return
		}
		metaInfo := map[string]any{
			"previousPage": nil, // setting them to nil make ensure matching the specicied types for the clients
			"startingFrom": nil,
		}
		if len(data) == 0 {
			writeJSONWithMeta(w, []logEntry{}, metaInfo)

			return
		}

		slices.Reverse(data)
		var previousPage interface{} = data[0].Time.UTC().Format(time.RFC3339Nano)

		metaInfo = map[string]any{
			"previousPage": previousPage,
			"startingFrom": starting,
		}
		writeJSONWithMeta(w, data, metaInfo)
	})
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		namespace := extractContextNamespace(r)
		instanceID := r.URL.Query().Get("instance")

		if instanceID == "" {
			http.Error(w, "Missing instance ID", http.StatusBadRequest)

			return
		}

		var logEntry map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&logEntry)
		if err != nil {
			writeInternalError(w, err)

			return
		}

		if _, ok := logEntry[string(core.LogTrackKey)]; !ok {
			writeBadrequestError(w, fmt.Errorf("missing 'track' field"))

			return
		}

		if v, ok := logEntry["namespace"].(string); !ok || v != namespace.Name {
			writeBadrequestError(w, fmt.Errorf("invalid or mismatched namespace"))

			return
		}

		msg, ok := logEntry["msg"].(string)
		if !ok {
			writeBadrequestError(w, fmt.Errorf("missing or invalid 'msg' field"))

			return
		}

		slogF := slog.Info
		if v, ok := logEntry["level"].(tracing.LogLevel); ok {
			switch v {
			case tracing.LevelDebug:
				slogF = slog.Debug
			case tracing.LevelInfo:
				slogF = slog.Info
			case tracing.LevelWarn:
				slogF = slog.Warn
			case tracing.LevelError:
				slogF = slog.Error
			}
		}

		delete(logEntry, "level")

		attr := make([]interface{}, 0, len(logEntry))
		for k, v := range logEntry {
			attr = append(attr, k, v)
		}

		slogF(msg, attr...)
		w.WriteHeader(http.StatusOK)
	})
}

func (m *logController) getNewer(ctx context.Context, t time.Time, params map[string]string) ([]logEntry, error) {
	var logs []core.LogEntry
	var err error

	// Determine the track based on the provided parameters
	stream, err := determineTrack(params)
	if err != nil {
		return []logEntry{}, err
	}

	// Call the appropriate LogStore method with cursorTime
	lastID, hasLastID := params["lastID"]
	_, isInstanceRequest := params["instance"]
	if hasLastID && isInstanceRequest {
		id, err := strconv.Atoi(lastID)
		if err != nil {
			return []logEntry{}, err
		}
		r, err := m.store.GetStartingIDUntilTimeInstance(ctx, stream, id, t)
		if err != nil {
			return []logEntry{}, err
		}
		logs = append(logs, r...)
	}
	if hasLastID && !isInstanceRequest {
		id, err := strconv.Atoi(lastID)
		if err != nil {
			return []logEntry{}, err
		}
		r, err := m.store.GetStartingIDUntilTime(ctx, stream, id, t)
		if err != nil {
			return []logEntry{}, err
		}
		logs = append(logs, r...)
	}

	if _, ok := params["instance"]; ok {
		r, err := m.store.GetNewerInstance(ctx, stream, t)
		if err != nil {
			return []logEntry{}, err
		}
		logs = append(logs, r...)
	} else {
		r, err := m.store.GetNewer(ctx, stream, t)
		if err != nil {
			return []logEntry{}, err
		}
		logs = append(logs, r...)
	}

	res := []logEntry{}
	for _, le := range logs {
		e := toFeatureLogEntry(le)
		res = append(res, e)
	}

	return res, nil
}

func (m *logController) getOlder(ctx context.Context, params map[string]string) ([]logEntry, time.Time, error) {
	var r []core.LogEntry
	var err error
	// Determine the track based on the provided parameters
	stream, err := determineTrack(params)
	if err != nil {
		return []logEntry{}, time.Time{}, err
	}

	starting := time.Now().UTC()
	if t, ok := params["before"]; ok {
		co, err := time.Parse(time.RFC3339Nano, t)
		if err != nil {
			return []logEntry{}, time.Time{}, err
		}
		starting = co
	}
	if _, ok := params["instance"]; ok {
		r, err = m.store.GetOlderInstance(ctx, stream, starting)
	} else {
		r, err = m.store.GetOlder(ctx, stream, starting)
	}
	if err != nil {
		return []logEntry{}, time.Time{}, err
	}
	res := []logEntry{}
	for _, le := range r {
		e := toFeatureLogEntry(le)
		res = append(res, e)
	}

	return res, starting, nil
}

// stream handles log streaming requests using Server-Sent Events (SSE).
// Clients subscribing to this endpoint will receive real-time log updates.
func (m *logController) stream(w http.ResponseWriter, r *http.Request) {
	// cursor is set to multiple seconds before the current time to mitigate data loss
	// that may occur due to delays between submitting and processing the request, or when a sequence of client requests is necessary.
	cursor := time.Now().UTC().Add(-time.Second * 3)

	// TODO: we may need to replace with a SSE-Server library instead of using our custom implementation.
	params := extractLogRequestParams(r)

	// Set the appropriate headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Create a channel to send SSE messages
	messageChannel := make(chan Event)

	var getCursoredStyle sseHandle = func(ctx context.Context, cursorTime time.Time) ([]CoursoredEvent, error) {
		logs, err := m.getNewer(ctx, cursorTime, params)
		if err != nil {
			return nil, err
		}
		res := make([]CoursoredEvent, 0, len(logs))
		for _, fle := range logs {
			b, err := json.Marshal(fle)
			if err != nil {
				return nil, err
			}
			dst := &bytes.Buffer{}
			if err := json.Compact(dst, b); err != nil {
				return nil, err
			}

			e := Event{
				ID:   strconv.Itoa(fle.ID),
				Data: dst.String(),
				Type: "message",
			}
			res = append(res, CoursoredEvent{
				Event: e,
				Time:  fle.Time,
			})
		}

		return res, nil
	}

	worker := seeWorker{
		Get:      getCursoredStyle,
		Interval: time.Second,
		Ch:       messageChannel,
		Cursor:   cursor,
	}
	go worker.start(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case message := <-messageChannel:
			_, err := io.Copy(w, strings.NewReader(fmt.Sprintf("id: %v\nevent: %v\ndata: %v\n\n", message.ID, message.Type, message.Data)))
			if err != nil {
				slog.Error("serve to SSE", "err", err)
			}

			f, ok := w.(http.Flusher)
			if !ok {
				// TODO Handle case where response writer is not a http.Flusher
				return
			}
			if f != nil {
				f.Flush()
			}
		}
	}
}

// determineTrack determines the log track based on provided parameters.
// It constructs a track string used for filtering logs in datastore queries.
func determineTrack(params map[string]string) (string, error) {
	if p, ok := params["instance"]; ok {
		return "flow.instance." + "%" + p + "%", nil
	} else if p, ok := params["route"]; ok {
		return "flow.route." + params["namespace"] + "." + p, nil
	} else if p, ok := params["activity"]; ok {
		return "flow.activity." + p, nil
	} else if p, ok := params["namespace"]; ok {
		return "flow.namespace." + p, nil
	} else if p, ok := params["trace"]; ok {
		return "flow.trace" + p, nil
	}

	return "", fmt.Errorf("requested logs for an unknown type")
}

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
	ID        int                   `json:"id"`
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
	Path     interface{} `json:"workflow"`
	CalledAs interface{} `json:"calledAs"`
	Instance interface{} `json:"instance"`
}

type ActivityEntryContext struct {
	ID interface{} `json:"id,omitempty"`
}
type RouteEntryContext struct {
	Path interface{} `json:"path,omitempty"`
}

func toFeatureLogEntry(e core.LogEntry) logEntry {
	featureLogEntry := logEntry{
		ID:    e.ID,
		Time:  e.Time,
		Msg:   e.Data["msg"],
		Level: e.Data["level"],
	}
	featureLogEntry.Error = e.Data["error"]
	featureLogEntry.Trace = e.Data["trace"]
	featureLogEntry.Span = e.Data["span"]
	featureLogEntry.Namespace = e.Data["namespace"]
	wfLogCtx := WorkflowEntryContext{}
	wfLogCtx.State = e.Data["state"]
	wfLogCtx.Path = e.Data["workflow"]
	wfLogCtx.Instance = e.Data["instance"]
	wfLogCtx.CalledAs = e.Data["calledAs"]
	wfLogCtx.Status = e.Data["status"]
	wfLogCtx.Branch = e.Data["branch"]
	featureLogEntry.Workflow = &wfLogCtx
	if wfLogCtx.Path == nil && wfLogCtx.Instance == nil {
		featureLogEntry.Workflow = nil
	}
	if id, ok := e.Data["activity"]; ok && id != nil {
		featureLogEntry.Activity = &ActivityEntryContext{ID: id}
	}
	if path, ok := e.Data["route"]; ok && path != nil {
		featureLogEntry.Route = &RouteEntryContext{Path: path}
	}

	return featureLogEntry
}
