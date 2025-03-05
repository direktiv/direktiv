package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/tracing"
	"github.com/go-chi/chi/v5"
)

type logController struct {
	store       datastore.LogStore
	logsBackend string
}

type logParams struct {
	namespace string
	after     time.Time
	track     string
}

func (m *logController) mountRouter(r chi.Router, config *core.Config) {
	r.Get("/subscribe", m.stream)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		params := extractLogRequestParams(r)

		now := time.Now().UTC()
		q := newLogQuerier(params.track)

		// handle different params here
		q = q.withDateSortAsc().beforeDate(now)

		logs, err := m.get(q.string())

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
			"previousPage": logs[0].Time.Format(time.RFC3339Nano),
			"startingFrom": now.Format(time.RFC3339Nano),
		}

		// writeJSONWithMeta(w, logs, metaInfo)
		writeJSONWithMeta(w, []logEntry{}, metaInfo)
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

// func (m *logController) getNewer(ctx context.Context, t time.Time, params map[string]string) ([]logEntry, error) {
// 	var logs []core.LogEntry
// 	var err error

// 	// Determine the track based on the provided parameters
// 	stream, err := determineTrack(params)
// 	if err != nil {
// 		return []logEntry{}, err
// 	}

// 	// Call the appropriate LogStore method with cursorTime
// 	lastID, hasLastID := params["lastID"]
// 	_, isInstanceRequest := params["instance"]
// 	if hasLastID && isInstanceRequest {
// 		id, err := strconv.Atoi(lastID)
// 		if err != nil {
// 			return []logEntry{}, err
// 		}
// 		r, err := m.store.GetStartingIDUntilTimeInstance(ctx, stream, id, t)
// 		if err != nil {
// 			return []logEntry{}, err
// 		}
// 		logs = append(logs, r...)
// 	}
// 	if hasLastID && !isInstanceRequest {
// 		id, err := strconv.Atoi(lastID)
// 		if err != nil {
// 			return []logEntry{}, err
// 		}
// 		r, err := m.store.GetStartingIDUntilTime(ctx, stream, id, t)
// 		if err != nil {
// 			return []logEntry{}, err
// 		}
// 		logs = append(logs, r...)
// 	}

// 	if _, ok := params["instance"]; ok {
// 		r, err := m.store.GetNewerInstance(ctx, stream, t)
// 		if err != nil {
// 			return []logEntry{}, err
// 		}
// 		logs = append(logs, r...)
// 	} else {
// 		r, err := m.store.GetNewer(ctx, stream, t)
// 		if err != nil {
// 			return []logEntry{}, err
// 		}
// 		logs = append(logs, r...)
// 	}

// 	res := []logEntry{}
// 	// for _, le := range logs {
// 	// 	// e := toFeatureLogEntry(le)
// 	// 	// res = append(res, e)
// 	// }

// 	return res, nil
// }

