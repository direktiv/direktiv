package victoriametrics

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/metastore"
)

// LogStore implements LogStore for VictoriaMetrics Logs backend.
type LogStore struct {
	Endpoint string       // VictoriaMetrics Logs API endpoint
	Client   *http.Client // HTTP client for making requests
}

// NewVictoriaMetricsLogStore creates a new log store instance.
func NewVictoriaMetricsLogStore(endpoint string, timeout *time.Duration) *LogStore {
	client := &http.Client{} // Ensure the client is always initialized

	if timeout != nil {
		client.Timeout = *timeout
	}

	return &LogStore{
		Endpoint: endpoint,
		Client:   client,
	}
}

// fetchLogs is a helper function that sends a request to VictoriaMetrics
// and either returns a slice of logs (for Get) or streams logs to a channel (for Stream).
func (v *LogStore) fetchLogs(ctx context.Context, options metastore.LogQueryOptions, endpoint string, add func(entry metastore.LogEntry) error) error {
	query := "*"
	if len(options.Keywords) != 0 {
		query = options.Keywords
	}

	formData := url.Values{}
	formData.Set("query", query)
	if options.Limit > 0 {
		formData.Set("limit", strconv.Itoa(options.Limit))
	}
	if len(options.Metadata) > 0 {
		filters, err := json.Marshal(options.Metadata)
		if err != nil {
			return err
		}
		formData.Set("extra_filters", string(filters))
	}
	if options.StartTime != nil {
		formData.Set("start", strconv.FormatInt(options.StartTime.UnixNano(), 10))
	}
	if options.EndTime != nil {
		formData.Set("end", strconv.FormatInt(options.EndTime.UnixNano(), 10))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Retry logic with exponential backoff
	maxRetries := 5
	delay := time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := v.Client.Do(req)
		if err != nil {
			// Network errors, context canceled, etc.
			if ctx.Err() != nil {
				return ctx.Err()
			}
			slog.Error("HTTP request failed", "attempt", attempt, "error", err)
		} else {
			defer resp.Body.Close()

			// If response is successful, process logs
			if resp.StatusCode == http.StatusOK {
				return processLogs(resp.Body, add)
			}

			// Retry on retriable HTTP errors (502, 503, 504)
			if resp.StatusCode != http.StatusBadGateway && resp.StatusCode != http.StatusServiceUnavailable && resp.StatusCode != http.StatusGatewayTimeout {
				bodyBytes, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed to fetch logs, status: %v, response: %s", resp.StatusCode, string(bodyBytes))
			}
		}

		// Apply exponential backoff before retrying
		if attempt < maxRetries {
			slog.Warn("Retrying request due to bad gateway", "attempt", attempt+1, "delay", delay)
			time.Sleep(delay)
			delay += 1 // Double the delay for next retry

			continue
		}

		return fmt.Errorf("fetchLogs failed after %d attempts", maxRetries)
	}

	return nil
}

func processLogs(body io.Reader, add func(entry metastore.LogEntry) error) error {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		rawLog := map[string]any{}
		line := scanner.Text()
		if err := json.Unmarshal([]byte(line), &rawLog); err != nil {
			return fmt.Errorf("failed to parse JSON line: %s, error: %w", line, err)
		}
		var entry metastore.LogEntry
		keyMapping := map[string]string{
			"_time": "time",
			"_msg":  "msg",
		}
		for k, v := range keyMapping {
			rawLog[v] = rawLog[k]
		}
		entryBytes, err := json.Marshal(rawLog)
		if err != nil {
			return fmt.Errorf("failed to marshal raw log: %w", err)
		}

		if err := json.Unmarshal(entryBytes, &entry); err != nil {
			return fmt.Errorf("failed to unmarshal into LogEntry: %w", err)
		}
		entry.Time = entry.Time.UTC()
		if err := add(entry); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}

	return nil
}

// Get fetches logs from VictoriaMetrics.
func (v *LogStore) Get(ctx context.Context, options metastore.LogQueryOptions) ([]metastore.LogEntry, error) {
	logs := []metastore.LogEntry{}
	add := func(entry metastore.LogEntry) error {
		logs = append(logs, entry)
		return nil
	}
	// Fetch logs and collect them into a slice
	err := v.fetchLogs(ctx, options, fmt.Sprintf("%s/select/logsql/query", v.Endpoint), add)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

// Stream streams logs from VictoriaMetrics in real-time.
func (v *LogStore) Stream(ctx context.Context, options metastore.LogQueryOptions, ch chan<- metastore.LogEntry) error {
	defer close(ch)

	add := func(entry metastore.LogEntry) error {
		select {
		case ch <- entry:
		case <-ctx.Done():
			return ctx.Err()
		}

		return nil
	}

	// Call fetchLogs and pass the add function for processing the entries
	err := v.fetchLogs(ctx, options, fmt.Sprintf("%s/select/logsql/tail", v.Endpoint), add)
	if err != nil {
		slog.Error("failed fetching logs", "err", err)
	}

	return err
}
