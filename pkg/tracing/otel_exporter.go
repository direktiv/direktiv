package tracing

import (
	"context"
	"encoding/json"
	"time"

	"github.com/direktiv/direktiv/pkg/datastore"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Exporter struct {
	RemoteExporter sdktrace.SpanExporter
	Store          datastore.TracesStore
}

// ExportSpans exports spans to both the trace store and remote exporter.
func (e *Exporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	var traces []datastore.Trace

	for _, span := range spans {
		rawTrace, err := MarshalSpan(span)
		if err != nil {
			return err
		}

		trace := datastore.Trace{
			TraceID:   span.SpanContext().TraceID().String(),
			SpanID:    span.SpanContext().SpanID().String(),
			StartTime: span.StartTime(),
			EndTime:   span.EndTime(),
			RawTrace:  rawTrace,
		}

		if span.Parent().IsValid() {
			parentID := span.Parent().SpanID().String()
			trace.ParentSpanID = &parentID
		}

		traces = append(traces, trace)
	}

	// Store traces in DB
	if len(traces) > 0 {
		if err := e.Store.Append(ctx, traces...); err != nil {
			return err
		}
	}

	// Export remotely if applicable
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

// MarshalSpan converts an OpenTelemetry span to JSON.
func MarshalSpan(span sdktrace.ReadOnlySpan) ([]byte, error) {
	traceData := TraceData{
		Attributes: make(map[string]any),
		Events:     []SpanEvent{},
		Links:      []SpanLink{},
		Library: InstrumentationLibrary{
			Name:    span.InstrumentationLibrary().Name,
			Version: span.InstrumentationLibrary().Version,
		},
	}

	// Extract Attributes
	for _, attr := range span.Attributes() {
		traceData.Attributes[string(attr.Key)] = attr.Value.AsInterface()
	}

	// Extract Events
	for _, event := range span.Events() {
		spanEvent := SpanEvent{
			Name:       event.Name,
			Timestamp:  event.Time,
			Attributes: make(map[string]any),
		}
		for _, attr := range event.Attributes {
			spanEvent.Attributes[string(attr.Key)] = attr.Value.AsInterface()
		}
		traceData.Events = append(traceData.Events, spanEvent)
	}

	// Extract Status
	status := span.Status()
	traceData.Status = SpanStatus{
		Code:        status.Code.String(),
		Description: status.Description,
	}

	// Extract Links
	for _, link := range span.Links() {
		traceData.Links = append(traceData.Links, SpanLink{
			TraceID: link.SpanContext.TraceID().String(),
			SpanID:  link.SpanContext.SpanID().String(),
		})
	}

	// Extract Span Kind
	traceData.SpanKind = span.SpanKind().String()

	return json.Marshal(traceData)
}

func UnmarshalSpan(data []byte) (TraceData, error) {
	var traceData TraceData
	err := json.Unmarshal(data, &traceData)

	return traceData, err
}

// TraceData stores additional trace details.
type TraceData struct {
	Attributes map[string]any         `json:"attributes"`
	Events     []SpanEvent            `json:"events"`
	SpanKind   string                 `json:"span_kind"`
	Status     SpanStatus             `json:"status"`
	Links      []SpanLink             `json:"links"`
	Library    InstrumentationLibrary `json:"library"`
}

// SpanEvent represents an event inside a span.
type SpanEvent struct {
	Name       string         `json:"name"`
	Timestamp  time.Time      `json:"timestamp"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// SpanStatus represents the success or failure state.
type SpanStatus struct {
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
}

// SpanLink represents links to other spans.
type SpanLink struct {
	TraceID string `json:"trace_id"`
	SpanID  string `json:"span_id"`
}

// InstrumentationLibrary stores the library that created the span.
type InstrumentationLibrary struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}
