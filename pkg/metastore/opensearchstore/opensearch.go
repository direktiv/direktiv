package opensearchstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/direktiv/direktiv/pkg/metastore"
	"github.com/opensearch-project/opensearch-go"
	"github.com/testcontainers/testcontainers-go"
	tcopensearch "github.com/testcontainers/testcontainers-go/modules/opensearch"
)

var _ metastore.Store = &opensearchMetaStore{}

type opensearchMetaStore struct {
	client *opensearch.Client
	co     Config
}

// TimelineStore implements metastore.Store.
func (o *opensearchMetaStore) TimelineStore() metastore.TimelineStore {
	return NewTimelinesStore(o.client, o.co)
}

func NewMetaStore(ctx context.Context, client *opensearch.Client, co Config) (metastore.Store, error) {
	store := &opensearchMetaStore{
		client: client,
		co:     co,
	}
	if co.EventsInit {
		err := store.EventsStore().Init(ctx)
		if err != nil {
			return nil, err
		}
	}
	if co.MetricsInit {
		err := store.MetricsStore().Init(ctx)
		if err != nil {
			return nil, err
		}
	}
	if co.LogInit {
		err := store.LogStore().Init(ctx)
		if err != nil {
			return nil, err
		}
	}

	return store, nil
}

// EventsStore implements metastore.Store.
func (o *opensearchMetaStore) EventsStore() metastore.EventsStore {
	return NewOpenSearchEventsStore(o.client, o.co)
}

func (o *opensearchMetaStore) LogStore() metastore.LogStore {
	return NewLogStore(o.client, o.co)
}

func (o *opensearchMetaStore) MetricsStore() metastore.MetricsStore {
	return NewMetricsStore(o.client, o.co)
}

type Config struct {
	LogIndex            string
	LogDeleteAfter      string
	LogInit             bool
	EventsIndex         string
	EventsDeleteAfter   string
	EventsInit          bool
	TimelineIndex       string
	TimelineDeleteAfter string
	MetricsIndex        string
	MetricsInit         bool
	MetricsDeleteAfter  string
}

func NewTestDataStore(t *testing.T) (metastore.Store, func(), error) {
	t.Helper()

	ctx := context.TODO()
	t.Log("starting OpenSearch container...")
	ctr, err := tcopensearch.Run(ctx, "opensearchproject/opensearch:2.11.1")
	if err != nil {
		return nil, func() {}, err
	}
	cleanup := func() {
		// t.Log("Cleaning up container...")
		testcontainers.CleanupContainer(t, ctr)
	}
	address, err := ctr.Address(ctx)
	if err != nil {
		return nil, cleanup, err
	}
	t.Logf("openSearch container address: %s", address)

	client, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{
			address,
		},
	})
	if err != nil {
		return nil, cleanup, err
	}

	t.Log("openSearch client created successfully.")
	co := Config{
		EventsIndex:       "test-events",
		EventsDeleteAfter: "7d",
	}
	err = NewOpenSearchEventsStore(client, co).Init(ctx)
	if err != nil {
		return nil, cleanup, err
	}

	return &opensearchMetaStore{
		client: client,
		co:     co,
	}, cleanup, nil
}

