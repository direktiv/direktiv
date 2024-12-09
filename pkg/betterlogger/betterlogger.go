package betterlogger

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/tracing"
)

type TrackAble interface {
	ShowInInstanceView() UserLogger
	ShowInNamespaceView() UserLogger
	ShowInGatewayView() UserLogger
	ShowInMirrorView() UserLogger
	ConsoleLogs() UserLogger
}

type UserLogger interface {
	DebugContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
}

func WithNamespace(attr coreNamespaceAttributes) TrackAble {
	return &logUtil{
		coreNamespaceAttributes: attr,
	}
}

func WithInstance(attr InstanceAttributes) TrackAble {
	return &logUtil{
		coreInstanceAttr:        attr.coreInstanceAttr,
		coreNamespaceAttributes: attr.coreNamespaceAttributes,
	}
}

func WithInstanceMemory(attr InstanceMemoryAttributes) TrackAble {
	return &logUtil{
		coreInstanceMemoryAttributes: attr.coreInstanceMemoryAttributes,
		coreInstanceAttr:             attr.coreInstanceAttr,
		coreNamespaceAttributes:      attr.coreNamespaceAttributes,
	}
}

func WithInstanceAction(attr InstanceMemoryAttributes) TrackAble {
	return &logUtil{
		coreInstanceMemoryAttributes: attr.coreInstanceMemoryAttributes,
		coreInstanceAttr:             attr.coreInstanceAttr,
		coreNamespaceAttributes:      attr.coreNamespaceAttributes,
	}
}

func WithMirror(attr CloudEventBusAttributes) TrackAble {
	return &logUtil{
		coreCloudEventBusAttributes: attr.coreCloudEventBusAttributes,
		coreNamespaceAttributes:     attr.coreNamespaceAttributes,
	}
}

func WithGatewayRoutes(attr GatewayAttributes) TrackAble {
	return &logUtil{
		coreNamespaceAttributes: attr.coreNamespaceAttributes,
		coreGatewayAttributes:   attr.coreGatewayAttributes,
	}
}

func WithEventProcessing(attr SyncAttributes) TrackAble {
	return &logUtil{
		coreNamespaceAttributes: attr.coreNamespaceAttributes,
		coreSyncAttributes:      attr.coreSyncAttributes,
	}
}

func TODO() UserLogger {
	return &logUtil{}
}

var _ UserLogger = &logUtil{}

var _ TrackAble = &logUtil{}

type logUtil struct {
	coreNamespaceAttributes
	coreInstanceAttr
	coreInstanceMemoryAttributes
	coreCloudEventBusAttributes
	coreGatewayAttributes
	coreSyncAttributes
	coreInstanceActionAttr
	track string // TODO: ensure thread safe handling when building the logger
}

// DebugContext implements logger.
func (l *logUtil) DebugContext(ctx context.Context, msg string, args ...any) {
	ctx = l.ctxBuilder(ctx)
	slog.DebugContext(ctx, msg, args...)
}

// ErrorContext implements logger.
func (l *logUtil) ErrorContext(ctx context.Context, msg string, args ...any) {
	ctx = l.ctxBuilder(ctx)
	slog.ErrorContext(ctx, msg, args...)
}

// InfoContext implements logger.
func (l *logUtil) InfoContext(ctx context.Context, msg string, args ...any) {
	ctx = l.ctxBuilder(ctx)
	slog.InfoContext(ctx, msg, args...)
}

// WarnContext implements logger.
func (l *logUtil) WarnContext(ctx context.Context, msg string, args ...any) {
	ctx = l.ctxBuilder(ctx)
	slog.WarnContext(ctx, msg, args...)
}

// ShowInGatewayView implements TrackAble.
func (l *logUtil) ShowInGatewayView() UserLogger {
	l.track = fmt.Sprintf("%v.%v", "route", l.Route)
	return l
}

// ShowInMirrorView implements TrackAble.
func (l *logUtil) ShowInMirrorView() UserLogger {
	l.track = fmt.Sprintf("%v.%v", "activity", l.SyncID)
	return l
}

// ShowInInstanceView implements TrackAble.
func (l *logUtil) ShowInInstanceView() UserLogger {
	l.track = fmt.Sprintf("%v.%v", "instance", l.CallPath)
	return l
}

// ShowInNamespaceView implements TrackAble.
func (l *logUtil) ShowInNamespaceView() UserLogger {
	l.track = fmt.Sprintf("%v.%v", "namespace", l.Namespace)
	return l
}

// ConsoleLogs implements TrackAble.
func (l *logUtil) ConsoleLogs() UserLogger {
	return l
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

type coreNamespaceAttributes struct {
	Namespace string
}

// GatewayAttributes holds metadata specific to a gateway component, helpful for logging.
type GatewayAttributes struct {
	coreNamespaceAttributes
	coreGatewayAttributes
}

type coreGatewayAttributes struct {
	coreNamespaceAttributes
	Plugin string // Optional. Name for the gateway-plugin
	Route  string // Endpoint of the gateway
}

// CloudEventBusAttributes holds metadata specific to the cloud event bus.
type CloudEventBusAttributes struct {
	coreNamespaceAttributes
	coreCloudEventBusAttributes
}
type coreCloudEventBusAttributes struct {
	EventID   string // Unique identifier for the event
	Source    string // Source of the event
	Subject   string // Subject of the event
	EventType string // Type of the event
}

// InstanceMemoryAttributes holds common metadata for an instance, which is helpful for logging.
type InstanceMemoryAttributes struct {
	coreNamespaceAttributes
	coreInstanceAttr
	coreInstanceMemoryAttributes
}

type InstanceActionAttributes struct {
	coreNamespaceAttributes
	coreInstanceAttr
	coreInstanceMemoryAttributes
	coreInstanceActionAttr
}

// InstanceAttributes holds common metadata for an instance, which is helpful for logging.
type InstanceAttributes struct {
	coreNamespaceAttributes
	coreInstanceAttr
}

type coreInstanceAttr struct {
	InstanceID   string // Unique identifier for the instance
	WorkflowPath string // Path of the workflow the instance belongs to
	CallPath     string // Identifies the log-stream, legacy feature from the old engine
}

type coreInstanceActionAttr struct {
	ActionID string // Unique identifier for the instance action
}

type coreInstanceMemoryAttributes struct {
	State string // Memory state of the instance
}

// SyncAttributes holds metadata specific to a Sync component, helpful for logging.
type SyncAttributes struct {
	coreSyncAttributes
	coreNamespaceAttributes
}

// SyncAttributes holds metadata specific to a Sync component, helpful for logging.
type coreSyncAttributes struct {
	SyncID string // Unique identifier for the Sync
}
