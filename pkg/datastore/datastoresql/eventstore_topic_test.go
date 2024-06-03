package datastoresql_test

import (
	"context"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/datastore/datastoresql"
	"github.com/google/uuid"
)

func Test_TopicAddGet(t *testing.T) {
	// Create a mock database
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unexpected NewMockGorm() error: %v", err)
	}

	// Create a new SQL data store
	store := datastoresql.NewSQLStore(db, "some key")

	// Generate test data
	ns := uuid.New()
	nsName := ns.String()
	eID := uuid.New()
	listeningForEventType := "a"
	topicName := ns.String() + "-" + listeningForEventType

	// Add event listener
	err = addEventListener(t, store.EventListener(), eID, ns)
	if err != nil {
		t.Fatal(err)
	}

	// Add topic
	err = addTopic(t, store.EventListenerTopics(), ns, nsName, eID, topicName, "")
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve listeners for the topic
	res, err := store.EventListenerTopics().GetListeners(context.Background(), topicName)
	if err != nil {
		t.Fatal(err)
	}

	// Assert results
	assertListeners(t, res, ns)
}

func addEventListener(t *testing.T, listenerStore datastore.EventListenerStore, eID uuid.UUID, ns uuid.UUID) error {
	err := listenerStore.Append(context.Background(), &datastore.EventListener{
		ID:                          eID,
		CreatedAt:                   time.Now().UTC(),
		UpdatedAt:                   time.Now().UTC(),
		Deleted:                     false,
		NamespaceID:                 ns,
		ListeningForEventTypes:      []string{"a"},
		ReceivedEventsForAndTrigger: make([]*datastore.Event, 0),
		LifespanOfReceivedEvents:    10000,
		TriggerType:                 1,
		TriggerWorkflow:             uuid.New().String(),
	})
	if err != nil {
		t.Errorf("failed to add event listener: %v", err)
		return err
	}
	return nil
}

func addTopic(t *testing.T, topicStore datastore.EventTopicsStore, ns uuid.UUID, nsName string, eID uuid.UUID, topicName string, extraInfo string) error {
	err := topicStore.Append(context.Background(), ns, nsName, eID, topicName, extraInfo)
	if err != nil {
		t.Errorf("failed to add topic: %v", err)
		return err
	}
	return nil
}

func assertListeners(t *testing.T, listeners []*datastore.EventListener, expectedNamespace uuid.UUID) {
	if len(listeners) == 0 {
		t.Error("got no results")
	}

	for _, el := range listeners {
		if el.NamespaceID != expectedNamespace {
			t.Error("got wrong namespace")
		}
	}
}
