package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/direktiv/direktiv/pkg/metastore"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

// NewOpenSearchLogStore creates a new OpenSearchLogStore with the specified settings.
func NewOpenSearchLogStore(client *opensearch.Client, co config) *LogStore {
	return &LogStore{
		client:      client,
		logIndex:    co.LogIndex,
		deleteAfter: co.LogDeleteAfter,
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

type LogStore struct {
	client      *opensearch.Client
	logIndex    string
	deleteAfter string // e.g., "30d" for 30 days
}

func (store *LogStore) Append(ctx context.Context, log metastore.LogEntry) error {
	if log.ID == "" {
		return fmt.Errorf("log entry ID is required")
	}

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
			"match_all": map[string]interface{}{},
		},
	}

	// query := map[string]interface{}{
	// 	"query": map[string]interface{}{
	// 		// "range": map[string]interface{}{
	// 		// 	"timestamp": map[string]interface{}{
	// 		// 		"gt": options.StartTime.UTC().UnixMilli(),
	// 		// 		"lt": options.EndTime.UTC().UnixMilli(),
	// 		// 	},
	// 		// },
	// 		"term": map[string]interface{}{
	// 			"level": options.Levels[0],
	// 		},
	// 	},
	// }

	// // Add level filters if provided
	// if len(options.Levels) > 0 {
	// 	queryMap, ok := query["query"].(map[string]interface{})
	// 	if !ok {
	// 		return nil, fmt.Errorf("invalid type for 'query', expected map[string]interface{}")
	// 	}
	// 	boolMap, ok := queryMap["bool"].(map[string]interface{})
	// 	if !ok {
	// 		return nil, fmt.Errorf("invalid type for 'bool', expected map[string]interface{}")
	// 	}
	// 	mustSlice, ok := boolMap["must"].([]map[string]interface{})
	// 	if !ok {
	// 		return nil, fmt.Errorf("invalid type for 'must', expected []map[string]interface{}")
	// 	}
	// 	mustSlice = append(mustSlice, map[string]interface{}{
	// 		"terms": map[string]interface{}{
	// 			"level": options.Levels, // Add array of levels
	// 		},
	// 	})
	// 	boolMap["must"] = mustSlice
	// }

	// Create the search request using the OpenSearch client
	searchRes, err := store.client.Search(
		store.client.Search.WithContext(ctx),
		store.client.Search.WithIndex(store.logIndex),
		// store.client.Search.WithQuery(),
		store.client.Search.WithBody(bytes.NewReader(mustJSON(query))),
		store.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	if searchRes.IsError() {
		responseBody, _ := io.ReadAll(searchRes.Body)
		slog.Error("Search failed", "status", searchRes.Status(), "response", string(responseBody))

		return nil, fmt.Errorf("error executing search: %s, response: %s", searchRes.Status(), string(responseBody))
	}
	defer searchRes.Body.Close()
	slog.Error(string(mustJSON(query)))
	// Check if the search was successful
	if searchRes.IsError() {
		responseBody, _ := io.ReadAll(searchRes.Body)
		return nil, fmt.Errorf("error executing search: %s, response: %s", searchRes.Status(), string(responseBody))
	}
	// responseBody, _ := io.ReadAll(searchRes.Body)
	// panic(string(responseBody))

	// Parse the response body
	var searchResult struct {
		Hits struct {
			Hits []struct {
				Source metastore.LogEntry `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(searchRes.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}
	// Extract log entries from the search result
	logs := make([]metastore.LogEntry, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		logs = append(logs, hit.Source)
	}

	return logs, nil
}

// mustJSON is a utility function to safely marshal the query to JSON, panicking if it fails
func mustJSON(v interface{}) []byte {
	body, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal query to JSON: %v", err))
	}

	return body
}

var _ metastore.LogStore = &LogStore{}

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
			"refresh_interval":   "1s", // TODO: tune & adjust refresh interval
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"timestamp": map[string]interface{}{
					"type":   "date",
					"format": "epoch_millis", // Handles Unix time in milliseconds
				},
				"level":    map[string]interface{}{"type": "text"},
				"message":  map[string]interface{}{"type": "text"},
				"metadata": map[string]interface{}{"type": "object"},
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
		// Log the response for debugging
		responseBody, _ := io.ReadAll(createRes.Body)
		return fmt.Errorf("error creating index: %s, response: %s", createRes.String(), string(responseBody))
	}

	return nil
}

func (store *LogStore) ensureDeletionPolicy(ctx context.Context) error {
	slog.Debug("define the ISM policy")
	policy := map[string]interface{}{
		"policy": map[string]interface{}{
			"description":   "Log retention policy",
			"default_state": "delete",
			"states": []map[string]interface{}{
				{
					"name": "delete",
					"actions": []map[string]interface{}{
						{
							"delete": map[string]interface{}{},
						},
					},
					"transitions": []map[string]interface{}{},
				},
			},
		},
	}

	slog.Debug("marshal the policy to JSON")
	body, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to marshal ISM policy: %w", err)
	}

	slog.Debug("apply the ISM policy")
	endpoint := fmt.Sprintf("/_plugins/_ism/policies/%s_policy", store.logIndex)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create ISM policy request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := store.client.Transport.Perform(req)
	if err != nil {
		return fmt.Errorf("failed to execute ISM policy request: %w", err)
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)
	if res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("error applying ISM policy: %s, response: %s", res.Status, string(bodyBytes))
	}

	slog.Debug("attach the ISM policy to the index")
	attachPolicyBody := map[string]interface{}{
		"policy_id": fmt.Sprintf("%s_policy", store.logIndex),
	}
	attachBody, err := json.Marshal(attachPolicyBody)
	if err != nil {
		return fmt.Errorf("failed to marshal attach policy body: %w", err)
	}

	attachEndpoint := fmt.Sprintf("/_plugins/_ism/add/%s", store.logIndex)
	attachReq, err := http.NewRequestWithContext(ctx, http.MethodPost, attachEndpoint, bytes.NewReader(attachBody))
	if err != nil {
		return fmt.Errorf("failed to create attach policy request: %w", err)
	}
	attachReq.Header.Set("Content-Type", "application/json")

	attachRes, err := store.client.Transport.Perform(attachReq)
	if err != nil {
		return fmt.Errorf("failed to execute attach policy request: %w", err)
	}
	defer attachRes.Body.Close()

	attachBodyBytes, _ := io.ReadAll(attachRes.Body)
	if attachRes.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("error attaching ISM policy: %s, response: %s", attachRes.Status, string(attachBodyBytes))
	}

	return nil
}
