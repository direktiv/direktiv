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
	"github.com/stretchr/testify/require"
)

// TestWithGroup checks if a new handler with the new group is created
func TestIntegrated(t *testing.T) {
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

	circuit.Start(func() error {
		err := worker.Start(circuit)
		if err != nil {
			return fmt.Errorf("logs worker, err: %w", err)
		}

		return nil
	})

	// Create handlers
	jsonHandler := tracing.NewContextHandler(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	channelHandler := tracing.NewChannelHandler(logCh, nil, "default", slog.LevelDebug)

	// Combine handlers using a TeeHandler
	compositeHandler := tracing.TeeHandler{
		jsonHandler,
		channelHandler,
	}
	// Set up the default logger
	slogger := slog.New(compositeHandler)
	slog.SetDefault(slogger)

	slog.Info("test log message")
	slog.Info("test log message 2")

	// Ensure that the logs were flushed
	time.Sleep(2 * time.Second)

	logs, err := store.LogStore().Get(context.Background(), metastore.LogQueryOptions{
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   time.Now().Add(time.Hour),
	})
	require.NoError(t, err)
	require.Len(t, logs, 2)
	require.EqualValues(t, "test log message", logs[0].Message)
	require.EqualValues(t, "test log message 2", logs[1].Message)
	close(logCh)
}
