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
