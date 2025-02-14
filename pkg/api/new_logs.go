package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/engine"
	"github.com/direktiv/direktiv/pkg/metastore"
	"github.com/direktiv/direktiv/pkg/tracing"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type newLogsCtr struct {
	meta metastore.LogStore
	db   *database.DB // TODO remove once UI is patched
}

func (c *newLogsCtr) mountRouter(r chi.Router) {
	r.Get("/", c.get)
	r.Get("/subscribe", c.subscribe)
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

func (c *newLogsCtr) get(w http.ResponseWriter, r *http.Request) {
	params := extractLogRequestParams(r)
	options, err := c.getOptions(r.Context(), params)
	if err != nil {
		writeBadrequestError(w, err)
		return
	}
	logs, err := c.meta.Get(r.Context(), *options)
	if err != nil {
		writeDataStoreError(w, err)
		return
	}
	res := []logEntry{}
	for _, log := range logs {
		entry := logEntry{
			ID:        log.Time.UnixNano(),
			Time:      log.Time,
			Msg:       log.Msg,
			Level:     log.Level,
			Namespace: log.Namespace,
			Trace:     log.Trace,
			Span:      log.Span,
			Error:     log.Error,
		}
		if len(log.Workflow) != 0 && len(log.State) != 0 {
			entry.Workflow = &WorkflowEntryContext{
				Status:   log.Status,
				State:    log.State,
				Path:     log.Workflow,
				Workflow: log.Workflow,
				CalledAs: log.CalledAs,
				Instance: log.Instance,
			}
		}
		if len(log.Activity) > 0 {
			entry.Activity = &ActivityEntryContext{
				ID: log.Activity,
			}
		}
		if len(log.Route) > 0 {
			entry.Route = &RouteEntryContext{
				Path: log.Route,
			}
		}
		res = append(res, entry)
	}
	if len(res) == 0 {
		metaInfo := map[string]any{
			"previousPage": nil,
			"startingFrom": nil,
		}
		writeJSONWithMeta(w, res, metaInfo)

		return
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})
	var previousPage interface{} = res[0].Time.UTC().Format(time.RFC3339Nano)

	metaInfo := map[string]any{
		"previousPage": previousPage,
		"startingFrom": res[len(res)-1].Time.UTC().Format(time.RFC3339Nano),
	}

	writeJSONWithMeta(w, res, metaInfo)
}

func (c *newLogsCtr) getOptions(ctx context.Context, params map[string]string) (*metastore.LogQueryOptions, error) {
	options := metastore.LogQueryOptions{
		Metadata: make(map[string]string),
	}
	if v, ok := params["before"]; ok && len(v) > 0 {
		co, err := time.Parse(time.RFC3339Nano, v)
		if err != nil {
			return nil, err
		}
		options.EndTime = &co
	}
	if v, ok := params["after"]; ok && len(v) > 0 {
		co, err := time.Parse(time.RFC3339Nano, v)
		if err != nil {
			return nil, err
		}
		options.StartTime = &co
	}
	if v, ok := params["lastID"]; ok && len(v) > 0 {
		uTime, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		co := time.Unix(int64(uTime), 0)
		options.StartTime = &co
	}
	if v, ok := params["instance"]; ok && len(v) > 0 {
		id, err := uuid.Parse(v)
		if err != nil {
			return nil, err
		}
		data, err := c.db.InstanceStore().ForInstanceID(id).GetSummary(ctx)
		if err != nil {
			return nil, err
		}
		x, err := engine.ParseInstanceData(data)
		if err != nil {
			return nil, err
		}
		traceID, err := tracing.TraceParentToTraceID(ctx, x.TelemetryInfo.TraceParent)
		if err != nil {
			return nil, err
		}
		options.Metadata["trace"] = traceID
		// TODO switch to spanID?
		if options.StartTime == nil {
			options.StartTime = &x.Instance.CreatedAt
		}
		if options.StartTime.Before(x.Instance.CreatedAt) {
			options.StartTime = &x.Instance.CreatedAt
		}

		params["instance"] = ""
	}
	if v, ok := params["activity"]; ok && len(v) > 0 {
		params["namespace"] = "" // TODO remove when mirror logger is fixed
	}
	filterParams := []string{
		"route",
		"trace",
		"span",
		"activity",
		"level",
		"branch",
		"namespace",
	}
	for _, k := range filterParams {
		if v, ok := params[k]; ok && len(v) > 0 {
			options.Metadata[k] = v
		}
	}

	return &options, nil
}

func (c *newLogsCtr) subscribe(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters for filtering logs
	params := extractLogRequestParams(r)
	options, err := c.getOptions(r.Context(), params)
	if err != nil {
		writeBadrequestError(w, err)
		return
	}

	// Set up response headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create a cancellation context
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Create a channel to receive log entries
	messageChannel := make(chan metastore.LogEntry)

	// Goroutine for writing log messages to the HTTP response
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case log := <-messageChannel:
				entry := logEntry{
					ID:        log.Time.UnixNano(),
					Time:      log.Time,
					Msg:       log.Msg,
					Level:     log.Level,
					Namespace: log.Namespace,
					Trace:     log.Trace,
					Span:      log.Span,
					Error:     log.Error,
				}
				if len(log.Workflow) != 0 && len(log.State) != 0 {
					entry.Workflow = &WorkflowEntryContext{
						Status:   log.Status,
						State:    log.State,
						Path:     log.Workflow,
						Workflow: log.Workflow,
						CalledAs: log.CalledAs,
						Instance: log.Instance,
					}
				}
				if len(log.Activity) > 0 {
					entry.Activity = &ActivityEntryContext{
						ID: fmt.Sprint(log.Time.UnixNano()),
					}
				}
				if len(log.Route) > 0 {
					entry.Route = &RouteEntryContext{
						Path: log.Route,
					}
				}

				b, err := json.Marshal(entry)
				if err != nil {
					slog.Error("serve to SSE", "err", err)
					return
				}

				dst := &bytes.Buffer{}
				if err := json.Compact(dst, b); err != nil {
					slog.Error("serve to SSE", "err", err)
					return
				}

				// Write the log entry to the response
				logBytes := dst.Bytes()
				_, err = io.Copy(w, strings.NewReader(fmt.Sprintf("id: %v\nevent: %v\ndata: %v\n\n", log.Time.UnixNano(), "message", string(logBytes))))
				if err != nil {
					slog.Error("serve to SSE", "err", err)
					return
				}

				// Flush the response if the connection supports it
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			}
		}
	}()

	// Stream logs from the database
	err = c.meta.Stream(r.Context(), *options, messageChannel)
	if err != nil {
		slog.Error("failed to stream", "error", err)
		return
	}
}
