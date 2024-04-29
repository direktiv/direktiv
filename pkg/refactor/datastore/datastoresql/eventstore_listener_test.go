package datastoresql_test

import (
	"context"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/datastore/datastoresql"
	"github.com/google/uuid"
)

func TestEventListenerOperations(t *testing.T) {
	t.Run("AddDeleteAndGet", testAddDeleteAndGet)
	t.Run("Update", testUpdate)
	t.Run("PaginationAndBoundaryCheck", testPaginationAndBoundaryCheck)
	t.Run("DeleteByWorkflow", testDeleteByWorkflow)
}

func testAddDeleteAndGet(t *testing.T) {
	listenerStore, listener, _, _ := setupTest(t)
	addTestEventListener(t, listenerStore, listener)
	verifyListenerAdded(t, listenerStore, listener)

	retrievedListener := getListenerByID(t, listenerStore, listener.ID, listener.NamespaceID)
	verifyListenerRetrieved(t, retrievedListener, listener)

	verifyAllListenersForNamespace(t, listenerStore, listener.NamespaceID, listener)

	deleteAndUpdateListener(t, listenerStore, listener)
	verifyListenerDeleted(t, listenerStore, listener.ID)
}

func testUpdate(t *testing.T) {
	listenerStore, listener, _, _ := setupTest(t)
	addTestEventListener(t, listenerStore, listener)

	listener.UpdatedAt = time.Now().UTC()
	updateErrs := listenerStore.UpdateOrDelete(context.Background(), []*datastore.EventListener{listener})
	for _, err := range updateErrs {
		if err != nil {
			t.Errorf("error updating listener: %v", err)
		}
	}

	_, getErr := listenerStore.GetByID(context.Background(), listener.ID)
	if getErr != nil {
		t.Errorf("error getting updated listener: %v", getErr)
	}
}

func testPaginationAndBoundaryCheck(t *testing.T) {
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
		err = listeners.Append(context.Background(), &datastore.EventListener{
			ID:                          eID,
			CreatedAt:                   time.Now().UTC(),
			UpdatedAt:                   time.Now().UTC(),
			Deleted:                     false,
			NamespaceID:                 ns,
			ListeningForEventTypes:      []string{"a"},
			ReceivedEventsForAndTrigger: make([]*datastore.Event, 0),
			LifespanOfReceivedEvents:    10000,
			TriggerType:                 1,
			TriggerWorkflow:             wf.String(),
		})
		if err != nil {
			t.Errorf("failed to append listener %d: %v", i, err)
		}
	}

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

	offset = 10
	limit = 10
	got, count, err = listeners.Get(context.Background(), ns, offset, limit)
	if err != nil {
		t.Error(err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 result, got %d", len(got))
	}
	if count != 11 {
		t.Errorf("expected total count to be 11, got %d", count)
	}

	for _, listener := range got {
		listener.UpdatedAt = time.Now().UTC()
		listener.Deleted = true
		errs := listeners.UpdateOrDelete(context.Background(), []*datastore.EventListener{listener})
		for _, err := range errs {
			if err != nil {
				t.Error(err)
				return
			}
		}
	}
}

func testDeleteByWorkflow(t *testing.T) {
	ns := uuid.New()
	eID := uuid.New()
	wf := uuid.New()
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unexpected NewMockGorm() error = %v", err)
	}
	store := datastoresql.NewSQLStore(db, "some key")
	listeners := store.EventListener()
	err = listeners.Append(context.Background(), &datastore.EventListener{
		ID:                          eID,
		CreatedAt:                   time.Now().UTC(),
		UpdatedAt:                   time.Now().UTC(),
		Deleted:                     false,
		NamespaceID:                 ns,
		ListeningForEventTypes:      []string{"a"},
		ReceivedEventsForAndTrigger: make([]*datastore.Event, 0),
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

func setupTest(t *testing.T) (datastore.EventListenerStore, *datastore.EventListener, uuid.UUID, string) {
	ns := uuid.New()
	nsName := ns.String()
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unexpected NewMockGorm() error = %v", err)
	}
	listenerStore := datastoresql.NewSQLStore(db, "some key").EventListener()

	listener := createTestEventListener(ns)

	return listenerStore, listener, ns, nsName
}

func createTestEventListener(ns uuid.UUID) *datastore.EventListener {
	return &datastore.EventListener{
		ID:                          uuid.New(),
		CreatedAt:                   time.Now().UTC(),
		UpdatedAt:                   time.Now().UTC(),
		Deleted:                     false,
		NamespaceID:                 ns,
		ListeningForEventTypes:      []string{"test-type"},
		ReceivedEventsForAndTrigger: []*datastore.Event{},
		LifespanOfReceivedEvents:    10000,
		TriggerType:                 datastore.StartSimple,
		TriggerWorkflow:             uuid.New().String(),
	}
}

func addTestEventListener(t *testing.T, listenerStore datastore.EventListenerStore, listener *datastore.EventListener) {
	err := listenerStore.Append(context.Background(), listener)
	if err != nil {
		t.Fatalf("error appending test listener: %v", err)
	}
}

func verifyListenerAdded(t *testing.T, store datastore.EventListenerStore, listener *datastore.EventListener) {
	_, err := store.GetByID(context.Background(), listener.ID)
	if err != nil {
		t.Errorf("error verifying listener added: %v", err)
	}
}

func getListenerByID(t *testing.T, store datastore.EventListenerStore, listenerID, nsID uuid.UUID) *datastore.EventListener {
	retrievedListener, err := store.GetByID(context.Background(), listenerID)
	if err != nil {
		t.Errorf("error retrieving listener by ID: %v", err)
	}
	return retrievedListener
}

func verifyListenerRetrieved(t *testing.T, retrievedListener, expectedListener *datastore.EventListener) {
	if retrievedListener.ID != expectedListener.ID {
		t.Error("retrieved listener ID does not match expected")
	}
	// Add more checks for other fields if necessary
}

func verifyAllListenersForNamespace(t *testing.T, store datastore.EventListenerStore, nsID uuid.UUID, listener *datastore.EventListener) {
	listeners, count, err := store.Get(context.Background(), nsID, 0, 0)
	if err != nil {
		t.Errorf("error retrieving listeners for namespace: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count of listeners to be 1, got %d", count)
	}
	if len(listeners) != 1 {
		t.Errorf("expected 1 listener, got %d", len(listeners))
	}
	// TODO: Add more checks for content if necessary
}

func deleteAndUpdateListener(t *testing.T, store datastore.EventListenerStore, listener *datastore.EventListener) {
	listener.UpdatedAt = time.Now().UTC()
	listener.Deleted = true
	errs := store.UpdateOrDelete(context.Background(), []*datastore.EventListener{listener})
	for _, err := range errs {
		if err != nil {
			t.Errorf("error updating or deleting listener: %v", err)
		}
	}
}

func verifyListenerDeleted(t *testing.T, store datastore.EventListenerStore, listenerID uuid.UUID) {
	_, err := store.GetByID(context.Background(), listenerID)
	if err == nil {
		t.Error("expected listener to be deleted")
	}
}
