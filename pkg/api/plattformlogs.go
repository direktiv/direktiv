package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
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

	if l.direction == "" || l.direction == "asc" {
		l.direction = "asc"
	} else {
		l.direction = "desc"
	}
	queryParts = append(queryParts, fmt.Sprintf("sort by (_time) %s", l.direction))

	timeSelector := ""
	if l.after != "" {
		timeSelector = fmt.Sprintf(" _time:>%s ", l.after)
	} else if l.before != "" {
		timeSelector = fmt.Sprintf(" _time:<%s ", l.before)
	}

	return fmt.Sprintf("query=scope:=%s id:=%s%s| %s",
		l.scope, l.id, timeSelector, strings.Join(queryParts, " | "))
}

func (m *logController) mountRouter(r chi.Router) {
	r.Get("/subscribe", m.stream)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		params := extractLogRequestParams(r)
		fmt.Println(params.toQuery())
		logs, err := m.get(params.toQuery())
		if err != nil {
			slog.Error("fetching logs for request", "err", err)
			writeInternalError(w, err)

			return
		}

		if len(logs) == 0 {
			writeJSONWithMeta(w, []logEntry{}, map[string]any{
				"previousPage": nil,
				"startingFrom": nil,
			})

			return
		}

		metaInfo := map[string]any{
			"previousPage": nil,
			"startingFrom": nil,
		}

		writeJSONWithMeta(w, logs, metaInfo)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		// namespace := extractContextNamespace(r)
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
			telemetry.LogInstanceError(ctx, logObject.Msg, fmt.Errorf(logObject.Msg))
		}

		w.WriteHeader(http.StatusOK)
	})
}

func (m *logController) get(query string) ([]logEntry, error) {
	var err error

	logs, err := m.fetchFromBackend(query)
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

	// TODO: we may need to replace with a SSE-Server library instead of using our custom implementation.
	params := extractLogRequestParams(r)

	// if nothing is set, we do the events from now
	if params.after == "" {
		params.after = time.Now().Format("2006-01-02T15:04:05.000000000Z")
	}

	rc := http.NewResponseController(w)

	queryAndSend := func(queryString string) (time.Time, error) {
		fmt.Println("QUERY AND SEND!")
		fmt.Println(queryString)
		logs, err := m.get(queryString)
		if err != nil {
			return time.Now(), err
		}

		for i := range logs {
			l := logs[i]
			b, _ := json.Marshal(l)
			_, err = fmt.Fprintf(w, fmt.Sprintf("id: %s\nevent: %s\ndata: %s\n\n", l.Time.Format("2006-01-02T15:04:05.000000000Z"), "message", string(b)))
			if err != nil {
				return time.Now(), err
			}
		}

		// check time of last log
		if len(logs) > 0 {
			lastLog := logs[len(logs)-1]
			cursor = lastLog.Time.UTC()
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
	cursor, err = queryAndSend(params.toQuery())
	if err != nil {
		writeError(w, &Error{
			Code:    "log request failed",
			Message: err.Error(),
		})

		return
	}
	params.after = cursor.Format("2006-01-02T15:04:05.000000000Z")
	// params.last = "100"

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
			cursor, err = queryAndSend(params.toQuery())
			if err != nil {
				writeError(w, &Error{
					Code:    "log request failed",
					Message: err.Error(),
				})
			}
			params.after = cursor.Format("2006-01-02T15:04:05.000000000Z")
		}
	}
}

// nolint:canonicalheader
func extractLogRequestParams(r *http.Request) logParams {
	// params := map[string]string{}

	var logParams logParams

	if v := chi.URLParam(r, "namespace"); v != "" {
		// params["namespace"] = v
		logParams.namespace = v

		// set track to namespace first, we can change it later
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
	}

	if r.URL.Query().Get("activity") != "" {
		logParams.scope = "activity"
		logParams.id = r.URL.Query().Get("activity")
	}

	// } else if p, ok := params["route"]; ok {
	// 	params["track"] = "route." + p

	return logParams
}

type logEntry struct {
	Time time.Time `json:"time"`
	// ID        string                `json:"id"`
	Msg       interface{}           `json:"msg"`
	Level     interface{}           `json:"level"`
	Namespace interface{}           `json:"namespace"`
	Workflow  *WorkflowEntryContext `json:"workflow,omitempty"`
	Activity  *ActivityEntryContext `json:"activity,omitempty"`
	Route     *RouteEntryContext    `json:"route,omitempty"`
	Error     string                `json:"error,omitempty"`
}

type WorkflowEntryContext struct {
	Status string `json:"status"`
	State  string `json:"state"`
	Path   string `json:"workflow"`
}

type ActivityEntryContext struct {
	ID interface{} `json:"id,omitempty"`
}
type RouteEntryContext struct {
	Path interface{} `json:"path,omitempty"`
}

func toFeatureLogEntry(e logEntryBackend) logEntry {
	featureLogEntry := logEntry{
		Time: e.Time,
		// ID:        e.Time.Format("2006-01-02T15:04:05.000000000Z"),
		Msg:       e.Msg,
		Level:     e.Level,
		Namespace: e.Namespace,
		Error:     e.Error,
	}

	// workflow data if instance
	if e.Scope == string(telemetry.LogScopeInstance) {
		featureLogEntry.Workflow = &WorkflowEntryContext{
			Status: e.Status,
			Path:   e.Path,
			State:  e.State,
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

func (m *logController) fetchFromBackend(query string) ([]logEntryBackend, error) {
	var ret []logEntryBackend

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost,
		fmt.Sprintf("http://%s/select/logsql/query", net.JoinHostPort(m.logsBackend, "9428")),
		bytes.NewBufferString(query))
	if err != nil {
		return ret, err
	}

	// set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	cli := &http.Client{Timeout: 30 * time.Second}

	// cc, _ := httputil.DumpRequest(req, true)
	// fmt.Println(string(cc))

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
