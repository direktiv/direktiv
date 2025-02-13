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

func (m *newLogsCtr) mountRouter(r chi.Router) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now().UTC().Add(-time.Hour)
		end := time.Now().UTC()
		logs, err := m.meta.Get(r.Context(), metastore.LogQueryOptions{
			StartTime: &start, // 1 hour ago
			EndTime:   &end,   // Now
			Metadata: map[string]string{
				"namespace": "test",
			},
			Limit: 10000,
		})
		if err != nil {
			writeDataStoreError(w, err)
			return
		}
		writeJSON(w, logs)
	})
	r.Get("/stream", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now().UTC().Add(-time.Hour)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		messageChannel := make(chan metastore.LogEntry)

		// Goroutine for writing messages to the HTTP response
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case message := <-messageChannel:
					b, err := json.Marshal(message)
					if err != nil {
						slog.Error("serve to SSE", "err", err)
					}

					dst := &bytes.Buffer{}
					if err := json.Compact(dst, b); err != nil {
						slog.Error("serve to SSE", "err", err)
					}

					// Writing to the response
					_, err = io.Copy(w, strings.NewReader(fmt.Sprintf("id: %v\nevent: %v\ndata: %v\n\n", uuid.NewString(), "message", dst.String())))
					if err != nil {
						slog.Error("serve to SSE", "err", err)
					}

					// Flush the response to the client
					if f, ok := w.(http.Flusher); ok {
						f.Flush()
					}
				}
			}
		}()

		// Stream logs from VictoriaMetrics
		err := m.meta.Stream(r.Context(), metastore.LogQueryOptions{
			StartTime: &start, // 1 hour ago
			Metadata: map[string]string{
				"namespace": "test",
			},
		}, messageChannel)
		if err != nil {
			writeDataStoreError(w, err)
			return
		}
	})
}
