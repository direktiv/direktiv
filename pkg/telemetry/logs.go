package telemetry

import (
	"context"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/core"
)

type LogLevel int
type LogScope string

type LogObject struct {
	Namespace string   `json:"namespace"`
	ID        string   `json:"id"`
	Scope     LogScope `json:"scope"`
	InstanceInfo
}

type InstanceInfo struct {
	Invoker   string         `json:"invoker,omitempty"`
	Path      string         `json:"path,omitempty"`
	State     string         `json:"state,omitempty"`
	Status    core.LogStatus `json:"status,omitempty"`
	SpanScope string         `json:"spanscope,omitempty"`
}

// HTTPInstanceInfo is used to post logs from the sidecar
// fluentbit can not be used because it is picking up the logs
// of newly created pods too late
type HTTPInstanceInfo struct {
	LogObject
	Msg   string   `json:"msg"`
	Level LogLevel `json:"level"`
}

func (ii *HTTPInstanceInfo) GetLogObject() LogObject {
	return ii.LogObject
}

const (
	LogLevelDebug LogLevel = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelError

	LogScopeInstance  LogScope = "instance"
	LogScopeNamespace LogScope = "namespace"
	LogScopeActivity  LogScope = "activity"
	LogScopeRoute     LogScope = "route"

	errorKey = "error"

	logObjectCtx = "log-ctx"
)

func LogRoute(level LogLevel, namespace, route, msg string) {
	ctx := context.WithValue(context.Background(), logObjectCtx, LogObject{
		Namespace: namespace,
		ID:        route,
		Scope:     LogScopeRoute,
	})
	logPublic(ctx, level, msg)
}

func LogRouteError(namespace, route, msg string, err error) {
	ctx := context.WithValue(context.Background(), logObjectCtx, LogObject{
		Namespace: namespace,
		ID:        route,
		Scope:     LogScopeRoute,
	})
	logPublic(ctx, LogLevelError, msg, slog.Any(errorKey, err.Error()))
}

func LogActivity(level LogLevel, namespace, pid, msg string) {
	ctx := context.WithValue(context.Background(), logObjectCtx, LogObject{
		Namespace: namespace,
		ID:        pid,
		Scope:     LogScopeActivity,
	})
	logPublic(ctx, level, msg)
}

func LogActivityError(namespace, pid, msg string, err error) {
	ctx := context.WithValue(context.Background(), logObjectCtx, LogObject{
		Namespace: namespace,
		ID:        pid,
		Scope:     LogScopeActivity,
	})
	logPublic(ctx, LogLevelError, msg, slog.Any(errorKey, err.Error()))
}

func LogNamespace(level LogLevel, namespace, msg string) {
	ctx := context.WithValue(context.Background(), logObjectCtx, LogObject{
		Namespace: namespace,
		ID:        namespace,
		Scope:     LogScopeNamespace,
	})
	logPublic(ctx, level, msg)
}

func LogNamespaceError(namespace, msg string, err error) {
	ctx := context.WithValue(context.Background(), logObjectCtx, LogObject{
		Namespace: namespace,
		ID:        namespace,
		Scope:     LogScopeNamespace,
	})
	logPublic(ctx, LogLevelError, msg, slog.Any(errorKey, err.Error()))
}

func LogInstance(ctx context.Context, level LogLevel, msg string) {
	logPublic(ctx, level, msg)
}

func LogInstanceError(ctx context.Context, msg string, err error) {
	logPublic(ctx, LogLevelError, msg, slog.Any(errorKey, err.Error()))
}

func logPublic(ctx context.Context, level LogLevel, msg string, attrs ...slog.Attr) {
	switch level {
	case LogLevelDebug:
		slog.DebugContext(ctx, msg)
	case LogLevelError:
		slog.ErrorContext(ctx, msg)
	case LogLevelWarn:
		slog.WarnContext(ctx, msg)
	default:
		slog.InfoContext(ctx, msg)
	}
}

func LogInitCtx(ctx context.Context, logObject LogObject) context.Context {
	return context.WithValue(ctx, logObjectCtx, logObject)
}
