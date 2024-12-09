// Package betterlogger provides a structured userlog system.
//
// Usage Example:
//
//	logger, err := betterlogger.WithNamespace(coreNamespaceAttributes{Namespace: "example-namespace"}).ShowInNamespaceView()
//	if err != nil {
//		panic(error.Error())
//	}
//	logger.InfoContext(ctx, "Namespace log entry", "key", "value")
package betterlogger

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/tracing"
)

// WithNamespace initializes a TrackAble logger with namespace-specific attributes.
//
// Parameters:
// - attr: Attributes associated with the namespace.
//
// Returns:
// - A TrackAble logger configured with the provided namespace attributes.
func WithNamespace(attr coreNamespaceAttributes) TrackAble {
	return newLogUtil(logUtil{coreNamespaceAttributes: attr})
}

// WithInstance initializes a TrackAble logger for an instance.
//
// Parameters:
// - attr: Attributes specific to an instance.
//
// Returns:
// - A TrackAble logger configured with the provided instance attributes.
func WithInstance(attr InstanceAttributes) TrackAble {
	return newLogUtil(logUtil{}, attr.coreInstanceAttr, attr.coreNamespaceAttributes)
}

// WithInstanceMemory initializes a TrackAble logger for an instance memory context.
//
// Parameters:
// - attr: Attributes specific to instance memory.
//
// Returns:
// - A TrackAble logger configured with the provided memory attributes.
func WithInstanceMemory(attr InstanceMemoryAttributes) TrackAble {
	return newLogUtil(logUtil{}, attr.coreInstanceMemoryAttributes, attr.coreInstanceAttr, attr.coreNamespaceAttributes)
}

// WithInstanceAction initializes a TrackAble logger for an instance action context.
//
// Parameters:
// - attr: Attributes specific to an instance action.
//
// Returns:
// - A TrackAble logger configured with the provided action attributes.
func WithInstanceAction(attr InstanceActionAttributes) TrackAble {
	return newLogUtil(logUtil{}, attr.coreInstanceActionAttr, attr.coreInstanceMemoryAttributes, attr.coreInstanceAttr, attr.coreNamespaceAttributes)
}

// WithMirror initializes a TrackAble logger for a Cloud Event Bus context.
//
// Parameters:
// - attr: Attributes specific to the Cloud Event Bus.
//
// Returns:
// - A TrackAble logger configured with the provided event bus attributes.
func WithMirror(attr CloudEventBusAttributes) TrackAble {
	return newLogUtil(logUtil{}, attr.coreCloudEventBusAttributes, attr.coreNamespaceAttributes)
}

// WithGatewayRoutes initializes a TrackAble logger for a gateway route context.
//
// Parameters:
// - attr: Attributes specific to the gateway routes.
//
// Returns:
// - A TrackAble logger configured with the provided gateway attributes.
func WithGatewayRoutes(attr GatewayAttributes) TrackAble {
	return newLogUtil(logUtil{}, attr.coreGatewayAttributes, attr.coreNamespaceAttributes)
}

// WithEventProcessing initializes a TrackAble logger for event processing.
//
// Parameters:
// - attr: Attributes specific to event synchronization.
//
// Returns:
// - A TrackAble logger configured with the provided synchronization attributes.
func WithEventProcessing(attr SyncAttributes) TrackAble {
	return newLogUtil(logUtil{}, attr.coreSyncAttributes, attr.coreNamespaceAttributes)
}

// TODO creates a default UserLogger with no specific attributes.
//
// Returns:
// - A UserLogger instance for general-purpose logging.
func TODO() UserLogger {
	return &logUtilWithTrack{logUtil: &logUtil{}}
}

var _ UserLogger = &logUtilWithTrack{}

var _ TrackAble = &logUtil{}

type logUtil struct {
	coreNamespaceAttributes
	coreInstanceAttr
	coreInstanceMemoryAttributes
	coreCloudEventBusAttributes
	coreGatewayAttributes
	coreSyncAttributes
	coreInstanceActionAttr
}

type logUtilWithTrack struct {
	*logUtil
	track string
}

func newLogUtil(base logUtil, additionalAttributes ...any) *logUtil {
	// Create a new instance of logUtil and apply any additional attributes
	newLogger := base
	for _, attr := range additionalAttributes {
		switch v := attr.(type) {
		case coreInstanceAttr:
			newLogger.coreInstanceAttr = v
		case coreInstanceMemoryAttributes:
			newLogger.coreInstanceMemoryAttributes = v
		case coreCloudEventBusAttributes:
			newLogger.coreCloudEventBusAttributes = v
		case coreGatewayAttributes:
			newLogger.coreGatewayAttributes = v
		case coreSyncAttributes:
			newLogger.coreSyncAttributes = v
		case coreInstanceActionAttr:
			newLogger.coreInstanceActionAttr = v
		}
	}

	return &newLogger
}

// DebugContext implements logger.
func (l *logUtilWithTrack) DebugContext(ctx context.Context, msg string, args ...any) {
	ctx = l.ctxBuilder(ctx)
	ctx = tracing.WithTrack(ctx, l.track)
	slog.DebugContext(ctx, msg, args...)
}

// ErrorContext implements logger.
func (l *logUtilWithTrack) ErrorContext(ctx context.Context, msg string, args ...any) {
	ctx = l.ctxBuilder(ctx)
	ctx = tracing.WithTrack(ctx, l.track)
	slog.ErrorContext(ctx, msg, args...)
}

// InfoContext implements logger.
func (l *logUtilWithTrack) InfoContext(ctx context.Context, msg string, args ...any) {
	ctx = l.ctxBuilder(ctx)
	ctx = tracing.WithTrack(ctx, l.track)
	slog.InfoContext(ctx, msg, args...)
}

// WarnContext implements logger.
func (l *logUtilWithTrack) WarnContext(ctx context.Context, msg string, args ...any) {
	ctx = l.ctxBuilder(ctx)
	ctx = tracing.WithTrack(ctx, l.track)
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
		ctx = tracing.AddInstanceMemoryAttr(ctx, tracing.InstanceAttributes{
			Namespace:    l.Namespace,
			InstanceID:   l.InstanceID,
			Invoker:      "todo_dummy_value",
			WorkflowPath: l.WorkflowPath,
			Status:       core.LogUnknownStatus,
			Callpath:     l.CallPath,
		}, l.State)
	} else if len(l.CallPath) != 0 {
		ctx = tracing.AddInstanceAttr(ctx, tracing.InstanceAttributes{
			Namespace:    l.Namespace,
			InstanceID:   l.InstanceID,
			Invoker:      "todo_dummy_value",
			WorkflowPath: l.WorkflowPath,
			Status:       core.LogUnknownStatus,
			Callpath:     l.CallPath,
		})
	} else if len(l.Namespace) != 0 {
		ctx = tracing.AddNamespace(ctx, l.Namespace)
	}
	if len(l.ActionID) != 0 {
		ctx = tracing.AddActionID(ctx, l.ActionID)
	}

	return ctx
}
