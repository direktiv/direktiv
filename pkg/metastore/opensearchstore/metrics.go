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
)

const metricsISMPolicyName = "otel-metrics-policy"

const maxRetries = 3

const retryDelay = 2 * time.Second

type MetricsStore struct {
	client      *opensearch.Client
	index       string
	deleteAfter string // e.g., "30d" for 30 days
}

func NewMetricsStore(client *opensearch.Client, co Config) metastore.MetricsStore {
	return &MetricsStore{
		client:      client,
		index:       co.MetricsIndex,
		deleteAfter: co.MetricsDeleteAfter,
	}
}

func (m *MetricsStore) Init(ctx context.Context) error {
	var err error
	for i := 1; i <= maxRetries; i++ {
		err = checkAndDeleteISMPolicy(ctx, m.client, logISMPolicyName, true)
		if err != nil {
			continue
		}
		err = ensureISMPolicy(ctx, m.client, metricsISMPolicyName, m.index, m.deleteAfter)
		if err == nil {
			return nil
		}
		slog.Warn(fmt.Sprintf("Failed to initialize ISM policy (attempt %d/%d): %v", i, maxRetries, err))
		time.Sleep(retryDelay) // Wait before retrying
	}

	return fmt.Errorf("failed to initialize ISM policy after %d attempts: %w", maxRetries, err)
}

func (m *MetricsStore) Get(ctx context.Context, metricName string, options metastore.MetricsQueryOptions) ([]map[string]any, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"name.keyword": metricName, // Correct field for metric name
						},
					},
				},
			},
		},
	}

	// Apply limit if specified
	if options.Limit > 0 {
		query["size"] = options.Limit
	}

	// Execute the OpenSearch query
	searchRes, err := m.client.Search(
		m.client.Search.WithContext(ctx),
		m.client.Search.WithIndex(m.index),
		m.client.Search.WithBody(bytes.NewReader(mustJSON(query))),
		m.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer searchRes.Body.Close()

	// Handle OpenSearch errors
	if searchRes.IsError() {
		responseBody, _ := io.ReadAll(searchRes.Body)
		slog.Error("search failed, status: %s, response: %s", searchRes.Status(), string(responseBody))

		return nil, fmt.Errorf("error executing search: %s, response: %s", searchRes.Status(), string(responseBody))
	}

	// Parse the response
	var searchResult struct {
		Hits struct {
			Hits []struct {
				Source map[string]any `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(searchRes.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	// Extract metrics from the response
	metrics := make([]map[string]any, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		metrics = append(metrics, hit.Source)
	}

	return metrics, nil
}

func (m *MetricsStore) GetAll(ctx context.Context, limit int) ([]string, error) {
	query := map[string]interface{}{
		"size": 0, // Don't return documents, just aggregations
		"aggs": map[string]interface{}{
			"unique_metric_names": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "name.keyword", // Correct field for metric names
					"size":  limit,          // Adjust based on expected number of metrics
				},
			},
		},
	}

	searchRes, err := m.client.Search(
		m.client.Search.WithContext(ctx),
		m.client.Search.WithIndex(m.index),
		m.client.Search.WithBody(bytes.NewReader(mustJSON(query))),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer searchRes.Body.Close()

	if searchRes.IsError() {
		responseBody, _ := io.ReadAll(searchRes.Body)
		slog.Error("search failed, status: %s, response: %s", searchRes.Status(), string(responseBody))

		return nil, fmt.Errorf("error executing search: %s, response: %s", searchRes.Status(), string(responseBody))
	}

	// Parse the response
	var searchResult struct {
		Aggregations struct {
			UniqueMetricNames struct {
				Buckets []struct {
					Key string `json:"key"`
				} `json:"buckets"`
			} `json:"unique_metric_names"`
		} `json:"aggregations"`
	}

	if err := json.NewDecoder(searchRes.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode aggregation response: %w", err)
	}

	// Extract metric names
	metricNames := make([]string, 0, len(searchResult.Aggregations.UniqueMetricNames.Buckets))
	for _, bucket := range searchResult.Aggregations.UniqueMetricNames.Buckets {
		metricNames = append(metricNames, bucket.Key)
	}

	return metricNames, nil
}
