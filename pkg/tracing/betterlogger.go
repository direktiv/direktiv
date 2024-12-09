package tracing

import (
	"context"

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
