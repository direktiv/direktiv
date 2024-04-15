package events_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
	"github.com/direktiv/direktiv/pkg/refactor/events"
	"github.com/google/uuid"
)

type triggerMock struct {
	events []*cloudevents.Event
	wf     uuid.UUID
	inst   uuid.UUID
	step   int
}

func Test_Add_Get_Complex_Context(t *testing.T) {
	ns := uuid.New()
	wfID1 := uuid.New()
	wfID2 := uuid.New()

	listeners := make([]*events.EventListener, 0)
	listeners = append(listeners,
		&events.EventListener{
			ID:                     uuid.New(),
			CreatedAt:              time.Now().UTC(),
			UpdatedAt:              time.Now().UTC(),
			Deleted:                false,
			NamespaceID:            ns,
			ListeningForEventTypes: []string{"test-topic"},
			TriggerType:            events.StartSimple,
			TriggerWorkflow:        wfID1.String(),
			GlobGatekeepers: map[string]string{
				"test-topic-id": "some id",
			},
		},
		&events.EventListener{
			ID:                     uuid.New(),
			CreatedAt:              time.Now().UTC(),
			UpdatedAt:              time.Now().UTC(),
			Deleted:                false,
			NamespaceID:            ns,
			ListeningForEventTypes: []string{"test-topic"},
			TriggerType:            events.StartSimple,
			TriggerWorkflow:        wfID2.String(),
			GlobGatekeepers: map[string]string{
				"test-topic-id": "some other id",
			},
		},
	)
	resultsForEngine := make(chan triggerMock, 1)
	var engine events.EventProcessing = events.EventEngine{
		WorkflowStart: func(workflowID uuid.UUID, events ...*cloudevents.Event) {
			resultsForEngine <- triggerMock{events: events, wf: workflowID}
		},
		WakeInstance: func(instanceID uuid.UUID, step int, events []*cloudevents.Event) {
			resultsForEngine <- triggerMock{events: events, inst: instanceID, step: step}
		},
		GetListenersByTopic: func(ctx context.Context, s string) ([]*events.EventListener, error) {
			return listeners, nil
		},
		UpdateListeners: func(ctx context.Context, listener []*events.EventListener) []error {
			for i, el := range listener {
				if el.Deleted {
					listener = append(listener[:i], listener[i+1:]...)
				}
			}

			return []error{}
		},
	}
	ev1 := newEventWithMeta("test-sub1", "test-topic", uuid.New(), map[string]any{"id": "some id"})
	ev2 := newEventWithMeta("test-sub1", "test-topic", uuid.New(), map[string]any{"id": "some other id"})
	engine.ProcessEvents(context.Background(), ns, []event.Event{*ev1}, func(template string, args ...interface{}) {})
	tr, err := waitForTrigger(t, resultsForEngine)
	if err != nil {
		t.Error("got no results")
		return
	}
	if tr.wf != wfID1 {
		t.Error("workflow should be triggered")
	}
	engine.ProcessEvents(context.Background(), ns, []event.Event{*ev2}, func(template string, args ...interface{}) {})
	tr, err = waitForTrigger(t, resultsForEngine)
	if err != nil {
		t.Error("got no results")
		return
	}
	if tr.wf != wfID2 {
		t.Error("workflow should be triggered")
	}
	ev3 := newEventWithMeta("test-sub1", "test-topic", uuid.New(), map[string]any{"id": "some id 2"})
	engine.ProcessEvents(context.Background(), ns, []event.Event{*ev3}, func(template string, args ...interface{}) {})
	tr, err = waitForTrigger(t, resultsForEngine)
	if err == nil {
		t.Error("Expected no workflow trigger due to mismatched event metadata")
		return
	}
}

