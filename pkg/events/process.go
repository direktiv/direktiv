package events

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/google/uuid"
	"github.com/ryanuber/go-glob"
)

// EventProcessing defines the contract for processing CloudEvents within your system.
// Implementations of this interface determine how events are handled, including event dispatching,
// listener filtering, and the triggering of associated actions.
type EventProcessing interface {
	ProcessEvents(
		ctx context.Context, // The context for the event processing operation TODO: document tracing requirements.
		namespace uuid.UUID, // The namespace to which the events belong.
		cloudevents []cloudevents.Event, // The CloudEvents to be processed.
		logErrors func(template string, args ...interface{}), // A function for logging errors encountered during event processing.
	)
}

type (
	// eventHandler represents a generic function type for handling Events.
	eventHandler func(ctx context.Context, events ...*datastore.Event)
	// WorkflowStart is a function type that signals the initiation of a new workflow,
	// providing the workflow's unique ID and related CloudEvents.
	WorkflowStart func(ctx context.Context, workflowID uuid.UUID, events ...*cloudevents.Event)
	// WorkflowStart is a function type that signals the initiation of a new workflow,
	// providing the namespace names and the workflow's unique path and related CloudEvents.
	WorkflowStartByPath func(namespace, workflow string, events ...*cloudevents.Event)
	// WakeEventsWaiter is a function type responsible for handling events that trigger
	// the continuation of a workflow instance at a specific step.
	WakeEventsWaiter func(ctx context.Context, instanceID uuid.UUID, events []*cloudevents.Event)
)

// EventEngine is the central coordinator for processing CloudEvents, dispatching them
// to event handlers, and triggering workflow actions.
type EventEngine struct {
	// WorkflowStart is a callback triggered when a new workflow begins.
	WorkflowStart WorkflowStart
	// WakeInstance is a callback triggered to resume a waiting workflow instance.
	WakeInstance WakeEventsWaiter
	// GetListenersByTopic retrieves EventListeners associated with a specified topic.
	GetListenersByTopic func(context.Context, string) ([]*datastore.EventListener, error)
	// UpdateListeners updates a set of EventListeners, returning any errors encountered.
	UpdateListeners func(ctx context.Context, listener []*datastore.EventListener) []error
}

// ProcessEvents dispatches CloudEvents to handlers in sequence. Event handlers are
// responsible for filtering events based on their specific criteria, collecting relevant
// events, and potentially triggering actions when all conditions are met.
func (ee EventEngine) ProcessEvents(
	ctx context.Context,
	namespace uuid.UUID,
	cloudevents []cloudevents.Event,
	handleErrors func(template string, args ...interface{}),
) {
	// 1. Extract Topics: Retrieves relevant event topics from the provided CloudEvents
	//    within the specified namespace.
	topics := ee.getTopics(ctx, namespace, cloudevents)
	// 2. Fetch Listeners: Retrieves EventListeners that are subscribed to the extracted topics.
	listeners, err := ee.getListeners(ctx, topics...)
	if err != nil {
		handleErrors("error getListeners %v", err)
	}
	// 3. Build Event Handlers: for each listener genererate a event handler.
	handelerChain := ee.getEventHandlers(ctx, listeners)
	// 4. Process Events: Dispatches all received CloudEvents to all event handlers in the chain.
	ee.handleEvents(ctx, namespace, cloudevents, handelerChain)
	// 5. Post-Processing: Executes post-processing logic for all listeners.
	//    Thereby storing any state-changes of the listeners, like marking it as deleted.
	err = ee.usePostProcessingEvents(ctx, listeners)
	if err != nil {
		handleErrors("error usePostProcessingEvents %v", err)
	}
	// TODO: Add metrics for tracking processing time and potential errors per listener or event type
}

// getListeners retrieves EventListeners that are subscribed to the specified topics.
func (ee EventEngine) getListeners(ctx context.Context, topics ...string) ([]*datastore.EventListener, error) {
	res := make([]*datastore.EventListener, 0)

	for _, topic := range topics {
		listeners, err := ee.GetListenersByTopic(ctx, topic)
		if err != nil {
			return nil, err
		}
		res = append(res, listeners...)
	}

	return res, nil
}

// getTopics extracts relevant event topics from the provided CloudEvents within the specified namespace.
func (EventEngine) getTopics(ctx context.Context, namespace uuid.UUID, cloudevents []cloudevents.Event) []string {
	_ = ctx // todo otel
	topics := make(map[string]string)
	for _, cloudevent := range cloudevents {
		topic := namespace.String() + "-" + cloudevent.Type()
		topics[topic] = ""
	}
	topicls := make([]string, 0, len(topics))
	for topic := range topics {
		topicls = append(topicls, topic)
	}

	return topicls
}