func (m *logController) get(query string) ([]logEntry, error) {
	// var r []core.LogEntry
	var err error

	// fmt.Println(query)
	// Determine the track based on the provided parameters
	// stream, err := determineTrack(params)
	// if err != nil {
	// 	return []logEntry{}, time.Time{}, err
	// }

	// starting := time.Now().UTC()
	// if t, ok := params["before"]; ok {
	// 	co, err := time.Parse(time.RFC3339Nano, t)
	// 	if err != nil {
	// 		return []logEntry{}, time.Time{}, err
	// 	}
	// 	starting = co
	// }

	// compare := ">="
	// if older {
	// 	compare = "<="
	// }

	// out := fmt.Sprintf("query=track:=%s _time:%s%s  | sort by (_time) asc",
	// 	stream, compare, starting.Format("2006-01-02T15:04:05.000Z"))
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
	// cursor is set to multiple seconds before the current time to mitigate data loss
	// that may occur due to delays between submitting and processing the request, or when a sequence of client requests is necessary.
	cursor := time.Now().Add(-48 * time.Hour)
	// var err error

	// Last-Event-ID header
	lastEventID := r.Header.Get("Last-Event-ID")
	fmt.Printf("LAST EVENT ID >%v<\n", lastEventID)

	// // TODO: we may need to replace with a SSE-Server library instead of using our custom implementation.
	params := extractLogRequestParams(r)

	rc := http.NewResponseController(w)

	queryAndSend := func(queryString string) (time.Time, error) {
		logs, err := m.get(queryString)
		if err != nil {
			return time.Now(), err
		}

		for i := range logs {
			l := logs[i]
			b, _ := json.Marshal(l)
			fmt.Println("WRITE LOGS")
			_, err = fmt.Fprintf(w, fmt.Sprintf("id: %v\nevent: %v\ndata: %v\n\n", l.ID, "message", string(b)))
			if err != nil {
				return time.Now(), err
			}
		}

		fmt.Printf("LOGS %d\n", len(logs))

		// check time of last log
		// checkTime := time.Now().UTC()
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

	// send initial data
	var err error
	q := newLogQuerier(params.track).withDateSortAsc()
	cursor, err = queryAndSend(q.string())
	if err != nil {
		// LOG ERROR
		fmt.Println(err)
		return
	}

	clientGone := r.Context().Done()

	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case <-clientGone:
			// disconnected
			fmt.Println("DISCONNECTED!!")
			return
		case <-t.C:
			var err error
			// Send an event to the client
			// Here we send only the "data" field, but there are few others
			// _, err := fmt.Fprintf(w, "data: The time is %s\n\n", time.Now().Format(time.UnixDate))
			// if err != nil {
			// 	return
			// }
			// fmt.Println("DONE!!!")
			// fmt.Println(cursor)
			// logs, _, err := m.get(r.Context(), false, params)
			// for i := range logs {
			// 	fmt.Println(logs[i])
			// }
			q := newLogQuerier(params.track).afterDate(cursor).withDateSortAsc()
			fmt.Println(q.string())
			cursor, err = queryAndSend(q.string())
			if err != nil {
				fmt.Println(err)
				return
			}

			// logs, err := m.get(q.string())
			// if err != nil {
			// 	return
			// }

			// fmt.Println("LOGS!!!!!!!!!!!!!!!!!!!!!!")

			// for i := range logs {
			// 	l := logs[i]
			// 	fmt.Println(l)
			// 	// toFeatureLogEntry(l)

			// 	// b, _ := json.Marshal(l)

			// 	_, err := fmt.Fprintf(w, fmt.Sprintf("id: %v\nevent: %v\ndata: %v\n\n", l.ID, "message", string(b)))
			// 	// if err != nil {
			// 	// 	return
			// 	// }
			// }

			// cursor = time.Now().UTC()
			// handle different params here
			// q = q.withDateSortAsc().beforeDate(now)

			// err = rc.Flush()
			// if err != nil {
			// 	return
			// }
		}
	}
	// // Create a context with cancellation
	// ctx, cancel := context.WithCancel(r.Context())
	// defer cancel()

	// // Create a channel to send SSE messages
	// messageChannel := make(chan Event)

	// var getCursoredStyle sseHandle = func(ctx context.Context, cursorTime time.Time) ([]CoursoredEvent, error) {
	// 	logs, err := m.getNewer(ctx, cursorTime, params)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	res := make([]CoursoredEvent, 0, len(logs))
	// 	for _, fle := range logs {
	// 		b, err := json.Marshal(fle)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		dst := &bytes.Buffer{}
	// 		if err := json.Compact(dst, b); err != nil {
	// 			return nil, err
	// 		}

	// 		e := Event{
	// 			ID:   strconv.Itoa(fle.ID),
	// 			Data: dst.String(),
	// 			Type: "message",
	// 		}
	// 		res = append(res, CoursoredEvent{
	// 			Event: e,
	// 			Time:  fle.Time,
	// 		})
	// 	}

	// 	return res, nil
	// }

	// worker := seeWorker{
	// 	Get:      getCursoredStyle,
	// 	Interval: time.Second,
	// 	Ch:       messageChannel,
	// 	Cursor:   cursor,
	// }
	// go worker.start(ctx)

	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		return
	// 	case message := <-messageChannel:beforeDate
	// 			f.Flush()
	// 		}
	// 	}
	// }
}

// determineTrack determines the log track based on provided parameters.
// It constructs a track string used for filtering logs in datastore queries.
func determineTrack(params map[string]string) (string, error) {

	return "", fmt.Errorf("requested logs for an unknown type")
}

