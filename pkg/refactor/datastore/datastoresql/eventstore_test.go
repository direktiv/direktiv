package datastoresql_test

import (
	"context"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore/datastoresql"
	"github.com/direktiv/direktiv/pkg/refactor/events"
	"github.com/google/uuid"
)

func Test_EventStoreAddGet(t *testing.T) {
	ns := uuid.New()
	eID := uuid.New()
	e2ID := uuid.New()
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	subj := "subject"
	hist := datastoresql.NewSQLStore(db, "some key").EventHistory()
	ev := newEvent(subj, "test-type", eID, ns)
	ev2 := newEvent(subj, "test-type", e2ID, ns)

	ls := make([]*events.Event, 0)
	ls = append(ls, &ev, &ev2)
	_, errs := hist.Append(context.Background(), ls)
	for _, err := range errs {
		if err != nil {
			t.Error(err)

			return
		}
	}

	gotEvents, err := hist.GetAll(context.Background())
	if err != nil {
		t.Error(err)

		return
	}
	if len(gotEvents) == 0 {
		t.Error("got no results")
	}
	if len(gotEvents) != 2 {
		t.Error("missing results")

		return
	}
	for _, e := range gotEvents {
		if e.Event.Type() != "test-type" {
			t.Error("Event had wrong type")
		}
	}
	res, c, err := hist.Get(context.Background(), 0, 0, ns)
	if err != nil {
		t.Error(err)

		return
	}
	if len(res) == 0 {
		t.Error("got not results")
	}

	if c != len(res) {
		t.Error("total count is off")
	}

	res, c, err = hist.Get(context.Background(), 1, 0, ns)
	if err != nil {
		t.Error(err)

		return
	}
	if len(res) == 0 {
		t.Error("got not results")
	}
	if len(res) != 1 {
		t.Error("limit was not applied is off")
	}
	if c == 1 {
		t.Error("count is off")
	}

	e, err := hist.GetByID(context.Background(), eID.String())
	if err != nil {
		t.Error(err)
	}
	if e.Namespace != ns {
		t.Error("returned event contains wrong ns")
	}
}

func newEvent(subj, t string, id, ns uuid.UUID) events.Event {
	ev := events.Event{
		Event: &cloudevents.Event{
			Context: &event.EventContextV03{
				Type: t,
				ID:   id.String(),
				Time: &types.Timestamp{
					Time: time.Now().UTC(),
				},
				Subject: &subj,
				Source:  *types.ParseURIRef("test.com"),
			},
		},
		Namespace:  ns,
		ReceivedAt: time.Now().UTC(),
	}

	return ev
}

func Test_sqlEventHistoryStore_Append(t *testing.T) {
	ns := uuid.New()
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unexpected NewMockGorm() error = %v", err)
	}
	eventHistoryStore := datastoresql.NewSQLStore(db, "some key").EventHistory()

	ev := newEvent("subject", "test-type", uuid.New(), ns)
	ev2 := newEvent("subject", "test-type", uuid.New(), ns)

	appendEvents := []*events.Event{&ev, &ev2}
	appendedEvents, appendErrs := eventHistoryStore.Append(context.Background(), appendEvents)
	for _, err := range appendErrs {
		if err != nil {
			t.Error(err)
			return
		}
	}
	if len(appendedEvents) != 2 {
		t.Error("appended event count mismatch")
	}
}

func Test_sqlEventListenerStore_UpdateOrDelete_Update(t *testing.T) {
	ns := uuid.New()
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unexpected NewMockGorm() error = %v", err)
	}
	listenerStore := datastoresql.NewSQLStore(db, "some key").EventListener()

	// Create a test event listener
	listener := &events.EventListener{
		ID:                          uuid.New(),
		CreatedAt:                   time.Now().UTC(),
		UpdatedAt:                   time.Now().UTC(),
		Deleted:                     false,
		NamespaceID:                 ns,
		ListeningForEventTypes:      []string{"test-type"},
		ReceivedEventsForAndTrigger: []*events.Event{},
		LifespanOfReceivedEvents:    10000,
		TriggerType:                 events.StartSimple,
		TriggerWorkflow:             uuid.New().String(),
	}

	// Append the test event listener
	err = listenerStore.Append(context.Background(), listener)
	if err != nil {
		t.Fatalf("error appending test listener: %v", err)
	}

	// Create a cloudevents.Event instance
	event := cloudevents.NewEvent()
	event.SetType("test-event")
	event.SetSource("http://example.com")
	event.SetID("123456")
	event.SetTime(time.Now())

	// Add data to the event's payload
	data := struct {
		Message string `json:"message"`
	}{
		Message: "Hello, CloudEvents!",
	}

	event.SetData(cloudevents.ApplicationJSON, data)

	// Update the listener's properties
	listener.UpdatedAt = time.Now().UTC()
	listener.ReceivedEventsForAndTrigger = append(listener.ReceivedEventsForAndTrigger, &events.Event{
		Event:         &event,
		Namespace:     ns,
		NamespaceName: "test-ns",
		ReceivedAt:    time.Now().UTC(), // Set current time,
	})

	// Call the UpdateOrDelete method
	updateErrs := listenerStore.UpdateOrDelete(context.Background(), []*events.EventListener{listener})
	for _, err := range updateErrs {
		if err != nil {
			t.Errorf("error updating or deleting listener: %v", err)
			return
		}
	}

	// Retrieve the updated listener by ID
	updatedListener, getErr := listenerStore.GetByID(context.Background(), listener.ID)
	if getErr != nil {
		t.Errorf("error getting updated listener: %v", getErr)
		return
	}

	// Validate that the updated fields match the changes
	if updatedListener.ReceivedEventsForAndTrigger[0].Event.ID() != event.ID() {
		t.Errorf("expected updated event to match, but they differ")
	}
}

