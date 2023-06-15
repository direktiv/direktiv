package events

import (
	"context"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/uuid"
	"github.com/ryanuber/go-glob"
)

type EventProcessing interface {
	ProcessEvents(
		ctx context.Context,
		namespace uuid.UUID,
		cloudevents []cloudevents.Event,
	)
}

type (
	eventHandler     func(ctx context.Context, events ...*Event)
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
	topics := ee.getTopics(ctx, namespace, cloudevents)
	listeners, _ := ee.getListeners(ctx, topics...)
	// TODO log err
	h := ee.getEventHandlers(ctx, listeners)
	// TODO log errors
	ee.handleEvents(ctx, namespace, cloudevents, h)
	_ = ee.usePostProcessingEvents(ctx, listeners)
}

func (ee EventEngine) getListeners(ctx context.Context, topics ...string) ([]*EventListener, error) {
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

func (ee EventEngine) getEventHandlers(ctx context.Context,
	listeners []*EventListener,
) []eventHandler {
	_ = ctx // todo otel

	handlers := make([]eventHandler, 0, len(listeners))
	for _, l := range listeners {
		handlers = append(handlers, ee.createEventHandler(l))
	}

	return handlers
}

func (ee EventEngine) createEventHandler(l *EventListener) eventHandler {
	if l.Deleted {
		return func(ctx context.Context, events ...*Event) {}
	}
	switch l.TriggerType {
	case StartAnd:
		return ee.eventAndHandler(l, false)
	case WaitAnd:
		return ee.eventAndHandler(l, true)
	case StartSimple:
		return ee.eventSimpleHandler(l, false)
	case WaitSimple:
		return ee.eventSimpleHandler(l, true)
	case StartOR:
		return ee.eventSimpleHandler(l, false)
	case WaitOR:
		return ee.eventSimpleHandler(l, true)
	case StartXOR:
		return ee.eventSimpleHandler(l, false)
	case WaitXOR:
		return ee.eventSimpleHandler(l, true)
	}

	return func(ctx context.Context, events ...*Event) {
		// TODO metrics
	}
}

func (ee EventEngine) usePostProcessingEvents(ctx context.Context,
	listeners []*EventListener,
) error {
	errs := ee.UpdateListeners(ctx, listeners)
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil // TODO
}

func eventPassedGatekeeper(globPatterns []string, event cloudevents.Event) bool {
	// adding source for comparison
	m := event.Context.GetExtensions()

	// if there is none, we need to create one for source
	if m == nil {
		m = make(map[string]interface{})
	}

	m["source"] = event.Context.GetSource()

	for _, f := range globPatterns {
		if v, ok := m[f]; ok {
			vs, ok2 := v.(string)

			// if both are strings we can glob
			if ok && ok2 && !glob.Glob(f, vs) {
				return false
			}
		} else {
			return false
		}
	}

	return true // todo
}

func (EventEngine) handleEvents(ctx context.Context,
	namespace uuid.UUID,
	cloudevents []cloudevents.Event, h []eventHandler,
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

func (ee EventEngine) eventAndHandler(l *EventListener, waitType bool) eventHandler {
	return func(ctx context.Context, events ...*Event) {
		for _, event := range events {
			if l.Deleted {
				return
			}
			if event.Namespace != l.NamespaceID {
				continue
			}
			types := l.ListeningForEventTypes
			// TODO metrics
			if !typeMatches(types, event) {
				continue
			}
			if eventTypeAlreadyPresent(l, event) {
				continue
			}

			l.ReceivedEventsForAndTrigger = append(l.ReceivedEventsForAndTrigger, event)
			ces := make([]*cloudevents.Event, 0, len(l.ReceivedEventsForAndTrigger)+1)
			for i := range l.ReceivedEventsForAndTrigger {
				e := l.ReceivedEventsForAndTrigger[i]
				ces = append(ces, e.Event)
			}
			// TODO metrics
			if canTriggerAction(l, types) {
				ee.triggerAction(waitType, l, ces)
			}
		}
	}
}

func canTriggerAction(l *EventListener, types []string) bool {
	typeMatch := make([]bool, len(l.ReceivedEventsForAndTrigger))
	for i := range l.ReceivedEventsForAndTrigger {
		e := l.ReceivedEventsForAndTrigger[i]
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

	return hasAll
}

func eventTypeAlreadyPresent(l *EventListener, event *Event) bool {
	for _, r := range l.ReceivedEventsForAndTrigger {
		if r.Event.Type() == event.Event.Type() {
			return true
		}
	}

	return false
}

func (ee EventEngine) eventSimpleHandler(l *EventListener, waitType bool) eventHandler {
	return func(ctx context.Context, events ...*Event) {
		for _, event := range events {
			if l.Deleted {
				return
			}
			if event.Namespace != l.NamespaceID {
				continue
			}
			types := l.ListeningForEventTypes
			match := typeMatches(types, event)
			if !match {
				continue
			}
			if !eventPassedGatekeeper(l.Trigger.GlobGatekeepers, *event.Event) {
				continue
			}
			ee.triggerAction(waitType, l, []*cloudevents.Event{event.Event})
		}
	}
}

func (ee EventEngine) triggerAction(waitType bool, l *EventListener, ces []*event.Event) {
	if waitType {
		l.Deleted = true
		go ee.WakeInstance(l.Trigger.InstanceID, l.Trigger.Step, ces)

		return
	}
	go ee.WorkflowStart(l.Trigger.WorkflowID, ces...)
}

func typeMatches(types []string, event *Event) bool {
	match := false
	for _, t := range types {
		if event.Event.Type() == t {
			match = true
		}
	}

	return match
}