// getEventHandlers generates event handlers based on the provided EventListeners.
func (ee EventEngine) getEventHandlers(ctx context.Context,
	listeners []*datastore.EventListener,
) []eventHandler {
	_ = ctx // todo otel

	handlers := make([]eventHandler, 0, len(listeners))
	for _, l := range listeners {
		handlers = append(handlers, ee.createEventHandler(l))
	}

	return handlers
}

// createEventHandler creates an event handler function tailored to a specific EventListener.
func (ee EventEngine) createEventHandler(l *datastore.EventListener) eventHandler {
	if l.Deleted {
		return func(ctx context.Context, events ...*datastore.Event) {}
	}
	switch l.TriggerType {
	case datastore.StartAnd:
		return ee.multiConditionEventAndHandler(l, false)
	case datastore.WaitAnd:
		return ee.multiConditionEventAndHandler(l, true)
	case datastore.StartSimple:
		return ee.singleConditionEventHandler(l, false)
	case datastore.WaitSimple:
		return ee.singleConditionEventHandler(l, true)
	case datastore.StartOR:
		return ee.singleConditionEventHandler(l, false)
	case datastore.WaitOR:
		return ee.singleConditionEventHandler(l, true)
	}

	return func(ctx context.Context, events ...*datastore.Event) {
		// TODO: Add metrics for event filtering/handling logic (events processed, events dropped, etc.)
	}
}

