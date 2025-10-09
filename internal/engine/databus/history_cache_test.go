package databus

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/direktiv/direktiv/internal/engine"
	"github.com/google/uuid"
)

// helper to build a rich InstanceEvent for testing.
func makeEvent(ns string, instID uuid.UUID, seq uint64) engine.InstanceEvent {
	meta := map[string]string{"k": "v", "x": "y"}
	in := json.RawMessage(`{"in":1}`)
	mem := json.RawMessage(`{"mem":true}`)
	out := json.RawMessage(`{"out":"ok"}`)
	now := time.Now().UTC()

	return engine.InstanceEvent{
		EventID:    uuid.New(),
		InstanceID: instID,
		Namespace:  ns,
		Metadata:   meta,
		Type:       "test",
		Time:       now,

		Script:   "script",
		Mappings: "maps",
		Fn:       "fn",
		Input:    in,
		Memory:   mem,
		Output:   out,
		Error:    "",

		Sequence: seq,
	}
}

func TestInsertAndSnapshot_FilteringAndDeepCopyOnInsert(t *testing.T) {
	c := NewHistoryCache()
	instA := uuid.New()
	instB := uuid.New()

	// Create events
	e1 := makeEvent("ns1", instA, 1)
	e2 := makeEvent("ns1", instA, 2)
	e3 := makeEvent("ns2", instA, 3)
	e4 := makeEvent("ns1", instB, 4)

	// Insert a pointer to e1, then mutate the original to ensure the cache cloned it.
	c.Insert(&e1)
	e1.Metadata["k"] = "mutated"
	e1.Input = json.RawMessage(`{"in":999}`)
	e1.Script = "mutated-script"

	// Insert others normally
	c.Insert(&e2)
	c.Insert(&e3)
	c.Insert(&e4)

	// Snapshot only ns1 + instA
	got := c.Snapshot("ns1", instA)
	if len(got) != 2 {
		t.Fatalf("expected 2 events in snapshot, got %d", len(got))
	}

	// Ensure order by insertion (since DeleteNamespace not involved)
	// and that the first snapshot entry is NOT affected by e1 mutations after Insert.
	// Find the one with Sequence == 1 (e1)
	var snapE1 *engine.InstanceEvent
	for _, ev := range got {
		if ev.Sequence == 1 {
			snapE1 = ev
			break
		}
	}
	if snapE1 == nil {
		t.Fatalf("did not find event with sequence 1 in snapshot")
	}

	if snapE1.Metadata["k"] != "v" {
		t.Fatalf("expected cached metadata['k'] to be 'v', got %q", snapE1.Metadata["k"])
	}
	if string(snapE1.Input) != `{"in":1}` {
		t.Fatalf("expected cached Input to be %s, got %s", `{"in":1}`, string(snapE1.Input))
	}
	if snapE1.Script != "script" {
		t.Fatalf("expected cached Script to be 'script', got %q", snapE1.Script)
	}

	// Ensure snapshot only includes ns1/instA
	for _, ev := range got {
		if ev.Namespace != "ns1" || ev.InstanceID != instA {
			t.Fatalf("snapshot contains unexpected event: ns=%s id=%s", ev.Namespace, ev.InstanceID)
		}
	}
}

func TestSnapshot_ReturnsClones_NotBackedByCache(t *testing.T) {
	c := NewHistoryCache()
	inst := uuid.New()
	e := makeEvent("ns", inst, 1)
	c.Insert(&e)

	// Take snapshot and mutate the returned objects/slices.
	snap1 := c.Snapshot("ns", inst)
	if len(snap1) != 1 {
		t.Fatalf("expected 1 event, got %d", len(snap1))
	}
	clone := snap1[0]

	// Mutate clone's map and raw messages
	clone.Metadata["new"] = "field"
	if len(clone.Input) > 0 {
		clone.Input[0] = '{' // destructive in-place change
	}
	clone.Script = "changed"

	// Take another snapshot to verify cache wasn't affected.
	snap2 := c.Snapshot("ns", inst)
	if len(snap2) != 1 {
		t.Fatalf("expected 1 event, got %d", len(snap2))
	}
	again := snap2[0]

	// The map should not contain "new"
	if _, ok := again.Metadata["new"]; ok {
		t.Fatalf("cache was mutated via snapshot clone (map)")
	}
	// The bytes should be unchanged
	if string(again.Input) != `{"in":1}` {
		t.Fatalf("cache was mutated via snapshot clone (raw message). got %s", string(again.Input))
	}
	// String fields (value-copied) should remain original
	if again.Script != "script" {
		t.Fatalf("cache was mutated via snapshot clone (string). got %q", again.Script)
	}
}

func TestDeleteNamespace_RemovesOnlyMatchingNamespace(t *testing.T) {
	c := NewHistoryCache()
	inst := uuid.New()

	events := []engine.InstanceEvent{
		makeEvent("keep", inst, 1),
		makeEvent("drop", inst, 2),
		makeEvent("keep", inst, 3),
		makeEvent("drop", inst, 4),
	}
	for i := range events {
		ev := events[i]
		c.Insert(&ev)
	}

	// Delete "drop" namespace
	c.DeleteNamespace("drop")

	// Snapshot "drop" => should be empty
	gotDrop := c.Snapshot("drop", inst)
	if len(gotDrop) != 0 {
		t.Fatalf("expected 0 events after DeleteNamespace for 'drop', got %d", len(gotDrop))
	}

	// Snapshot "keep" => should contain 2 items with seq 1 and 3
	gotKeep := c.Snapshot("keep", inst)
	if len(gotKeep) != 2 {
		t.Fatalf("expected 2 events for 'keep', got %d", len(gotKeep))
	}
	wantSeqs := []uint64{1, 3}
	gotSeqs := []uint64{gotKeep[0].Sequence, gotKeep[1].Sequence}
	if !reflect.DeepEqual(wantSeqs, gotSeqs) {
		t.Fatalf("expected sequences %v, got %v", wantSeqs, gotSeqs)
	}
}

func TestInsert_NilPointerDoesPanic(t *testing.T) {
	c := NewHistoryCache()
	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Fatalf("Insert(nil) didn't panic")
	}()
	c.Insert(nil)

	// cache should remain empty
	if got := c.Snapshot("any", uuid.New()); len(got) != 0 {
		t.Fatalf("expected empty snapshot after Insert(nil), got %d", len(got))
	}
}
