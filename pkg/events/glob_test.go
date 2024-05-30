package events_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/events"
	"github.com/google/uuid"
)

func TestEventPassedGatekeeper(t *testing.T) {
	t.Run("SimpleEventPasses", simpleEventPasses)
	t.Run("EventFailsDueToMismatch", eventFailsDueToMismatch)

	t.Run("EventFailsWithMultipleConditions", eventFailsWithMultipleConditions)
	t.Run("EventPassesWithMultipleContexts", eventPassesWithMultipleContexts)
	t.Run("EventPassesWithMultipleConditions", eventPassesWithMultipleContexts)
}

func simpleEventPasses(t *testing.T) {
	event := newEventWithMeta("mysub", "mytop", uuid.New(), map[string]interface{}{
		"id": "some-id",
	})
	patterns := map[string]string{
		"id": "some-id",
	}
	if !events.EventPassedGatekeeper(patterns, *event) {
		t.Error("Expected event to pass the gatekeeper")
	}
}

func eventPassesWithMultipleConditions(t *testing.T) {
	event := newEventWithMeta("mysub", "mytop", uuid.New(), map[string]interface{}{
		"id":  "some-id",
		"id2": "some-other-id",
	})
	patterns := map[string]string{
		"id":  "some-id",
		"id2": "some-other-id",
	}
	if !events.EventPassedGatekeeper(patterns, *event) {
		t.Error("Expected event to pass the gatekeeper")
	}
}

func eventFailsWithMultipleConditions(t *testing.T) {
	event := newEventWithMeta("mysub", "mytop", uuid.New(), map[string]interface{}{
		"id": "some-id",
	})
	patterns := map[string]string{
		"id":  "some-id",
		"id2": "some-other-id",
	}
	if events.EventPassedGatekeeper(patterns, *event) {
		t.Error("Expected event to fail the gatekeeper due to mismatch")
	}
}

func eventPassesWithMultipleContexts(t *testing.T) {
	event := newEventWithMeta("mysub", "mytop", uuid.New(), map[string]interface{}{
		"id":  "some-id",
		"id2": "some-other-id",
	})
	patterns := map[string]string{
		"id": "some-id",
	}
	if !events.EventPassedGatekeeper(patterns, *event) {
		t.Error("Expected event to pass the gatekeeper")
	}
}

func eventFailsDueToMismatch(t *testing.T) {
	event := newEventWithMeta("mysub", "mytop", uuid.New(), map[string]interface{}{
		"id": "wrong-id",
	})
	patterns := map[string]string{
		"id": "some-id",
	}
	if events.EventPassedGatekeeper(patterns, *event) {
		t.Error("Expected event to fail the gatekeeper due to mismatch")
	}
}
