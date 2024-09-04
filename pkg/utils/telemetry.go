package utils

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/direktiv/direktiv/pkg/middlewares"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	otlp "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	otlpgrpc "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

var TelemetryMiddleware = func(h http.Handler) http.Handler {
	return h
}

var globalGRPCDialOptions []grpc.DialOption

func AddGlobalGRPCDialOption(opt grpc.DialOption) {
	globalGRPCDialOptions = append(globalGRPCDialOptions, opt)
}

var globalGRPCServerOptions []grpc.ServerOption

func AddGlobalGRPCServerOption(opt grpc.ServerOption) {
	globalGRPCServerOptions = append(globalGRPCServerOptions, opt)
}

var instrumentationName string

func InitTelemetry(cirCtx context.Context, addr string, svcName, imName string) (func(), error) {
	slog.Debug("Initializing telemetry.", "instrumentationName", imName)
	instrumentationName = imName

	prop := propagation.TraceContext{}
	otel.SetTracerProvider(otel.GetTracerProvider())
	otel.SetTextMapPropagator(prop)

	if addr == "" {
		return func() {}, nil
	}
	var exp sdktrace.SpanExporter
	var err error

	slog.Debug("Creating OTLP gRPC client.", "endpoint", addr)
	driver := otlpgrpc.NewClient(
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithEndpoint(addr),
		otlpgrpc.WithDialOption(grpc.WithBlock()),
	)

	slog.Debug("Setting up OTLP exporter.")
	exp, err = otlp.New(cirCtx, driver)
	if err != nil {
		slog.Error("Failed to create OTLP exporter.", "error", err)
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	slog.Debug("Creating resource with service name.", "serviceName", svcName)
	res, err := resource.New(cirCtx, resource.WithAttributes(semconv.ServiceNameKey.String(svcName)))
	if err != nil {
		slog.Error("Failed to create resource.", "error", err)
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	slog.Debug("Setting up SimpleSpanProcessor with no-op exporter.")
	bsp := sdktrace.NewSimpleSpanProcessor(exp)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // Always sample spans, generate trace IDs
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	slog.Debug("Setting tracer provider.")
	otel.SetTracerProvider(tp)

	slog.Debug("Registering HTTP telemetry middleware.")
	TelemetryMiddleware = func(h http.Handler) http.Handler {
		return &telemetryHandler{
			imName: imName,
			next:   h,
		}
	}

	middlewares.RegisterHTTPMiddleware(TelemetryMiddleware)

	slog.Debug("Telemetry initialization completed.")

	return telemetryWaiter(tp, bsp), nil
}

type telemetryHandler struct {
	imName string
	next   http.Handler
}

func (h *telemetryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	prop := otel.GetTextMapPropagator()
	ctx := prop.Extract(r.Context(), &httpCarrier{
		r: r,
	})

	tp := otel.GetTracerProvider()
	tr := tp.Tracer(h.imName)
	route := "apiv2"

	if mux.CurrentRoute(r) != nil {
		route = mux.CurrentRoute(r).GetName()
	}
	ctx, span := tr.Start(ctx, route, trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	subr := r.WithContext(ctx)

	h.next.ServeHTTP(w, subr)
}

func telemetryWaiter(tp *sdktrace.TracerProvider, bsp sdktrace.SpanProcessor) func() {
	return func() {
		t := time.Now().UTC()

		deadline := t.Add(25 * time.Second)

		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()

		err := bsp.ForceFlush(ctx)
		if err != nil {
			fmt.Printf("Failed to flush telemetry data: %v\n", err)
			return
		}

		err = tp.Shutdown(ctx)
		if err != nil {
			fmt.Printf("Failed to shutdown telemetry: %v\n", err)
			return
		}
	}
}

func Trace(ctx context.Context, msg string) {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return
	}

	span.AddEvent(msg)
}

type httpCarrier struct {
	r *http.Request
}

func (c *httpCarrier) Get(key string) string {
	return c.r.Header.Get(key)
}

// nolint:canonicalheader
func (c *httpCarrier) Keys() []string {
	return c.r.Header.Values("oteltmckeys")
}

// nolint:canonicalheader
func (c *httpCarrier) Set(key, val string) {
	prev := c.Get(key)
	if prev == "" {
		c.r.Header.Add("oteltmckeys", key)
	}
	c.r.Header.Set(key, val)
}

func TraceHTTPRequest(ctx context.Context, r *http.Request) (cleanup func()) {
	tp := otel.GetTracerProvider()
	tr := tp.Tracer(instrumentationName)
	ctx, span := tr.Start(ctx, "function", trace.WithSpanKind(trace.SpanKindClient))

	prop := otel.GetTextMapPropagator()
	prop.Inject(ctx, &httpCarrier{
		r: r,
	})

	return func() { span.End() }
}

func TraceGWHTTPRequest(ctx context.Context, r *http.Request, instrumentationName string) (cleanup func()) {
	tp := otel.GetTracerProvider()
	tr := tp.Tracer(instrumentationName)
	ctx, span := tr.Start(ctx, "gateway", trace.WithSpanKind(trace.SpanKindClient))

	prop := otel.GetTextMapPropagator()
	prop.Inject(ctx, &httpCarrier{
		r: r,
	})

	return func() { span.End() }
}

type GenericTelemetryCarrier struct {
	Trace map[string]string
}

func (c *GenericTelemetryCarrier) Get(key string) string {
	v := c.Trace[key]
	return v
}

func (c *GenericTelemetryCarrier) Keys() []string {
	var keys []string
	for k := range c.Trace {
		keys = append(keys, k)
	}
	return keys
}

func (c *GenericTelemetryCarrier) Set(key, val string) {
	c.Trace[key] = val
}

func TransplantTelemetryContextInformation(a, b context.Context) context.Context {
	carrier := &GenericTelemetryCarrier{
		Trace: make(map[string]string),
	}
	prop := otel.GetTextMapPropagator()
	prop.Inject(a, carrier)
	ctx := prop.Extract(b, carrier)
	return ctx
}
