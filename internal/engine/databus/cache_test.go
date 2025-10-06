package databus

import (
	"testing"
	"time"

	"github.com/direktiv/direktiv/internal/engine"
	"github.com/google/uuid"
)

// helper to build an InstanceStatus quickly
func mk(ns string, id uuid.UUID, created time.Time, seq uint64) *engine.InstanceStatus {
	return &engine.InstanceStatus{
		InstanceID:      id,
		Namespace:       ns,
		CreatedAt:       created,
		HistorySequence: seq,
	}
}

func TestUpsert_InsertAndUpdateNewerIgnoreOlder(t *testing.T) {
	c := NewStatusCache()
	id := uuid.New()
	now := time.Now()

	// insert
	s1 := mk("ns", id, now, 1)
	c.Upsert(s1)

	got := c.Snapshot("", uuid.Nil)
	if len(got) != 1 {
		t.Fatalf("expected 1 item, got %d", len(got))
	}
	if got[0].InstanceID != id || got[0].HistorySequence != 1 {
		t.Fatalf("unexpected first insert: %+v", got[0])
	}

	// update with newer HistorySequence
	s2 := mk("ns", id, now.Add(1*time.Minute), 2)
	c.Upsert(s2)

	got = c.Snapshot("", uuid.Nil)
	if len(got) != 1 {
		t.Fatalf("expected 1 item after update, got %d", len(got))
	}
	if got[0].HistorySequence != 2 {
		t.Fatalf("expected seq=2 after newer update, got %d", got[0].HistorySequence)
	}

	// attempt to "downgrade" with older HistorySequence (should be ignored)
	s3 := mk("ns", id, now.Add(2*time.Minute), 1)
	c.Upsert(s3)

	got = c.Snapshot("", uuid.Nil)
	if got[0].HistorySequence != 2 {
		t.Fatalf("expected seq to remain 2 after older upsert, got %d", got[0].HistorySequence)
	}
}

func TestSnapshot_FilteringAndOrdering(t *testing.T) {
	c := NewStatusCache()
	now := time.Now()

	id1 := uuid.New()
	id2 := uuid.New()
	id3 := uuid.New()

	// Intentionally insert out of CreatedAt order
	c.Upsert(mk("a", id1, now, 1))
	c.Upsert(mk("b", id2, now, 1))
	c.Upsert(mk("b", id3, now, 1))

	// All items
	all := c.Snapshot("", uuid.Nil)
	if len(all) != 3 {
		t.Fatalf("expected 3 items, got %d", len(all))
	}
	if all[0].InstanceID != id3 || all[1].InstanceID != id2 || all[2].InstanceID != id1 {
		t.Fatalf("unexpected items in all: %+v", all)
	}

	// Filter by namespace
	nsA := c.Snapshot("a", uuid.Nil)
	if len(nsA) != 1 {
		t.Fatalf("expected 2 items in ns=a, got %d", len(nsA))
	}
	for _, s := range nsA {
		if s.Namespace != "a" {
			t.Fatalf("unexpected namespace in filter: %s", s.Namespace)
		}
	}

	// Filter by instance ID
	byID := c.Snapshot("", id2)
	if len(byID) != 1 || byID[0].InstanceID != id2 {
		t.Fatalf("expected exactly the item with id2, got: %+v", byID)
	}
}

func TestSnapshotPage_LimitOffsetAndTotal(t *testing.T) {
	c := NewStatusCache()
	base := time.Now()

	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	// Insert in created order so post-sort won't change relative order
	for i := 0; i < len(ids); i++ {
		c.Upsert(mk("ns", ids[i], base.Add(time.Duration(i)*time.Minute), uint64(i+1)))
	}

	// Ask for limit=2, offset=1 within namespace "ns"
	page, total := c.SnapshotPage("ns", uuid.Nil, 2, 1)
	if total != 5 {
		t.Fatalf("expected total=5, got %d", total)
	}
	if len(page) != 2 {
		t.Fatalf("expected 2 items on page, got %d", len(page))
	}

	// Expect the second and third items by CreatedAt (since we skipped one)
	if !page[0].CreatedAt.Equal(base.Add(3 * time.Minute)) {
		t.Fatalf("expected first page item to be +3m, got %v", page[0].CreatedAt)
	}
	if !page[1].CreatedAt.Equal(base.Add(2 * time.Minute)) {
		t.Fatalf("expected second page item to be +2m, got %v", page[1].CreatedAt)
	}
}

func TestDeleteNamespace_RemovesAndRebuildsIndex(t *testing.T) {
	c := NewStatusCache()
	now := time.Now()

	idA1 := uuid.New()
	idA2 := uuid.New()
	idB := uuid.New()

	c.Upsert(mk("a", idA1, now, 1))
	c.Upsert(mk("a", idA2, now.Add(1*time.Minute), 1))
	c.Upsert(mk("b", idB, now.Add(2*time.Minute), 1))

	c.DeleteNamespace("a")

	left := c.Snapshot("", uuid.Nil)
	if len(left) != 1 {
		t.Fatalf("expected 1 item after deleting ns 'a', got %d", len(left))
	}
	if left[0].Namespace != "b" || left[0].InstanceID != idB {
		t.Fatalf("unexpected remaining item: %+v", left[0])
	}

	// Ensure index was rebuilt correctly by performing an update on the remaining item.
	c.Upsert(mk("b", idB, now.Add(3*time.Minute), 2))
	after := c.Snapshot("", uuid.Nil)
	if len(after) != 1 || after[0].HistorySequence != 2 {
		t.Fatalf("expected remaining item to update via index, got: %+v", after[0])
	}
}

func TestSnapshotReturnsClones_Immutability(t *testing.T) {
	c := NewStatusCache()
	id := uuid.New()
	now := time.Now()

	c.Upsert(mk("ns", id, now, 1))

	// mutate the snapshot copy
	snap := c.Snapshot("", uuid.Nil)
	if len(snap) != 1 {
		t.Fatalf("expected 1 item, got %d", len(snap))
	}
	snap[0].HistorySequence = 999 // should not affect the cache

	again := c.Snapshot("", uuid.Nil)
	if again[0].HistorySequence != 1 {
		t.Fatalf("expected cache to be immutable via snapshots, got %d", again[0].HistorySequence)
	}
}
