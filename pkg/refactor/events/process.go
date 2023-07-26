package events

import (
	"context"
	"strings"
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
		logErrors func(template string, args ...interface{}),
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
	handleErrors func(template string, args ...interface{}),
) {
	topics := ee.getTopics(ctx, namespace, cloudevents)
	listeners, err := ee.getListeners(ctx, topics...)
	if err != nil {
		handleErrors("error getListeners %v", err)
	}
	h := ee.getEventHandlers(ctx, listeners)
	ee.handleEvents(ctx, namespace, cloudevents, h)
	err = ee.usePostProcessingEvents(ctx, listeners)
	if err != nil {
		handleErrors("error usePostProcessingEvents %v", err)
	}
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

func eventPassedGatekeeper(globPatterns map[string]string, event cloudevents.Event) bool {
	if len(globPatterns) == 0 {
		return true
	}
	// adding source for comparison
	m := event.Context.GetExtensions()

	// if there is none, we need to create one for source
	if m == nil {
		m = make(map[string]interface{})
	}

	m["source"] = event.Context.GetSource()
	match := false
	for k, f := range globPatterns {
		x := strings.TrimPrefix(k, event.Type()+"-")
		if v, ok := m[x]; ok {
			vs, ok2 := v.(string)
			if !ok2 {
				continue
			}
			match = match || glob.Glob(f, vs)
			// if both are strings we can glob
			// return !glob.Glob(f, event.Type()+"-"+vs)
		}
	}

	return match // todo
}

func (EventEngine) handleEvents(ctx context.Context,
	namespace uuid.UUID,
	cloudevents []cloudevents.Event, h []eventHandler,
) {
	events := make([]*Event, 0, len(cloudevents))

	for _, e := range cloudevents {
		eCopy := e.Clone()
		events = append(events, &Event{
			Namespace:  namespace,
			ReceivedAt: time.Now(),
			Event:      &eCopy,
		})
	}
	// panic(len(h))
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
			if !eventPassedGatekeeper(l.GlobGatekeepers, *event.Event) {
				continue
			}
			l.ReceivedEventsForAndTrigger = append(l.ReceivedEventsForAndTrigger, event)
			ces := make([]*cloudevents.Event, 0, len(l.ReceivedEventsForAndTrigger)+1)
			ces = removeExpired(l, ces)
			// TODO metrics
			if canTriggerAction(ces, types) {
				tr := triggerActionArgs{
					WorkflowID: l.TriggerWorkflow,
					InstanceID: l.TriggerInstance,
					Step:       l.TriggerInstanceStep,
				}
				ee.triggerAction(waitType, tr, ces)
				l.ReceivedEventsForAndTrigger = []*Event{}
				if waitType {
					l.Deleted = true
				}
			}
		}
	}
}

func removeExpired(l *EventListener, ces []*event.Event) []*event.Event {
	for i := range l.ReceivedEventsForAndTrigger {
		e := l.ReceivedEventsForAndTrigger[i]
		if l.LifespanOfReceivedEvents != 0 && e.ReceivedAt.Add(time.Duration(l.LifespanOfReceivedEvents)*time.Millisecond).Before(time.Now()) {
			continue
		}
		ces = append(ces, e.Event)
	}

	return ces
}

func canTriggerAction(l []*cloudevents.Event, types []string) bool {
	if len(types) < len(l) {
		return false
	}
	typeMatch := make(map[string]bool)
	for _, v := range types {
		typeMatch[v] = false
	}
	for _, e := range l {
		typeMatch[e.Type()] = true
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
			tr := triggerActionArgs{
				WorkflowID: l.TriggerWorkflow,
				InstanceID: l.TriggerInstance,
				Step:       l.TriggerInstanceStep,
			}
			if !eventPassedGatekeeper(l.GlobGatekeepers, *event.Event) {
				continue
			}
			ee.triggerAction(waitType, tr, []*cloudevents.Event{event.Event})
			if waitType {
				l.Deleted = true
			}
		}
	}
}

type triggerActionArgs struct {
	WorkflowID uuid.UUID // the id of the workflow.
	InstanceID uuid.UUID // optional fill for instance-waiting trigger.
	Step       int       // optional fill for instance-waiting trigger.
}

func (ee EventEngine) triggerAction(waitType bool, t triggerActionArgs, ces []*event.Event) {
	if waitType {
		go ee.WakeInstance(t.InstanceID, t.Step, ces)

		return
	}
	go ee.WorkflowStart(t.WorkflowID, ces...)
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
