package flow

import (
	"context"
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	"github.com/google/uuid"
)

const (
	// BroadcastEventTypeCreate is an event type for listening to 'create'.
	BroadcastEventTypeCreate string = "create"
	// BroadcastEventTypeUpdate is an event type for listening to 'update'.
	BroadcastEventTypeUpdate string = "update"
	// BroadcastEventTypeDelete is an event type for listenting to 'delete'.
	BroadcastEventTypeDelete string = "delete"

	BroadcastEventTypeInstanceStarted string = "started"
	BroadcastEventTypeInstanceFailed  string = "failed"
	BroadcastEventTypeInstanceSuccess string = "success"
)

const (
	// BroadcastEventScopeWorkflow is the scope in which you want to listen for events.
	BroadcastEventScopeWorkflow string = "workflow"
	// BroadcastEventScopeNamespace is the scope in which you want to listen for events.
	BroadcastEventScopeNamespace string = "namespace"
	// BroadcastEventScopeInstance is the scope in which you want to listen for events.
	BroadcastEventScopeInstance string = "instance"
)

const (
	// BroadcastEventPrefixWorkflow is the event prefix that is being broadcasted.
	BroadcastEventPrefixWorkflow string = "workflow"
	// BroadcastEventPrefixDirectory is the event prefix that is being broadcasted.
	BroadcastEventPrefixDirectory string = "directory"
	// BroadcastEventPrefixVariable is the event prefix that is being broadcasted.
	BroadcastEventPrefixVariable string = "variable"
	// BroadcastEventPrefixInstance is the event prefix that is being broadcasted.
	BroadcastEventPrefixInstance string = "instance"
)

type broadcastWorkflowInput struct {
	Name   string
	Path   string
	Parent string
	Live   bool
}

func (flow *flow) BroadcastWorkflow(ctx context.Context, eventType string, input broadcastWorkflowInput, ns *ent.Namespace) error {
	// BROADCAST EVENT

	target := fmt.Sprintf("%s.%s", BroadcastEventPrefixWorkflow, eventType)
	cfg, err := loadNSConfig([]byte(ns.Config))
	if err != nil {
		return fmt.Errorf("failed to load namespace config: %w", err)
	}

	// skip if broad target is not enabled
	if !cfg.broadcastEnabled(target) {
		return nil
	}

	event := cloudevents.NewEvent(cloudevents.VersionV03)
	uid := uuid.New()
	event.SetID(uid.String())
	event.SetType(target)
	event.SetSource("direktiv")
	err = event.SetData("application/json", input)
	if err != nil {
		return fmt.Errorf("failed to create CloudEvent: %w", err)
	}

	return flow.events.BroadcastCloudevent(ctx, ns, &event, 60)
}

type broadcastDirectoryInput struct {
	Path   string
	Parent string
}

func (flow *flow) BroadcastDirectory(ctx context.Context, eventType string, input broadcastDirectoryInput, ns *ent.Namespace) error {
	// BROADCAST EVENT
	target := fmt.Sprintf("%s.%s", BroadcastEventPrefixDirectory, eventType)
	cfg, err := loadNSConfig([]byte(ns.Config))
	if err != nil {
		return fmt.Errorf("failed to load namespace config: %w", err)
	}

	// skip if broad target is not enabled
	if !cfg.broadcastEnabled(target) {
		return nil
	}

	event := cloudevents.NewEvent(cloudevents.VersionV03)
	uid := uuid.New()
	event.SetID(uid.String())
	event.SetType(target)
	event.SetSource("direktiv")
	err = event.SetData("application/json", input)
	if err != nil {
		return fmt.Errorf("failed to create CloudEvent: %w", err)
	}

	return flow.events.BroadcastCloudevent(ctx, ns, &event, 60)
}

type broadcastVariableInput struct {
	WorkflowPath string
	InstanceID   string
	Key          string
	TotalSize    int64
	Scope        string
}

func (flow *flow) BroadcastVariable(ctx context.Context, eventType string, eventScope string, input broadcastVariableInput, ns *ent.Namespace) error {
	// BROADCAST EVENT
	target := fmt.Sprintf("%s.%s.%s", eventScope, BroadcastEventPrefixVariable, eventType)
	cfg, err := loadNSConfig([]byte(ns.Config))
	if err != nil {
		return fmt.Errorf("failed to load namespace config: %w", err)
	}

	// skip if broad target is not enabled
	if !cfg.broadcastEnabled(target) {
		return nil
	}

	event := cloudevents.NewEvent(cloudevents.VersionV03)
	uid := uuid.New()
	event.SetID(uid.String())
	event.SetType(target)
	event.SetSource("direktiv")
	err = event.SetData("application/json", input)
	if err != nil {
		return fmt.Errorf("failed to create CloudEvent: %w", err)
	}

	return flow.events.BroadcastCloudevent(ctx, ns, &event, 60)
}

type broadcastInstanceInput struct {
	WorkflowPath string
	InstanceID   string
	Caller       string
}

func (flow *flow) BroadcastInstance(eventType string, ctx context.Context, input broadcastInstanceInput, ns *ent.Namespace) error {
	// BROADCAST EVENT
	target := fmt.Sprintf("%s.%s", BroadcastEventPrefixInstance, eventType)
	cfg, err := loadNSConfig([]byte(ns.Config))
	if err != nil {
		return fmt.Errorf("failed to load namespace config: %w", err)
	}

	// skip if broad target is not enabled
	if !cfg.broadcastEnabled(target) {
		return nil
	}

	event := cloudevents.NewEvent(cloudevents.VersionV03)
	uid := uuid.New()
	event.SetID(uid.String())
	event.SetType(target)
	event.SetSource("direktiv")
	err = event.SetData("application/json", input)
	if err != nil {
		return fmt.Errorf("failed to create CloudEvent: %w", err)
	}

	return flow.events.BroadcastCloudevent(ctx, ns, &event, 60)
}
