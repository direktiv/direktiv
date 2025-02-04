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

func TestSimpleOpenSearchEventStore(t *testing.T) {
	// Create a new test data store
	store, cleanup, err := opensearchstore.NewTestDataStore(t)
	defer cleanup()
	require.NoError(t, err)

	now := time.Now()
	id := uuid.NewString()

	err = store.EventsStore().Append(context.Background(), metastore.EventEntry{
		ID:         id,
		ReceivedAt: now.UnixMilli(),
		CloudEvent: `{"type":"test.event", "source":"test", "data":"hello world"}`,
		Namespace:  "default",
		Metadata:   map[string]string{"key1": "value1"},
	})
	require.NoError(t, err)

	// Allow OpenSearch some time to index the event
	time.Sleep(time.Second * 2)

	events, err := store.EventsStore().Get(context.Background(), metastore.EventQueryOptions{
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
	})
	require.NoError(t, err)
	require.Len(t, events, 1)

	require.ElementsMatch(t, []metastore.EventEntry{{
		ID:         id,
		ReceivedAt: now.UnixMilli(),
		CloudEvent: `{"type":"test.event", "source":"test", "data":"hello world"}`,
		Namespace:  "default",
		Metadata:   map[string]string{"key1": "value1"},
	}}, events)
}

func TestOpenSearchEventStore(t *testing.T) {
	// Create a new test data store
	store, cleanup, err := opensearchstore.NewTestDataStore(t)
	defer cleanup()
	require.NoError(t, err)

	now := time.Now()
	thirdEventTimeStamp := now.Add(time.Minute)

	// Append multiple events
	eventEntries := []metastore.EventEntry{
		{
			ID:         uuid.NewString(),
			ReceivedAt: now.Add(-time.Minute).UnixMilli(),
			CloudEvent: `{"type":"event.first", "source":"service1", "data":"first test event"}`,
			Namespace:  "default",
			Metadata:   map[string]string{"service": "auth", "env": "prod"},
		},
		{
			ID:         uuid.NewString(),
			ReceivedAt: now.UnixMilli(),
			CloudEvent: `{"type":"event.second", "source":"service2", "data":"second event with keyword"}`,
			Namespace:  "default",
			Metadata:   map[string]string{"service": "billing", "env": "dev"},
		},
		{
			ID:         uuid.NewString(),
			ReceivedAt: thirdEventTimeStamp.UnixMilli(),
			CloudEvent: `{"type":"event.third", "source":"service3", "data":"third event"}`,
			Namespace:  "default",
			Metadata:   map[string]string{"service": "auth", "env": "staging"},
		},
	}

	for _, entry := range eventEntries {
		err := store.EventsStore().Append(context.Background(), entry)
		require.NoError(t, err)
	}

	// Allow indexing time
	time.Sleep(time.Second * 2)

	// Retrieve all events within a time range
	events, err := store.EventsStore().Get(context.Background(), metastore.EventQueryOptions{
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
	})
	require.NoError(t, err)
	require.Len(t, events, 3)

	// Test retrieval by metadata filter
	events, err = store.EventsStore().Get(context.Background(), metastore.EventQueryOptions{
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		Metadata:  map[string]string{"service": "auth"},
	})
	require.NoError(t, err)
	require.Len(t, events, 2)

	// Test retrieval by keywords
	events, err = store.EventsStore().Get(context.Background(), metastore.EventQueryOptions{
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		Keywords:  []string{"keyword"},
	})
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.JSONEq(t, `{"type":"event.second", "source":"service2", "data":"second event with keyword"}`, events[0].CloudEvent)

	// Test retrieval by limit (fetch only 1 event)
	events, err = store.EventsStore().Get(context.Background(), metastore.EventQueryOptions{
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		Limit:     1,
	})
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.JSONEq(t, `{"type":"event.first", "source":"service1", "data":"first test event"}`, events[0].CloudEvent)

	// Test retrieval of last event by timestamp
	events, err = store.EventsStore().Get(context.Background(), metastore.EventQueryOptions{
		StartTime: thirdEventTimeStamp,
		EndTime:   now.Add(time.Hour),
		Limit:     1,
	})
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.JSONEq(t, `{"type":"event.third", "source":"service3", "data":"third event"}`, events[0].CloudEvent)
}

func TestOpenSearchEventStore_AppendBatch(t *testing.T) {
	// Create a new test data store
	store, cleanup, err := opensearchstore.NewTestDataStore(t)
	defer cleanup()
	require.NoError(t, err)

	now := time.Now()

	// Prepare multiple events for bulk append
	eventEntries := []metastore.EventEntry{
		{
			ID:         uuid.NewString(),
			ReceivedAt: now.Add(-time.Minute).UnixMilli(),
			CloudEvent: `{"type":"bulk.event.first", "source":"service1", "data":"first bulk event"}`,
			Namespace:  "default",
			Metadata:   map[string]string{"batch": "1"},
		},
		{
			ID:         uuid.NewString(),
			ReceivedAt: now.UnixMilli(),
			CloudEvent: `{"type":"bulk.event.second", "source":"service2", "data":"second bulk event"}`,
			Namespace:  "default",
			Metadata:   map[string]string{"batch": "1"},
		},
		{
			ID:         uuid.NewString(),
			ReceivedAt: now.Add(time.Minute).UnixMilli(),
			CloudEvent: `{"type":"bulk.event.third", "source":"service3", "data":"third bulk event"}`,
			Namespace:  "default",
			Metadata:   map[string]string{"batch": "1"},
		},
	}

	// Append events in bulk
	err = store.EventsStore().AppendBatch(context.Background(), eventEntries...)
	require.NoError(t, err)

	// Allow OpenSearch some time to index the events
	time.Sleep(time.Second * 2)

	// Retrieve all events within a time range
	events, err := store.EventsStore().Get(context.Background(), metastore.EventQueryOptions{
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
	})
	require.NoError(t, err)
	require.Len(t, events, 3)

	// Validate that all appended events exist
	require.ElementsMatch(t, eventEntries, events)
}

func TestOpenSearchEventStore_GetByID(t *testing.T) {
	// Create a new test data store
	store, cleanup, err := opensearchstore.NewTestDataStore(t)
	defer cleanup()
	require.NoError(t, err)

	now := time.Now()
	eventID := uuid.NewString()

	// Append an event
	eventEntry := metastore.EventEntry{
		ID:         eventID,
		ReceivedAt: now.UnixMilli(),
		CloudEvent: `{"type":"test.getbyid", "source":"test", "data":"retrieve me"}`,
		Namespace:  "default",
		Metadata:   map[string]string{"key": "value"},
	}

	err = store.EventsStore().Append(context.Background(), eventEntry)
	require.NoError(t, err)

	// Allow OpenSearch some time to index the event
	time.Sleep(time.Second * 2)

	// Retrieve the event by ID
	retrievedEvent, err := store.EventsStore().GetByID(context.Background(), eventID)
	require.NoError(t, err)

	// Validate that the retrieved event matches the stored event
	require.Equal(t, eventEntry, retrievedEvent)
}
