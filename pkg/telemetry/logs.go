package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/core"
	"go.opentelemetry.io/otel/trace"
)

type LogLevel int

type InstanceInfo struct {
	Namespace string         `json:"namespace"`
	Instance  string         `json:"instance"`
	Invoker   string         `json:"invoker"`
	Callpath  string         `json:"callpath"`
	Path      string         `json:"path"`
	State     string         `json:"state"`
	Status    core.LogStatus `json:"status"`
	Trace     string         `json:"trace"`
	Span      string         `json:"span"`
}

type NamespaceInfo struct {
	Namespace string `json:"namespace"`
}

type HTTPInstanceInfo struct {
	InstanceInfo
	Msg   string   `json:"msg"`
	Level LogLevel `json:"level"`
}

func (ii *HTTPInstanceInfo) GetInstanceInfo() InstanceInfo {
	return ii.InstanceInfo
}

const (
	LogLevelDebug LogLevel = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelError

	errorKey     = "error"
	scopeKey     = "scope"
	namespaceKey = "namespace"
	instanceKey  = "instance"

	DirektivInstance = "instance-ctx"
)

func LogInitInstance(ctx context.Context, info InstanceInfo) context.Context {
	// add opentelemtry if it exists
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().TraceID().IsValid() {
		info.Trace = span.SpanContext().TraceID().String()
		info.Span = span.SpanContext().SpanID().String()
	}

	// check if required fields are set
	return context.WithValue(ctx, DirektivInstance, info)
}

func logInstance(ctx context.Context, msg string, lvl LogLevel, err error) {
	instanceID := ""
	i := ctx.Value(DirektivInstance)

	if i == nil {
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!IT IS NIL")
	}

	if i != nil {
		info, ok := i.(InstanceInfo)
		if !ok {
			slog.Error("instance info not the expected type")
			return
		}
		instanceID = info.Instance
	}

	if instanceID == "" {
		slog.Error("instance id is empty")
		return
	}

	switch lvl {
	case LogLevelInfo:
		slog.InfoContext(ctx, msg, slog.String(scopeKey, "instance."+instanceID))
	case LogLevelDebug:
		slog.DebugContext(ctx, msg, slog.String(scopeKey, "instance."+instanceID))
	case LogLevelWarn:
		slog.WarnContext(ctx, msg, slog.String(scopeKey, "instance."+instanceID))
	case LogLevelError:
		slog.ErrorContext(ctx, msg, slog.Any(errorKey, err.Error()), slog.String(scopeKey, "instance."+instanceID))
	}
}

func LogInstanceError(ctx context.Context, msg string, err error) {
	logInstance(ctx, msg, LogLevelError, err)
}

func LogInstanceInfo(ctx context.Context, msg string) {
	logInstance(ctx, msg, LogLevelInfo, nil)
}

func LogInstanceDebug(ctx context.Context, msg string) {
	logInstance(ctx, msg, LogLevelDebug, nil)
}

func LogInstanceWarn(ctx context.Context, msg string) {
	logInstance(ctx, msg, LogLevelWarn, nil)
}

func LogNamespaceInfo(ctx context.Context, msg, namespace string) {
	slog.InfoContext(ctx, msg, slog.String(namespaceKey, namespace), slog.String(scopeKey, "namespace."+namespace))
}

func LogNamespaceDebug(ctx context.Context, msg, namespace string) {
	slog.DebugContext(ctx, msg, slog.String(namespaceKey, namespace), slog.String(scopeKey, "namespace."+namespace))
}

func LogNamespaceWarn(ctx context.Context, msg, namespace string) {
	slog.WarnContext(ctx, msg, slog.String(namespaceKey, namespace), slog.String(scopeKey, "namespace."+namespace))
}

func LogNamespaceError(ctx context.Context, msg, namespace string, err error) {
	if err == nil {
		err = fmt.Errorf("%s", msg)
	}
	slog.ErrorContext(ctx, msg, slog.String(namespaceKey, namespace),
		slog.Any(errorKey, err.Error()), slog.String(scopeKey, "namespace."+namespace))
}

func LogActivityInfo(msg, namespace, pid string) {
	slog.Info(msg, slog.String(namespaceKey, namespace),
		slog.String(instanceKey, pid), slog.String(scopeKey, "activity."+pid))
}

func LogActivityDebug(msg, namespace, pid string) {
	slog.Debug(msg, slog.String(namespaceKey, namespace), slog.String(instanceKey, pid),
		slog.String(scopeKey, "activity."+pid))
}

func LogActivityWarn(msg, namespace, pid string) {
	slog.Warn(msg, slog.String(namespaceKey, namespace), slog.String(instanceKey, pid),
		slog.String(scopeKey, "activity."+pid))
}

func LogActivityError(msg, namespace, pid string, err error) {
	if err == nil {
		err = fmt.Errorf("%s", msg)
	}
	slog.Error(msg, slog.String(namespaceKey, namespace), slog.String(instanceKey, pid),
		slog.Any(errorKey, err.Error()), slog.String(scopeKey, "activity."+pid))
}
