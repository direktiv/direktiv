package events

import (
	"context"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

type (
	EventHandler func(ctx context.Context, events ...*Event)
	//	ChainedEventHandlers []EventHandler

	WorkflowStart    func(workflowID uuid.UUID, events ...*cloudevents.Event)
	WakeEventsWaiter func(instanceID uuid.UUID, step int, events []*cloudevents.Event)
)

type EventEngine struct {
	WorkflowStart       WorkflowStart
	WakeInstance        WakeEventsWaiter
	UpdateListeners     func(ctx context.Context, listener []*EventListener) []error
	GetListenersByTopic func(context.Context, string) ([]*EventListener, error)
}

func process(ctx context.Context,
	namespace uuid.UUID,
	cloudevents []cloudevents.Event, h []EventHandler,
) {
	events := make([]*Event, 0, len(cloudevents))
	for _, e := range cloudevents {
		events = append(events, &Event{
			Namespace:  namespace,
			ReceivedAt: time.Now(),
			Event:      &e,
		})
	}
	for _, eh := range h {
		eh(ctx, events...)
	}
}

func (ee EventEngine) GetEventHandlers(ctx context.Context,
	namespace uuid.UUID,
	cloudevents []cloudevents.Event,
) ([]EventHandler, error) {
	handlers := make([]EventHandler, 0)
	for _, cloudevent := range cloudevents {

		topic := namespace.String() + "-" + cloudevent.Type()
		listeners, err := ee.GetListenersByTopic(ctx, topic)
		if err != nil {
			return nil, err
		}
		for _, l := range listeners {
			handlers = append(handlers, ee.createEventHandler(l))
		}
	}
	return handlers, nil
}

func (ee EventEngine) createEventHandler(l *EventListener) EventHandler {
	if l.Deleted {
		return func(ctx context.Context, events ...*Event) {}
	}
	switch l.TriggerType {
	case StartAnd:
		return func(ctx context.Context, events ...*Event) {
			for _, event := range events {
				if l.Deleted {
					return
				}
				if event.Namespace != l.NamespaceID {
					continue
				}
				types := l.ListeningForEventTypes
				match := false
				for _, t := range types {
					if event.Event.Type() == t {
						match = true
					}
				}
				if !match {
					// TODO metrics
					continue
				}
				for _, r := range l.ReceivedEventsForAndTrigger {
					if r.Event.Type() == event.Event.Type() {
						continue
					}
				}
				l.ReceivedEventsForAndTrigger = append(l.ReceivedEventsForAndTrigger, event)
				typeMatch := make([]bool, len(l.ReceivedEventsForAndTrigger))
				ces := make([]*cloudevents.Event, 0, len(l.ReceivedEventsForAndTrigger)+1)
				for i := range l.ReceivedEventsForAndTrigger {
					e := l.ReceivedEventsForAndTrigger[i]
					ces = append(ces, e.Event)
					for _, t := range types {
						if e.Event.Type() == t {
							typeMatch[i] = true
						}
					}
				}
				hasAll := true
				for _, h := range typeMatch {
					hasAll = hasAll && h
				}
				if !hasAll {
					// TODO metrics
					continue
				}
				go ee.WorkflowStart(l.Trigger.WorkflowID, ces...)
			}
		}
	case WaitAnd:
		return func(ctx context.Context, events ...*Event) {
			for _, event := range events {
				if l.Deleted {
					return
				}
				if event.Namespace != l.NamespaceID {
					continue
				}
				types := l.ListeningForEventTypes
				match := false
				for _, t := range types {
					if event.Event.Type() == t {
						match = true
					}
				}
				if !match {
					// TODO metrics
					continue
				}
				for _, r := range l.ReceivedEventsForAndTrigger {
					if r.Event.Type() == event.Event.Type() {
						continue
					}
				}
				l.ReceivedEventsForAndTrigger = append(l.ReceivedEventsForAndTrigger, event)
				typeMatch := make([]bool, len(l.ReceivedEventsForAndTrigger))
				ces := make([]*cloudevents.Event, 0, len(l.ReceivedEventsForAndTrigger)+1)
				for i := range l.ReceivedEventsForAndTrigger {
					e := l.ReceivedEventsForAndTrigger[i]
					ces = append(ces, e.Event)
					for _, t := range types {
						if e.Event.Type() == t {
							typeMatch[i] = true
						}
					}
				}
				hasAll := true
				for _, h := range typeMatch {
					hasAll = hasAll && h
				}
				if !hasAll {
					// TODO metrics
					continue
				}
				l.Deleted = true
				go ee.WorkflowStart(l.Trigger.WorkflowID, ces...)
			}
		}
	case StartSimple:
		return func(ctx context.Context, events ...*Event) {
			for _, event := range events {
				if l.Deleted {
					return
				}
				if event.Namespace != l.NamespaceID {
					continue
				}
				types := l.ListeningForEventTypes
				match := false
				for _, t := range types {
					if event.Event.Type() == t {
						match = true
					}
				}
				if !EventPassedGatekeeper(l.Trigger.GlobGatekeepers, *event.Event) {
					continue
				}
				if !match {
					// TODO metrics
					continue
				}
				go ee.WorkflowStart(l.Trigger.WorkflowID, event.Event)
			}
		}
	case WaitSimple:
		return func(ctx context.Context, events ...*Event) {
			for _, event := range events {
				if l.Deleted {
					return
				}
				if event.Namespace != l.NamespaceID {
					continue
				}
				types := l.ListeningForEventTypes
				match := false
				for _, t := range types {
					if event.Event.Type() == t {
						match = true
					}
				}
				if !match {
					// TODO metrics
					continue
				}
				l.Deleted = true
				go ee.WakeInstance(l.Trigger.InstanceID, l.Trigger.Step, []*cloudevents.Event{event.Event})
			}
		}
	case StartOR:
		return func(ctx context.Context, events ...*Event) {
			for _, event := range events {
				if l.Deleted {
					return
				}
				if event.Namespace != l.NamespaceID {
					continue
				}
				types := l.ListeningForEventTypes
				match := false
				for _, t := range types {
					if event.Event.Type() == t {
						match = true
					}
				}
				if !match {
					// TODO metrics
					continue
				}
				go ee.WorkflowStart(l.Trigger.WorkflowID, event.Event)
			}
		}
	case WaitOR:
		return func(ctx context.Context, events ...*Event) {
			for _, event := range events {
				if l.Deleted {
					return
				}
				if event.Namespace != l.NamespaceID {
					continue
				}
				types := l.ListeningForEventTypes
				match := false
				for _, t := range types {
					if event.Event.Type() == t {
						match = true
					}
				}
				if !match {
					// TODO metrics
					continue
				}
				l.Deleted = true
				go ee.WakeInstance(l.Trigger.InstanceID, l.Trigger.Step, []*cloudevents.Event{event.Event})
			}
		}
	case StartXOR:
		return func(ctx context.Context, events ...*Event) {
			for _, event := range events {
				if l.Deleted {
					return
				}
				if event.Namespace != l.NamespaceID {
					continue
				}
				types := l.ListeningForEventTypes
				match := false
				for _, t := range types {
					if event.Event.Type() == t {
						match = true
					}
				}
				if !match {
					// TODO metrics
					continue
				}
				go ee.WorkflowStart(l.Trigger.WorkflowID, event.Event)
			}
		}
	case WaitXOR:
		return func(ctx context.Context, events ...*Event) {
			for _, event := range events {
				if l.Deleted {
					return
				}
				if event.Namespace != l.NamespaceID {
					continue
				}
				types := l.ListeningForEventTypes
				match := false
				for _, t := range types {
					if event.Event.Type() == t {
						match = true
					}
				}
				if !match {
					// TODO metrics
					continue
				}
				l.Deleted = true
				go ee.WakeInstance(l.Trigger.InstanceID, l.Trigger.Step, []*cloudevents.Event{event.Event})
			}
		}
	}
	return func(ctx context.Context, events ...*Event) {
		// TODO metrics
	}
}

func (ee EventEngine) UseMePostProcessingAEvent(ctx context.Context, listeners []*EventListener) error {
	return fmt.Errorf("unimplemented")
}

func EventPassedGatekeeper(globPatterns []string, event cloudevents.Event) bool {
	return true // todo
}
