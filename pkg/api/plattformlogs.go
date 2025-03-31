package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/telemetry"
	"github.com/go-chi/chi/v5"
)

type logController struct {
	logsBackend string
}

type logParams struct {
	namespace     string
	callpath      string
	scope         string
	id            string
	limit         string
	direction     string
	after, before string
	last, first   string
}

func (l logParams) toQuery() string {
	queryParts := make([]string, 0)

	limiter := ""
	if l.limit != "" {
		limiter = fmt.Sprintf("limit %s", l.limit)
	} else if l.first != "" {
		limiter = fmt.Sprintf("first %s by (_time)", l.first)
	} else if l.last != "" {
		limiter = fmt.Sprintf("last %s by (_time)", l.last)
	}

	if limiter != "" {
		queryParts = append(queryParts, limiter)
	}

	timeSelector := ""
	if l.after != "" {
		timeSelector = fmt.Sprintf(" _time:>%s ", l.after)
	} else if l.before != "" {
		timeSelector = fmt.Sprintf(" _time:<%s ", l.before)
	}

	idQuery := fmt.Sprintf("id:=%s", l.id)
	if l.callpath != "" {
		idQuery = fmt.Sprintf("callpath:/%s/*", l.id)
	}

	pipe := ""
	if len(queryParts) > 0 {
		pipe = fmt.Sprintf("| %s", strings.Join(queryParts, " | "))
	}

	// if there is a namespace (should), add it to the query
	// important for route requests
	nsQuery := ""
	if l.namespace != "" {
		nsQuery = fmt.Sprintf("namespace:=%s", l.namespace)
	}

	return fmt.Sprintf("query=scope:=%s %s %s%s %s",
		l.scope, nsQuery, idQuery, timeSelector, pipe)
}

func (m *logController) mountRouter(r chi.Router) {
	r.Get("/subscribe", m.stream)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		params := extractLogRequestParams(r)
		logs, err := m.get(r.Context(), params.toQuery())
		if err != nil {
			slog.Error("fetching logs for request", "err", err)
			writeInternalError(w, err)

			return
		}

		// sort and last/first does not work well with victoria logs.
		// we sort manually.
		if params.direction == "" || params.direction == "asc" {
			sort.Slice(logs, func(i, j int) bool {
				return logs[i].Time.Before(logs[j].Time)
			})
		} else {
			sort.Slice(logs, func(i, j int) bool {
				return logs[i].Time.After(logs[j].Time)
			})
		}

		writeJSON(w, logs)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		instanceID := r.URL.Query().Get("instance")

		if instanceID == "" {
			http.Error(w, "missing instance id", http.StatusBadRequest)

			return
		}

		// var logEntry map[string]interface{}
		var logObject telemetry.HTTPInstanceInfo
		err := json.NewDecoder(r.Body).Decode(&logObject)
		if err != nil {
			writeInternalError(w, err)

			return
		}

		ctx := telemetry.LogInitCtx(r.Context(), logObject.LogObject)

		switch logObject.Level {
		case telemetry.LogLevelDebug:
			telemetry.LogInstance(ctx, telemetry.LogLevelDebug, logObject.Msg)
		case telemetry.LogLevelInfo:
			telemetry.LogInstance(ctx, telemetry.LogLevelInfo, logObject.Msg)
		case telemetry.LogLevelWarn:
			telemetry.LogInstance(ctx, telemetry.LogLevelWarn, logObject.Msg)
		case telemetry.LogLevelError:
			telemetry.LogInstanceError(ctx, logObject.Msg, fmt.Errorf("%s", logObject.Msg))
		}

		w.WriteHeader(http.StatusOK)
	})
}

func (m *logController) get(ctx context.Context, query string) ([]logEntry, error) {
	var err error

	logs, err := m.fetchFromBackend(ctx, query)
	if err != nil {
		return []logEntry{}, err
	}

	res := []logEntry{}
	for i := range logs {
		e := toFeatureLogEntry(logs[i])
		res = append(res, e)
	}

	return res, nil
}

// stream handles log streaming requests using Server-Sent Events (SSE).
// Clients subscribing to this endpoint will receive real-time log updates.
func (m *logController) stream(w http.ResponseWriter, r *http.Request) {
	cursor := time.Now().UTC()

	params := extractLogRequestParams(r)

	// if nothing is set, we do the events from now
	if params.after == "" {
		params.after = time.Now().UTC().Format("2006-01-02T15:04:05.000000000Z")
	}

	rc := http.NewResponseController(w)

	queryAndSend := func(ctx context.Context, queryString string) (time.Time, error) {
		logs, err := m.get(ctx, queryString)
		if err != nil {
			return time.Now().UTC(), err
		}

		for i := range logs {
			l := logs[i]
			b, _ := json.Marshal(l)

			_, err = fmt.Fprintf(w, "id: %s\nevent: %s\ndata: %s\n\n", l.Time.UTC().Format("2006-01-02T15:04:05.000000000Z"), "message", string(b))
			if err != nil {
				return time.Now().UTC(), err
			}
		}

		// check time of last log
		if len(logs) > 0 {
			lastLog := logs[len(logs)-1]
			cursor = lastLog.Time.UTC()
		} else {
			cursor, err = parseQueryTime(params.after)
			// can not do much about it, use `now`
			if err != nil {
				slog.Error("can not parse params.after", slog.Any("error", err))
				cursor = time.Now().UTC()
			}
		}

		return cursor, rc.Flush()
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")
	rc.Flush()

	// Last-Event-ID header
	lastEventID := r.Header.Get("Last-Event-ID")
	if lastEventID != "" {
		params.after = lastEventID
	}

	// send initial data
	var err error
	cursor, err = queryAndSend(r.Context(), params.toQuery())
	if err != nil {
		slog.Error("error querying data", slog.Any("error", err))
		http.Error(w, "error querying data", http.StatusInternalServerError)

		return
	}
	params.after = cursor.Format("2006-01-02T15:04:05.000000000Z")

	clientGone := r.Context().Done()

	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case <-clientGone:
			// disconnected
			slog.Debug("client disconnected")
			return
		case <-t.C:
			var err error
			params.last = ""
			cursor, err = queryAndSend(r.Context(), params.toQuery())
			if err != nil {
				slog.Error("error querying data", slog.Any("error", err))
				http.Error(w, "error querying data", http.StatusInternalServerError)

				return
			}
			params.after = cursor.Format("2006-01-02T15:04:05.000000000Z")
		}
	}
}

