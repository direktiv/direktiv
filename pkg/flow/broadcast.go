package flow

import (
	"context"
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/direktiv/direktiv/pkg/flow/ent"
)

const (
	BroadcastEventTypeCreate string = "create"
	BroadcastEventTypeUpdate string = "update"
	BroadcastEventTypeDelete string = "delete"

	BroadcastEventTypeInstanceStarted string = "started"
	BroadcastEventTypeInstanceFailed  string = "failed"
	BroadcastEventTypeInstanceSuccess string = "success"
)

const (
	BroadcastEventScopeWorkflow  string = "workflow"
	BroadcastEventScopeNamespace string = "namespace"
	BroadcastEventScopeInstance  string = "instance"
)

const (
	BroadcastEventPrefixWorkflow  string = "workflow"
	BroadcastEventPrefixDirectory string = "directory"
	BroadcastEventPrefixVariable  string = "variable"
	BroadcastEventPrefixInstance  string = "instance"
)

type broadcastWorkflowInput struct {
	Name   string
	Path   string
	Parent string
	Live   bool
}

func (flow *flow) BroadcastWorkflow(eventType string, ctx context.Context, input broadcastWorkflowInput, ns *ent.Namespace) error {
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

func (flow *flow) BroadcastDirectory(eventType string, ctx context.Context, input broadcastDirectoryInput, ns *ent.Namespace) error {
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

func (flow *flow) BroadcastVariable(eventType string, eventScope string, ctx context.Context, input broadcastVariableInput, ns *ent.Namespace) error {
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
