package events

import (
	"context"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

type EventProcessing interface {
	ProcessEvents(
		ctx context.Context,
		namespace uuid.UUID,
		cloudevents []cloudevents.Event,
	)
}

type (
	EventHandler     func(ctx context.Context, events ...*Event)
	WorkflowStart    func(workflowID uuid.UUID, events ...*cloudevents.Event)
	WakeEventsWaiter func(instanceID uuid.UUID, step int, events []*cloudevents.Event)
)

type EventEngine struct {
	WorkflowStart       WorkflowStart
	WakeInstance        WakeEventsWaiter
	GetListenersByTopic func(context.Context, string) ([]*EventListener, error)
	UpdateListeners     func(ctx context.Context, listener []*EventListener) []error
}

func (ee EventEngine) ProcessEvents(
	ctx context.Context,
	namespace uuid.UUID,
	cloudevents []cloudevents.Event,
) {
	topics := ee.GetTopics(ctx, namespace, cloudevents)
	listeners, _ := ee.GetListeners(ctx, topics...)
	// TODO log err
	h, _ := ee.GetEventHandlers(ctx, listeners)
	// TODO log errors
	ee.HandleEvents(ctx, namespace, cloudevents, h)
	_ = ee.UsePostProcessingEvents(ctx, listeners)
}

func (ee EventEngine) GetListeners(ctx context.Context, topics ...string) ([]*EventListener, error) {
	res := make([]*EventListener, 0)

	for _, topic := range topics {
		listeners, err := ee.GetListenersByTopic(ctx, topic)
		if err != nil {
			return nil, err
		}
		res = append(res, listeners...)
	}

	return res, nil
}

func (EventEngine) GetTopics(ctx context.Context, namespace uuid.UUID, cloudevents []cloudevents.Event) []string {
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

func (ee EventEngine) GetEventHandlers(ctx context.Context,
	listeners []*EventListener,
) ([]EventHandler, error) {
	handlers := make([]EventHandler, 0, len(listeners))
	for _, l := range listeners {
		handlers = append(handlers, ee.CreateEventHandler(l))
	}

	return handlers, nil
}

func (ee EventEngine) CreateEventHandler(l *EventListener) EventHandler {
	if l.Deleted {
		return func(ctx context.Context, events ...*Event) {}
	}
	switch l.TriggerType {
	case StartAnd:
		return ee.EventAndHandler(l, false)
	case WaitAnd:
		return ee.EventAndHandler(l, true)
	case StartSimple:
		return ee.EventSimpleHandler(l, false)
	case WaitSimple:
		return ee.EventSimpleHandler(l, true)
	case StartOR:
		return ee.EventSimpleHandler(l, false)
	case WaitOR:
		return ee.EventSimpleHandler(l, true)
	case StartXOR:
		return ee.EventSimpleHandler(l, false)
	case WaitXOR:
		return ee.EventSimpleHandler(l, true)
	}
	return func(ctx context.Context, events ...*Event) {
		// TODO metrics
	}
}

func (ee EventEngine) UsePostProcessingEvents(ctx context.Context,
	listeners []*EventListener,
) error {
	errs := ee.UpdateListeners(ctx, listeners)
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	
	return fmt.Errorf("unimplemented")
}

func EventPassedGatekeeper(globPatterns []string, event cloudevents.Event) bool {
	return true // todo
}

func (EventEngine) HandleEvents(ctx context.Context,
	namespace uuid.UUID,
	cloudevents []cloudevents.Event, h []EventHandler,
) {
	events := make([]*Event, 0, len(cloudevents))
	for i := range cloudevents {
		e := &cloudevents[i]
		events = append(events, &Event{
			Namespace:  namespace,
			ReceivedAt: time.Now(),
			Event:      e,
		})
	}
	for _, eh := range h {
		eh(ctx, events...)
	}
}

func (ee EventEngine) EventAndHandler(l *EventListener, waitType bool) EventHandler {
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
			if waitType {
				l.Deleted = true
				go ee.WakeInstance(l.Trigger.InstanceID, l.Trigger.Step, ces)

				return
			}
			go ee.WorkflowStart(l.Trigger.WorkflowID, ces...)
		}
	}
}

func (ee EventEngine) EventSimpleHandler(l *EventListener, waitType bool) EventHandler {
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
			if waitType {
				l.Deleted = true
				go ee.WakeInstance(l.Trigger.InstanceID, l.Trigger.Step, []*cloudevents.Event{event.Event})
				return
			}
			go ee.WorkflowStart(l.Trigger.WorkflowID, event.Event)
		}
	}
}