// usePostProcessingEvents executes post-processing logic for EventListeners
// (such as storing state changes) and returns any errors encountered.
func (ee EventEngine) usePostProcessingEvents(ctx context.Context,
	listeners []*datastore.EventListener,
) error {
	errs := ee.UpdateListeners(ctx, listeners)
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

// EventPassedGatekeeper determines if an event satisfies a set of glob-based filtering patterns.
// These patterns can target extensions within the CloudEvent, such as its source.
//
// Parameters:
//   - globPatterns: A map where keys are attribute names and values are glob patterns.
//   - event: The CloudEvent being evaluated.
//
// Returns:
//   - true if the event matches all patterns or if there are no patterns to check.
//   - false if the event fails to match any relevant pattern.
//
// Example:
// ```go
// // Sample globPatterns with event type prefixes and context keys:
//
//	patterns := map[string]string{
//	    "alarm":   "*", // Match any alarm type
//	    "vmid":    "vm-12345",  // Specific vmid
//	    "clusterid": "cluster-abcd", // Specific clusterid
//	}
func EventPassedGatekeeper(context map[string]string, event cloudevents.Event) bool {
	// Early return if there are no gatekeeper patterns.
	if len(context) == 0 {
		return true
	}

	// Prepare extensions, including the event source.
	extensions := event.Context.GetExtensions()
	if extensions == nil {
		extensions = make(map[string]interface{})
	}
	extensions["source"] = event.Context.GetSource()

	// Check each relevant pattern against the event extensions.
	for patternKey, pattern := range context {
		if !extensionMatchesPattern(extensions, patternKey, pattern) {
			return false // Pattern mismatch, event failed gatekeeper.
		}
	}

	return true // Event passed all gatekeepers.
}

// extensionMatchesPattern checks if an event extension matches a given glob pattern.
func extensionMatchesPattern(extensions map[string]interface{}, extensionKey, pattern string) bool {
	extensionValue, found := extensions[extensionKey]
	if !found {
		return false
	}
	valueStr := fmt.Sprintf("%v", extensionValue)

	return glob.Glob(pattern, valueStr)
}

// handleEvents dispatches CloudEvents to the provided event handlers.
func (EventEngine) handleEvents(ctx context.Context,
	namespace uuid.UUID,
	cloudevents []cloudevents.Event, h []eventHandler,
) {
	events := make([]*datastore.Event, 0, len(cloudevents))

	for _, e := range cloudevents {
		eCopy := e.Clone()
		events = append(events, &datastore.Event{
			NamespaceID: namespace,
			ReceivedAt:  time.Now().UTC(),
			Event:       &eCopy,
		})
	}
	// panic(len(h))
	for _, eh := range h {
		eh(ctx, events...)
	}
}

// multiConditionEventAndHandler creates an event handler for "And" type triggers...
func (ee EventEngine) multiConditionEventAndHandler(l *datastore.EventListener, waitType bool) eventHandler {
	return func(ctx context.Context, events ...*datastore.Event) {
		for _, event := range events {
			if l.Deleted {
				return // Skip processing for deleted listeners.
			}
			if event.NamespaceID != l.NamespaceID {
				continue
			}
			types := l.ListeningForEventTypes
			// TODO Add metrics collection points

			removeExpired(l) // Removing any events already collected events that are expired from the EventListener.

			if !typeMatches(types, event) {
				continue // Check if event type matches listener's interest.
			}
			// Apply additional glob-based filtering.
			if !PassEventContextFilters(l, event) {
				continue
			}

			if eventTypeAlreadyPresent(l, event) {
				continue // Prevent duplicate event processing.
			}

			// Event collection for "and" trigger logic.
			l.ReceivedEventsForAndTrigger = append(l.ReceivedEventsForAndTrigger, event)
			ces := make([]*cloudevents.Event, 0, len(l.ReceivedEventsForAndTrigger)+1)
			for _, e := range l.ReceivedEventsForAndTrigger {
				ces = append(ces, e.Event)
			}

			// TODO Add metrics collection points
			if canTriggerAction(ces, types) {
				tr := triggerActionArgs{
					WorkflowID: l.TriggerWorkflow,
					InstanceID: l.TriggerInstance,
				}
				ee.triggerAction(ctx, waitType, tr, ces)

				// Reset event collection and mark for deletion if needed
				l.ReceivedEventsForAndTrigger = []*datastore.Event{}
				if waitType {
					l.Deleted = true
				}
			}
		}
	}
}

func PassEventContextFilters(l *datastore.EventListener, event *datastore.Event) bool {
	for _, filter := range l.EventContextFilters {
		if filter.Type == event.Event.Type() {
			return EventPassedGatekeeper(filter.Context, *event.Event)
		}
	}

	return true
}

// removeExpired removes expired events from an EventListener's collection.
func removeExpired(l *datastore.EventListener) {
	var validEvents []*datastore.Event
	for _, e := range l.ReceivedEventsForAndTrigger {
		if l.LifespanOfReceivedEvents == 0 || e.ReceivedAt.Add(time.Duration(l.LifespanOfReceivedEvents)*time.Millisecond).After(time.Now().UTC()) {
			validEvents = append(validEvents, e)
		}
	}
	l.ReceivedEventsForAndTrigger = validEvents
}

// canTriggerAction determines if sufficient events have been collected to trigger an action.
func canTriggerAction(l []*cloudevents.Event, types []string) bool {
	// Early return if not enough collected events for a potential trigger.
	if len(types) < len(l) {
		return false
	}
	// Build a map to track if all required event types have been received.
	typeMatch := make(map[string]bool)
	for _, v := range types {
		typeMatch[v] = false
	}
	// Update the tracker based on collected events.
	for _, e := range l {
		typeMatch[e.Type()] = true
	}
	// Update the tracker based on collected events.
	hasAll := true
	for _, h := range typeMatch {
		hasAll = hasAll && h // Check if every value in the map is true.
	}
	// TODO: Add metrics to track how often triggers are evaluated and how often they succeed.
	return hasAll
}

// eventTypeAlreadyPresent checks if an event type has already been received for an "And" trigger.
func eventTypeAlreadyPresent(l *datastore.EventListener, event *datastore.Event) bool {
	for _, r := range l.ReceivedEventsForAndTrigger {
		if r.Event.Type() == event.Event.Type() {
			return true
		}
	}

	return false
}

// singleConditionEventHandler creates an event handler for "Simple" type triggers.
func (ee EventEngine) singleConditionEventHandler(l *datastore.EventListener, waitType bool) eventHandler {
	return func(ctx context.Context, events ...*datastore.Event) {
		for _, event := range events {
			if l.Deleted {
				return // Skip processing for deleted listeners.
			}
			if event.NamespaceID != l.NamespaceID {
				continue // Filter for relevant namespace.
			}

			// Check if the event type matches the listener's interest.
			if !typeMatches(l.ListeningForEventTypes, event) {
				continue
			}
			// Apply additional glob-based filtering.
			if !PassEventContextFilters(l, event) {
				continue
			}

			// Construct trigger arguments.
			tr := triggerActionArgs{
				WorkflowID: l.TriggerWorkflow,
				InstanceID: l.TriggerInstance,
			}
			// Trigger the action (note: single event passed).
			ee.triggerAction(ctx, waitType, tr, []*cloudevents.Event{event.Event})
			if waitType {
				l.Deleted = true // Mark listener for deletion if applicable.
			}
		}
	}
}

// triggerActionArgs encapsulates arguments for triggering workflow actions.
type triggerActionArgs struct {
	WorkflowPath string // the path of the workflow.
	WorkflowID   string // the id of the workflow. to be removed.
	InstanceID   string // optional fill for instance-waiting trigger.
}

// triggerAction triggers a workflow (start or resume) based on the waitType flag.
func (ee EventEngine) triggerAction(ctx context.Context, waitType bool, t triggerActionArgs, ces []*event.Event) {
	if waitType {
		id, err := uuid.Parse(t.InstanceID)
		if err != nil {
			slog.Error("failed to parse a instance id in the event-engine while processing an event")
			return
		}
		go ee.WakeInstance(ctx, id, ces)

		return
	}
	id, err := uuid.Parse(t.WorkflowID)
	if err != nil {
		slog.Error("failed to parse a workflow id in the event-engine while processing an event")
		return
	}
	go ee.WorkflowStart(ctx, id, ces...)
}

// typeMatches checks if an event's type matches any of the provided types.
func typeMatches(types []string, event *datastore.Event) bool {
	match := false
	for _, t := range types {
		if event.Event.Type() == t {
			match = true
		}
	}

	return match
}
