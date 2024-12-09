package opensearch

import (
	"context"
	"testing"

	"github.com/direktiv/direktiv/pkg/metastore"
	"github.com/opensearch-project/opensearch-go"
	"github.com/testcontainers/testcontainers-go"
	tcopensearch "github.com/testcontainers/testcontainers-go/modules/opensearch"
)

type opensearchMetaStore struct {
	client *opensearch.Client
}

func NewTestDataStore(t *testing.T) (metastore.Store, func(), error) {
	t.Helper()

	ctx := context.TODO()
	t.Log("Starting OpenSearch container...")
	ctr, err := tcopensearch.Run(ctx, "opensearchproject/opensearch:2.11.1")
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		t.Log("Cleaning up container...")
		testcontainers.CleanupContainer(t, ctr)
	}
	address, err := ctr.Address(ctx)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	t.Logf("OpenSearch container address: %s", address)

	client, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{
			address,
		},
	})
	if err != nil {
		cleanup()
		return nil, nil, err
	}

	t.Log("OpenSearch client created successfully.")

	return &opensearchMetaStore{
		client: client,
	}, cleanup, nil
}
