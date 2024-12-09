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
	ctx := context.Background()

	ctr, err := tcopensearch.Run(ctx, "opensearchproject/opensearch:2.11.1")
	if err != nil {
		panic(err.Error())
	}
	cleanup := func() {
		testcontainers.CleanupContainer(t, ctr)
	}
	address, err := ctr.Address(ctx)
	if err != nil {
		panic(err.Error())
	}
	client, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{
			address,
		},
	})
	if err != nil {
		panic(err.Error())
	}

	return &opensearchMetaStore{
		client: client,
	}, cleanup, nil
}
