package opensearch_test

import (
	"context"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/metastore"
	"github.com/direktiv/direktiv/pkg/metastore/opensearch"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOpenSearchLogsStore(t *testing.T) {
	// Create a new test data store
	store, cleanup, err := opensearch.NewTestDataStore(t)
	defer cleanup()
	require.NoError(t, err)

	now := time.Now()
	id := uuid.NewString()
	err = store.LogStore().Append(context.Background(), metastore.LogEntry{
		ID:        id,
		Timestamp: now.UnixMilli(),
		Level:     "ERROR",
		Message:   "test log message",
	})
	require.NoError(t, err)
	time.Sleep(time.Second * 2)
	logs, err := store.LogStore().Get(context.Background(), metastore.LogQueryOptions{
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		//Levels:    []string{"ERROR"},
	})

	require.Len(t, logs, 1)

	require.ElementsMatch(t, []metastore.LogEntry{{
		ID:        id,
		Timestamp: now.UnixMilli(),
		Level:     "ERROR",
		Message:   "test log message",
	}}, logs)
}