// nolint:canonicalheader
func extractLogRequestParams(r *http.Request) logParams {
	// params := map[string]string{}

	var logParams logParams

	if v := chi.URLParam(r, "namespace"); v != "" {
		// params["namespace"] = v
		logParams.namespace = v

		// set track to namespace first, we can change it later
		logParams.track = "namespace." + v
	}

	// values := r.URL.Query()
	// for k, v := range values {
	// 	if len(v) > 0 {
	// 		params[k] = v[0]
	// 	}
	// }

	// if p, ok := params["instance"]; ok {
	// 	params["track"] = "instance." + p
	// } else if p, ok := params["route"]; ok {
	// 	params["track"] = "route." + p
	// } else if p, ok := params["activity"]; ok {
	// 	params["track"] = "activity." + p
	// } else if p, ok := params["namespace"]; ok {
	// 	params["track"] = "namespace." + p
	// } else if p, ok := params["trace"]; ok {
	// 	params["track"] = "trace." + p
	// }

	return logParams
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

// func toFeatureLogEntry(e core.LogEntry) logEntry {
func toFeatureLogEntry(e Log) logEntry {
	featureLogEntry := logEntry{
		ID:        int(e.Time.UnixMilli()),
		Time:      e.Time,
		Msg:       e.Msg,
		Level:     e.Level,
		Trace:     e.Trace,
		Span:      e.Span,
		Namespace: e.Namespace,
		// Workflow: &WorkflowEntryContext{
		// 	State:    e.State,
		// 	Path:     e.Workflow,
		// 	Instance: e.Instance,
		// 	// CalledAs: ,
		// 	Status: e.Status,
		// 	// Branch: ,
		// },
	}
	// featureLogEntry.Error = e.Data["error"]

	// wfLogCtx := WorkflowEntryContext{}
	// wfLogCtx.State = e.Data["state"]
	// wfLogCtx.Path = e.Data["workflow"]
	// wfLogCtx.Instance = e.Data["instance"]
	// wfLogCtx.CalledAs = e.Data["calledAs"]
	// wfLogCtx.Status = e.Data["status"]
	// wfLogCtx.Branch = e.Data["branch"]
	// featureLogEntry.Workflow = &wfLogCtx
	// if wfLogCtx.Path == nil && wfLogCtx.Instance == nil {
	// 	featureLogEntry.Workflow = nil
	// }
	// if id, ok := e.Data["activity"]; ok && id != nil {
	// 	featureLogEntry.Activity = &ActivityEntryContext{ID: id}
	// }
	// if path, ok := e.Data["route"]; ok && path != nil {
	// 	featureLogEntry.Route = &RouteEntryContext{Path: path}
	// }

	return featureLogEntry
}

type Log struct {
	Time      time.Time `json:"_time"`
	StreamID  string    `json:"_stream_id"`
	Stream    string    `json:"_stream"`
	Msg       string    `json:"_msg"`
	P         string    `json:"_p"`
	Stream2   string    `json:"stream"`
	Date      time.Time `json:"date"`
	Instance  string    `json:"instance"`
	Invoker   string    `json:"invoker"`
	Level     string    `json:"level"`
	Namespace string    `json:"namespace"`
	Span      string    `json:"span"`
	State     string    `json:"state"`
	Status    string    `json:"status"`
	Trace     string    `json:"trace"`
	Track     string    `json:"track"`
	Workflow  string    `json:"workflow"`
}

func (m *logController) fetchFromBackend(query string) ([]Log, error) {
	var ret []Log

	req, err := http.NewRequest(http.MethodPost,
		fmt.Sprintf("http://%s:9428/select/logsql/query", m.logsBackend),
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
		var log Log
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

type logQuerier struct {
	query string
	sort  string
}

func newLogQuerier(track string) logQuerier {
	return logQuerier{
		query: "query=track:=" + track,
	}
}

func (l logQuerier) beforeDate(t time.Time) logQuerier {
	l.query = l.query + " _time:<" + t.Format("2006-01-02T15:04:05.000000000Z")
	return l
}

func (l logQuerier) afterDate(t time.Time) logQuerier {
	l.query = l.query + " _time:>" + t.Format("2006-01-02T15:04:05.000000000Z")
	return l
}

func (l logQuerier) withDateSortAsc() logQuerier {
	l.sort = l.sort + " | sort by (_time) asc"
	return l
}

func (l logQuerier) withDateSortDesc() logQuerier {
	l.sort = l.sort + " | sort by (_time) desc"
	return l
}

func (l logQuerier) string() string {
	return fmt.Sprintf("%s %s", l.query, l.sort)
}
