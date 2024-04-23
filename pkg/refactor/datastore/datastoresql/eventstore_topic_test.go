package datastoresql_test

import (
	"context"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore/datastoresql"
	"github.com/direktiv/direktiv/pkg/refactor/events"
	"github.com/google/uuid"
)

func Test_TopicAddGet(t *testing.T) {
	ns := uuid.New()
	nsName := ns.String()
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
	err = topics.Append(context.Background(), ns, nsName, eID, ns.String()+"-a", "")
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
