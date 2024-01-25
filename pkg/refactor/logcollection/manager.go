package logcollection

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
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

func (m *Manager) Get(ctx context.Context, offset int, params map[string]string) ([]core.FeatureLogEntry, error) {
	var r []LogEntry
	var err error
	//nolint:nestif
	if p, ok := params["root-instance-id"]; ok {
		stream := "flow." + p
		r, err = m.store.GetInstanceLogs(ctx, stream, p, offset)
		if err != nil {
			return []core.FeatureLogEntry{}, err
		}
	} else if p, ok := params["namespace"]; ok {
		stream := "flow." + p
		r, err = m.store.Get(ctx, stream, offset)
		if err != nil {
			return []core.FeatureLogEntry{}, err
		}
	} else if p, ok := params["route"]; ok {
		stream := "flow.gateway." + p
		r, err = m.store.Get(ctx, stream, offset)
		if err != nil {
			return []core.FeatureLogEntry{}, err
		}
	} else {
		return []core.FeatureLogEntry{}, fmt.Errorf("requested logs for a unknown typ")
	}
	res := loglist{}
	for _, le := range r {
		e, err := le.toFeatureLogEntry()
		if err != nil {
			return []core.FeatureLogEntry{}, err
		}
		res = append(res, e)
	}
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

// TODO move this to an other pkg.
func (m *Manager) Stream(params map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the appropriate headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Create a context with cancellation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		// Create a channel to send SSE messages
		messageChannel := make(chan []byte)
		defer close(messageChannel)

		// Start a goroutine to listen for messages and send them to the client
		go func() {
			for {
				select {
				case <-ctx.Done():
					return // Exit goroutine on context cancellation
				case message := <-messageChannel:
					fmt.Fprintf(w, "data: %s\n\n", message)
					f, ok := w.(http.Flusher)
					if !ok {
						// TODO Handle case where response writer is not a http.Flusher
						slog.Error("Response writer is not a http.Flusher")

						return
					}
					f.Flush()
				}
			}
		}()
		worker := logStoreWorker{
			Get:      m.Get,
			Interval: time.Second,
			LogCh:    messageChannel,
			Params:   params,
		}
		worker.start(ctx)
	}
}

func (e LogEntry) toFeatureLogEntry() (core.FeatureLogEntry, error) {
	entry := e.Data["entry"]
	m, ok := entry.(map[string]interface{})
	if !ok {
		return core.FeatureLogEntry{}, fmt.Errorf("log-entry format is corrupt")
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

func (e *loglist) filterByBranch(_ string) {
	// TODO
}

func (e *loglist) filterByState(_ string) {
	// TODO
}

func (e *loglist) filterByLevel(_ string) {
	// TODO
}