func ensureISMPolicy(ctx context.Context, client *opensearch.Client, policyName, indexName, retentionPeriod string) error {
	slog.Debug("defining ISM policy", "index", indexName)
	policy := map[string]interface{}{
		"policy": map[string]interface{}{
			"description":   "Auto-delete old data",
			"default_state": "hot",
			"states": []map[string]interface{}{
				{
					"name": "hot",
					"actions": []map[string]interface{}{
						{
							"rollover": map[string]interface{}{
								//	"min_size":      "50gb", // Optional, use if needed
								"min_index_age": retentionPeriod,
							},
						},
					},
					"transitions": []map[string]interface{}{
						{
							"state_name": "delete",
							"conditions": map[string]interface{}{
								"min_index_age": retentionPeriod,
							},
						},
					},
				},
				{
					"name": "delete",
					"actions": []map[string]interface{}{
						{"delete": map[string]interface{}{}},
					},
				},
			},
		},
	}

	// Convert policy to JSON
	body, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to marshal ISM policy %s: %w", indexName, err)
	}

	slog.Debug("creating ISM policy for logs in OpenSearch")
	policyEndpoint := fmt.Sprintf("/_plugins/_ism/policies/%s", policyName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, policyEndpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create ISM policy request %s: %w", indexName, err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Transport.Perform(req)
	if err != nil {
		return fmt.Errorf("failed to execute ISM policy request %s: %w", indexName, err)
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)
	if res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("error applying ISM policy %s: %s, response: %s", indexName, res.Status, string(bodyBytes))
	}

	slog.Info("ISM policy applied successfully")

	// Attach the ISM policy to the index alias
	attachPolicyBody := map[string]interface{}{
		"policy_id": policyName,
	}
	attachBody, err := json.Marshal(attachPolicyBody)
	if err != nil {
		return fmt.Errorf("failed to marshal attach policy %s body: %w", indexName, err)
	}

	attachEndpoint := fmt.Sprintf("/_plugins/_ism/add/%s", indexName)
	attachReq, err := http.NewRequestWithContext(ctx, http.MethodPost, attachEndpoint, bytes.NewReader(attachBody))
	if err != nil {
		return fmt.Errorf("failed to create attach policy %s request: %w", indexName, err)
	}
	attachReq.Header.Set("Content-Type", "application/json")

	attachRes, err := client.Transport.Perform(attachReq)
	if err != nil {
		return fmt.Errorf("failed to execute attach policy %s request: %w", indexName, err)
	}
	defer attachRes.Body.Close()

	attachBodyBytes, _ := io.ReadAll(attachRes.Body)
	if attachRes.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("error attaching %s ISM policy: %s, response: %s", indexName, attachRes.Status, string(attachBodyBytes))
	}
	slog.Info("ISM policy successfully attached", "index", indexName)

	return nil
}

func checkAndDeleteISMPolicy(ctx context.Context, client *opensearch.Client, policyName string, deleteIfExists bool) error {
	slog.Debug("checking if ISM policy exists", "policy", policyName)

	policyEndpoint := fmt.Sprintf("/_plugins/_ism/policies/%s", policyName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, policyEndpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create ISM policy check request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Transport.Perform(req)
	if err != nil {
		return fmt.Errorf("failed to execute ISM policy check request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		slog.Debug("ISM policy does not exist", "policy", policyName)
		return nil
	}

	if res.StatusCode >= http.StatusBadRequest {
		bodyBytes, _ := io.ReadAll(res.Body)
		return fmt.Errorf("error checking ISM policy: %s, response: %s", res.Status, string(bodyBytes))
	}

	slog.Debug("ISM policy exists", "policy", policyName)

	if deleteIfExists {
		slog.Debug("deleting existing ISM policy", "policy", policyName)
		delReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, policyEndpoint, nil)
		if err != nil {
			return fmt.Errorf("failed to create ISM policy delete request: %w", err)
		}
		delReq.Header.Set("Content-Type", "application/json")

		delRes, err := client.Transport.Perform(delReq)
		if err != nil {
			return fmt.Errorf("failed to execute ISM policy delete request: %w", err)
		}
		defer delRes.Body.Close()

		if delRes.StatusCode >= http.StatusBadRequest {
			bodyBytes, _ := io.ReadAll(delRes.Body)
			return fmt.Errorf("error deleting ISM policy: %s, response: %s", delRes.Status, string(bodyBytes))
		}

		slog.Info("ISM policy deleted successfully", "policy", policyName)

		return nil
	}

	return nil
}

// GetMapping retrieves the mapping of the specified index.
func (o *opensearchMetaStore) GetMapping(ctx context.Context, index string) (map[string]interface{}, error) {
	// Make a request to OpenSearch to get the mapping
	mappingRes, err := o.client.Indices.GetMapping(
		o.client.Indices.GetMapping.WithContext(ctx),
		o.client.Indices.GetMapping.WithIndex(index),
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
