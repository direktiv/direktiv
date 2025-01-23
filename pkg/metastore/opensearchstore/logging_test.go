package opensearchstore_test

import (
	"context"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/metastore"
	"github.com/direktiv/direktiv/pkg/metastore/opensearchstore"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSimpleOpenSearchLogsStore(t *testing.T) {
	// Create a new test data store
	store, cleanup, err := opensearchstore.NewTestDataStore(t)
	defer cleanup()
	require.NoError(t, err)

	now := time.Now()
	id := uuid.NewString()
	err = store.LogStore().Append(context.Background(), metastore.LogEntry{
		ID:        id,
		Timestamp: now.UnixMilli(),
		Level:     8,
		Message:   "test log message",
	})
	require.NoError(t, err)
	time.Sleep(time.Second * 2)
	logs, err := store.LogStore().Get(context.Background(), metastore.LogQueryOptions{
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		Level:     8,
	})

	require.Len(t, logs, 1)

	require.ElementsMatch(t, []metastore.LogEntry{{
		ID:        id,
		Timestamp: now.UnixMilli(),
		Level:     8,
		Message:   "test log message",
	}}, logs)
}

func TestOpenSearchLogsStore(t *testing.T) {
	// Create a new test data store
	store, cleanup, err := opensearchstore.NewTestDataStore(t)
	defer cleanup()
	require.NoError(t, err)

	now := time.Now()
	thirgLogEntryTimeStamp := now.Add(time.Minute)
	// Append multiple log entries
	logEntries := []metastore.LogEntry{
		{
			ID:        uuid.NewString(),
			Timestamp: now.Add(-time.Minute).UnixMilli(),
			Level:     8,
			Message:   "first test log message",
			Metadata:  map[string]string{"service": "auth", "env": "prod"},
		},
		{
			ID:        uuid.NewString(),
			Timestamp: now.UnixMilli(),
			Level:     6,
			Message:   "second test log with keyword",
			Metadata:  map[string]string{"service": "billing", "env": "dev"},
		},
		{
			ID:        uuid.NewString(),
			Timestamp: thirgLogEntryTimeStamp.UnixMilli(),
			Level:     4,
			Message:   "third log message",
			Metadata:  map[string]string{"service": "auth", "env": "staging"},
		},
	}

	for _, entry := range logEntries {
		err := store.LogStore().Append(context.Background(), entry)
		require.NoError(t, err)
	}
	logs, err := store.LogStore().Get(context.Background(), metastore.LogQueryOptions{
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
	})
	require.NoError(t, err)
	require.Len(t, logs, 3)
	// Test retrieval by time range and level
	logs, err = store.LogStore().Get(context.Background(), metastore.LogQueryOptions{
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		Level:     8,
	})
	require.NoError(t, err)
	require.Len(t, logs, 1)

	// Test retrieval by time range and level
	logs, err = store.LogStore().Get(context.Background(), metastore.LogQueryOptions{
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		Level:     4,
	})
	require.NoError(t, err)
	require.Len(t, logs, 3)

	// Test retrieval by metadata
	logs, err = store.LogStore().Get(context.Background(), metastore.LogQueryOptions{
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		Metadata:  map[string]string{"service": "auth"},
	})
	require.NoError(t, err)
	require.Len(t, logs, 2)

	// Test retrieval by keywords
	logs, err = store.LogStore().Get(context.Background(), metastore.LogQueryOptions{
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		Keywords:  []string{"keyword"},
	})
	require.NoError(t, err)
	require.Len(t, logs, 1)
	require.Equal(t, "second test log with keyword", logs[0].Message)

	// Test retrieval by Limit
	logs, err = store.LogStore().Get(context.Background(), metastore.LogQueryOptions{
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		Limit:     1,
	})
	require.NoError(t, err)
	require.Len(t, logs, 1)
	require.Equal(t, "first test log message", logs[0].Message)

	// Test retrieval by offset and Limit
	logs, err = store.LogStore().Get(context.Background(), metastore.LogQueryOptions{
		StartTime: thirgLogEntryTimeStamp,
		EndTime:   now.Add(time.Hour),
		Limit:     1,
	})
	require.NoError(t, err)
	require.Len(t, logs, 1)
	require.Equal(t, "third log message", logs[0].Message)
}
