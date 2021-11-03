package flow

import (
	"context"
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	"github.com/google/uuid"
)

const (
	// BroadcastEventTypeCreate...
	BroadcastEventTypeCreate string = "create"
	// BroadcastEventTypeUpdate...
	BroadcastEventTypeUpdate string = "update"
	// BroadcastEventTypeDelete...
	BroadcastEventTypeDelete string = "delete"

	BroadcastEventTypeInstanceStarted string = "started"
	BroadcastEventTypeInstanceFailed  string = "failed"
	BroadcastEventTypeInstanceSuccess string = "success"
)

const (
	// BroadcastEventScopeWorkflow...
	BroadcastEventScopeWorkflow string = "workflow"
	// BroadcastEventScopeNamespace...
	BroadcastEventScopeNamespace string = "namespace"
	// BroadcastEventScopeInstance...
	BroadcastEventScopeInstance string = "instance"
)

const (
	// BroadcastEventPrefixWorkflow...
	BroadcastEventPrefixWorkflow string = "workflow"
	// BroadcastEventPrefixDirectory...
	BroadcastEventPrefixDirectory string = "directory"
	// BroadcastEventPrefixVariable...
	BroadcastEventPrefixVariable string = "variable"
	// BroadcastEventPrefixInstance...
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
	event.SetData("application/json", input)

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
	event.SetData("application/json", input)

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
	event.SetData("application/json", input)

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
	event.SetData("application/json", input)

	return flow.events.BroadcastCloudevent(ctx, ns, &event, 60)
}
