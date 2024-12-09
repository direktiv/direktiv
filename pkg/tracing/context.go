package tracing

import (
	"context"
	"fmt"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/engine"
	"go.opentelemetry.io/otel/trace"
)

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

// LogContextKey defines the keys used for storing values in the context for structured logging.
type LogContextKey int

const (
	NamespaceKey LogContextKey = iota // Key for namespace
	InstanceKey                       // Key for instance ID
	InvokerKey                        // Key for the invoker
	CallpathKey                       // Key for the call path
	WorkflowKey                       // Key for the workflow path
	StateKey                          // Key for state information
	LogTrackKey                       // Key for log tracking information
	TraceKey                          // Key for trace information (OpenTelemetry)
	SpanKey                           // Key for span information (OpenTelemetry)
	LevelKey                          // Key for log level
	StatusKey                         // Key for log status
	MsgKey                            // Key for log message
	ActionKey                         // Key for action ID
	BranchKey                         // Key for branch information
)

// AddNamespace adds the namespace to the context for tracing and structured logging.
func AddNamespace(ctx context.Context, namespaceName string) context.Context {
	return context.WithValue(ctx, NamespaceKey, namespaceName)
}

// AddActionID adds an action identifier to the context, useful for tracing specific operations.
func AddActionID(ctx context.Context, action string) context.Context {
	return context.WithValue(ctx, ActionKey, action)
}

// AddBranch adds the branch identifier to the context for workflows with conditional branches.
func AddBranch(ctx context.Context, branch int) context.Context {
	return context.WithValue(ctx, BranchKey, branch)
}

// InstanceAttributes holds common metadata for an instance, which is helpful for logging.
//
//	type InstanceAttributes struct {
//		Namespace    string         // Namespace where the instance belongs
//		InstanceID   string         // Unique identifier for the instance
//		Invoker      string         // The invoker triggering the instance
//		Callpath     string         // The callpath for tracing the instance
//		WorkflowPath string         // Path of the workflow the instance belongs to
//		Status       core.LogStatus // Current status of the instance
//	}
//
// AddInstanceMemoryAttr adds instance-specific attributes and state information to the context for logging.
func AddInstanceMemoryAttr(ctx context.Context, attrs InstanceAttributes, state string) context.Context {
	ctx = AddInstanceAttr(ctx, InstanceAttributes{
		Namespace:    attrs.Namespace,
		InstanceID:   attrs.InstanceID,
		WorkflowPath: attrs.WorkflowPath,
		Callpath:     attrs.Callpath,
	})
	ctx = AddStateAttr(ctx, state)
	ctx = AddStatus(ctx, attrs.Status)

	return ctx
}

// AddInstanceAttr adds the core instance attributes to the context for logging.
func AddInstanceAttr(ctx context.Context, attrs InstanceAttributes) context.Context {
	ctx = AddNamespace(ctx, attrs.Namespace)
	ctx = context.WithValue(ctx, InstanceKey, attrs.InstanceID)
	ctx = context.WithValue(ctx, InvokerKey, attrs.Invoker)
	ctx = context.WithValue(ctx, CallpathKey, attrs.Callpath)
	ctx = context.WithValue(ctx, WorkflowKey, attrs.WorkflowPath)

	return ctx
}

// AddStateAttr adds the state information of the instance to the context.
func AddStateAttr(ctx context.Context, state string) context.Context {
	return context.WithValue(ctx, StateKey, state)
}

// AddLoseInstanceIDAttr adds a new instance ID to the context.
func AddLoseInstanceIDAttr(ctx context.Context, instanceID string) context.Context {
	return context.WithValue(ctx, InstanceKey, instanceID)
}

// AddStatus adds the status of the instance to the context, indicating the current state.
func AddStatus(ctx context.Context, status core.LogStatus) context.Context {
	return context.WithValue(ctx, StatusKey, status)
}

// WithTrack adds a unique tracking identifier to the context, useful for correlating logs.
func WithTrack(ctx context.Context, track string) context.Context {
	return context.WithValue(ctx, LogTrackKey, track)
}

// GetCoreAttributes retrieves the core attributes from the context for structured logging.
func GetCoreAttributes(ctx context.Context) map[string]interface{} {
	tags := make(map[string]interface{})

	// Retrieve core attributes using context keys
	if namespace, ok := ctx.Value(NamespaceKey).(string); ok {
		tags["namespace"] = namespace
	}
	if instance, ok := ctx.Value(InstanceKey).(string); ok {
		tags["instance"] = instance
	}
	if invoker, ok := ctx.Value(InvokerKey).(string); ok {
		tags["invoker"] = invoker
	}
	if callpath, ok := ctx.Value(CallpathKey).(string); ok {
		tags["callpath"] = callpath
	}
	if workflow, ok := ctx.Value(WorkflowKey).(string); ok {
		tags["workflow"] = workflow
	}
	if state, ok := ctx.Value(StateKey).(string); ok {
		tags["state"] = state
	}
	if status, ok := ctx.Value(StatusKey).(core.LogStatus); ok {
		tags["status"] = status
	}
	if action, ok := ctx.Value(ActionKey).(string); ok {
		tags["action"] = action
	}
	if branch, ok := ctx.Value(BranchKey).(string); ok {
		tags["branch"] = branch
	}
	if trackValue, ok := ctx.Value(LogTrackKey).(string); ok {
		tags[string(core.LogTrackKey)] = trackValue
	}

	return tags
}

func GetStatus(ctx context.Context) core.LogStatus {
	if status, ok := ctx.Value(StatusKey).(core.LogStatus); ok {
		return status
	}

	return core.LogUnknownStatus
}

// GetRawLogEntryWithStatus creates a log entry containing the status, level, and message, enriched with context attributes.
func GetRawLogEntryWithStatus(ctx context.Context, level core.LogLevel, msg string, status core.LogStatus) map[string]interface{} {
	tags := GetAttributes(ctx)
	tags["status"] = status
	tags["level"] = level.String()
	tags["msg"] = msg

	return tags
}

// BuildNamespaceTrack constructs a unique track identifier for the namespace.
func BuildNamespaceTrack(namespace string) string {
	return fmt.Sprintf("%v.%v", "namespace", namespace)
}

// BuildInstanceTrack constructs a unique track identifier for the instance.
func BuildInstanceTrack(instance *engine.Instance) string {
	callpath := instance.Instance.ID.String()
	if instance.DescentInfo == nil {
		return fmt.Sprintf("%v.%v", "instance", callpath)
	}
	for _, v := range instance.DescentInfo.Descent {
		callpath += "/" + v.ID.String()
	}

	return fmt.Sprintf("%v.%v", "instance", callpath)
}

// BuildInstanceTrackViaCallpath constructs a unique track identifier using the callpath.
func BuildInstanceTrackViaCallpath(callpath string) string {
	return fmt.Sprintf("%v.%v", "instance", callpath)
}

// CreateCallpath constructs a callpath string based on an instance's descent information.
func CreateCallpath(instance *engine.Instance) string {
	callpath := ""
	for _, v := range instance.DescentInfo.Descent {
		callpath += "/" + v.ID.String()
	}

	return callpath
}

// GetAttributes retrieves both core attributes and trace-specific information (trace and span IDs) from the context.
func GetAttributes(ctx context.Context) map[string]interface{} {
	tags := GetCoreAttributes(ctx)

	// Add OpenTelemetry trace and span information if available
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().TraceID().IsValid() {
		tags["trace"] = span.SpanContext().TraceID().String()
		tags["span"] = span.SpanContext().SpanID().String()
	}

	return tags
}
