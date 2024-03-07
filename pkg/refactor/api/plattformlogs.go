package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/plattformlogs"
	"github.com/go-chi/chi/v5"
)

type logController struct {
	store plattformlogs.LogStore
}

func NewLogManager(store plattformlogs.LogStore) core.LogCollectionManager {
	return logController{
		store: store,
	}
}

func (m *logController) mountRouter(r chi.Router) {
	r.Get("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		params := extractLogRequestParams(r)
		m.Stream(params).ServeHTTP(w, r)
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		params := extractLogRequestParams(r)

		// Call the Get method with the cursor instead of offset
		data, err := m.GetOlder(r.Context(), params)
		if err != nil {
			slog.Error("Get logs", "error", err)
			writeInternalError(w, err)

			return
		}
		if len(data) == 0 {
			data = nil
			writeJSONWithNextPage(w, data, nil)

			return
		}
		previousPage := data[len(data)-1].Time.UTC().Format(time.RFC3339Nano)
		if len(data) < 200 {
			previousPage = ""
		}
		writeJSONWithNextPage(w, data, previousPage)
	})
}

func (m logController) GetNewer(ctx context.Context, t time.Time, params map[string]string) ([]core.PlattformLogEntry, error) {
	var logs []plattformlogs.LogEntry
	var err error

	// Determine the stream based on the provided parameters
	stream, err := determineStream(params)
	if err != nil {
		return []core.PlattformLogEntry{}, err
	}

	// Call the appropriate LogStore method with cursorTime
	lastID, hasLastID := params["lastID"]
	_, isInstanceRequest := params["instance"]
	if hasLastID && isInstanceRequest {
		id, err := strconv.Atoi(lastID)
		if err != nil {
			return []core.PlattformLogEntry{}, err
		}
		r, err := m.store.GetStartingIDUntilTimeInstance(ctx, stream, id, t)
		if err != nil {
			return []core.PlattformLogEntry{}, err
		}
		logs = append(logs, r...)
	}
	if hasLastID && !isInstanceRequest {
		id, err := strconv.Atoi(lastID)
		if err != nil {
			return []core.PlattformLogEntry{}, err
		}
		r, err := m.store.GetStartingIDUntilTime(ctx, stream, id, t)
		if err != nil {
			return []core.PlattformLogEntry{}, err
		}
		logs = append(logs, r...)
	}

	if _, ok := params["instance"]; ok {
		r, err := m.store.GetNewerInstance(ctx, stream, t)
		if err != nil {
			return []core.PlattformLogEntry{}, err
		}
		logs = append(logs, r...)
	} else {
		r, err := m.store.GetNewer(ctx, stream, t)
		if err != nil {
			return []core.PlattformLogEntry{}, err
		}
		logs = append(logs, r...)
	}

	res := loglist{}
	for _, le := range logs {
		e, err := plattformlogs.ToFeatureLogEntry(le)
		if err != nil {
			return []core.PlattformLogEntry{}, err
		}
		res = append(res, e)
	}

	// Apply filters based on additional parameters
	if p, ok := params["level"]; ok {
		res.filterByLevel(p)
	}
	if p, ok := params["branch"]; ok {
		res.filterByBranch(p)
	}
	if p, ok := params["state"]; ok {
		res.filterByState(p)
	}

	return res, nil
}

func (m logController) GetOlder(ctx context.Context, params map[string]string) ([]core.PlattformLogEntry, error) {
	var r []plattformlogs.LogEntry
	var err error

	// Determine the stream based on the provided parameters
	stream, err := determineStream(params)
	if err != nil {
		return []core.PlattformLogEntry{}, err
	}
	starting := time.Now().UTC()
	if t, ok := params["before"]; ok {
		co, err := time.Parse(time.RFC3339Nano, t)
		if err != nil {
			return []core.PlattformLogEntry{}, err
		}
		starting = co
	}
	if _, ok := params["instance"]; ok {
		r, err = m.store.GetOlderInstance(ctx, stream, starting)
	} else {
		r, err = m.store.GetOlder(ctx, stream, starting)
	}
	if err != nil {
		return []core.PlattformLogEntry{}, err
	}
	res := loglist{}
	for _, le := range r {
		e, err := plattformlogs.ToFeatureLogEntry(le)
		if err != nil {
			return []core.PlattformLogEntry{}, err
		}
		res = append(res, e)
	}

	// Apply filters based on additional parameters
	if p, ok := params["level"]; ok {
		res.filterByLevel(p)
	}
	if p, ok := params["branch"]; ok {
		res.filterByBranch(p)
	}
	if p, ok := params["state"]; ok {
		res.filterByState(p)
	}

	return res, nil
}