func Test_Add_Get_And(t *testing.T) {
	ns := uuid.New()
	wfID := uuid.New()

	listeners := make([]*events.EventListener, 0)
	listeners = append(listeners,
		&events.EventListener{
			ID:                     uuid.New(),
			CreatedAt:              time.Now().UTC(),
			UpdatedAt:              time.Now().UTC(),
			Deleted:                false,
			NamespaceID:            ns,
			ListeningForEventTypes: []string{"test-topic", "test-topic2"},
			TriggerType:            events.StartAnd,
			TriggerWorkflow:        wfID.String(),
			GlobGatekeepers: map[string]string{
				"test-topic2-id": "some id",
			},
		},
	)
	resultsForEngine := make(chan triggerMock, 1)
	var engine events.EventProcessing = events.EventEngine{
		WorkflowStart: func(workflowID uuid.UUID, events ...*cloudevents.Event) {
			resultsForEngine <- triggerMock{events: events, wf: workflowID}
		},
		WakeInstance: func(instanceID uuid.UUID, step int, events []*cloudevents.Event) {
			resultsForEngine <- triggerMock{events: events, inst: instanceID, step: step}
		},
		GetListenersByTopic: func(ctx context.Context, s string) ([]*events.EventListener, error) {
			return listeners, nil
		},
		UpdateListeners: func(ctx context.Context, listener []*events.EventListener) []error {
			for i, el := range listener {
				if el.Deleted {
					listener = append(listener[:i], listener[i+1:]...)
				}
			}

			return []error{}
		},
	}
	eID := uuid.New()
	ev := newEvent("test-sub1", "test-topic", eID)
	ev2 := newEvent("test-sub1", "test-topic2", uuid.New())
	engine.ProcessEvents(context.Background(), ns, []event.Event{*ev}, func(template string, args ...interface{}) {})
	engine.ProcessEvents(context.Background(), ns, []event.Event{*ev2}, func(template string, args ...interface{}) {})
	tr, err := waitForTrigger(t, resultsForEngine)
	if err == nil {
		t.Error("Expected no workflow trigger due to mismatched event metadata")
		return
	}
	eID = uuid.New()
	ev = newEventWithMeta("test-sub1", "test-topic", eID, map[string]any{"id": "some id"})
	engine.ProcessEvents(context.Background(), ns, []event.Event{*ev}, func(template string, args ...interface{}) {})
	tr, err = waitForTrigger(t, resultsForEngine)
	if err == nil {
		t.Error("Expected no workflow trigger due to mismatched event metadata")
		return
	}
	ev = newEventWithMeta("test-sub1", "test-topic2", eID, map[string]any{"id": "some id"})
	engine.ProcessEvents(context.Background(), ns, []event.Event{*ev}, func(template string, args ...interface{}) {})
	tr, err = waitForTrigger(t, resultsForEngine)
	if err != nil {
		t.Error("got no results")
		return
	}
	if tr.wf != wfID {
		t.Error("workflow should be triggered")
	}
}

func Test_Add_Get_GatekeeperSimple(t *testing.T) {
	ns := uuid.New()
	wfID := uuid.New()

	listeners := make([]*events.EventListener, 0)
	listeners = append(listeners,
		&events.EventListener{
			ID:                     uuid.New(),
			CreatedAt:              time.Now().UTC(),
			UpdatedAt:              time.Now().UTC(),
			Deleted:                false,
			NamespaceID:            ns,
			ListeningForEventTypes: []string{"test-topic"},
			TriggerType:            events.StartSimple,
			TriggerWorkflow:        wfID.String(),
			GlobGatekeepers: map[string]string{
				"test-topic-id": "some id",
			},
		},
	)
	resultsForEngine := make(chan triggerMock, 1)
	var engine events.EventProcessing = events.EventEngine{
		WorkflowStart: func(workflowID uuid.UUID, events ...*cloudevents.Event) {
			resultsForEngine <- triggerMock{events: events, wf: workflowID}
		},
		WakeInstance: func(instanceID uuid.UUID, step int, events []*cloudevents.Event) {
			resultsForEngine <- triggerMock{events: events, inst: instanceID, step: step}
		},
		GetListenersByTopic: func(ctx context.Context, s string) ([]*events.EventListener, error) {
			return listeners, nil
		},
		UpdateListeners: func(ctx context.Context, listener []*events.EventListener) []error {
			for i, el := range listener {
				if el.Deleted {
					listener = append(listener[:i], listener[i+1:]...)
				}
			}

			return []error{}
		},
	}
	eID := uuid.New()
	ev := newEvent("test-sub1", "test-topic", eID)
	engine.ProcessEvents(context.Background(), ns, []event.Event{*ev}, func(template string, args ...interface{}) {})
	tr, err := waitForTrigger(t, resultsForEngine)
	if err == nil {
		t.Error("Expected no workflow trigger due to mismatched event metadata")
		return
	}
	eID = uuid.New()
	ev = newEventWithMeta("test-sub1", "test-topic", eID, map[string]any{"id": "some id"})
	engine.ProcessEvents(context.Background(), ns, []event.Event{*ev}, func(template string, args ...interface{}) {})
	tr, err = waitForTrigger(t, resultsForEngine)
	if err != nil {
		t.Error("got no results")
		return
	}
	if tr.wf != wfID {
		t.Error("workflow should be triggered")
	}
}

