package tracing

// Package tracing provides a structured userlog system.
//
// Usage Example:
//
//	logger, err := tracing.WithNamespace(coreNamespaceAttributes{Namespace: "example-namespace"}).ShowInNamespaceView()
//	if err != nil {
//		panic(error.Error())
//	}
//	logger.InfoContext(ctx, "Namespace log entry", "key", "value")
import (
	"context"
	"fmt"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/core"
)

// WithNamespace initializes a TrackAble logger with namespace-specific attributes.
func WithNamespace(namespace string) TrackAble {
	return newLogUtil(logUtil{
		Namespace: namespace,
	})
}

// WithInstance initializes a TrackAble logger for an instance.
func WithInstance(attr InstanceAttributes) TrackAble {
	return newLogUtil(logUtil{
		Namespace:    attr.Namespace,
		InstanceID:   attr.InstanceID,
		WorkflowPath: attr.WorkflowPath,
		CallPath:     attr.Callpath,
	})
}

// WithInstanceMemory initializes a TrackAble logger for an instance memory context.
func WithInstanceMemory(attr InstanceMemoryAttributes) TrackAble {
	return newLogUtil(logUtil{
		Namespace:    attr.Namespace,
		InstanceID:   attr.InstanceID,
		WorkflowPath: attr.WorkflowPath,
		CallPath:     attr.Callpath,
		State:        attr.State,
	})
}

// WithInstanceAction initializes a TrackAble logger for an instance action context.
func WithInstanceAction(attr InstanceActionAttributes) TrackAble {
	return newLogUtil(logUtil{
		Namespace:    attr.Namespace,
		InstanceID:   attr.InstanceID,
		WorkflowPath: attr.WorkflowPath,
		CallPath:     attr.Callpath,
		State:        attr.State,
		ActionID:     attr.ActionID,
	})
}

// WithMirror initializes a TrackAble logger for a Cloud Event Bus context.
func WithMirror(attr CloudEventBusAttributes) TrackAble {
	return newLogUtil(logUtil{
		Namespace: attr.Namespace,
		EventID:   attr.EventID,
		Source:    attr.Source,
		Subject:   attr.Subject,
		EventType: attr.EventType,
	})
}

// WithGatewayRoutes initializes a TrackAble logger for a gateway route context.
func WithGatewayRoutes(attr GatewayAttributes) TrackAble {
	return newLogUtil(logUtil{
		Namespace: attr.Namespace,
		Plugin:    attr.Plugin,
		Route:     attr.Route,
	})
}

// WithEventProcessing initializes a TrackAble logger for event processing.
func WithEventProcessing(attr SyncAttributes) TrackAble {
	return newLogUtil(logUtil{
		Namespace: attr.Namespace,
		SyncID:    attr.SyncID,
	})
}

// TODO creates a default UserLogger with no specific attributes.
func TODO() UserLogger {
	return &logUtilWithTrack{logUtil: &logUtil{}}
}

var _ UserLogger = &logUtilWithTrack{}

var _ TrackAble = &logUtil{}

type logUtil struct {
	Namespace    string
	InstanceID   string
	WorkflowPath string
	CallPath     string
	State        string
	EventID      string
	Source       string
	Subject      string
	EventType    string
	Plugin       string
	Route        string
	SyncID       string
	ActionID     string
}

type logUtilWithTrack struct {
	*logUtil
	track string
}

func newLogUtil(base logUtil) *logUtil {
	return &base
}

// DebugContext implements logger.
func (l *logUtilWithTrack) DebugContext(ctx context.Context, msg string, args ...any) {
	ctx = l.ctxBuilder(ctx)
	ctx = WithTrack(ctx, l.track)
	slog.DebugContext(ctx, msg, args...)
}

// ErrorContext implements logger.
func (l *logUtilWithTrack) ErrorContext(ctx context.Context, msg string, args ...any) {
	ctx = l.ctxBuilder(ctx)
	ctx = WithTrack(ctx, l.track)
	slog.ErrorContext(ctx, msg, args...)
}

// InfoContext implements logger.
func (l *logUtilWithTrack) InfoContext(ctx context.Context, msg string, args ...any) {
	ctx = l.ctxBuilder(ctx)
	ctx = WithTrack(ctx, l.track)
	slog.InfoContext(ctx, msg, args...)
}

// WarnContext implements logger.
func (l *logUtilWithTrack) WarnContext(ctx context.Context, msg string, args ...any) {
	ctx = l.ctxBuilder(ctx)
	ctx = WithTrack(ctx, l.track)
	slog.WarnContext(ctx, msg, args...)
}

// ShowInGatewayView implements TrackAble.
func (l *logUtil) ShowInGatewayView() (UserLogger, error) {
	return &logUtilWithTrack{
		track:   fmt.Sprintf("%v.%v", "route", l.Route),
		logUtil: l,
	}, nil
}

// ShowInMirrorView implements TrackAble.
func (l *logUtil) ShowInMirrorView() (UserLogger, error) {
	if len(l.SyncID) == 0 {
		return nil, fmt.Errorf("failed to build logger")
	}

	return &logUtilWithTrack{
		track:   fmt.Sprintf("%v.%v", "activity", l.SyncID),
		logUtil: l,
	}, nil
}

// ShowInInstanceView implements TrackAble.
func (l *logUtil) ShowInInstanceView() (UserLogger, error) {
	if len(l.CallPath) == 0 {
		return nil, fmt.Errorf("failed to build logger")
	}

	return &logUtilWithTrack{
		track:   fmt.Sprintf("%v.%v", "instance", l.CallPath),
		logUtil: l,
	}, nil
}

// ShowInNamespaceView implements TrackAble.
func (l *logUtil) ShowInNamespaceView() (UserLogger, error) {
	if len(l.Namespace) == 0 {
		return nil, fmt.Errorf("failed to build logger")
	}

	return &logUtilWithTrack{
		track:   fmt.Sprintf("%v.%v", "namespace", l.Namespace),
		logUtil: l,
	}, nil
}

// ConsoleLogs implements TrackAble.
func (l *logUtil) ConsoleLogs() (UserLogger, error) {
	return &logUtilWithTrack{
		logUtil: l,
	}, nil
}

func (l *logUtil) ctxBuilder(ctx context.Context) context.Context {
	if len(l.State) != 0 {
		ctx = AddInstanceMemoryAttr(ctx, InstanceAttributes{
			Namespace:    l.Namespace,
			InstanceID:   l.InstanceID,
			Invoker:      "todo_dummy_value",
			WorkflowPath: l.WorkflowPath,
			Status:       core.LogUnknownStatus,
			Callpath:     l.CallPath,
		}, l.State)
	} else if len(l.CallPath) != 0 {
		ctx = AddInstanceAttr(ctx, InstanceAttributes{
			Namespace:  l.Namespace,
			InstanceID: l.InstanceID,
			// Invoker:      "todo_dummy_value",
			WorkflowPath: l.WorkflowPath,
			Callpath:     l.CallPath,
		})
	} else if len(l.Namespace) != 0 {
		ctx = AddNamespace(ctx, l.Namespace)
	}
	if len(l.ActionID) != 0 {
		ctx = AddActionID(ctx, l.ActionID)
	}

	return ctx
}
