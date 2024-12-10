package opensearch_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/metastore/opensearch"
	"github.com/stretchr/testify/require"
)

func TestOpenSearchMetaStore(t *testing.T) {
	// Create a new test data store
	_, cleanup, err := opensearch.NewTestDataStore(t)
	defer cleanup()
	require.NoError(t, err)
}
