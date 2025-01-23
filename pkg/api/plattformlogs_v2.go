package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/metastore"
	"github.com/direktiv/direktiv/pkg/tracing"
	"github.com/go-chi/chi/v5"
)

type logControllerV2 struct {
	metaLogStore metastore.LogStore
}

func (m *logControllerV2) mountRouter(r chi.Router) {
	r.Get("/subscribe", m.stream)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		params := extractLogRequestParams(r)
		slog.Info("Log endpoint request start")
		// Call the Get method with the cursor instead of offset
		data, starting, err := m.getOlder(r.Context(), params)
		if err != nil {
			slog.Error("Fetching logs for request.", "err", err)
			writeInternalError(w, err)

			return
		}
		metaInfo := map[string]any{
			"previousPage": nil,
			"startingFrom": nil,
		}
		if len(data) == 0 {
			slog.Info("Log endpoint request empty")
			writeJSONWithMeta(w, []logEntry{}, metaInfo)

			return
		}

		var previousPage interface{} = data[0].Time.UTC().Format(time.RFC3339Nano)

		metaInfo = map[string]any{
			"previousPage": previousPage,
			"startingFrom": starting,
		}
		slog.Info("Log endpoint request data", "data", data)

		writeJSONWithMeta(w, data, metaInfo)
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

func (m *logControllerV2) getOlder(ctx context.Context, params map[string]string) ([]logEntry, time.Time, error) {
	var r []logEntry
	var err error
	// Determine the track based on the provided parameters
	// stream, err := determineTrack(params)
	// if err != nil {
	// 	return []logEntry{}, time.Time{}, err
	// }

	starting := time.Now().UTC().Add(-time.Hour + 2)
	if t, ok := params["before"]; ok {
		co, err := time.Parse(time.RFC3339Nano, t)
		if err != nil {
			return []logEntry{}, time.Time{}, err
		}
		starting = co
	}

	r, err = getOlder(ctx, m.metaLogStore, params["namespace"], starting)
	if err != nil {
		return []logEntry{}, time.Time{}, err
	}

	return r, starting, nil
}

// TODO: stream handles log streaming requests using Server-Sent Events (SSE).
// Clients subscribing to this endpoint will receive real-time log updates.
func (m *logControllerV2) stream(w http.ResponseWriter, r *http.Request) {}

func toFeatureLogEntryV2(e metastore.LogEntry) logEntry {
	// Create a new feature log entry and map the relevant fields.
	featureLogEntry := logEntry{
		ID:        0,                           // TODO: Map LogEntry ID from the metastore to feature log entry
		Time:      time.UnixMilli(e.Timestamp), // Convert Unix timestamp to time.Time
		Msg:       e.Message,                   // Directly map the Message field
		Level:     e.Level,                     // Map the Level field
		Namespace: e.Metadata["namespace"],     // Assuming "namespace" is present in metadata
		Trace:     e.Metadata["trace"],         // Assuming "trace" is present in metadata
		Span:      e.Metadata["span"],          // Assuming "span" is present in metadata
	}

	// Map optional contextual fields if available
	if workflowPath, ok := e.Metadata["workflow"]; ok {
		featureLogEntry.Workflow = &WorkflowEntryContext{
			Path: workflowPath,
		}
	}
	if activityID, ok := e.Metadata["activity"]; ok {
		featureLogEntry.Activity = &ActivityEntryContext{
			ID: activityID,
		}
	}
	if routePath, ok := e.Metadata["route"]; ok {
		featureLogEntry.Route = &RouteEntryContext{
			Path: routePath,
		}
	}

	// Return the constructed feature log entry
	return featureLogEntry
}

// Helper function to get logs starting from a specific ID until a given time, without instance context.
func getStartingIDUntilTime(ctx context.Context, store metastore.LogStore, id string, untilTime time.Time) ([]logEntry, error) {
	// Construct query options to filter logs by ID and time range.
	queryOptions := metastore.LogQueryOptions{
		StartTime: time.Time{}, // Adjust start time as needed
		EndTime:   untilTime,
		Level:     0, // Adjust level if needed
	}

	logs, err := store.Get(ctx, queryOptions)
	if err != nil {
		return nil, err
	}

	var filteredLogs []logEntry
	for _, log := range logs {
		if log.ID >= id { // Filter by ID
			// Convert metastore.LogEntry to logEntry and append
			featureLogEntry := toFeatureLogEntryV2(log)
			filteredLogs = append(filteredLogs, featureLogEntry)
		}
	}

	return filteredLogs, nil
}

// Helper function to get newer logs by instance and timestamp.
func getNewerInstance(ctx context.Context, store metastore.LogStore, timestamp time.Time) ([]logEntry, error) {
	// Construct query options to filter logs by instance and timestamp.
	queryOptions := metastore.LogQueryOptions{
		StartTime: timestamp,
		EndTime:   time.Now(), // Current time as end time
		// Metadata: map[string]string{
		// 	"instance": instanceID,
		// },
	}

	logs, err := store.Get(ctx, queryOptions)
	if err != nil {
		return nil, err
	}

	// Convert the logs to logEntry before returning.
	var featureLogs []logEntry
	for _, log := range logs {
		featureLogs = append(featureLogs, toFeatureLogEntryV2(log))
	}

	return featureLogs, nil
}

// Helper function to get newer logs without instance context.
func getNewer(ctx context.Context, store metastore.LogStore, timestamp time.Time) ([]logEntry, error) {
	// Construct query options to filter logs by timestamp.
	queryOptions := metastore.LogQueryOptions{
		StartTime: timestamp,
		EndTime:   time.Now(), // Current time as end time
	}

	logs, err := store.Get(ctx, queryOptions)
	if err != nil {
		return nil, err
	}

	// Convert the logs to logEntry before returning.
	var featureLogs []logEntry
	for _, log := range logs {
		featureLogs = append(featureLogs, toFeatureLogEntryV2(log))
	}

	return featureLogs, nil
}

// getOlder retrieves older logs by stream and timestamp (without instance context).
func getOlder(ctx context.Context, store metastore.LogStore, namespace string, starting time.Time) ([]logEntry, error) {
	// Construct query options to filter logs by stream and time range.
	queryOptions := metastore.LogQueryOptions{
		StartTime: starting.UTC(),
		EndTime:   time.Now().UTC(), // Current time as end time
		Metadata: map[string]string{
			"namespace": namespace,
		},
	}

	logs, err := store.Get(ctx, queryOptions)
	if err != nil {
		return nil, err
	}

	// Convert the logs to logEntry before returning.
	var featureLogs []logEntry
	for _, log := range logs {
		featureLogs = append(featureLogs, toFeatureLogEntryV2(log))
	}

	return featureLogs, nil
}
