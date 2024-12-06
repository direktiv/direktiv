/*
Package tracing provides telemetry and structured logging capabilities for workflows,
including integration with OpenTelemetry and context-aware logging.

# Initialization

To initialize telemetry:

1. **Telemetry**:

	```go
	telEnd, err := tracing.InitTelemetry(context.Background(), srv.config.OpenTelemetry, "direktiv/flow", "direktiv")
	if err != nil {
		return nil, fmt.Errorf("Telemetry init failed: %w", err)
	} // call telEnd() on systems shutdown signal
	```

> **Note**: The logging section is deprecated and will be removed in future versions.

# Usage

## Workflow

1. Start a new trace span with `NewSpan`.
2. Use the `betterlogger` system to log within the span's context.

## Example Usage

	```go
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
*/
package tracing
