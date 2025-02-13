package victoriametrics

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
func NewVictoriaMetricsLogStore(endpoint string, timeout time.Duration) *LogStore {
	return &LogStore{
		Endpoint: endpoint,
		Client:   &http.Client{Timeout: timeout},
	}
}

// fetchLogs is a helper function that sends a request to VictoriaMetrics
// and either returns a slice of logs (for Get) or streams logs to a channel (for Stream).
func (v *LogStore) fetchLogs(ctx context.Context, options metastore.LogQueryOptions, endpoint string, ch chan<- metastore.LogEntry) ([]metastore.LogEntry, error) {
	query := "*"
	if len(options.Keywords) != 0 {
		query = options.Keywords
	}

	// Prepare request body
	formData := url.Values{}
	formData.Set("query", query)
	formData.Set("limit", strconv.Itoa(options.Limit))
	if len(options.Metadata) > 0 {
		filters, err := json.Marshal(options.Metadata)
		if err != nil {
			return nil, err
		}
		formData.Set("extra_filters", string(filters))
	}
	if options.StartTime != nil {
		formData.Set("start", fmt.Sprint(options.StartTime.Unix()))
	}
	if options.EndTime != nil {
		formData.Set("end", fmt.Sprint(options.EndTime.Unix()))
	}

	// Create a POST request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute the request
	resp, err := v.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch logs, status: %v, response: %s", resp.StatusCode, string(bodyBytes))
	}

	// Read logs from response
	var logs []metastore.LogEntry
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		var entry metastore.LogEntry
		line := scanner.Text()

		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("failed to parse JSON line: %s, error: %w", line, err)
		}

		if ch != nil {
			// If streaming, send to the channel
			select {
			case ch <- entry:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		} else {
			// Otherwise, collect for batch response
			logs = append(logs, entry)
		}
	}

	// Handle scanner errors
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	return logs, nil
}

// Get fetches logs from VictoriaMetrics.
func (v *LogStore) Get(ctx context.Context, options metastore.LogQueryOptions) ([]metastore.LogEntry, error) {
	return v.fetchLogs(ctx, options, fmt.Sprintf("%s/select/logsql/query", v.Endpoint), nil)
}

// Stream streams logs from VictoriaMetrics in real-time.
func (v *LogStore) Stream(ctx context.Context, options metastore.LogQueryOptions, ch chan<- metastore.LogEntry) error {
	defer close(ch)
	_, err := v.fetchLogs(ctx, options, fmt.Sprintf("%s/select/logsql/tail", v.Endpoint), ch)

	return err
}
