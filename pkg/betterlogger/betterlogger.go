package betterlogger

import (
	"context"
)

type TrackAble interface {
	// ShowInInstanceView configures the logger to focus on instance-related logs.
	//
	// Returns:
	// - A UserLogger with tracking information for instance logs.
	ShowInInstanceView() UserLogger
	// ShowInNamespaceView configures the logger to focus on namespace-related logs.
	//
	// Returns:
	// - A UserLogger with tracking information for namespace logs.
	ShowInNamespaceView() UserLogger
	// ShowInGatewayView configures the logger to focus on gateway-related logs.
	//
	// Returns:
	// - A UserLogger with tracking information for gateway logs.
	ShowInGatewayView() UserLogger
	// ShowInMirrorView configures the logger to focus on mirror-related logs.
	//
	// Returns:
	// - A UserLogger with tracking information for mirror logs.
	ShowInMirrorView() UserLogger
	// ConsoleLogs configures the logger for console-specific logging.
	//
	// Returns:
	// - A UserLogger for console logs.
	ConsoleLogs() UserLogger
}

type UserLogger interface {
	DebugContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
}

type coreNamespaceAttributes struct {
	Namespace string
}

// GatewayAttributes holds metadata specific to a gateway component, helpful for logging.
type GatewayAttributes struct {
	coreNamespaceAttributes
	coreGatewayAttributes
}

// CloudEventBusAttributes holds metadata specific to the cloud event bus.
type CloudEventBusAttributes struct {
	coreNamespaceAttributes
	coreCloudEventBusAttributes
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

// SyncAttributes holds metadata specific to a Sync component, helpful for logging.
type SyncAttributes struct {
	coreSyncAttributes
	coreNamespaceAttributes
}

type coreInstanceAttr struct {
	InstanceID   string // Unique identifier for the instance
	WorkflowPath string // Path of the workflow the instance belongs to
	CallPath     string // Identifies the log-stream, legacy feature from the old engine
}

type coreGatewayAttributes struct {
	coreNamespaceAttributes
	Plugin string // Optional. Name for the gateway-plugin
	Route  string // Endpoint of the gateway
}
type coreCloudEventBusAttributes struct {
	EventID   string // Unique identifier for the event
	Source    string // Source of the event
	Subject   string // Subject of the event
	EventType string // Type of the event
}
type coreInstanceActionAttr struct {
	ActionID string // Unique identifier for the instance action
}

type coreInstanceMemoryAttributes struct {
	State string // Memory state of the instance
}

// SyncAttributes holds metadata specific to a Sync component, helpful for logging.
type coreSyncAttributes struct {
	SyncID string // Unique identifier for the Sync
}
