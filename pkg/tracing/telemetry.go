// nolint:unused
package tracing

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

const (
	flushingTimeoutSeconds = 25
	maxRetries             = 3
)

var (
	instrumentationName string
	meter               otelmetric.Meter
	requestCounter      otelmetric.Int64Counter
	requestDuration     otelmetric.Float64Histogram
)

// InitTelemetry initializes tracing and metrics with OTLP.
func InitTelemetry(ctx context.Context, addr, svcName, imName string) (func(), error) {
	slog.Debug("Initializing telemetry.", slog.String("instrumentationName", imName))
	instrumentationName = imName

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceNameKey.String(svcName)))
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Use the retry-enabled setupMetrics for metrics
	if err := setupMetrics(ctx, addr, res); err != nil {
		return nil, err
	}

	tracerProvider, spanProcessor, err := setupTracing(ctx, addr, res)
	if err != nil {
		return nil, err
	}

	return telemetryWaiter(ctx, tracerProvider, spanProcessor), nil
}

// setupMetrics initializes OpenTelemetry metrics.
func setupMetrics(ctx context.Context, addr string, res *resource.Resource) error {
	slog.Debug("setting up OpenTelemetry metric provider.")

	// Create the custom MetricExporter
	metricExporter := &MetricExporter{}
	var err error

	// Initialize the RemoteExporter if an address is provided
	if addr != "" {
		metricExporter.RemoteExporter, err = retry(func() (*otlpmetricgrpc.Exporter, error) {
			return otlpmetricgrpc.New(ctx,
				otlpmetricgrpc.WithInsecure(),
				otlpmetricgrpc.WithEndpoint(addr),
			)
		})
		if err != nil {
			return fmt.Errorf("failed to create OTLP metric exporter: %w", err)
		}
	}

	// Use the custom MetricExporter (not just the RemoteExporter) with the PeriodicReader
	reader := sdkmetric.NewPeriodicReader(metricExporter)

	// Create and set the MeterProvider
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(res),
	)
	// Initialize the meter here
	meter = otel.Meter(instrumentationName)

	// Initialize requestCounter (counter for requests)
	requestCounter, err = meter.Int64Counter("requests", otelmetric.WithDescription("Counts the number of requests"))
	if err != nil {
		return err
	}
	// Initialize requestDuration (histogram for request durations)
	requestDuration, err = meter.Float64Histogram("request_duration", otelmetric.WithDescription("Records the duration of requests"))
	if err != nil {
		return err
	}

	otel.SetMeterProvider(meterProvider)

	return nil
}

// setupTracing initializes OpenTelemetry tracing.
func setupTracing(ctx context.Context, addr string, res *resource.Resource) (*sdktrace.TracerProvider, sdktrace.SpanProcessor, error) {
	slog.Debug("setting up OpenTelemetry tracing.")

	exporter := &Exporter{}
	var err error

	if addr != "" {
		slog.Info("adding OTLP Exporter.")
		driver := otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(addr),
			otlptracegrpc.WithInsecure(),
		)
		exporter.remoteExporter, err = retry(func() (sdktrace.SpanExporter, error) {
			return otlptrace.New(ctx, driver)
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
		}
	}

	sampler := sdktrace.ParentBased(sdktrace.AlwaysSample())
	spanProcessor := sdktrace.NewBatchSpanProcessor(exporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(spanProcessor),
	)

	otel.SetTracerProvider(tracerProvider)

	return tracerProvider, spanProcessor, nil
}

// retry attempts an operation multiple times with backoff.
func retry[T any](fn func() (T, error)) (T, error) {
	var lastErr error
	var result T
	for attempt := range maxRetries {
		result, lastErr = fn()
		if lastErr == nil {
			return result, nil
		}
		slog.Warn("operation failed, retrying...", slog.Int("attempt", attempt+1), slog.Any("error", lastErr))
		time.Sleep(time.Duration(attempt+1) * time.Second)
	}
	slog.Error("operation failed after retries.", slog.Any("error", lastErr))

	return result, lastErr
}

// telemetryWaiter ensures all telemetry data is flushed and the provider is shut down gracefully.
func telemetryWaiter(ctx context.Context, tp *sdktrace.TracerProvider, bsp sdktrace.SpanProcessor) func() {
	return func() {
		ctx, cancel := context.WithTimeout(ctx, flushingTimeoutSeconds*time.Second)
		defer cancel()

		slog.Info("flushing telemetry data before shutdown.")
		if err := bsp.ForceFlush(ctx); err != nil {
			slog.Error("failed to flush telemetry data.", "error", err)
		}

		slog.Info("shutting down telemetry.")
		if err := tp.Shutdown(ctx); err != nil {
			slog.Error("failed to shutdown telemetry.", "error", err)
		}
	}
}
