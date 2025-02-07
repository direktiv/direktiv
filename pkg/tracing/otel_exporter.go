package tracing

import (
	"context"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Exporter struct {
	RemoteExporter sdktrace.SpanExporter
}

// ExportSpans exports spans to both console and remote (if available).
func (e *Exporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	// for _, span := range spans {
	// 	fmt.Printf("Span: %s | TraceID: %s | ParentID: %s | Duration: %s\n",
	// 		span.Name(), span.SpanContext().TraceID(), span.Parent().SpanID(), span.EndTime().Sub(span.StartTime()),
	// 	)
	// }

	if e.RemoteExporter != nil {
		return e.RemoteExporter.ExportSpans(ctx, spans)
	}

	return nil
}

// Shutdown ensures proper cleanup.
func (e *Exporter) Shutdown(ctx context.Context) error {
	if e.RemoteExporter != nil {
		return e.RemoteExporter.Shutdown(ctx)
	}

	return nil
}

type MetricExporter struct {
	RemoteExporter sdkmetric.Exporter
}

// Export logs metrics and forwards them if a remote exporter is set.
func (e *MetricExporter) Export(ctx context.Context, rm *metricdata.ResourceMetrics) error {
	// Log metrics to console
	// for _, scopeMetrics := range rm.ScopeMetrics {
	// 	for _, m := range scopeMetrics.Metrics {
	// 		fmt.Printf("Metric: %s | Description: %s | Data: %v\n",
	// 			m.Name, m.Description, m.Data,
	// 		)
	// 	}
	// }

	// Forward metrics to remote exporter if available
	if e.RemoteExporter != nil {
		return e.RemoteExporter.Export(ctx, rm)
	}

	return nil
}

// ForceFlush forwards flush requests to the remote exporter if available.
func (e *MetricExporter) ForceFlush(ctx context.Context) error {
	if e.RemoteExporter != nil {
		return e.RemoteExporter.ForceFlush(ctx)
	}

	return nil
}

// Shutdown forwards shutdown requests to the remote exporter if available.
func (e *MetricExporter) Shutdown(ctx context.Context) error {
	if e.RemoteExporter != nil {
		return e.RemoteExporter.Shutdown(ctx)
	}

	return nil
}

// Aggregation provides default aggregation for supported instruments.
func (e *MetricExporter) Aggregation(kind sdkmetric.InstrumentKind) sdkmetric.Aggregation {
	if e.RemoteExporter != nil {
		// Delegate to RemoteExporter if available.
		return e.RemoteExporter.Aggregation(kind)
	}

	return sdkmetric.DefaultAggregationSelector(kind)
}

// Temporality provides default temporality for supported instruments.
func (e *MetricExporter) Temporality(kind sdkmetric.InstrumentKind) metricdata.Temporality {
	if e.RemoteExporter != nil {
		// Delegate to RemoteExporter
		return e.RemoteExporter.Temporality(kind)
	}

	return sdkmetric.DefaultTemporalitySelector(kind)
}
