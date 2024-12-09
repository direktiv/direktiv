package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/direktiv/direktiv/pkg/metastore"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type LogStore struct {
	client      *opensearch.Client
	logIndex    string
	deleteAfter string // e.g., "30d" for 30 days
}

func (store *LogStore) Append(ctx context.Context, log metastore.LogEntry) error {
	if log.ID == "" {
		return fmt.Errorf("log entry ID is required")
	}

	if log.Timestamp.IsZero() {
		return fmt.Errorf("log entry timestamp is required")
	}

	// Serialize log entry
	body, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	req := opensearchapi.IndexRequest{
		Index:      store.logIndex,
		DocumentID: log.ID,
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, store.client)
	if err != nil {
		return fmt.Errorf("failed to append log entry: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error appending log entry, status: %s", res.String())
	}

	return nil
}

// Get implements metastore.LogStore.
func (store *LogStore) Get(ctx context.Context, options metastore.LogQueryOptions) ([]metastore.LogEntry, error) {
	// Construct the query
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"range": map[string]interface{}{
							"timestamp": map[string]interface{}{
								"gte": options.StartTime.Format("2006-01-02T15:04:05.000Z"),
								"lte": options.EndTime.Format("2006-01-02T15:04:05.000Z"),
							},
						},
					},
				},
			},
		},
	}

	if len(options.Levels) > 0 {
		queryMap, ok := query["query"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid type for 'query', expected map[string]interface{}")
		}
		boolMap, ok := queryMap["bool"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid type for 'bool', expected map[string]interface{}")
		}
		mustSlice, ok := boolMap["must"].([]map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid type for 'must', expected []map[string]interface{}")
		}
		mustSlice = append(mustSlice, map[string]interface{}{
			"terms": map[string]interface{}{
				"level": options.Levels, // Add array of levels
			},
		})
		boolMap["must"] = mustSlice
	}

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	// Generate the search request
	req := opensearchapi.SearchRequest{
		Index: []string{store.logIndex},
		Body:  bytes.NewReader(body),
	}

	// Execute the search request
	res, err := req.Do(ctx, store.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error executing search: %s", res.String())
	}

	// Parse the response
	var searchResult struct {
		Hits struct {
			Hits []struct {
				Source metastore.LogEntry `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	// Extract log entries
	logs := make([]metastore.LogEntry, len(searchResult.Hits.Hits))
	for i, hit := range searchResult.Hits.Hits {
		logs[i] = hit.Source
	}

	return logs, nil
}

var _ metastore.LogStore = &LogStore{}

// NewOpenSearchLogStore creates a new OpenSearchLogStore with the specified settings.
func NewOpenSearchLogStore(client *opensearch.Client, logIndex string, deleteAfter string) *LogStore {
	return &LogStore{
		client:      client,
		logIndex:    logIndex,
		deleteAfter: deleteAfter,
	}
}

// Init ensures the index and lifecycle policies are created.
func (store *LogStore) Init(ctx context.Context) error {
	// Ensure the index exists
	if err := store.ensureIndex(ctx); err != nil {
		return fmt.Errorf("failed to ensure index: %w", err)
	}

	// Ensure lifecycle policies
	if err := store.ensureDeletionPolicy(ctx); err != nil {
		return fmt.Errorf("failed to ensure deletion policy: %w", err)
	}

	return nil
}

func (store *LogStore) ensureIndex(ctx context.Context) error {
	// Check if the index exists
	req := opensearchapi.IndicesExistsRequest{Index: []string{store.logIndex}}
	res, err := req.Do(ctx, store.client)
	if err != nil {
		return fmt.Errorf("failed to check if index exists: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		// Index already exists
		return nil
	}

	// Define index mappings and settings
	indexSettings := map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 1,
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"timestamp": map[string]interface{}{"type": "date"},
				"level":     map[string]interface{}{"type": "keyword"},
				"message":   map[string]interface{}{"type": "text"},
				"metadata":  map[string]interface{}{"type": "object"},
			},
		},
	}

	body, err := json.Marshal(indexSettings)
	if err != nil {
		return fmt.Errorf("failed to marshal index settings: %w", err)
	}

	// Create the index
	createReq := opensearchapi.IndicesCreateRequest{
		Index: store.logIndex,
		Body:  bytes.NewReader(body),
	}

	createRes, err := createReq.Do(ctx, store.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		return fmt.Errorf("error creating index: %s", createRes.String())
	}

	return nil
}

func (store *LogStore) ensureDeletionPolicy(ctx context.Context) error {
	// Define a policy for automatic log deletion
	policy := map[string]interface{}{
		"policy": map[string]interface{}{
			"description": "Log retention policy",
			"phases": map[string]interface{}{
				"delete": map[string]interface{}{
					"min_age": store.deleteAfter,
					"actions": map[string]interface{}{
						"delete": map[string]interface{}{},
					},
				},
			},
		},
	}

	body, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to marshal policy: %w", err)
	}

	// Construct the HTTP request
	endpoint := fmt.Sprintf("/_ilm/policy/%s_policy", store.logIndex)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set necessary headers
	req.Header.Set("Content-Type", "application/json")

	// Use the client's transport to execute the request
	res, err := store.client.Transport.Perform(req)
	if err != nil {
		return fmt.Errorf("failed to execute ILM request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusOK {
		return fmt.Errorf("error applying lifecycle policy: %s", res.Status)
	}

	// Attach the policy to the index
	attachPolicyBody := fmt.Sprintf(`{"index.lifecycle.name": "%s_policy"}`, store.logIndex)
	attachReq := opensearchapi.IndicesPutSettingsRequest{
		Index: []string{store.logIndex},
		Body:  strings.NewReader(attachPolicyBody),
	}

	attachRes, err := attachReq.Do(ctx, store.client)
	if err != nil {
		return fmt.Errorf("failed to attach lifecycle policy: %w", err)
	}
	defer attachRes.Body.Close()

	if attachRes.IsError() {
		return fmt.Errorf("error attaching lifecycle policy: %s", attachRes.String())
	}

	return nil
}
