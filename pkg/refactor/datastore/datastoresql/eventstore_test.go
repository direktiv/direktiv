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

func Test_Add_Get(t *testing.T) {
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

	e, err := hist.GetByID(context.Background(), eID)
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
					Time: time.Now(),
				},
				Subject: &subj,
				Source:  *types.ParseURIRef("test.com"),
			},
		},
		Namespace:  ns,
		ReceivedAt: time.Now(),
	}

	return ev
}

func Test_Topic_Add_Get(t *testing.T) {
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
		CreatedAt:                   time.Now(),
		UpdatedAt:                   time.Now(),
		Deleted:                     false,
		NamespaceID:                 ns,
		ListeningForEventTypes:      []string{"a"},
		ReceivedEventsForAndTrigger: make([]*events.Event, 0),
		LifespanOfReceivedEvents:    10000,
		TriggerType:                 1,
		TriggerWorkflow:             uuid.New(),
	})
	if err != nil {
		t.Error(err)
	}
	err = topics.Append(context.Background(), ns, eID, ns.String()+"-a")
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

func Test_Listener_Add_Delete_Get(t *testing.T) {
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
		CreatedAt:                   time.Now(),
		UpdatedAt:                   time.Now(),
		Deleted:                     false,
		NamespaceID:                 ns,
		ListeningForEventTypes:      []string{"a"},
		ReceivedEventsForAndTrigger: make([]*events.Event, 0),
		LifespanOfReceivedEvents:    10000,
		TriggerType:                 1,
		TriggerWorkflow:             uuid.New(),
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
	if got[0].TriggerWorkflow != wf {
		t.Error("trigger info was not correct")
	}
	got[0].UpdatedAt = time.Now()
	got[0].Deleted = true
	errs := listeners.Update(context.Background(), []*events.EventListener{got[0]})
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
	if count != 1 {
		t.Error("got wrong count")
	}
	if len(got) != 1 {
		t.Error("got wrong results")
	}
	if got[0].ID != eID {
		t.Error("got wrong entry")
	}
	if !got[0].Deleted {
		t.Error("entry was not updated")
	}
	err = listeners.Delete(context.Background())
	if err != nil {
		t.Error(err)
	}
	_, err = listeners.GetByID(context.Background(), eID)
	if err == nil {
		t.Error("entry was excepted to be deleted")
	}
}

func Test_Listener_Add_Delete_ByWf(t *testing.T) {
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
		CreatedAt:                   time.Now(),
		UpdatedAt:                   time.Now(),
		Deleted:                     false,
		NamespaceID:                 ns,
		ListeningForEventTypes:      []string{"a"},
		ReceivedEventsForAndTrigger: make([]*events.Event, 0),
		LifespanOfReceivedEvents:    10000,
		TriggerType:                 1,
		TriggerWorkflow:             uuid.New(),
	})
	if err != nil {
		t.Error(err)
	}

	err = listeners.DeleteAllForWorkflow(context.Background(), wf)
	if err != nil {
		t.Error(err)
	}
	_, err = listeners.GetByID(context.Background(), eID)
	if err == nil {
		t.Error("expected this listener to be deleted")
	}
}
