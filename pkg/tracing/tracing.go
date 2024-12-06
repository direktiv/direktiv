/*
Package tracing provides telemetry and structured logging capabilities for workflows,
including integration with OpenTelemetry and context-aware logging.

# Initialization

To initialize telemetry and logging:

1. **Telemetry**:

	```go
	telEnd, err := tracing.InitTelemetry(context.Background(), srv.config.OpenTelemetry, "direktiv/flow", "direktiv")
	if err != nil {
		return nil, fmt.Errorf("Telemetry init failed: %w", err)
	} // call telEnd() on systems shutdown signal
	```

2. **Logging**:

	```go
	handlers := tracing.NewContextHandler(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slogger := slog.New(
		tracing.TeeHandler{
			handlers,
			tracing.EventHandler{},
		},
	)

	slog.SetDefault(slogger)
	```

# Usage

## Workflow

1. Add metadata to a context using helpers like `AddInstanceAttr`.
2. Start a new trace span with `NewSpan`.
3. Add the "track" to the context using the helpers.
4. Use `slog` to log within the span's context.

## Example Usage

	```go
	// Add attributes to the context
	ctx = tracing.AddInstanceAttr(ctx, tracing.InstanceAttributes{
		Namespace:    args.Namespace.Name,
		InstanceID:   args.ID.String(),
		Invoker:      args.Invoker,
		Callpath:     args.TelemetryInfo.CallPath,
		WorkflowPath: args.CalledAs,
		Status:       core.LogRunningStatus,
	})
	// use "track" for log correlation without a trace provider
	ctx = tracing.WithTrack(ctx, tracing.BuildInstanceTrackViaCallpath(args.TelemetryInfo.CallPath))
	// Start a new span
	ctx, cleanup, err := tracing.NewSpan(ctx, "creating a new Instance: "+args.ID.String()+", workflow: "+args.CalledAs)
	if err != nil {
		slog.Debug("failed to create new span", "error", err)
		// Depending on severity, either return, retry, or continue without span
	}
	defer cleanup() // Ensures telemetry data is flushed and resources are freed

	// Log within the span's context
	slog.DebugContext(ctx, "Initializing new instance creation.")
	```

## Setting Span Error

The `SetSpanError` method allows you to mark a span as an error and provide additional context with an error message and description.

### Example Usage:

	```go
	ctx, cleanup, _ := tracing.NewSpan(ctx, "example-operation")
	defer cleanup()

	// Simulate an error
	err := fmt.Errorf("operation failed")
	tracing.SetSpanError(ctx, err, "An error occurred during example operation")

	// Continue execution or handle the error accordingly
	```

The `SetSpanError` method ensures that:

1. The span's status is set to `Error`.
2. Metadata such as `error.message` and a custom `error.description` are attached as attributes.
3. An event with error details is recorded for richer tracing.

slog will internally ensure that the log-entry is properly ingested and redirected to the proper stream/track.
It also internally ingests proper telemetry information and metrics.

# Notes

  - "track" represents a flat unique identifier for a chain of logs associated with a resource call,
    including its nested dependencies, like workflow instances. The track ties together logs from different parts
    of the workflow, whereas spans track individual operations or tasks.
  - "track" and "traceID"-"spanID" serve the same purpose, yet using "track"
    allows the system to function without a proper-tracing-provider.
*/
package tracing
