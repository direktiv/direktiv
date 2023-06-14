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
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	subj := "subject"
	hist := datastoresql.NewSQLStore(db, "some key").EventHistory()
	ev := events.Event{
		Event: &cloudevents.Event{
			Context: &event.EventContextV03{
				Type: "test",
				ID:   uuid.NewString(),
				Time: &types.Timestamp{
					Time: time.Now(),
				},
				Subject: &subj,
				Source:  *types.ParseURIRef("test.com"),
			},
		},
		Namespace:  uuid.New(),
		ReceivedAt: time.Now(),
	}

	ls := make([]*events.Event, 0)
	ls = append(ls, &ev)
	_, err = hist.Append(context.Background(), &events.Event{
		Event: &cloudevents.Event{
			Context: &event.EventContextV03{
				Type: "test",
				ID:   eID.String(),
				Time: &types.Timestamp{
					Time: time.Now(),
				},
				Subject: &subj,
				Source:  *types.ParseURIRef("test.com"),
			},
		},
		Namespace:  ns,
		ReceivedAt: time.Now(),
	}, ls...,
	)
	if err != nil {
		t.Error(err)

		return
	}

	events, err := hist.GetAll(context.Background())
	if err != nil {
		t.Error(err)

		return
	}
	if len(events) == 0 {
		t.Error("got no results")
	}
	for _, e := range events {
		if e.Event.Type() != "test" {
			t.Error("Event had no type")
		}
	}
	var c int
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

	e, err := hist.GetByID(context.Background(), eID)
	if err != nil {
		t.Error(err)
	}
	if e.Namespace != ns {
		t.Error("returned event contains wrong ns")
	}
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
		Trigger:                     events.TriggerInfo{WorkflowID: uuid.New()},
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
		Trigger:                     events.TriggerInfo{WorkflowID: uuid.New()},
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
	got[0].UpdatedAt = time.Now()
	got[0].Deleted = true
	err, errs := listeners.Update(context.Background(), got[0])
	if err != nil {
		t.Error(err)
	}
	if len(errs) != 0 {
		t.Error("got unexpected errs")
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
