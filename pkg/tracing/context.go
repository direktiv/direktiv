package tracing

import (
	"context"
	"fmt"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/engine"
	"go.opentelemetry.io/otel/trace"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (level LogLevel) String() string {
	switch level {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "DEBUG"
	}
}

type LogContextKey int

const (
	NamespaceKey LogContextKey = iota
	InstanceKey
	InvokerKey
	CallpathKey
	WorkflowKey
	StateKey
	LogTrackKey
	TraceKey
	SpanKey
	LevelKey
	StatusKey
	MsgKey
	ActionKey
	BranchKey
)

func AddNamespace(ctx context.Context, namespaceName string) context.Context {
	return context.WithValue(ctx, NamespaceKey, namespaceName)
}

func AddActionID(ctx context.Context, action string) context.Context {
	return context.WithValue(ctx, ActionKey, action)
}

func AddBranch(ctx context.Context, branch int) context.Context {
	return context.WithValue(ctx, BranchKey, branch)
}

// InstanceAttributes holds the common attributes for an instance.
type InstanceAttributes struct {
	Namespace    string
	InstanceID   string
	Invoker      string
	Callpath     string
	WorkflowPath string
	Status       core.LogStatus
}

func AddInstanceMemoryAttr(ctx context.Context, attrs InstanceAttributes, state string) context.Context {
	ctx = AddInstanceAttr(ctx, attrs)
	ctx = AddStateAttr(ctx, state)
	ctx = AddStatus(ctx, attrs.Status)

	return ctx
}

func AddInstanceAttr(ctx context.Context, attrs InstanceAttributes) context.Context {
	ctx = AddNamespace(ctx, attrs.Namespace)
	ctx = context.WithValue(ctx, InstanceKey, attrs.InstanceID)
	ctx = context.WithValue(ctx, InvokerKey, attrs.Invoker)
	ctx = context.WithValue(ctx, CallpathKey, attrs.Callpath)
	ctx = context.WithValue(ctx, WorkflowKey, attrs.WorkflowPath)

	return ctx
}

func AddStateAttr(ctx context.Context, state string) context.Context {
	return context.WithValue(ctx, StateKey, state)
}

func AddLoseInstanceIDAttr(ctx context.Context, instanceID string) context.Context {
	return context.WithValue(ctx, InstanceKey, instanceID)
}

func AddStatus(ctx context.Context, status core.LogStatus) context.Context {
	return context.WithValue(ctx, StatusKey, status)
}

func WithTrack(ctx context.Context, track string) context.Context {
	return context.WithValue(ctx, LogTrackKey, track)
}

func GetCoreAttributes(ctx context.Context) map[string]interface{} {
	tags := make(map[string]interface{})

	// Retrieve core attributes using enum keys
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

func GetAttributes(ctx context.Context) map[string]interface{} {
	tags := GetCoreAttributes(ctx)

	// Add trace information if available
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().TraceID().IsValid() {
		tags["trace"] = span.SpanContext().TraceID().String()
		tags["span"] = span.SpanContext().SpanID().String()
	}

	return tags
}

func GetRawLogEntryWithStatus(ctx context.Context, level LogLevel, msg string, status core.LogStatus) map[string]interface{} {
	tags := GetAttributes(ctx)
	tags["status"] = status
	tags["level"] = level.String()
	tags["msg"] = msg

	return tags
}

func BuildNamespaceTrack(namespace string) string {
	return fmt.Sprintf("%v.%v", "namespace", namespace)
}

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

func BuildInstanceTrackViaCallpath(callpath string) string {
	return fmt.Sprintf("%v.%v", "instance", callpath)
}

// CreateCallpath builds the callpath string from the instance's descent information.
func CreateCallpath(instance *engine.Instance) string {
	callpath := ""
	for _, v := range instance.DescentInfo.Descent {
		callpath += "/" + v.ID.String()
	}

	return callpath
}
