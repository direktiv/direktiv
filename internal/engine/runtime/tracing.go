package runtime

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type tracingPack struct {
	ctx, parentCtx   context.Context
	tracer           trace.Tracer
	span, parentSpan trace.Span

	namespace, id, invoker, path string

	thrownError error
}

func newTracingPack(ctx context.Context, namespace, id, invoker, path string) *tracingPack {
	ctx = telemetry.SetupInstanceLogs(ctx, namespace, id, invoker, path)

	tracer := otel.Tracer(telemetry.OtelServiceName)
	ctx, span := tracer.Start(ctx, "start")

	spanCtx := span.SpanContext()
	slog.Info("start tracing instance", slog.String("tracing-id", spanCtx.TraceID().String()))

	tp := &tracingPack{
		tracer:     tracer,
		parentSpan: span,
		parentCtx:  ctx,
		ctx:        ctx,

		namespace: namespace,
		id:        id,
		invoker:   invoker,
		path:      path,
	}

	tp.span = tp.setAttributes(span)

	// we set the parent always on ok, changes when we have e.g. flow timeouts
	// that would be a case to set the error here.
	tp.span.SetStatus(codes.Ok, codes.Ok.String())

	telemetry.LogInstance(ctx, telemetry.LogLevelInfo,
		fmt.Sprintf("flow starting, trace-id %s", spanCtx.TraceID().String()))

	return tp
}

func (tp *tracingPack) tracingStart(fName string) {
	tp.ctx = telemetry.SetInstanceLogState(tp.ctx, fName)
	tp.parentCtx = telemetry.SetInstanceLogState(tp.parentCtx, fName)

	childCtx, span := tp.tracer.Start(tp.parentCtx, fName)
	tp.span = tp.setAttributes(span)
	tp.ctx = childCtx
}

func (tp *tracingPack) tracingTransition(fName string) {
	telemetry.LogInstance(tp.ctx, telemetry.LogLevelInfo,
		fmt.Sprintf("transitioning to '%s'", fName))

	tp.span.SetStatus(codes.Ok, codes.Ok.String())

	tp.span.End()
	tp.tracingStart(fName)
}

func (tp *tracingPack) handleError(err error) {
	tp.ctx = telemetry.SetInstanceLogStatus(tp.ctx, core.LogErrStatus)
	tp.parentCtx = telemetry.SetInstanceLogStatus(tp.parentCtx, core.LogErrStatus)

	// if that is nil then it has been thrown in an action
	// or http request
	if tp.thrownError == nil {
		tp.span.SetStatus(codes.Error, err.Error())
	} else {
		err = tp.thrownError	
	}

	telemetry.LogInstance(tp.ctx, telemetry.LogLevelError,
		fmt.Sprintf("error during flow: %s", err.Error()))

	tp.span.End()
}

func (tp *tracingPack) tracingFinish() {
	tp.ctx = telemetry.SetInstanceLogStatus(tp.ctx, core.LogCompletedStatus)
	tp.parentCtx = telemetry.SetInstanceLogStatus(tp.parentCtx, core.LogCompletedStatus)

	tp.span.SetStatus(codes.Ok, codes.Ok.String())
	tp.span.End()
}

func (tp *tracingPack) finish() {
	// if !tp.isError {
	tp.parentSpan.SetStatus(codes.Ok, codes.Ok.String())
	// }

	telemetry.LogInstance(tp.ctx, telemetry.LogLevelInfo, "flow finished")

	tp.parentSpan.End()
}

func (tp *tracingPack) trace(action string) trace.Span {
	_, span := tp.tracer.Start(tp.ctx, action)
	tp.setAttributes(span)
	return span
}

func (tp *tracingPack) setAttributes(span trace.Span) trace.Span {
	span.SetAttributes(
		attribute.KeyValue{
			Key:   "namespace",
			Value: attribute.StringValue(tp.namespace),
		},
		attribute.KeyValue{
			Key:   "path",
			Value: attribute.StringValue(tp.path),
		},
		attribute.KeyValue{
			Key:   "invoker",
			Value: attribute.StringValue(tp.invoker),
		},
		attribute.KeyValue{
			Key:   "instance",
			Value: attribute.StringValue(tp.id),
		},
		attribute.KeyValue{
			Key:   "scope",
			Value: attribute.StringValue(string(telemetry.LogScopeInstance)),
		},
	)

	return span
}