// Stream handles the SSE endpoint.
func (m logController) Stream(params map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the appropriate headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Create a context with cancellation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		// Create a channel to send SSE messages
		messageChannel := make(chan Event)
		// Adjust the logStoreWorker to use cursor instead of offset
		worker := logStoreWorker{
			Get:      m.GetNewer,
			Interval: time.Second,
			Ch:       messageChannel,
			Params:   params,
			Cursor:   time.Now().UTC(),
		}
		go worker.start(ctx)

		for {
			select {
			case <-ctx.Done():
				slog.Info("context done")

				return
			case message := <-messageChannel:
				slog.Info("data", "message", message)
				_, err := io.Copy(w, strings.NewReader(fmt.Sprintf("id: %v\nevent: %v\ndata: %v\n\n", message.ID, message.Type, message.Data)))
				if err != nil {
					slog.Error("copy", "error", err)
				}

				f, ok := w.(http.Flusher)
				if !ok {
					// TODO Handle case where response writer is not a http.Flusher
					slog.Error("Response writer is not a http.Flusher")

					return
				}
				if f != nil {
					f.Flush()
				}
			}
		}
	}
}

type loglist []core.PlattformLogEntry

func (e *loglist) filterByBranch(branch string) {
	// TODO revisit this implementation
	filteredEntries := make(loglist, 0)
	for _, entry := range *e {
		if entry.Branch == branch {
			filteredEntries = append(filteredEntries, entry)
		}
	}
	*e = filteredEntries
}

func (e *loglist) filterByState(state string) {
	// TODO revisit this implementation
	filteredEntries := make(loglist, 0)
	for _, entry := range *e {
		if entry.State == state {
			filteredEntries = append(filteredEntries, entry)
		}
	}
	*e = filteredEntries
}

func (e *loglist) filterByLevel(level string) {
	// TODO revisit this implementation
	filteredEntries := make(loglist, 0)
	for _, entry := range *e {
		if entry.Level == level {
			filteredEntries = append(filteredEntries, entry)
		}
	}
	*e = filteredEntries
}

func determineStream(params map[string]string) (string, error) {
	if p, ok := params["instance"]; ok {
		return "flow.instance." + "%" + p + "%", nil
	} else if p, ok := params["route"]; ok {
		return "flow.route." + p, nil
	} else if p, ok := params["activity"]; ok {
		return "flow.mirror." + p, nil
	} else if p, ok := params["namespace"]; ok {
		return "flow.namespace." + p, nil
	} else if p, ok := params["trace"]; ok {
		return "flow.trace" + p, nil
	}

	return "", fmt.Errorf("requested logs for an unknown type")
}

type Event struct {
	ID   string
	Data string
	Type string
}

// LogStoreWorker manages the log polling and channel communication.
type logStoreWorker struct {
	Get      func(ctx context.Context, cursorTime time.Time, params map[string]string) ([]core.PlattformLogEntry, error)
	Interval time.Duration
	Ch       chan Event
	Params   map[string]string
	Cursor   time.Time // Cursor instead of Offset
}

// Start starts the log polling worker.
func (lw *logStoreWorker) start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(lw.Interval)
		defer ticker.Stop()
		defer close(lw.Ch)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				slog.Info("data", "message", lw.Params)
				logs, err := lw.Get(ctx, lw.Cursor, lw.Params)
				if err != nil {
					slog.Error("TODO: should we quit with an error?", "error", err)

					continue
				}
				for _, fle := range logs {
					b, err := json.Marshal(fle)
					if err != nil {
						slog.Error("TODO: should we quit with an error?", "error", err)

						continue
					}
					dst := &bytes.Buffer{}
					if err := json.Compact(dst, b); err != nil {
						slog.Error("TODO: should we quit with an error?", "error", err)
					}

					slog.Info("data", "message", string(b))
					e := Event{
						ID:   fle.ID,
						Data: dst.String(),
						Type: "message",
					}
					lw.Ch <- e
				}

				// Update cursorTime for the next iteration
				if len(logs) > 0 {
					lw.Cursor = logs[len(logs)-1].Time
				}
			}
		}
	}()
}

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
