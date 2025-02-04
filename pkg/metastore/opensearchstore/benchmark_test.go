package opensearchstore_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/metastore"
	"github.com/direktiv/direktiv/pkg/metastore/opensearchstore"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var globStore metastore.Store
var cleanup func()

func setupOpensearch(b *testing.B, reset bool) {
	var err error
	if reset || globStore == nil {
		globStore, cleanup, err = opensearchstore.NewTestDataStoreB(b)
		require.NoError(b, err)
	}
}
func BenchmarkAppendPerformance(b *testing.B) {
	setupOpensearch(b, false)

	now := time.Now()
	events := make([]metastore.EventEntry, 100)
	for i := 0; i < len(events); i++ {
		events[i] = metastore.EventEntry{
			ID:         uuid.NewString(),
			ReceivedAt: now.UnixMilli(),
			CloudEvent: fmt.Sprintf(`{"type":"benchmark.event", "source":"test", "data":"performance test %s"}`, uuid.NewString()),
			Namespace:  "default",
			Metadata:   map[string]string{"batch": "benchmark", "instanceID": uuid.NewString()},
		}
	}

	b.ResetTimer()

	// Increase b.N to perform more iterations for continuous load
	for i := 0; i < b.N; i++ {
		err := globStore.EventsStore().AppendBatch(context.Background(), events...)
		require.NoError(b, err)
	}

	// Calculate and log QPS and QPM
	qps := float64(b.N) / b.Elapsed().Seconds()
	qpm := qps * 60
	b.Logf("AppendBatch: %.2f QPS, %.2f QPM", qps, qpm)

	b.StopTimer()
}

func BenchmarkQueryPerformance(b *testing.B) {
	setupOpensearch(b, false)

	now := time.Now()

	for i := 0; i < 5000; i++ {
		err := globStore.EventsStore().Append(context.Background(), metastore.EventEntry{
			ID:         uuid.NewString(),
			ReceivedAt: now.UnixMilli(),
			CloudEvent: fmt.Sprintf(`{"type":"benchmark.event", "source":"test", "data":"performance test %s"}`, uuid.NewString()),
			Namespace:  "default",
			Metadata:   map[string]string{"batch": "benchmark", "instanceID": uuid.NewString()},
		})
		require.NoError(b, err)
	}

	time.Sleep(5 * time.Second)

	queryCases := []struct {
		name      string
		queryOpts metastore.EventQueryOptions
	}{
		{
			name: "Simple Time Range Query",
			queryOpts: metastore.EventQueryOptions{
				StartTime: now.Add(-time.Hour),
				EndTime:   now.Add(time.Hour),
			},
		},
		{
			name: "Complex Metadata Query",
			queryOpts: metastore.EventQueryOptions{
				StartTime: now.Add(-time.Hour),
				EndTime:   now.Add(time.Hour),
				Metadata:  map[string]string{"batch": "benchmark"},
			},
		},
	}

	for _, qc := range queryCases {
		b.Run(qc.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := globStore.EventsStore().Get(context.Background(), qc.queryOpts)
				require.NoError(b, err)
			}

			qps := float64(b.N) / b.Elapsed().Seconds()
			qpm := qps * 60
			b.Logf("%s: %.2f QPS, %.2f QPM", qc.name, qps, qpm)

			b.StopTimer()
		})
	}
}
func BenchmarkGetByIDPerformance(b *testing.B) {
	setupOpensearch(b, false)

	var eventID string
	{
		now := time.Now()
		event := metastore.EventEntry{
			ID:         uuid.NewString(),
			ReceivedAt: now.UnixMilli(),
			CloudEvent: fmt.Sprintf(`{"type":"benchmark.event", "source":"test", "data":"performance test %s"}`, uuid.NewString()),
			Namespace:  "default",
			Metadata:   map[string]string{"batch": "benchmark", "instanceID": uuid.NewString()},
		}
		err := globStore.EventsStore().Append(context.Background(), event)
		require.NoError(b, err)
		eventID = event.ID
	}

	time.Sleep(5 * time.Second)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := globStore.EventsStore().GetByID(context.Background(), eventID)
		require.NoError(b, err)
	}

	qps := float64(b.N) / b.Elapsed().Seconds()
	qpm := qps * 60
	b.Logf("GetByID: %.2f QPS, %.2f QPM", qps, qpm)

	b.StopTimer()
}

func cleanupOpensearch() {
	if cleanup != nil {
		cleanup()
	}
}