var formats = []string{"2006-01-02T15:04:05.000000000Z", "2006-01-02T15:04:05.000Z"}

func parseQueryTime(input string) (time.Time, error) {
	for _, format := range formats {
		t, err := time.Parse(format, input)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, errors.New("unrecognized time format")
}

// nolint:canonicalheader
func extractLogRequestParams(r *http.Request) logParams {
	var logParams logParams

	if v := chi.URLParam(r, "namespace"); v != "" {
		logParams.namespace = v

		// set scope to namespace first, we can change it later
		logParams.scope = "namespace"
		logParams.id = v
	}

	// fetch all possible query params
	logParams.limit = r.URL.Query().Get("limit")
	logParams.direction = r.URL.Query().Get("direction")
	logParams.after = r.URL.Query().Get("after")
	logParams.before = r.URL.Query().Get("before")
	logParams.last = r.URL.Query().Get("last")
	logParams.first = r.URL.Query().Get("first")

	// use last event id if send on reconnect by a client
	if r.URL.Query().Get("lastEventId") != "" {
		logParams.after = r.URL.Query().Get("lastEventId")
	}

	if r.URL.Query().Get("instance") != "" {
		logParams.scope = "instance"
		logParams.id = r.URL.Query().Get("instance")
		logParams.callpath = r.URL.Query().Get("instance")
	}

	if r.URL.Query().Get("activity") != "" {
		logParams.scope = "activity"
		logParams.id = r.URL.Query().Get("activity")
	}

	if r.URL.Query().Get("route") != "" {
		logParams.scope = "route"
		logParams.id = r.URL.Query().Get("route")
	}

	return logParams
}

type logEntry struct {
	Time      time.Time             `json:"time"`
	Msg       interface{}           `json:"msg"`
	Level     interface{}           `json:"level"`
	Namespace interface{}           `json:"namespace"`
	Workflow  *WorkflowEntryContext `json:"workflow,omitempty"`
	Activity  *ActivityEntryContext `json:"activity,omitempty"`
	Route     *RouteEntryContext    `json:"route,omitempty"`
	Error     string                `json:"error,omitempty"`
}

type WorkflowEntryContext struct {
	Status   string `json:"status"`
	State    string `json:"state"`
	Path     string `json:"workflow"`
	Instance string `json:"instance"`
	CallPath string `json:"callpath"`
}

type ActivityEntryContext struct {
	ID interface{} `json:"id,omitempty"`
}
type RouteEntryContext struct {
	Path interface{} `json:"path,omitempty"`
}

func toFeatureLogEntry(e logEntryBackend) logEntry {
	featureLogEntry := logEntry{
		Time:      e.Time.UTC(),
		Msg:       e.Msg,
		Level:     e.Level,
		Namespace: e.Namespace,
		Error:     e.Error,
	}

	// workflow data if instance
	if e.Scope == string(telemetry.LogScopeInstance) {
		featureLogEntry.Workflow = &WorkflowEntryContext{
			Status:   e.Status,
			Path:     e.Path,
			State:    e.State,
			Instance: e.ID,
			CallPath: e.CallPath,
		}
	}

	if strings.HasPrefix(e.Scope, "activity.") {
		featureLogEntry.Activity = &ActivityEntryContext{
			ID: e.ID,
		}
	}

	if strings.HasPrefix(e.Scope, "route.") {
		featureLogEntry.Route = &RouteEntryContext{}
	}

	return featureLogEntry
}

type logEntryBackend struct {
	Time        time.Time `json:"_time"`
	ID          string    `json:"id"`
	StreamID    string    `json:"_stream_id"`
	Stream      string    `json:"_stream"`
	Msg         string    `json:"_msg"`
	P           string    `json:"_p"`
	Invoker     string    `json:"invoker"`
	CallPath    string    `json:"callpath"`
	Level       string    `json:"level"`
	Namespace   string    `json:"namespace"`
	Status      string    `json:"status"`
	StreamValue string    `json:"stream"`
	Date        time.Time `json:"date"`
	Path        string    `json:"path"`
	Scope       string    `json:"scope"`
	State       string    `json:"state"`
	Error       string    `json:"error"`
}

func (m *logController) fetchFromBackend(ctx context.Context, query string) ([]logEntryBackend, error) {
	var ret []logEntryBackend

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("http://%s/select/logsql/query", net.JoinHostPort(m.logsBackend, "9428")),
		bytes.NewBufferString(query))
	if err != nil {
		return ret, err
	}

	// set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	cli := &http.Client{Timeout: 30 * time.Second}

	resp, err := cli.Do(req)
	if err != nil {
		return ret, err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	for {
		var log logEntryBackend
		err := dec.Decode(&log)
		if err == io.EOF {
			break
		}
		if err != nil {
			return ret, err
		}
		ret = append(ret, log)
	}

	return ret, err
}
