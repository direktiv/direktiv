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

// LogStore implements LogStore for VictoriaMetrics backend.
type LogStore struct {
	Endpoint string       // VictoriaMetrics API endpoint
	Client   *http.Client // HTTP client for making requests
}

// NewVictoriaMetricsLogStore creates a new log store instance.
func NewVictoriaMetricsLogStore(endpoint string, timeout time.Duration) *LogStore {
	return &LogStore{
		Endpoint: endpoint,
		Client:   &http.Client{Timeout: timeout},
	}
}

// Get fetches logs from VictoriaMetrics based on LogQueryOptions.
func (v *LogStore) Get(ctx context.Context, options metastore.LogQueryOptions) ([]metastore.LogEntry, error) {
	// query := v.buildLogQLQuery(options)
	query := "*"
	// Prepare request body
	formData := url.Values{}
	formData.Set("query", query)
	formData.Set("limit", strconv.Itoa(options.Limit))

	// Create a POST request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/select/logsql/query", v.Endpoint), strings.NewReader(formData.Encode()))
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

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch logs, status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Parse response JSON
	var response []metastore.LogEntry

	// Read the response as NDJSON (Newline-Delimited JSON)
	scanner := bufio.NewScanner(strings.NewReader(string(bodyBytes)))
	for scanner.Scan() {
		var entry metastore.LogEntry
		line := scanner.Text()

		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("failed to parse JSON line: %s, error: %w", line, err)
		}

		response = append(response, entry)
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	return response, nil
}

// buildLogQLQuery converts LogQueryOptions into a VictoriaMetrics LogQL query.
func (v *LogStore) buildLogQLQuery(options metastore.LogQueryOptions) string {
	var filters []string

	// Convert metadata filters (namespace, instance, workflow, etc.)
	for k, v := range options.Metadata {
		filters = append(filters, fmt.Sprintf(`%s="%s"`, k, v))
	}

	// Add log level filter if specified
	if options.Level != "" {
		filters = append(filters, fmt.Sprintf(`level="%s"`, options.Level))
	}

	// Construct the base query
	query := fmt.Sprintf("{%s}", strings.Join(filters, ","))

	// Add keyword search if specified
	for _, keyword := range options.Keywords {
		query += fmt.Sprintf(` |= "%s"`, keyword)
	}

	return query
}
