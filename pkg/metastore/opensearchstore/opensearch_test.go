package opensearchstore_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/metastore/opensearchstore"
	"github.com/stretchr/testify/require"
)

func TestOpenSearchMetaStore(t *testing.T) {
	// Create a new test data store
	_, cleanup, err := opensearchstore.NewTestDataStore(t)
	defer cleanup()
	require.NoError(t, err)
}
