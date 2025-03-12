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
	Trace     string   `json:"trace,omitempty"`
	Span      string   `json:"span,omitempty"`
	Scope     LogScope `json:"scope"`
	InstanceInfo
}

type InstanceInfo struct {
	// Namespace string         `json:"namespace2,omitempty"`
	// Instance  string         `json:"instance,omitempty"`
	Invoker  string         `json:"invoker,omitempty"`
	Callpath string         `json:"callpath,omitempty"`
	Path     string         `json:"path,omitempty"`
	State    string         `json:"state,omitempty"`
	Status   core.LogStatus `json:"status,omitempty"`
	// Trace    string         `json:"trace2,omitempty"`
	// Span     string         `json:"span2,omitempty"`
}

// type HTTPInstanceInfo struct {
// 	InstanceInfo
// 	Msg   string   `json:"msg"`
// 	Level LogLevel `json:"level"`
// }

// func (ii *HTTPInstanceInfo) GetInstanceInfo() InstanceInfo {
// 	return ii.InstanceInfo
// }

const (
	LogLevelDebug LogLevel = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelError

	LogScopeInstance  LogScope = "instance"
	LogScopeNamespace LogScope = "namespace"
	LogScopeActivity  LogScope = "activity"
	LogScopeRoute     LogScope = "route"

	errorKey     = "error"
	scopeKey     = "scope"
	namespaceKey = "namespace"
	traceKey     = "trace"
	spanKey      = "span"

	instanceKey = "instance"
	routeKey    = "route"

	DirektivInstance = "instance-ctx"
	LogObjectCtx     = "log-ctx"
)

func LogRoute(level LogLevel, namespace, route, msg string) {
	ctx := context.WithValue(context.Background(), LogObjectCtx, LogObject{
		Namespace: namespace,
		ID:        route,
		Scope:     LogScopeRoute,
	})
	logPublic(ctx, level, msg)
}

func LogActivity(level LogLevel, namespace, pid, msg string) {
	ctx := context.WithValue(context.Background(), LogObjectCtx, LogObject{
		Namespace: namespace,
		ID:        pid,
		Scope:     LogScopeActivity,
	})
	logPublic(ctx, level, msg)
}

func LogNamespace(level LogLevel, namespace, msg string) {
	ctx := context.WithValue(context.Background(), LogObjectCtx, LogObject{
		Namespace: namespace,
		ID:        namespace,
		Scope:     LogScopeNamespace,
	})
	logPublic(ctx, level, msg)
}

func LogInstance(ctx context.Context, level LogLevel, msg string) {
	logPublic(ctx, level, msg)
}

func logPublic(ctx context.Context, level LogLevel, msg string) {
	// set span and trace!!

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

func LogInitInstance(ctx context.Context, logObject LogObject) context.Context {
	// add opentelemtry if it exists
	// span := trace.SpanFromContext(ctx)
	// if span.SpanContext().TraceID().IsValid() {
	// 	// info.Trace = span.SpanContext().TraceID().String()
	// 	// info.Span = span.SpanContext().SpanID().String()
	// }

	// check if required fields are set
	return context.WithValue(ctx, LogObjectCtx, logObject)
}

// func logInstance(ctx context.Context, msg string, lvl LogLevel, err error) {
// 	instanceID := ""
// 	i := ctx.Value(DirektivInstance)

// 	if i == nil {
// 		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!IT IS NIL")
// 	}

// 	if i != nil {
// 		info, ok := i.(InstanceInfo)
// 		if !ok {
// 			slog.Error("instance info not the expected type")
// 			return
// 		}
// 		instanceID = info.Instance
// 	}

// 	if instanceID == "" {
// 		slog.Error("instance id is empty")
// 		return
// 	}

// 	switch lvl {
// 	case LogLevelInfo:
// 		slog.InfoContext(ctx, msg, slog.String(scopeKey, "instance."+instanceID))
// 	case LogLevelDebug:
// 		slog.DebugContext(ctx, msg, slog.String(scopeKey, "instance."+instanceID))
// 	case LogLevelWarn:
// 		slog.WarnContext(ctx, msg, slog.String(scopeKey, "instance."+instanceID))
// 	case LogLevelError:
// 		slog.ErrorContext(ctx, msg, slog.Any(errorKey, err.Error()), slog.String(scopeKey, "instance."+instanceID))
// 	}
// }

// func LogInstanceError(ctx context.Context, msg string, err error) {
// 	logInstance(ctx, msg, LogLevelError, err)
// }

// func LogInstanceInfo(ctx context.Context, msg string) {
// 	logInstance(ctx, msg, LogLevelInfo, nil)
// }

// func LogInstanceDebug(ctx context.Context, msg string) {
// 	logInstance(ctx, msg, LogLevelDebug, nil)
// }

// func LogInstanceWarn(ctx context.Context, msg string) {
// 	logInstance(ctx, msg, LogLevelWarn, nil)
// }
