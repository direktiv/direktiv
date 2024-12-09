package betterlogger

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/betterlogger/internal"
)

type UserLogger interface {
	DebugContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, err error, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
}

func ForNamespace(attr coreNamespaceAttributes) UserLogger {
	return &logUtil{
		coreNamespaceAttributes: attr,
	}
}

func ForInstance(attr InstanceAttributes) UserLogger {
	return &logUtil{
		coreInstanceAttr:        attr.coreInstanceAttr,
		coreNamespaceAttributes: attr.coreNamespaceAttributes,
	}
}

func ForInstanceMemory(attr InstanceMemoryAttributes) UserLogger {
	return &logUtil{
		coreInstanceMemoryAttributes: attr.coreInstanceMemoryAttributes,
		coreInstanceAttr:             attr.coreInstanceAttr,
		coreNamespaceAttributes:      attr.coreNamespaceAttributes,
	}
}

func ForMirror(attr CloudEventBusAttributes) UserLogger {
	return &logUtil{
		coreCloudEventBusAttributes: attr.coreCloudEventBusAttributes,
		coreNamespaceAttributes:     attr.coreNamespaceAttributes,
	}
}

func ForGatewayRoutes(attr GatewayAttributes) UserLogger {
	return &logUtil{
		coreNamespaceAttributes: attr.coreNamespaceAttributes,
		coreGatewayAttributes:   attr.coreGatewayAttributes,
	}
}

func ForEvenProcessing(attr SyncAttributes) UserLogger {
	return &logUtil{
		coreNamespaceAttributes: attr.coreNamespaceAttributes,
		coreSyncAttributes:      attr.coreSyncAttributes,
	}
}

var _ UserLogger = logUtil{}

type logUtil struct {
	coreNamespaceAttributes
	coreInstanceAttr
	coreInstanceMemoryAttributes
	coreCloudEventBusAttributes
	coreGatewayAttributes
	coreSyncAttributes
	track string
}

// DebugContext implements logger.
func (l logUtil) DebugContext(ctx context.Context, msg string, args ...any) {
	panic("unimplemented")
}

// ErrorContext implements logger.
func (l logUtil) ErrorContext(ctx context.Context, msg string, err error, args ...any) {
	panic("unimplemented")
}

// InfoContext implements logger.
func (l logUtil) InfoContext(ctx context.Context, msg string, args ...any) {
	panic("unimplemented")
}

// WarnContext implements logger.
func (l logUtil) WarnContext(ctx context.Context, msg string, args ...any) {
	panic("unimplemented")
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

type coreInstanceMemoryAttributes struct {
	State string // Memory state of the instance
}

// buildSyncAttributes constructs the base attributes for mirror logging.
func buildSyncAttributes(attr SyncAttributes, additionalAttrs ...slog.Attr) []slog.Attr {
	baseAttrs := []slog.Attr{
		slog.String("activity", attr.SyncID),
		slog.String("namespace", attr.Namespace),
		slog.String("track", fmt.Sprintf("%v.%v", "activity", attr.SyncID)),
	}

	return internal.MergeAttributes(baseAttrs, additionalAttrs...)
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

// buildInstanceAttributes constructs the base attributes for instance logging.
func buildInstanceAttributes(attr InstanceAttributes, additionalAttrs ...slog.Attr) []slog.Attr {
	baseAttrs := []slog.Attr{
		slog.String("namespace", attr.Namespace),
		slog.String("instance", attr.InstanceID),
		slog.String("workflow", attr.WorkflowPath),
		slog.String("track", fmt.Sprintf("%v.%v", "instance", attr.CallPath)),
	}

	return internal.MergeAttributes(baseAttrs, additionalAttrs...)
}

// buildInstanceMemoryAttributes constructs the base attributes for instance memory logging.
func buildInstanceMemoryAttributes(attr InstanceMemoryAttributes, additionalAttrs ...slog.Attr) []slog.Attr {
	baseAttrs := buildInstanceAttributes(InstanceAttributes{
		coreInstanceAttr:        attr.coreInstanceAttr,
		coreNamespaceAttributes: attr.coreNamespaceAttributes,
	})
	memoryAttr := slog.String("state", attr.State)

	return internal.MergeAttributes(baseAttrs, append(additionalAttrs, memoryAttr)...)
}

// buildCloudEventAttributes constructs the base attributes for cloud event logging.
func buildCloudEventAttributes(attr CloudEventBusAttributes, additionalAttrs ...slog.Attr) []slog.Attr {
	baseAttrs := []slog.Attr{
		slog.String("event_id", attr.EventID),
		slog.String("source", attr.Source),
		slog.String("subject", attr.Subject),
		slog.String("event_type", attr.EventType),
		slog.String("namespace", attr.Namespace),
		slog.String("track", fmt.Sprintf("%v.%v", "namespace", attr.Namespace)),
	}

	return internal.MergeAttributes(baseAttrs, additionalAttrs...)
}

// buildGatewayAttributes constructs the base attributes for gateway logging.
func buildGatewayAttributes(attr GatewayAttributes, additionalAttrs ...slog.Attr) []slog.Attr {
	baseAttrs := []slog.Attr{
		slog.String("component", "gateway"),
		slog.String("gateway_plugin", attr.Plugin),
		slog.String("namespace", attr.Namespace),
		slog.String("route", attr.Route),
		slog.String("track", fmt.Sprintf("%v.%v", "route", attr.Route)),
	}

	return internal.MergeAttributes(baseAttrs, additionalAttrs...)
}

// buildGatewayAttributes constructs the base attributes for gateway logging.
func buildNamespaceAttributes(namespace string, additionalAttrs ...slog.Attr) []slog.Attr {
	baseAttrs := []slog.Attr{
		slog.String("namespace", namespace),
		slog.String("track", fmt.Sprintf("%v.%v", "namespace", namespace)),
	}

	return internal.MergeAttributes(baseAttrs, additionalAttrs...)
}
