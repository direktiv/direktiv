package tracing

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/metastore"
)

// Worker processes logs from a channel and periodically flushes them in batches to a datastore.
type Worker struct {
	logCh         chan metastore.LogEntry // Channel receiving log records
	batch         []metastore.LogEntry    // Holds the current batch of log records
	flushInterval time.Duration           // Time interval to flush the logs
	logStore      metastore.LogStore      // DataStore instance to persist logs
	cachedLevel   int                     // Cached log level from settings
}

// WorkerArgs holds the parameters for initializing a Worker.
type WorkerArgs struct {
	LogCh         chan metastore.LogEntry // Channel to receive log records
	LogStore      metastore.LogStore      // DataStore for log storage
	MaxBatchSize  int                     // Maximum size of the batch to accumulate
	FlushInterval time.Duration           // Time interval to flush logs
	CachedLevel   int                     // Initial cached log level
}

// Worker handles the process of batching and flushing log records to a datastore. It reads log records from a
// channel, groups them into batches, and flushes them at regular intervals to a datastore, ensuring efficient log storage.
//
// Typical usage includes setting up the worker with a channel and datastore, then starting it to process logs
// asynchronously in the background.
func NewWorker(args WorkerArgs) *Worker {
	return &Worker{
		logCh:         args.LogCh,
		batch:         make([]metastore.LogEntry, args.MaxBatchSize),
		flushInterval: args.FlushInterval,
		logStore:      args.LogStore,
		cachedLevel:   args.CachedLevel,
	}
}

// Start launches the log processing loop in a goroutine, which batches and flushes logs periodically.
func (w *Worker) Start(circuit *core.Circuit) error {
	count := 0

	for {
		select {
		case record, ok := <-w.logCh:
			if !ok {
				// Channel is closed, flush remaining logs
				if count > 0 {
					w.processBatch(w.batch[:count]) //nolint:contextcheck
				}

				return nil
			}

			// Add log to batch
			w.batch[count] = record
			count++
			if count >= len(w.batch) {
				// Batch is full, flush it
				w.processBatch(w.batch[:count]) //nolint:contextcheck
				count = 0
			}

		case <-time.After(w.flushInterval):
			slog.Info("Flush ping")
			if count > 0 {
				w.processBatch(w.batch[:count]) //nolint:contextcheck
				count = 0
			}

		case <-circuit.Context().Done():
			// Context canceled, flush remaining logs
			if count > 0 {
				w.processBatch(w.batch[:count]) //nolint:contextcheck
				count = 0
			}
		}
	}
}

// processBatch flushes a batch of log records to the datastore if they meet the required log level.
func (w *Worker) processBatch(batch []metastore.LogEntry) {
	ctx := context.Background()
	slog.Info("processBatch ping")

	// Filter logs based on the required log level
	filteredBatch := make([]metastore.LogEntry, 0, len(batch))
	for _, logRecord := range batch {
		if logRecord.Level >= w.cachedLevel {
			filteredBatch = append(filteredBatch, logRecord)
		}
	}

	for _, l := range filteredBatch {
		slog.Info("Persist the filtered logs")
		// Persist the filtered logs
		if err := w.logStore.Append(ctx, l); err != nil {
			slog.ErrorContext(ctx, "failed to store", "error", fmt.Errorf("%w: error appending logs", err))
		}
	}
}
