package opensearchstore

import (
	"context"
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

func NewMetaStore(ctx context.Context, client *opensearch.Client, co Config) (metastore.Store, error) {
	store := &opensearchMetaStore{
		client: client,
		co:     co,
	}
	if co.LogInit {
		err := store.LogStore().Init(ctx)
		if err != nil {
			return nil, err
		}
	}

	return store, nil
}

// LogStore implements metastore.Store.
func (o *opensearchMetaStore) LogStore() metastore.LogStore {
	return NewOpenSearchLogStore(o.client, o.co)
}

type Config struct {
	LogIndex       string
	LogDeleteAfter string
	LogInit        bool
}

func NewTestDataStore(t *testing.T) (metastore.Store, func(), error) {
	t.Helper()

	ctx := context.TODO()
	t.Log("Starting OpenSearch container...")
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
	t.Logf("OpenSearch container address: %s", address)

	client, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{
			address,
		},
	})
	if err != nil {
		return nil, cleanup, err
	}

	t.Log("OpenSearch client created successfully.")
	co := Config{
		LogIndex:       "test",
		LogDeleteAfter: "7d",
	}
	err = NewOpenSearchLogStore(client, co).Init(ctx)
	if err != nil {
		return nil, cleanup, err
	}

	return &opensearchMetaStore{
		client: client,
		co:     co,
	}, cleanup, nil
}