func Test_Add_Get(t *testing.T) {
	ns := uuid.New()
	wfID := uuid.New()
	instID := uuid.New()

	waitListener := &events.EventListener{
		ID:                     uuid.New(),
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
		Deleted:                false,
		NamespaceID:            ns,
		ListeningForEventTypes: []string{"test-wait-topic"},
		TriggerType:            events.WaitSimple,
		TriggerInstance:        instID.String(),
	}
	listeners := make([]*events.EventListener, 0)
	listeners = append(listeners,
		&events.EventListener{
			ID:                     uuid.New(),
			CreatedAt:              time.Now().UTC(),
			UpdatedAt:              time.Now().UTC(),
			Deleted:                false,
			NamespaceID:            ns,
			ListeningForEventTypes: []string{"test-topic"},
			TriggerType:            events.StartSimple,
			TriggerWorkflow:        wfID.String(),
		},
		waitListener,
		&events.EventListener{
			ID:                     uuid.New(),
			CreatedAt:              time.Now().UTC(),
			UpdatedAt:              time.Now().UTC(),
			Deleted:                false,
			NamespaceID:            ns,
			ListeningForEventTypes: []string{"event-and-topic-a", "event-and-topic-b"},
			TriggerType:            events.StartAnd,
			TriggerWorkflow:        wfID.String(),
		},
	)
	resultsForEngine := make(chan triggerMock, 1)
	var engine events.EventProcessing = events.EventEngine{
		WorkflowStart: func(workflowID uuid.UUID, events ...*cloudevents.Event) {
			resultsForEngine <- triggerMock{events: events, wf: workflowID}
		},
		WakeInstance: func(instanceID uuid.UUID, step int, events []*cloudevents.Event) {
			resultsForEngine <- triggerMock{events: events, inst: instanceID, step: step}
		},
		GetListenersByTopic: func(ctx context.Context, s string) ([]*events.EventListener, error) {
			return listeners, nil
		},
		UpdateListeners: func(ctx context.Context, listener []*events.EventListener) []error {
			for i, el := range listener {
				if el.Deleted {
					listener = append(listener[:i], listener[i+1:]...)
				}
			}

			return []error{}
		},
	}
	// test simple case
	eID := uuid.New()
	ev := newEvent("test-sub1", "test-topic", eID)
	engine.ProcessEvents(context.Background(), ns, []event.Event{*ev}, func(template string, args ...interface{}) {})
	tr, err := waitForTrigger(t, resultsForEngine)
	if err != nil {
		t.Error("got no results")
		return
	}
	if tr.wf != wfID {
		t.Error("workflow should be triggered")
	}
	// test simple wait case
	eID = uuid.New()
	ev = newEvent("test-sub", "test-wait-topic", eID)
	engine.ProcessEvents(context.Background(), ns, []event.Event{*ev}, func(template string, args ...interface{}) {})
	tr, err = waitForTrigger(t, resultsForEngine)
	if err != nil {
		t.Error("got no results")
	}
	if tr.inst != instID {
		t.Error("workflow should be triggered")
	}
	if !waitListener.Deleted {
		t.Error("wait listeners should be marked as deleted after being triggered")
	}

	// test for event type that has no listener registered
	id := uuid.New()
	ev = newEvent("test-sub", "invalid-topic", id)
	engine.ProcessEvents(context.Background(), ns, []event.Event{*ev}, func(template string, args ...interface{}) {})
	_, err = waitForTrigger(t, resultsForEngine)
	if err == nil {
		t.Error("got unexpected results")

		return
	}
	// test andTrigger logic
	idA := uuid.New()
	evA := newEvent("test-sub", "event-and-topic-a", idA)
	idB := uuid.New()
	evB := newEvent("test-sub", "event-and-topic-b", idB)
	engine.ProcessEvents(context.Background(), ns, []event.Event{*evA}, func(template string, args ...interface{}) {})
	engine.ProcessEvents(context.Background(), ns, []event.Event{*evB}, func(template string, args ...interface{}) {})
	trAnd, err := waitForTrigger(t, resultsForEngine)
	if err != nil {
		t.Error("got no results")

		return
	}
	if trAnd.wf != wfID {
		t.Error("workflow should be triggered")
	}
	matchA := false
	matchB := false
	if len(trAnd.events) != 2 {
		t.Error("both events should be passed via trigger")
	}
	for _, e := range trAnd.events {
		if e.ID() == idA.String() {
			matchA = true
		}
		if e.ID() == idB.String() {
			matchB = true
		}
	}
	if !matchA || !matchB {
		t.Errorf("event where not properly passed to triggered action")
	}
	// test if andTrigger resets its state after being triggered
	idB = uuid.New()
	evB = newEvent("test-sub", "event-and-topic-b", idB)
	engine.ProcessEvents(context.Background(), ns, []event.Event{*evB}, func(template string, args ...interface{}) {})
	_, err = waitForTrigger(t, resultsForEngine)
	if err == nil {
		t.Error("expected no results")

		return
	}
	// the order of incoming type should not matter for andTrigger
	idA = uuid.New()
	evA = newEvent("test-sub", "event-and-topic-a", idA)
	engine.ProcessEvents(context.Background(), ns, []event.Event{*evA}, func(template string, args ...interface{}) {})
	_, err = waitForTrigger(t, resultsForEngine)
	if err != nil {
		t.Error("expected results")

		return
	}
}

