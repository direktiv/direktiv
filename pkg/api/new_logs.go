package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/metastore"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type newLogsCtr struct {
	meta metastore.LogStore
}

func (c *newLogsCtr) mountRouter(r chi.Router) {
	r.Get("/", c.getLogs)
	r.Get("/subscribe", c.stream)
}

func (c *newLogsCtr) getLogs(w http.ResponseWriter, r *http.Request) {
	params := extractLogRequestParamsV2(r)

	data, starting, err := c.fetchOlderLogs(r.Context(), params)
	if err != nil {
		slog.Error("Error fetching logs.", "err", err)
		writeInternalError(w, err)

		return
	}

	metaInfo := map[string]any{
		"previousPage": nil,
		"startingFrom": nil,
	}

	if len(data) == 0 {
		writeJSONWithMeta(w, []logEntry{}, metaInfo)
		return
	}

	metaInfo["previousPage"] = data[0].Time.UTC().Format(time.RFC3339Nano)
	metaInfo["startingFrom"] = starting

	writeJSONWithMeta(w, data, metaInfo)
}

func (c *newLogsCtr) stream(w http.ResponseWriter, r *http.Request) {
	cursor := time.Now().UTC().Add(-3 * time.Second)
	params := extractLogRequestParamsV2(r)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	messageChannel := make(chan Event)

	var getCursoredStyle sseHandle = func(ctx context.Context, cursorTime time.Time) ([]CoursoredEvent, error) {
		logs, err := c.fetchNewerLogs(ctx, cursorTime, params)
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
				ID:   uuid.NewString(),
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

	worker := sseWorker{
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

func (c *newLogsCtr) fetchOlderLogs(ctx context.Context, params map[string]string) ([]logEntry, time.Time, error) {
	before := time.Now().UTC()
	if t, ok := params["before"]; ok {
		parsedTime, err := time.Parse(time.RFC3339Nano, t)
		if err != nil {
			return nil, time.Time{}, err
		}
		before = parsedTime
	}
	logs, err := c.meta.Get(ctx, metastore.LogQueryOptions{
		EndTime:  before,
		Limit:    10000,
		Metadata: extractLogQueryParamsV2(params),
	})
	if err != nil {
		return nil, time.Time{}, err
	}

	var results []logEntry
	for _, log := range logs {
		results = append(results, convertLogEntry(log))
	}

	return results, before, nil
}

func (c *newLogsCtr) fetchNewerLogs(ctx context.Context, cursorTime time.Time, params map[string]string) ([]logEntry, error) {
	options := metastore.LogQueryOptions{
		StartTime: cursorTime,
		Limit:     10000,
		Metadata:  extractLogQueryParamsV2(params),
	}
	// TODO opensearch does not have a sequential id lastID, hasLastID := params["lastID"]
	logs, err := c.meta.Get(ctx, options)
	if err != nil {
		return nil, err
	}

	var results []logEntry
	for _, log := range logs {
		results = append(results, convertLogEntry(log))
	}

	return results, nil
}

func convertLogEntry(e metastore.LogEntry) logEntry {
	entry := logEntry{
		ID:        0,
		Time:      e.Time,
		Msg:       e.Msg,
		Level:     e.Level,
		Error:     e.Error,
		Trace:     e.Trace,
		Span:      e.Span,
		Namespace: e.Namespace,
	}

	if e.Workflow != "" || e.Instance != "" {
		entry.Workflow = &WorkflowEntryContext{
			State:    e.State,
			Path:     e.Workflow,
			Instance: e.Instance,
			CalledAs: e.CalledAs,
			Status:   e.Status,
			Branch:   e.Branch,
		}
	}

	if e.Activity != "" {
		entry.Activity = &ActivityEntryContext{ID: e.Activity}
	}

	if e.Route != "" {
		entry.Route = &RouteEntryContext{Path: e.Route}
	}

	return entry
}

func extractLogQueryParamsV2(params map[string]string) map[string]string {
	allowedKeys := map[string]bool{
		"namespace": true, "instance": true, "route": true, "activity": true,
		"trace": true, "span": true, "branch": true, "level": true, "state": true,
		"status": true,
	}

	filteredParams := make(map[string]string)
	for key, value := range params {
		if _, ok := allowedKeys[key]; ok && len(value) > 0 {
			filteredParams[key] = value
		}
	}

	return filteredParams
}

// nolint:canonicalheader
func extractLogRequestParamsV2(r *http.Request) map[string]string {
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
	if v := r.URL.Query().Get("workflow"); v != "" {
		params["workflow"] = v
	}
	if v := r.URL.Query().Get("state"); v != "" {
		params["state"] = v
	}
	if v := r.URL.Query().Get("status"); v != "" {
		params["status"] = v
	}
	if v := r.URL.Query().Get("error"); v != "" {
		params["error"] = v
	}

	return params
}
