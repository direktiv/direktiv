package util

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
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

func InitTelemetry(conf *Config, svcName, imName string) (func(), error) {

	var prop propagation.TextMapPropagator
	prop = propagation.TraceContext{}
	otel.SetTracerProvider(trace.NewNoopTracerProvider())
	otel.SetTextMapPropagator(prop)

	addr := conf.GetTelemetryBackendAddr()
	if addr == "" {
		return func() {}, nil
	}

	driver := otlpgrpc.NewDriver(
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithEndpoint(addr),
		otlpgrpc.WithDialOption(grpc.WithBlock()),
	)

	ctx := context.Background()

	exp, err := otlp.NewExporter(ctx, driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %v", err)
	}

	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceNameKey.String(svcName)))
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %v", err)
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

		ctx, span := tr.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
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

		ctx, span := tr.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		err = handler(ctx, ss)
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

	ctx := r.Context()

	prop := otel.GetTextMapPropagator()
	requestMetadata, _ := metadata.FromIncomingContext(ctx)
	metadataCopy := requestMetadata.Copy()
	carrier := &grpcMetadataTMC{&metadataCopy}
	ctx = prop.Extract(ctx, carrier)

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
