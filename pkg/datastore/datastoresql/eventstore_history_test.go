package datastoresql_test

import (
	"context"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/datastore/datastoresql"
	"github.com/google/uuid"
)

func setupEventHistoryStore(t *testing.T) (datastore.EventHistoryStore, uuid.UUID, string) {
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unexpected NewMockGorm() error: %v", err)
	}

	ns := uuid.New()
	nsName := ns.String()

	return datastoresql.NewSQLStore(db, "some key").EventHistory(), ns, nsName
}

// func Test_EventStoreAddGet(t *testing.T) {
// 	hist, ns, nsName := setupEventHistoryStore(t)

// 	eID := uuid.New()
// 	e2ID := uuid.New()

// 	ev := newEvent("subject", "test-type", eID, ns, nsName)
// 	ev2 := newEvent("subject", "test-type", e2ID, ns, nsName)

// 	ls := []*datastore.Event{&ev, &ev2}
// 	_, errs := hist.Append(context.Background(), ls)
// 	for _, err := range errs {
// 		if err != nil {
// 			t.Error(err)
// 			return
// 		}
// 	}

// 	// Assert that events were added successfully
// 	assertEventsAdded(t, hist, ns)

// 	// Test Get() method
// 	testGet(t, hist, ns)
// }

func Test_EventStoreAddGetNew(t *testing.T) {
	hist, ns, nsName := setupEventHistoryStore(t)

	eID := uuid.New()
	e2ID := uuid.New()

	ev := newEvent("subject", "test-type", eID, ns, nsName)
	ev2 := newEvent("subject", "test-type", e2ID, ns, nsName)

	ls := []*datastore.Event{&ev, &ev2}
	_, errs := hist.Append(context.Background(), ls)
	for _, err := range errs {
		if err != nil {
			t.Error(err)
			return
		}
	}

	// Assert that events were added successfully
	assertEventsAdded(t, hist, ns)

	// Test GetOld() method
	testGetOld(t, hist, ns)
}

func Test_DeleteOldEvents(t *testing.T) {
	hist, ns, _ := setupEventHistoryStore(t)

	// Add some events
	eID := uuid.New()
	ev := newEvent("subject", "test-type", eID, ns, "")
	_, errs := hist.Append(context.Background(), []*datastore.Event{&ev})
	for _, err := range errs {
		if err != nil {
			t.Error(err)
			return
		}
	}

	// Delete old events
	sinceWhen := time.Now().UTC().Add(time.Hour) // Delete events older than an hour
	err := hist.DeleteOld(context.Background(), sinceWhen)
	if err != nil {
		t.Error(err)
		return
	}

	// Verify that events were deleted
	res, err := hist.GetAll(context.Background())
	if err != nil {
		t.Error(err)
		return
	}

	if len(res) != 0 {
		t.Error("expected 0 events after deletion, but got", len(res))
	}
}

func Test_GetEventByID(t *testing.T) {
	hist, ns, _ := setupEventHistoryStore(t)

	// Add an event
	eID := uuid.New()
	ev := newEvent("subject", "test-type", eID, ns, "")
	_, errs := hist.Append(context.Background(), []*datastore.Event{&ev})
	for _, err := range errs {
		if err != nil {
			t.Error(err)
			return
		}
	}

	// Retrieve the event by ID
	retrievedEvent, err := hist.GetByID(context.Background(), eID.String())
	if err != nil {
		t.Error(err)
		return
	}

	// Verify that the retrieved event matches the original event
	if retrievedEvent == nil {
		t.Error("failed to retrieve event")
		return
	}

	if retrievedEvent.Event.ID() != eID.String() {
		t.Error("retrieved event ID does not match")
	}
}

func assertEventsAdded(t *testing.T, hist datastore.EventHistoryStore, ns uuid.UUID) {
	// Retrieve all events
	gotEvents, err := hist.GetAll(context.Background())
	if err != nil {
		t.Error(err)
		return
	}

	// Assert that events were added
	if len(gotEvents) != 2 {
		t.Error("expected 2 events, but got", len(gotEvents))
		return
	}
}

func testGetOld(t *testing.T, hist datastore.EventHistoryStore, ns uuid.UUID) {
	// Test GetOld() method
	res, err := hist.GetOld(context.Background(), ns.String(), time.Now().UTC())
	if err != nil {
		t.Error(err)
		return
	}

	// Assert results
	if len(res) != 2 {
		t.Error("expected 2 events, but got", len(res))
		return
	}
}

func assertResults(t *testing.T, res []*datastore.Event, c int) {
	if len(res) == 0 {
		t.Error("got no results")
		return
	}

	if c != len(res) {
		t.Error("total count is off")
	}
}

func newEvent(subj, t string, id, ns uuid.UUID, nsName string) datastore.Event {
	return datastore.Event{
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
		NamespaceID: ns,
		Namespace:   nsName,
		ReceivedAt:  time.Now().UTC(),
	}
}
