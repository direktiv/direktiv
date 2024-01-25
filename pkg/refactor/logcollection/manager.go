package logcollection

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

type Manager struct {
	store LogStore
}

func NewManger(store LogStore) Manager {
	return Manager{
		store: store,
	}
}

func (m *Manager) Get(ctx context.Context, cursorTime time.Time, params map[string]string) ([]core.FeatureLogEntry, error) {
	var r []LogEntry
	var err error

	// Determine the stream based on the provided parameters
	stream, err := determineStream(params)
	if err != nil {
		return []core.FeatureLogEntry{}, err
	}

	// Call the appropriate LogStore method with cursorTime
	if p, ok := params["root-instance-id"]; ok {
		r, err = m.store.GetInstanceLogs(ctx, stream, p, cursorTime)
	} else {
		r, err = m.store.Get(ctx, stream, cursorTime)
	}

	if err != nil {
		return []core.FeatureLogEntry{}, err
	}

	res := loglist{}
	for _, le := range r {
		e, err := le.toFeatureLogEntry()
		if err != nil {
			return []core.FeatureLogEntry{}, err
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
func (m *Manager) Stream(params map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the appropriate headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Create a context with cancellation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		// Extract cursor from request URL query
		cursorTimeStr := r.URL.Query().Get("cursor")
		cursorTime, err := time.Parse(time.RFC3339Nano, cursorTimeStr)
		if err != nil {
			// Handle error
			http.Error(w, "Invalid cursor parameter", http.StatusBadRequest)

			return
		}

		// Create a channel to send SSE messages
		messageChannel := make(chan string)
		// Adjust the logStoreWorker to use cursor instead of offset
		worker := logStoreWorker{
			Get:      m.Get,
			Interval: time.Second,
			LogCh:    messageChannel,
			Params:   params,
			Cursor:   cursorTime,
		}
		go worker.start(ctx)

		for {
			select {
			case <-ctx.Done():
				slog.Info("context  done")

				return
			case message := <-messageChannel:
				slog.Info("data", "message", message)
				_, err := io.Copy(w, strings.NewReader(fmt.Sprintf("data: %s\n\n", message)))
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

func (e LogEntry) toFeatureLogEntry() (core.FeatureLogEntry, error) {
	entry, ok := e.Data["entry"].(string)
	if !ok {
		return core.FeatureLogEntry{}, fmt.Errorf("log-entry format is corrupt")
	}

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(entry), &m); err != nil {
		return core.FeatureLogEntry{}, fmt.Errorf("failed to unmarshal log entry: %w", err)
	}

	return core.FeatureLogEntry{
		Time:     e.Time,
		Msg:      fmt.Sprint(m["msg"]),
		Level:    fmt.Sprint(m["level"]),
		Trace:    fmt.Sprint(m["trace"]),
		State:    fmt.Sprint(m["state"]),
		Branch:   fmt.Sprint(m["branch"]),
		Metadata: map[string]string{},
	}, nil
}

type loglist []core.FeatureLogEntry

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
	if p, ok := params["root-instance-id"]; ok {
		return "flow.instance." + p, nil
	} else if p, ok := params["namespace"]; ok {
		return "flow.namespace." + p, nil
	} else if p, ok := params["route"]; ok {
		return "flow.gateway." + p, nil
	}

	return "", fmt.Errorf("requested logs for an unknown type")
}
