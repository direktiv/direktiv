package opensearchstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/direktiv/direktiv/pkg/metastore"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type Timelines struct {
	client *opensearch.Client
	index  string
}

func NewTimelinesStore(client *opensearch.Client, co Config) *Timelines {
	return &Timelines{
		client: client,
		index:  co.TimelineIndex,
	}
}

func (t *Timelines) Get(ctx context.Context, traceID string, options metastore.TimelineQueryOptions) ([]map[string]any, error) {
	mustClauses := []map[string]interface{}{
		{"match": map[string]interface{}{"traceId": traceID}},
	}

	// Add time range filtering
	if !options.StartTime.IsZero() || !options.EndTime.IsZero() {
		timeRange := map[string]interface{}{}
		if !options.StartTime.IsZero() {
			timeRange["gte"] = options.StartTime.Format(time.RFC3339Nano)
		}
		if !options.EndTime.IsZero() {
			timeRange["lte"] = options.EndTime.Format(time.RFC3339Nano)
		}
		mustClauses = append(mustClauses, map[string]interface{}{
			"range": map[string]interface{}{
				"startTime": timeRange,
			},
		})
	}

	// Add span ID filter if provided
	if options.SpanID != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"match": map[string]interface{}{"spanId": options.SpanID},
		})
	}

	// Add additional metadata filters
	for key, value := range options.Metadata {
		// Match attributes either under `span.attributes` or `resource.attributes`
		attrQuery := map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"match": map[string]interface{}{"span.attributes." + key: value}},
					{"match": map[string]interface{}{"resource.attributes." + key: value}},
				},
				"minimum_should_match": 1,
			},
		}
		mustClauses = append(mustClauses, attrQuery)
	}
	if options.Limit == 0 {
		options.Limit = 10000
	}
	// Construct final query
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustClauses,
			},
		},
		"size": options.Limit,
	}

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req := opensearchapi.SearchRequest{
		Index: []string{t.index},
		Body:  bytes.NewReader(body),
	}

	resp, err := req.Do(ctx, t.client)
	if err != nil {
		slog.Error("Error executing OpenSearch query", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		responseBody, _ := io.ReadAll(resp.Body)
		slog.Error("OpenSearch search error", "status", resp.Status(), "response", string(responseBody))

		return nil, fmt.Errorf("error searching OpenSearch: %s, response: %s", resp.Status(), string(responseBody))
	}

	var result struct {
		Hits struct {
			Hits []struct {
				Source map[string]any `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Error("Error decoding OpenSearch response", "error", err)
		return nil, err
	}

	serviceMapData := make([]map[string]any, len(result.Hits.Hits))
	for i, hit := range result.Hits.Hits {
		serviceMapData[i] = hit.Source
	}

	return serviceMapData, nil
}

func (t *Timelines) GetMapping(ctx context.Context) (map[string]interface{}, error) {
	mappingRes, err := t.client.Indices.GetMapping(
		t.client.Indices.GetMapping.WithContext(ctx),
		t.client.Indices.GetMapping.WithIndex(t.index),
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

	var mapping map[string]interface{}
	if err := json.NewDecoder(mappingRes.Body).Decode(&mapping); err != nil {
		return nil, fmt.Errorf("failed to decode mapping response: %w", err)
	}

	return mapping, nil
}
