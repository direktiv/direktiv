package util

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	otlp "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	otlpgrpc "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var TelemetryMiddleware = func(h http.Handler) http.Handler {
	return h
}

var telemetryUnaryServerInterceptor grpc.UnaryServerInterceptor
var telemetryStreamServerInterceptor grpc.StreamServerInterceptor

var globalGRPCDialOptions []grpc.DialOption

func AddGlobalGRPCDialOption(opt grpc.DialOption) {
	globalGRPCDialOptions = append(globalGRPCDialOptions, opt)
}

var globalGRPCServerOptions []grpc.ServerOption

func AddGlobalGRPCServerOption(opt grpc.ServerOption) {
	globalGRPCServerOptions = append(globalGRPCServerOptions, opt)
}

type grpcMetadataTMC struct {
	md *metadata.MD
}

func (tmc *grpcMetadataTMC) Get(k string) string {
	array := tmc.md.Get(k)
	if len(array) == 0 {
		return ""
	}
	return array[0]
}

func (tmc *grpcMetadataTMC) Keys() []string {
	keys := tmc.md.Get("oteltmckeys")
	if keys == nil {
		keys = make([]string, 0)
	}
	return keys
}

func (tmc *grpcMetadataTMC) Set(k, v string) {
	newKey := len(tmc.md.Get(k)) == 0
	tmc.md.Set(k, v)
	if newKey {
		tmc.md.Append("oteltmckeys", k)
	}
}

var instrumentationName string

func InitTelemetry(conf *Config, svcName, imName string) (func(), error) {

	instrumentationName = imName

	var prop propagation.TextMapPropagator
	prop = propagation.TraceContext{}
	otel.SetTracerProvider(trace.NewNoopTracerProvider())
	otel.SetTextMapPropagator(prop)

	addr := conf.GetTelemetryBackendAddr()
	if addr == "" {
		return func() {}, nil
	}

	driver := otlpgrpc.NewClient(
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithEndpoint(addr),
		otlpgrpc.WithDialOption(grpc.WithBlock()),
	)

	ctx := context.Background()

	exp, err := otlp.New(ctx, driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceNameKey.String(svcName)))
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exp)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	otel.SetTracerProvider(tp)

	AddGlobalGRPCDialOption(grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		tp := otel.GetTracerProvider()
		tr := tp.Tracer(imName)

		requestMetadata, _ := metadata.FromOutgoingContext(ctx)
		metadataCopy := requestMetadata.Copy()

		name := method
		var span trace.Span
		ctx, span = tr.Start(
			ctx,
			name,
			trace.WithSpanKind(trace.SpanKindClient),
		)
		defer span.End()

		prop = otel.GetTextMapPropagator()
		prop.Inject(ctx, &grpcMetadataTMC{&metadataCopy})

		ctx = metadata.NewOutgoingContext(ctx, metadataCopy)

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			s, _ := status.FromError(err)
			span.SetStatus(codes.Code(s.Code()), s.Message())
		}

		return err

	}))

	AddGlobalGRPCDialOption(grpc.WithStreamInterceptor(func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {

		tp := otel.GetTracerProvider()
		tr := tp.Tracer(imName)

		requestMetadata, _ := metadata.FromOutgoingContext(ctx)
		metadataCopy := requestMetadata.Copy()

		name := method
		var span trace.Span
		ctx, span = tr.Start(
			ctx,
			name,
			trace.WithSpanKind(trace.SpanKindClient),
		)
		defer span.End()

		prop = otel.GetTextMapPropagator()
		prop.Inject(ctx, &grpcMetadataTMC{&metadataCopy})

		ctx = metadata.NewOutgoingContext(ctx, metadataCopy)

		cs, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			s, _ := status.FromError(err)
			span.SetStatus(codes.Code(s.Code()), s.Message())
		}

		return cs, nil

	}))

	telemetryUnaryServerInterceptor = func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

		prop := otel.GetTextMapPropagator()
		requestMetadata, _ := metadata.FromIncomingContext(ctx)
		metadataCopy := requestMetadata.Copy()
		carrier := &grpcMetadataTMC{&metadataCopy}
		ctx = prop.Extract(ctx, carrier)

		tp := otel.GetTracerProvider()
		tr := tp.Tracer(imName)

		var span trace.Span
		ctx, span = tr.Start(
			ctx,
			info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		resp, err = handler(ctx, req)
		if err != nil {
			s, _ := status.FromError(err)
			span.SetStatus(codes.Code(s.Code()), s.Message())
		}

		return resp, err

	}

	telemetryStreamServerInterceptor = func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		ctx := ss.Context()

		prop := otel.GetTextMapPropagator()
		requestMetadata, _ := metadata.FromIncomingContext(ctx)
		metadataCopy := requestMetadata.Copy()
		carrier := &grpcMetadataTMC{&metadataCopy}
		ctx = prop.Extract(ctx, carrier)

		tp := otel.GetTracerProvider()
		tr := tp.Tracer(imName)

		var span trace.Span
		_, span = tr.Start(
			ctx,
			info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		err = handler(srv, ss)
		if err != nil {
			s, _ := status.FromError(err)
			span.SetStatus(codes.Code(s.Code()), s.Message())
		}

		return err

	}

	TelemetryMiddleware = func(h http.Handler) http.Handler {
		return &telemetryHandler{
			imName: imName,
			next:   h,
		}
	}

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
	ctx, span := tr.Start(ctx, mux.CurrentRoute(r).GetName(), trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	subr := r.WithContext(ctx)

	h.next.ServeHTTP(w, subr)

}

func telemetryWaiter(tp *sdktrace.TracerProvider, bsp sdktrace.SpanProcessor) func() {

	return func() {

		t := time.Now()

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

func (c *httpCarrier) Keys() []string {
	return c.r.Header.Values("oteltmckeys")
}

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
