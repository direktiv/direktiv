package tracing_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/metastore"
	"github.com/direktiv/direktiv/pkg/metastore/opensearchstore"
	"github.com/direktiv/direktiv/pkg/tracing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWorker_BatchFlush(t *testing.T) {
	store, cleanup, err := opensearchstore.NewTestDataStore(t)
	defer cleanup()
	require.NoError(t, err)

	// Create log channel and worker
	logCh := make(chan metastore.LogEntry, 10)
	worker := tracing.NewWorker(tracing.WorkerArgs{
		LogCh:         logCh,
		LogStore:      store.LogStore(),
		MaxBatchSize:  1,
		FlushInterval: 1 * time.Millisecond,
		CachedLevel:   int(slog.LevelDebug),
	})

	circuit := core.NewCircuit(context.Background(), os.Interrupt)

	circuit.Start(func() error {
		err := worker.Start(circuit)
		if err != nil {
			return fmt.Errorf("logs worker, err: %w", err)
		}

		return nil
	})

	// Send logs to the worker
	logCh <- metastore.LogEntry{Message: "log 1", Level: int(slog.LevelInfo), Metadata: map[string]string{"test-key": "test-value"}, Timestamp: time.Now().UnixMilli(), ID: uuid.NewString()}
	logCh <- metastore.LogEntry{Message: "log 2", Level: int(slog.LevelInfo), Metadata: map[string]string{"test-key": "test-value"}, Timestamp: time.Now().UnixMilli(), ID: uuid.NewString()}

	// Ensure that the logs were flushed
	time.Sleep(2 * time.Second)
	logs, err := store.LogStore().Get(context.Background(), metastore.LogQueryOptions{
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   time.Now().Add(time.Hour),
	})
	require.NoError(t, err)
	require.Len(t, logs, 2)
	// Close the channel to trigger final flush
	close(logCh)
}
