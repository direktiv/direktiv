package telemetry

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/direktiv/direktiv/internal/core"
)

type (
	LogLevel int
	LogScope string
)

type LogObject struct {
	InstanceInfo

	Namespace string   `json:"namespace"`
	ID        string   `json:"id"`
	Scope     LogScope `json:"scope"`
}

func LogObjectFromHeader(ctx context.Context, header http.Header) LogObject {
	return LogObject{
		InstanceInfo: InstanceInfo{
			Invoker: header.Get(core.EngineHeaderInvoker),
			Path:    header.Get(core.EngineHeaderPath),
			State:   header.Get(core.EngineHeaderState),
			Status:  core.LogStatus(header.Get(core.EngineHeaderStatus)),
		},
		Namespace: header.Get(core.EngineHeaderNamespace),
		ID:        header.Get(core.EngineHeaderActionID),
		Scope:     LogScope(header.Get(core.EngineHeaderScope)),
	}
}

func (l LogObject) ToHeader(header *http.Header) {
	header.Set(core.EngineHeaderState, l.State)
	header.Set(core.EngineHeaderStatus, string(l.Status))
	header.Set(core.EngineHeaderScope, string(l.Scope))
	header.Set(core.EngineHeaderInvoker, l.Invoker)
	header.Set(core.EngineHeaderPath, l.Path)
	header.Set(core.EngineHeaderNamespace, l.Namespace)
	header.Set(core.EngineHeaderActionID, l.ID)
}

type InstanceInfo struct {
	Invoker string         `json:"invoker,omitempty"`
	Path    string         `json:"path,omitempty"`
	State   string         `json:"state,omitempty"`
	Status  core.LogStatus `json:"status,omitempty"`
}

// HTTPInstanceInfo is used to post logs from the sidecar.
// fluentbit can not be used because it is picking up the logs of newly created pods too late.
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
)

const (
	LogScopeInstance  LogScope = "instance"
	LogScopeNamespace LogScope = "namespace"
	LogScopeActivity  LogScope = "activity"
	LogScopeRoute     LogScope = "route"
)

const (
	errorKey            = "error"
	LogObjectIdentifier = "log-ctx"
)

type DirektivLogCtx string

func LogRoute(level LogLevel, namespace, route, msg string) {
	ctx := context.WithValue(context.TODO(), DirektivLogCtx(LogObjectIdentifier), LogObject{
		Namespace: namespace,
		ID:        route,
		Scope:     LogScopeRoute,
	})
	logPublic(ctx, level, msg)
}

func LogRouteError(namespace, route, msg string, err error) {
	ctx := context.WithValue(context.TODO(), DirektivLogCtx(LogObjectIdentifier), LogObject{
		Namespace: namespace,
		ID:        route,
		Scope:     LogScopeRoute,
	})
	logPublic(ctx, LogLevelError, msg, slog.Any(errorKey, err.Error()))
}

func LogActivity(level LogLevel, namespace, pid, msg string) {
	ctx := context.WithValue(context.TODO(), DirektivLogCtx(LogObjectIdentifier), LogObject{
		Namespace: namespace,
		ID:        pid,
		Scope:     LogScopeActivity,
	})
	logPublic(ctx, level, msg)
}

func LogActivityError(namespace, pid, msg string, err error) {
	ctx := context.WithValue(context.TODO(), DirektivLogCtx(LogObjectIdentifier), LogObject{
		Namespace: namespace,
		ID:        pid,
		Scope:     LogScopeActivity,
	})
	logPublic(ctx, LogLevelError, msg, slog.Any(errorKey, err.Error()))
}

func LogNamespace(level LogLevel, namespace, msg string) {
	ctx := context.WithValue(context.TODO(), DirektivLogCtx(LogObjectIdentifier), LogObject{
		Namespace: namespace,
		ID:        namespace,
		Scope:     LogScopeNamespace,
	})
	logPublic(ctx, level, msg)
}

func LogNamespaceError(namespace, msg string, err error) {
	ctx := context.WithValue(context.TODO(), DirektivLogCtx(LogObjectIdentifier), LogObject{
		Namespace: namespace,
		ID:        namespace,
		Scope:     LogScopeNamespace,
	})
	logPublic(ctx, LogLevelError, msg, slog.Any(errorKey, err.Error()))
}

func SetInstanceLogState(ctx context.Context, state string) context.Context {
	logObject := ctx.Value(DirektivLogCtx(LogObjectIdentifier)).(LogObject)
	logObject.State = state
	return context.WithValue(ctx, DirektivLogCtx(LogObjectIdentifier), logObject)
}

func SetInstanceLogStatus(ctx context.Context, status core.LogStatus) context.Context {
	logObject := ctx.Value(DirektivLogCtx(LogObjectIdentifier)).(LogObject)
	logObject.Status = status
	return context.WithValue(ctx, DirektivLogCtx(LogObjectIdentifier), logObject)
}

func SetupInstanceLogs(ctx context.Context, namespace, id, invoker, path string) context.Context {
	ctx = context.WithValue(ctx, DirektivLogCtx(LogObjectIdentifier),
		LogObject{
			Namespace: namespace,
			ID:        id,
			Scope:     LogScopeInstance,
			InstanceInfo: InstanceInfo{
				Invoker: invoker,
				Path:    path,
				Status:  core.LogRunningStatus,
				State:   "pre-run",
			},
		})

	return ctx
}

func LogInstance(ctx context.Context, level LogLevel, msg string) {
	logPublic(ctx, level, msg)
}

func LogInstanceError(ctx context.Context, msg string, err error) {
	logPublic(ctx, LogLevelError, msg, slog.Any(errorKey, err.Error()))
}

func logPublic(ctx context.Context, level LogLevel, msg string, attrs ...any) {
	switch level {
	case LogLevelDebug:
		slog.DebugContext(ctx, msg, attrs...)
	case LogLevelError:
		slog.ErrorContext(ctx, msg, attrs...)
	case LogLevelWarn:
		slog.WarnContext(ctx, msg, attrs...)
	case LogLevelInfo:
		slog.InfoContext(ctx, msg, attrs...)
	}
}

func LogInitCtx(ctx context.Context, logObj LogObject) context.Context {
	return context.WithValue(ctx, DirektivLogCtx(LogObjectIdentifier), logObj)
}
