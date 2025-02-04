package opensearchstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/direktiv/direktiv/pkg/metastore"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

const eventsISMPolicyName = "direktiv-events-policy"

const maxBatchSize = 500

// NewOpenSearchEventsStore creates a new OpenSearchLogStore with the specified settings.
func NewOpenSearchEventsStore(client *opensearch.Client, co Config) *EventStore {
	return &EventStore{
		client:      client,
		Index:       co.EventsIndex,
		deleteAfter: co.EventsDeleteAfter,
	}
}

// Init ensures the index and lifecycle policies are created.
func (store *EventStore) Init(ctx context.Context) error {
	// Ensure the index exists
	if err := store.ensureIndex(ctx); err != nil {
		return fmt.Errorf("failed to ensure index: %w", err)
	}
	err := checkAndDeleteISMPolicy(ctx, store.client, logISMPolicyName, true)
	if err != nil {
		return err
	}
	// Ensure lifecycle policies
	if err := ensureISMPolicy(ctx, store.client, eventsISMPolicyName, store.Index, store.deleteAfter); err != nil {
		return fmt.Errorf("failed to ensure deletion policy: %w", err)
	}

	return nil
}

type EventStore struct {
	client      *opensearch.Client
	Index       string
	deleteAfter string // e.g., "30d" for 30 days
}

func (store *EventStore) Append(ctx context.Context, e metastore.EventEntry) error {
	body, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	req := opensearchapi.IndexRequest{
		Index:      store.Index,
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

func (store *EventStore) AppendBatch(ctx context.Context, events ...metastore.EventEntry) error {
	var buf bytes.Buffer
	if len(events) > maxBatchSize {
		return fmt.Errorf("batch bigger then the maxium batch size")
	}
	for _, e := range events {
		meta := map[string]interface{}{"index": map[string]interface{}{"_index": store.Index}}
		metaBytes, _ := json.Marshal(meta)
		eventBytes, _ := json.Marshal(e)
		buf.Write(metaBytes)
		buf.WriteByte('\n')
		buf.Write(eventBytes)
		buf.WriteByte('\n')
	}

	req := opensearchapi.BulkRequest{
		Body: bytes.NewReader(buf.Bytes()),
	}

	res, err := req.Do(ctx, store.client)
	if err != nil {
		return fmt.Errorf("bulk index failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		responseBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bulk index error: %s, response: %s", res.Status(), string(responseBody))
	}

	return nil
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
		store.client.Search.WithIndex(store.Index),
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

// GetByID retrieves a single event by its unique ID
func (store *EventStore) GetByID(ctx context.Context, id string) (metastore.EventEntry, error) {
	var event metastore.EventEntry

	res, err := store.client.Get(store.Index, id)
	if err != nil {
		return event, fmt.Errorf("error retrieving event by ID: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return event, fmt.Errorf("event not found")
	}
	if res.StatusCode != http.StatusOK {
		return event, fmt.Errorf("unexpected response status: %d", res.StatusCode)
	}

	var response struct {
		Source metastore.EventEntry `json:"_source"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return event, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Source, nil
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
	req := opensearchapi.IndicesExistsRequest{Index: []string{store.Index}}
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
		Index: store.Index,
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
