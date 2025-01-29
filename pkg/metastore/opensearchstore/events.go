package opensearchstore

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

// NewOpenSearchEventsStore creates a new OpenSearchLogStore with the specified settings.
func NewOpenSearchEventsStore(client *opensearch.Client, co Config) *EventStore {
	return &EventStore{
		client:      client,
		eventIndex:  co.EventsIndex,
		deleteAfter: co.EventsDeleteAfter,
	}
}

// Init ensures the index and lifecycle policies are created.
func (store *EventStore) Init(ctx context.Context) error {
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

type EventStore struct {
	client      *opensearch.Client
	eventIndex  string
	deleteAfter string // e.g., "30d" for 30 days
}

func (store *EventStore) Append(ctx context.Context, e metastore.EventEntry) error {
	body, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	req := opensearchapi.IndexRequest{
		Index:      store.eventIndex,
		DocumentID: e.ID,
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, store.client)
	if err != nil {
		return fmt.Errorf("failed to index event: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		responseBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("error indexing event: %s, response: %s", res.Status(), string(responseBody))
	}

	return nil
}

// GetMapping retrieves the mapping of the specified index.
func (store *EventStore) GetMapping(ctx context.Context) (map[string]interface{}, error) {
	// Make a request to OpenSearch to get the mapping
	mappingRes, err := store.client.Indices.GetMapping(
		store.client.Indices.GetMapping.WithContext(ctx),
		store.client.Indices.GetMapping.WithIndex(store.eventIndex), // Use the index name from LogStore
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get index mapping: %w", err)
	}
	defer mappingRes.Body.Close()

	if mappingRes.IsError() {
		responseBody, _ := io.ReadAll(mappingRes.Body)
		slog.Error("Failed to retrieve mapping", "status", mappingRes.Status(), "response", string(responseBody))

		return nil, fmt.Errorf("error retrieving index mapping: %s, response: %s", mappingRes.Status(), string(responseBody))
	}

	// Decode the mapping response
	var mapping map[string]interface{}
	if err := json.NewDecoder(mappingRes.Body).Decode(&mapping); err != nil {
		return nil, fmt.Errorf("failed to decode mapping response: %w", err)
	}

	return mapping, nil
}

func (store *EventStore) Get(ctx context.Context, options metastore.EventQueryOptions) ([]metastore.EventEntry, error) {
	// Build the query
	filters := []map[string]interface{}{
		{
			"range": map[string]interface{}{
				"ReceivedAt": map[string]interface{}{
					"gte": options.StartTime.UTC().UnixMilli(),
					"lte": options.EndTime.UTC().UnixMilli(),
				},
			},
		},
	}

	// Add metadata filters
	for key, value := range options.Metadata {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				fmt.Sprintf("Metadata.%s", key): value,
			},
		})
	}

	// Add keyword search in CloudEvent JSON
	if len(options.Keywords) > 0 {
		shouldQueries := []map[string]interface{}{}
		for _, keyword := range options.Keywords {
			shouldQueries = append(shouldQueries, map[string]interface{}{
				"match": map[string]interface{}{
					"CloudEvent": keyword,
				},
			})
		}
		filters = append(filters, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": shouldQueries,
			},
		})
	}

	// Build the search query
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": filters,
			},
		},
		"sort": []map[string]interface{}{
			{
				"ReceivedAt": map[string]interface{}{
					"order": "asc",
				},
			},
		},
	}

	// Apply limit
	if options.Limit > 0 {
		query["size"] = options.Limit
	}

	// Execute the search
	searchRes, err := store.client.Search(
		store.client.Search.WithContext(ctx),
		store.client.Search.WithIndex(store.eventIndex),
		store.client.Search.WithBody(bytes.NewReader(mustJSON(query))),
		store.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer searchRes.Body.Close()

	if searchRes.IsError() {
		responseBody, _ := io.ReadAll(searchRes.Body)
		return nil, fmt.Errorf("error executing search: %s, response: %s", searchRes.Status(), string(responseBody))
	}

	// Parse response
	var searchResult struct {
		Hits struct {
			Hits []struct {
				Source metastore.EventEntry `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(searchRes.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	// Extract events from the response
	events := make([]metastore.EventEntry, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		events = append(events, hit.Source)
	}

	return events, nil
}

func mustJSON(v interface{}) []byte {
	body, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal query to JSON: %v", err))
	}

	return body
}

var _ metastore.EventsStore = &EventStore{}

func (store *EventStore) ensureIndex(ctx context.Context) error {
	// Check if the index exists
	req := opensearchapi.IndicesExistsRequest{Index: []string{store.eventIndex}}
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
			"refresh_interval":   "1s",
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"ID":         map[string]interface{}{"type": "keyword"},
				"ReceivedAt": map[string]interface{}{"type": "date", "format": "epoch_millis"},
				"CloudEvent": map[string]interface{}{"type": "text"},
				"Namespace":  map[string]interface{}{"type": "keyword"},
				"Metadata":   map[string]interface{}{"type": "object"},
			},
		},
	}

	body, err := json.Marshal(indexSettings)
	if err != nil {
		return fmt.Errorf("failed to marshal index settings: %w", err)
	}

	// Create the index
	createReq := opensearchapi.IndicesCreateRequest{
		Index: store.eventIndex,
		Body:  bytes.NewReader(body),
	}

	createRes, err := createReq.Do(ctx, store.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		responseBody, _ := io.ReadAll(createRes.Body)
		return fmt.Errorf("error creating index: %s, response: %s", createRes.String(), string(responseBody))
	}

	return nil
}

func (store *EventStore) ensureDeletionPolicy(ctx context.Context) error {
	policy := map[string]interface{}{
		"policy": map[string]interface{}{
			"description":   "Event retention policy",
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

	body, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to marshal ISM policy: %w", err)
	}

	endpoint := fmt.Sprintf("/_plugins/_ism/policies/%s_policy", store.eventIndex)
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

	// Attach the ISM policy to the index
	attachPolicyBody := map[string]interface{}{
		"policy_id": fmt.Sprintf("%s_policy", store.eventIndex),
	}
	attachBody, err := json.Marshal(attachPolicyBody)
	if err != nil {
		return fmt.Errorf("failed to marshal attach policy body: %w", err)
	}

	attachEndpoint := fmt.Sprintf("/_plugins/_ism/add/%s", store.eventIndex)
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