func Test_TopicAddGet(t *testing.T) {
	ns := uuid.New()
	eID := uuid.New()
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	store := datastoresql.NewSQLStore(db, "some key")
	topics := store.EventListenerTopics()
	listeners := store.EventListener()
	err = listeners.Append(context.Background(), &events.EventListener{
		ID:                          eID,
		CreatedAt:                   time.Now().UTC(),
		UpdatedAt:                   time.Now().UTC(),
		Deleted:                     false,
		NamespaceID:                 ns,
		ListeningForEventTypes:      []string{"a"},
		ReceivedEventsForAndTrigger: make([]*events.Event, 0),
		LifespanOfReceivedEvents:    10000,
		TriggerType:                 1,
		TriggerWorkflow:             uuid.New().String(),
	})
	if err != nil {
		t.Error(err)
	}
	err = topics.Append(context.Background(), ns, eID, ns.String()+"-a", "")
	if err != nil {
		t.Error(err)
	}
	res, err := topics.GetListeners(context.Background(), ns.String()+"-a")
	if err != nil {
		t.Error(err)
	}
	if len(res) == 0 {
		t.Error("got no results")
	}
	for _, el := range res {
		if el.NamespaceID != ns {
			t.Error("got wrong namespace")
		}
	}
}

func Test_ListenerAddDeleteGet(t *testing.T) {
	ns := uuid.New()
	eID := uuid.New()
	wf := uuid.New()
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	store := datastoresql.NewSQLStore(db, "some key")
	listeners := store.EventListener()
	err = listeners.Append(context.Background(), &events.EventListener{
		ID:                          eID,
		CreatedAt:                   time.Now().UTC(),
		UpdatedAt:                   time.Now().UTC(),
		Deleted:                     false,
		NamespaceID:                 ns,
		ListeningForEventTypes:      []string{"a"},
		ReceivedEventsForAndTrigger: make([]*events.Event, 0),
		LifespanOfReceivedEvents:    10000,
		TriggerType:                 1,
		TriggerWorkflow:             wf.String(),
	})
	if err != nil {
		t.Error(err)
	}
	res, err := listeners.GetByID(context.Background(), eID)
	if err != nil {
		t.Error(err)
	}
	if res.ID != eID {
		t.Error("got wrong entry")
	}
	got, count, err := listeners.Get(context.Background(), ns, 0, 0)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("got wrong count")
	}
	if len(got) != 1 {
		t.Error("got wrong results")
	}
	if got[0].ID != eID {
		t.Error("got wrong entry")
	}
	if got[0].TriggerWorkflow != wf.String() {
		t.Error("trigger info was not correct")
	}
	got[0].UpdatedAt = time.Now().UTC()
	got[0].Deleted = true
	errs := listeners.UpdateOrDelete(context.Background(), []*events.EventListener{got[0]})
	for _, err := range errs {
		if err != nil {
			t.Error(err)

			return
		}
	}
	got, count, err = listeners.Get(context.Background(), ns, 0, 0)
	if err != nil {
		t.Error(err)
	}
	if count != 0 {
		t.Error("got wrong count")
	}
	if len(got) != 0 {
		t.Error("got wrong results")
	}
	_, err = listeners.GetByID(context.Background(), eID)
	if err == nil {
		t.Error("entry was excepted to be deleted")
	}
}

