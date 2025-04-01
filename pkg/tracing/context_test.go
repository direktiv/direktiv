package tracing_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/engine"
	"github.com/direktiv/direktiv/pkg/instancestore"
	"github.com/direktiv/direktiv/pkg/tracing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestAddNamespace checks if the namespace tag is added correctly
func TestAddNamespace(t *testing.T) {
	ctx := context.Background()
	ctx = tracing.AddNamespace(ctx, "namespace1")

	tags := tracing.GetCoreAttributes(ctx)
	assert.Equal(t, "namespace1", tags["namespace"], "Expected 'namespace' to be 'namespace1'")
}

// TestAddInstanceAttr tests adding instance attributes to the context using the InstanceAttributes struct
func TestAddInstanceAttr(t *testing.T) {
	ctx := context.Background()
	attrs := tracing.InstanceAttributes{
		Namespace:    "namespace1",
		InstanceID:   "instance1",
		Invoker:      "invoker1",
		Callpath:     "callpath1",
		WorkflowPath: "workflow1",
		Status:       core.LogStatus("ok"),
	}
	ctx = tracing.AddInstanceAttr(ctx, attrs)

	tags := tracing.GetCoreAttributes(ctx)
	assert.Equal(t, "namespace1", tags["namespace"])
	assert.Equal(t, "instance1", tags["instance"])
	assert.Equal(t, "invoker1", tags["invoker"])
	assert.Equal(t, "callpath1", tags["callpath"])
	assert.Equal(t, "workflow1", tags["workflow"])
}

// TestAddInstanceMemoryAttr tests adding instance memory attributes (with state) to the context
func TestAddInstanceMemoryAttr(t *testing.T) {
	ctx := context.Background()
	attrs := tracing.InstanceAttributes{
		Namespace:    "namespace1",
		InstanceID:   "instance1",
		Invoker:      "invoker1",
		Callpath:     "callpath1",
		WorkflowPath: "workflow1",
		Status:       core.LogStatus("ok"),
	}
	state := "state1"
	ctx = tracing.AddInstanceMemoryAttr(ctx, attrs, state)

	tags := tracing.GetCoreAttributes(ctx)
	assert.Equal(t, "namespace1", tags["namespace"])
	assert.Equal(t, "instance1", tags["instance"])
	assert.Equal(t, "invoker1", tags["invoker"])
	assert.Equal(t, "callpath1", tags["callpath"])
	assert.Equal(t, "workflow1", tags["workflow"])
	assert.Equal(t, "state1", tags["state"])
	assert.Equal(t, core.LogStatus("ok"), tags["status"])
}

// TestAddStateAttr checks the behavior of adding state attributes
func TestAddStateAttr(t *testing.T) {
	ctx := context.Background()
	ctx = tracing.AddStateAttr(ctx, "state1")

	tags := tracing.GetCoreAttributes(ctx)
	assert.Equal(t, "state1", tags["state"], "Expected 'state' to be 'state1'")
}

// TestWithTrack checks if the tracking value is set properly
func TestWithTrack(t *testing.T) {
	ctx := context.Background()
	ctx = tracing.WithTrack(ctx, "track1")

	track := ctx.Value(tracing.LogTrackKey).(string)
	assert.Equal(t, "track1", track, "Expected 'track' to be 'track1'")
}

// TestGetRawLogEntryWithStatus ensures the raw log entry contains all expected fields
func TestGetRawLogEntryWithStatus(t *testing.T) {
	ctx := context.Background()
	attrs := tracing.InstanceAttributes{
		Namespace:    "namespace1",
		InstanceID:   "instance1",
		Invoker:      "invoker1",
		Callpath:     "callpath1",
		WorkflowPath: "workflow1",
		Status:       core.LogStatus("ok"),
	}
	ctx = tracing.AddInstanceMemoryAttr(ctx, attrs, "state1")
	ctx = tracing.WithTrack(ctx, "track1")

	logEntry := tracing.GetRawLogEntryWithStatus(ctx, tracing.LevelInfo, "message", core.LogStatus("ok"))
	assert.Equal(t, "namespace1", logEntry["namespace"])
	assert.Equal(t, "INFO", logEntry["level"].(string))
	assert.Equal(t, "message", logEntry["msg"])
	assert.Equal(t, core.LogStatus("ok"), logEntry["status"])
	assert.Equal(t, "track1", logEntry[string(core.LogTrackKey)])
}

// TestBuildNamespaceTrack ensures the namespace track is formatted correctly
func TestBuildNamespaceTrack(t *testing.T) {
	track := tracing.BuildNamespaceTrack("namespace1")
	assert.Equal(t, "namespace.namespace1", track, "Expected track to be 'namespace.namespace1'")
}

// TestBuildInstanceTrack checks instance track building
func TestBuildInstanceTrack(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	id3 := uuid.New()
	instance := &engine.Instance{
		Instance: &instancestore.InstanceData{
			ID: id1,
		},
		DescentInfo: &engine.InstanceDescentInfo{
			Descent: []engine.ParentInfo{
				{ID: id2},
				{ID: id3},
			},
		},
	}
	track := tracing.BuildInstanceTrack(instance)
	expected := fmt.Sprintf("instance.%s/%s/%s", id1, id2, id3)
	assert.Equal(t, expected, track, "Expected track to be '"+expected+"'")
}

// TestBuildInstanceTrackViaCallpath checks instance track via callpath building
func TestBuildInstanceTrackViaCallpath(t *testing.T) {
	track := tracing.BuildInstanceTrackViaCallpath("callpath1")
	assert.Equal(t, "instance.callpath1", track, "Expected track to be 'instance.callpath1'")
}

// TestLogLevelString tests the String method of LogLevel
func TestLogLevelString(t *testing.T) {
	assert.Equal(t, "DEBUG", tracing.LevelDebug.String())
	assert.Equal(t, "INFO", tracing.LevelInfo.String())
	assert.Equal(t, "WARN", tracing.LevelWarn.String())
	assert.Equal(t, "ERROR", tracing.LevelError.String())

	// Testing an invalid log level, should return default "DEBUG"
	var invalidLevel tracing.LogLevel = 999
	assert.Equal(t, "DEBUG", invalidLevel.String(), "Expected 'DEBUG' for invalid log level")
}
