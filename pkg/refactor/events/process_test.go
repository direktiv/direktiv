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
		TriggerInstance:        instID,
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
			TriggerWorkflow:        wfID,
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
			TriggerWorkflow:        wfID,
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

func waitForTrigger(t *testing.T, c chan triggerMock) (*triggerMock, error) {
	t.Helper()
	var count int
	for {
		select {
		case startedAction := <-c:
			return &startedAction, nil
		default:
			if count > 3 {
				return nil, fmt.Errorf("timeout")
			}
			time.Sleep(1 * time.Millisecond)
			count++
		}
	}
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
