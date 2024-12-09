package tracing

// Package tracing provides a structured logs system.
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

type TrackAble interface {
	// ShowInInstanceView configures the logger to focus on instance-related logs.
	//
	// Returns:
	// - A UserLogger with tracking information for instance logs.
	ShowInInstanceView() (UserLogger, error)
	// ShowInNamespaceView configures the logger to focus on namespace-related logs.
	//
	// Returns:
	// - A UserLogger with tracking information for namespace logs.
	ShowInNamespaceView() (UserLogger, error)
	// ShowInGatewayView configures the logger to focus on gateway-related logs.
	//
	// Returns:
	// - A UserLogger with tracking information for gateway logs.
	ShowInGatewayView() (UserLogger, error)
	// ShowInMirrorView configures the logger to focus on mirror-related logs.
	//
	// Returns:
	// - A UserLogger with tracking information for mirror logs.
	ShowInMirrorView() (UserLogger, error)
	// ConsoleLogs configures the logger for console-specific logging.
	//
	// Returns:
	// - A UserLogger for console logs.
	ConsoleLogs() (UserLogger, error)
}

type UserLogger interface {
	DebugContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
}

// GatewayAttributes holds metadata specific to a gateway component, helpful for logging.
type GatewayAttributes struct {
	Namespace string
	Plugin    string // Optional. Name for the gateway-plugin
	Route     string // Endpoint of the gateway
}

// CloudEventBusAttributes holds metadata specific to the cloud event bus.
type CloudEventBusAttributes struct {
	Namespace string
	EventID   string // Unique identifier for the event
	Source    string // Source of the event
	Subject   string // Subject of the event
	EventType string // Type of the event
}

// InstanceMemoryAttributes holds common metadata for an instance, which is helpful for logging.
type InstanceMemoryAttributes struct {
	Namespace    string
	InstanceID   string // Unique identifier for the instance
	WorkflowPath string // Path of the workflow the instance belongs to
	Callpath     string // Identifies the log-stream, legacy feature from the old engine
	State        string // Memory state of the instance
	Invoker      string
	Status       core.LogStatus
}

// InstanceActionAttributes holds metadata for an instance action.
type InstanceActionAttributes struct {
	Namespace    string
	InstanceID   string // Unique identifier for the instance
	WorkflowPath string // Path of the workflow the instance belongs to
	Callpath     string // Identifies the log-stream, legacy feature from the old engine
	State        string // Memory state of the instance
	ActionID     string // Unique identifier for the instance action
	Invoker      string
	Status       core.LogStatus
}

// InstanceAttributes holds common metadata for an instance, which is helpful for logging.
type InstanceAttributes struct {
	Namespace    string
	InstanceID   string // Unique identifier for the instance
	WorkflowPath string // Path of the workflow the instance belongs to
	Callpath     string // Identifies the log-stream, legacy feature from the old engine
	Invoker      string
	Status       core.LogStatus
}

// SyncAttributes holds metadata specific to a Sync component, helpful for logging.
type SyncAttributes struct {
	Namespace string
	SyncID    string // Unique identifier for the Sync
}

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
	Invoker      string
	State        string
	Status       core.LogStatus
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
			Invoker:      l.Invoker,
			WorkflowPath: l.WorkflowPath,
			Status:       l.Status,
			Callpath:     l.CallPath,
		}, l.State)
	} else if len(l.CallPath) != 0 {
		ctx = AddInstanceAttr(ctx, InstanceAttributes{
			Namespace:    l.Namespace,
			InstanceID:   l.InstanceID,
			Invoker:      l.Invoker,
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
