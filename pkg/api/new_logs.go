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

	"github.com/direktiv/direktiv/pkg/metastore"
	"github.com/go-chi/chi/v5"
)

type newLogsCtr struct {
	meta metastore.LogStore
}

func (c *newLogsCtr) mountRouter(r chi.Router) {
	r.Get("/", c.get)
	r.Get("/subscribe", c.subscribe)
}

func (c *newLogsCtr) get(w http.ResponseWriter, r *http.Request) {
	params := extractLogRequestParams(r)
	options, err := getOptions(params)
	if err != nil {
		writeBadrequestError(w, err)
	}
	logs, err := c.meta.Get(r.Context(), *options)
	if err != nil {
		writeDataStoreError(w, err)
	}
	res := []logEntry{}
	slog.Info("got logs", "logs", len(logs))
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
		if len(log.Workflow) != 0 {
			entry.Workflow = &WorkflowEntryContext{
				Status:   log.Status,
				State:    log.State,
				Branch:   log.Branch,
				Path:     log.Workflow,
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
	}
	var previousPage interface{} = res[0].Time.UTC().Format(time.RFC3339Nano)

	metaInfo := map[string]any{
		"previousPage": previousPage,
		"startingFrom": res[len(res)-1].Time.UTC().Format(time.RFC3339Nano),
	}

	writeJSONWithMeta(w, res, metaInfo)
}

func getOptions(params map[string]string) (*metastore.LogQueryOptions, error) {
	filterParams := []string{
		"instance",
		"route",
		"trace",
		"span",
		"activity",
		"level",
		"branch",
		"namespace",
	}
	options := metastore.LogQueryOptions{
		Metadata: make(map[string]string),
	}

	for _, k := range filterParams {
		if v, ok := params[k]; ok && len(v) > 0 {
			options.Metadata[k] = v
		}
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
	options.Limit = 4

	return &options, nil
}

func (c *newLogsCtr) subscribe(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters for filtering logs
	params := extractLogRequestParams(r)
	options, err := getOptions(params)
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
				if len(log.Workflow) != 0 {
					entry.Workflow = &WorkflowEntryContext{
						Status:   log.Status,
						State:    log.State,
						Branch:   log.Branch,
						Path:     log.Workflow,
						CalledAs: log.CalledAs,
						Instance: log.Instance,
					}
				}
				if len(log.Activity) > 0 {
					entry.Activity = &ActivityEntryContext{
						ID: 0,
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
		writeDataStoreError(w, err)
		return
	}
}
