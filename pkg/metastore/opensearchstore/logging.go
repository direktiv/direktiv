package opensearchstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/direktiv/direktiv/pkg/metastore"
	"github.com/opensearch-project/opensearch-go"
)

type LogStore struct {
	client   *opensearch.Client
	logIndex string
}

func NewLogStore(client *opensearch.Client, co Config) *LogStore {
	return &LogStore{
		client:   client,
		logIndex: co.LogIndex,
	}
}

func (store *LogStore) Get(ctx context.Context, options metastore.LogQueryOptions) ([]metastore.LogEntry, error) {
	// Construct the query filters
	filters := []map[string]interface{}{
		{
			"range": map[string]interface{}{
				"time": map[string]interface{}{
					"gte": options.StartTime.UTC().Format(time.RFC3339),
					"lte": options.EndTime.UTC().Format(time.RFC3339),
				},
			},
		},
	}

	// Add level filter (use keyword subfield)
	if options.Level != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"level.keyword": options.Level,
			},
		})
	}

	// Add metadata filters
	for key, value := range options.Metadata {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				fmt.Sprintf("%s.keyword", key): value,
			},
		})
	}

	// Add keyword (full-text search on msg field)
	if len(options.Keywords) > 0 {
		shouldQueries := []map[string]interface{}{}
		for _, keyword := range options.Keywords {
			shouldQueries = append(shouldQueries, map[string]interface{}{
				"match": map[string]interface{}{
					"msg": keyword,
				},
			})
		}
		filters = append(filters, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": shouldQueries,
			},
		})
	}

	// Build the OpenSearch query
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": filters,
			},
		},
		"sort": []map[string]interface{}{
			{
				"time": map[string]interface{}{
					"order": "asc",
				},
			},
		},
	}

	// Apply limit to the query
	if options.Limit > 0 {
		query["size"] = options.Limit
	}

	// Execute the OpenSearch query
	searchRes, err := store.client.Search(
		store.client.Search.WithContext(ctx),
		store.client.Search.WithIndex(store.logIndex),
		store.client.Search.WithBody(bytes.NewReader(mustJSON(query))),
		store.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer searchRes.Body.Close()

	// Handle errors
	if searchRes.IsError() {
		responseBody, _ := io.ReadAll(searchRes.Body)
		log.Printf("search failed, status: %s, response: %s", searchRes.Status(), string(responseBody))

		return nil, fmt.Errorf("error executing search: %s, response: %s", searchRes.Status(), string(responseBody))
	}

	// Parse the response
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

	// Extract logs from the response
	logs := make([]metastore.LogEntry, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		logs = append(logs, hit.Source)
	}

	return logs, nil
}
