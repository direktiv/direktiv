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
	ev := newEvent(subj, "test-type", eID, ns, ns.String())
	ev2 := newEvent(subj, "test-type", e2ID, ns, ns.String())

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

func Test_EventStoreAddGetNew(t *testing.T) {
	ns := uuid.New()
	eID := uuid.New()
	e2ID := uuid.New()
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	subj := "subject"
	hist := datastoresql.NewSQLStore(db, "some key").EventHistory()
	ev := newEvent(subj, "test-type", eID, ns, ns.String())
	ev2 := newEvent(subj, "test-type", e2ID, ns, ns.String())

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
	res, err := hist.GetOld(context.Background(), ns.String(), time.Now().UTC())
	if err != nil {
		t.Error(err)

		return
	}
	if len(res) == 0 {
		t.Error("got not results")
	}
	if len(res) != 2 {
		t.Error("missing results")
	}
	e, err := hist.GetByID(context.Background(), eID.String())
	if err != nil {
		t.Error(err)
	}
	if e.Namespace != ns {
		t.Error("returned event contains wrong ns")
	}
}

func newEvent(subj, t string, id, ns uuid.UUID, nsName string) events.Event {
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
		Namespace:     ns,
		NamespaceName: nsName,
		ReceivedAt:    time.Now().UTC(),
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

	ev := newEvent("subject", "test-type", uuid.New(), ns, ns.String())
	ev2 := newEvent("subject", "test-type", uuid.New(), ns, ns.String())

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