func Test_sqlEventListenerStore_UpdateOrDelete(t *testing.T) {
	ns := uuid.New()
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unexpected NewMockGorm() error = %v", err)
	}
	listenerStore := datastoresql.NewSQLStore(db, "some key").EventListener()

	// Create a test event listener
	listener := &events.EventListener{
		ID:                          uuid.New(),
		CreatedAt:                   time.Now().UTC(),
		UpdatedAt:                   time.Now().UTC(),
		Deleted:                     false,
		NamespaceID:                 ns,
		ListeningForEventTypes:      []string{"test-type"},
		ReceivedEventsForAndTrigger: []*events.Event{},
		LifespanOfReceivedEvents:    10000,
		TriggerType:                 events.StartSimple,
		TriggerWorkflow:             uuid.New().String(),
	}

	// Append the test event listener
	err = listenerStore.Append(context.Background(), listener)
	if err != nil {
		t.Fatalf("error appending test listener: %v", err)
	}

	// Modify the listener to be deleted
	listener.Deleted = true

	// Update or delete the listener
	errs := listenerStore.UpdateOrDelete(context.Background(), []*events.EventListener{listener})
	for _, err := range errs {
		if err != nil {
			t.Errorf("error updating or deleting listener: %v", err)
		}
	}

	// Verify that the listener has been deleted
	_, err = listenerStore.GetByID(context.Background(), listener.ID)
	if err == nil {
		t.Error("listener still exists after deletion")
	}
}

func Test_ListenerAddDeleteByWf(t *testing.T) {
	ns := uuid.New()
	eID := uuid.New()
	wf := uuid.New()
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	store := datastoresql.NewSQLStore(db, "some key")
	listeners := store.EventListener()
	err = listeners.Append(context.Background(), &events.EventListener{
		ID:                          eID,
		CreatedAt:                   time.Now().UTC(),
		UpdatedAt:                   time.Now().UTC(),
		Deleted:                     false,
		NamespaceID:                 ns,
		ListeningForEventTypes:      []string{"a"},
		ReceivedEventsForAndTrigger: make([]*events.Event, 0),
		LifespanOfReceivedEvents:    10000,
		TriggerType:                 1,
		TriggerWorkflow:             wf.String(),
	})
	if err != nil {
		t.Error(err)
	}

	ids, err := listeners.DeleteAllForWorkflow(context.Background(), wf)
	if err != nil {
		t.Error(err)
	}
	if *ids[0] != eID {
		t.Error("listenerid Was wrong")
	}
	_, err = listeners.GetByID(context.Background(), eID)
	if err == nil {
		t.Error("expected this listener to be deleted")
	}
}

func Test_ListenerAddDeleteGetWithPaginationAndBoundaryCheck(t *testing.T) {
	ns := uuid.New()
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unexpected NewMockGorm() error = %v", err)
	}
	store := datastoresql.NewSQLStore(db, "some key")
	listeners := store.EventListener()

	// Adding 11 entries
	for i := 0; i < 11; i++ {
		eID := uuid.New()
		wf := uuid.New()
		err = listeners.Append(context.Background(), &events.EventListener{
			ID:                          eID,
			CreatedAt:                   time.Now().UTC(),
			UpdatedAt:                   time.Now().UTC(),
			Deleted:                     false,
			NamespaceID:                 ns,
			ListeningForEventTypes:      []string{"a"},
			ReceivedEventsForAndTrigger: make([]*events.Event, 0),
			LifespanOfReceivedEvents:    10000,
			TriggerType:                 1,
			TriggerWorkflow:             wf.String(),
		})
		if err != nil {
			t.Errorf("failed to append listener %d: %v", i, err)
		}
	}

	// First pagination test with offset = 5, limit = 3
	offset := 5
	limit := 3
	got, count, err := listeners.Get(context.Background(), ns, limit, offset)
	if err != nil {
		t.Error(err)
	}
	if len(got) != limit {
		t.Errorf("expected %d results, got %d", limit, len(got))
	}
	if count != 11 {
		t.Errorf("expected total count to be 11, got %d", count)
	}

	// Second pagination test with offset = 10, limit = 10
	offset = 10
	limit = 10
	got, count, err = listeners.Get(context.Background(), ns, offset, limit)
	if err != nil {
		t.Error(err)
	}
	// Expecting 1 result because there are 11 entries and we are starting from the 10th index
	if len(got) != 1 {
		t.Errorf("expected 1 result, got %d", len(got))
	}
	if count != 11 {
		t.Errorf("expected total count to be 11, got %d", count)
	}
	// Clean-up by deleting the entries (optional in terms of this test case logic)
	for _, listener := range got {
		listener.UpdatedAt = time.Now().UTC()
		listener.Deleted = true
		errs := listeners.UpdateOrDelete(context.Background(), []*events.EventListener{listener})
		for _, err := range errs {
			if err != nil {
				t.Error(err)
				return
			}
		}
	}
}
