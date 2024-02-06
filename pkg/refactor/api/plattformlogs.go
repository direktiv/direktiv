package api

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
	"github.com/direktiv/direktiv/pkg/refactor/plattformlogs"
)

type logController struct {
	store plattformlogs.LogStore
}

func NewLogManager(store plattformlogs.LogStore) core.LogCollectionManager {
	return logController{
		store: store,
	}
}

func (m logController) Get(ctx context.Context, cursorTime time.Time, params map[string]string) ([]core.FeatureLogEntry, error) {
	var r []plattformlogs.LogEntry
	var err error

	// Determine the stream based on the provided parameters
	stream, err := determineStream(params)
	if err != nil {
		return []core.FeatureLogEntry{}, err
	}

	// Call the appropriate LogStore method with cursorTime
	if _, ok := params["instance-id"]; ok {
		r, err = m.store.GetInstanceLogs(ctx, stream, cursorTime)
	} else {
		r, err = m.store.Get(ctx, stream, cursorTime)
	}

	if err != nil {
		return []core.FeatureLogEntry{}, err
	}

	res := loglist{}
	for _, le := range r {
		e, err := le.ToFeatureLogEntry()
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
func (m logController) Stream(params map[string]string) http.HandlerFunc {
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
	if p, ok := params["instance-id"]; ok {
		return "flow.instance." + "%" + p + "%", nil
	} else if p, ok := params["namespace"]; ok {
		return "flow.namespace." + p, nil
	} else if p, ok := params["route"]; ok {
		return "flow.gateway." + p, nil
	}

	return "", fmt.Errorf("requested logs for an unknown type")
}

// LogStoreWorker manages the log polling and channel communication.
type logStoreWorker struct {
	Get      func(ctx context.Context, cursorTime time.Time, params map[string]string) ([]core.FeatureLogEntry, error)
	Interval time.Duration
	LogCh    chan string
	Params   map[string]string
	Cursor   time.Time // Cursor instead of Offset
}

// Start starts the log polling worker.
func (lw *logStoreWorker) start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(lw.Interval)
		defer ticker.Stop()
		defer close(lw.LogCh)
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
					slog.Info("data", "message", string(b))
					lw.LogCh <- string(b)
				}

				// Update cursorTime for the next iteration
				if len(logs) > 0 {
					lw.Cursor = logs[len(logs)-1].Time
				}
			}
		}
	}()
}