func Test_Add_GatekkeeperComplex(t *testing.T) {
	ns := uuid.New()
	wfID := uuid.New()

	listeners := make([]*events.EventListener, 0)
	listeners = append(listeners,
		&events.EventListener{
			ID:                     uuid.New(),
			CreatedAt:              time.Now().UTC(),
			UpdatedAt:              time.Now().UTC(),
			Deleted:                false,
			NamespaceID:            ns,
			ListeningForEventTypes: []string{"test-topic", "other-topic"},
			TriggerType:            events.StartAnd,
			TriggerWorkflow:        wfID.String(),
			GlobGatekeepers: map[string]string{
				"test-topic-id":  "some id",
				"other-topic-id": "some other id",
			},
		},
	)
	resultsForEngine := make(chan triggerMock, 1)
	var engine events.EventProcessing = events.EventEngine{
		WorkflowStart: func(workflowID uuid.UUID, events ...*cloudevents.Event) {
			resultsForEngine <- triggerMock{events: events, wf: workflowID}
		},
		WakeInstance: func(instanceID uuid.UUID, step int, events []*cloudevents.Event) {
			resultsForEngine <- triggerMock{events: events, inst: instanceID, step: step}
		},
		GetListenersByTopic: func(ctx context.Context, s string) ([]*events.EventListener, error) {
			return listeners, nil
		},
		UpdateListeners: func(ctx context.Context, listener []*events.EventListener) []error {
			for i, el := range listener {
				if el.Deleted {
					listener = append(listener[:i], listener[i+1:]...)
				}
			}

			return []error{}
		},
	}
	// test simple case
	eID := uuid.New()
	ev := newEventWithMeta("test-sub1", "test-topic", eID, map[string]any{
		"id": "some id",
	})
	engine.ProcessEvents(context.Background(), ns, []event.Event{*ev}, func(template string, args ...interface{}) {})
	tr, err := waitForTrigger(t, resultsForEngine)
	if err == nil {
		t.Fatal("should not be triggered")
		return
	}
	// test simple wait case
	eID = uuid.New()
	ev = newEventWithMeta("test-sub1", "other-topic", eID, map[string]any{
		"id": "some other id",
	})
	engine.ProcessEvents(context.Background(), ns, []event.Event{*ev}, func(template string, args ...interface{}) {})
	tr, err = waitForTrigger(t, resultsForEngine)
	if err != nil {
		t.Fatal("got no results")
	}
	if tr.wf != wfID {
		t.Error("workflow should be triggered")
	}
}

func waitForTrigger(t *testing.T, c chan triggerMock) (*triggerMock, error) {
	t.Helper()
	var count int
	for {
		select {
		case startedAction := <-c:
			return &startedAction, nil
		default:
			if count > 5 {
				return nil, fmt.Errorf("timeout")
			}
			time.Sleep(1 * time.Millisecond)
			count++
		}
	}
}

func newEventWithMeta(subj, t string, id uuid.UUID, context map[string]any) *cloudevents.Event {
	ev := &cloudevents.Event{
		Context: &event.EventContextV03{
			Type: t,
			ID:   id.String(),
			Time: &types.Timestamp{
				Time: time.Now().UTC(),
			},
			Subject:    &subj,
			Source:     *types.ParseURIRef("test.com"),
			Extensions: context,
		},
	}

	return ev
}

func newEvent(subj, t string, id uuid.UUID) *cloudevents.Event {
	ev := &cloudevents.Event{
		Context: &event.EventContextV03{
			Type: t,
			ID:   id.String(),
			Time: &types.Timestamp{
				Time: time.Now().UTC(),
			},
			Subject: &subj,
			Source:  *types.ParseURIRef("test.com"),
		},
	}

	return ev
}
