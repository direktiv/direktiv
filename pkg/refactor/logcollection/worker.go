package logcollection

import (
	"context"
	"time"
)

// LogStoreWorker manages the log polling and channel communication.
type LogStoreWorker struct {
	LogStore   LogStore
	Stream     string
	InstanceID string
	Interval   time.Duration
	StopCh     chan struct{}
	LogCh      chan []LogEntry
}

// NewLogStoreWorker creates a new LogStoreWorker.
func NewLogStoreWorker(logStore LogStore, stream, instanceID string, interval time.Duration) *LogStoreWorker {
	return &LogStoreWorker{
		LogStore:   logStore,
		Stream:     stream,
		InstanceID: instanceID,
		Interval:   interval,
		StopCh:     make(chan struct{}),
		LogCh:      make(chan []LogEntry),
	}
}

// Start starts the log polling worker.
func (lw *LogStoreWorker) Start(ctx context.Context) {
	go func() {
		defer close(lw.LogCh)
		ticker := time.NewTicker(lw.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				logs, err := lw.LogStore.Get(ctx, lw.Stream, 0)
				if err != nil {
					// Handle error
					continue
				}
				lw.LogCh <- logs
			}
		}
	}()
}

// StartInstanceLogsWorker starts the instance logs polling worker.
func (lw *LogStoreWorker) StartInstanceLogsWorker(ctx context.Context) {
	go func() {
		defer close(lw.LogCh)
		ticker := time.NewTicker(lw.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				logs, err := lw.LogStore.GetInstanceLogs(ctx, lw.Stream, lw.InstanceID, 0)
				if err != nil {
					// Handle error
					continue
				}
				lw.LogCh <- logs
			}
		}
	}()
}

// Stop stops the log polling worker.
func (lw *LogStoreWorker) Stop() {
	close(lw.StopCh)
}
