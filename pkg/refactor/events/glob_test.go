package events_test

import (
	"testing"

	events "github.com/direktiv/direktiv/pkg/refactor/events"
	"github.com/google/uuid"
)

func Test_eventPassedGatekeeper(t *testing.T) {
	t.Run("simple", simpleTest)
	t.Run("negative", negativeTest)

}

func simpleTest(t *testing.T) {
	event := newEventWithMeta("mysub", "mytop", uuid.New(), map[string]any{
		"id": "some-id",
	})
	patterns := map[string]string{
		"id": "some-id",
	}
	if !events.EventPassedGatekeeper(patterns, *event) {
		t.Error("Expected event to pass the gatekeeper")
	}
}

func negativeTest(t *testing.T) {
	event := newEventWithMeta("mysub", "mytop", uuid.New(), map[string]any{
		"id": "wrong-id",
	})
	patterns := map[string]string{
		"id": "some-id",
	}
	if events.EventPassedGatekeeper(patterns, *event) {
		t.Error("Expected event to fail the gatekeeper due to mismatch")
	}
}
