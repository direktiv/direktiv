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

// Helper function to retrieve tags from context
func getTagsFromContext(ctx context.Context) map[string]interface{} {
	return ctx.Value(core.LogTagsKey).(map[string]interface{})
}

// TestAddTag checks the behavior of the AddTag function
func TestAddTag(t *testing.T) {
	ctx := context.Background()
	ctx = tracing.AddTag(ctx, "key", "value")

	tags := getTagsFromContext(ctx)
	assert.Equal(t, "value", tags["key"], "Expected 'key' to have value 'value'")
}

// TestAddNamespace checks if the namespace tag is added correctly
func TestAddNamespace(t *testing.T) {
	ctx := context.Background()
	ctx = tracing.AddNamespace(ctx, "namespace1")

	tags := getTagsFromContext(ctx)
	assert.Equal(t, "namespace1", tags["namespace"], "Expected 'namespace' to be 'namespace1'")
}

// TestAddInstanceAttr tests adding instance attributes to the context
func TestAddInstanceAttr(t *testing.T) {
	ctx := context.Background()
	ctx = tracing.AddInstanceAttr(ctx, "instance1", "invoker1", "callpath1", "workflow1")

	tags := getTagsFromContext(ctx)
	assert.Equal(t, "instance1", tags["instance"])
	assert.Equal(t, "invoker1", tags["invoker"])
	assert.Equal(t, "callpath1", tags["callpath"])
	assert.Equal(t, "workflow1", tags["workflow"])
}

// TestAddStateAttr checks the behavior of adding state attributes
func TestAddStateAttr(t *testing.T) {
	ctx := context.Background()
	ctx = tracing.AddStateAttr(ctx, "state1")

	tags := getTagsFromContext(ctx)
	assert.Equal(t, "state1", tags["state"], "Expected 'state' to be 'state1'")
}

// TestWithTrack checks if the tracking value is set properly
func TestWithTrack(t *testing.T) {
	ctx := context.Background()
	ctx = tracing.WithTrack(ctx, "track1")

	track := ctx.Value(core.LogTrackKey).(string)
	assert.Equal(t, "track1", track, "Expected 'track' to be 'track1'")
}

// TestGetRawLogEntryWithStatus ensures the raw log entry contains all expected fields
func TestGetRawLogEntryWithStatus(t *testing.T) {
	ctx := context.Background()
	ctx = tracing.AddTag(ctx, "key", "value")
	ctx = tracing.WithTrack(ctx, "track1")

	logEntry := tracing.GetRawLogEntryWithStatus(ctx, tracing.LevelInfo, "message", core.LogStatus("ok"))
	assert.Equal(t, "value", logEntry["key"])
	assert.Equal(t, "INFO", logEntry["level"].(tracing.LogLevel).String())
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

// TestAddTagWithEmptyContext tests adding a tag to an empty context
func TestAddTagWithEmptyContext(t *testing.T) {
	ctx := context.Background()

	// Add a tag and ensure it was added
	ctx = tracing.AddTag(ctx, "newKey", "newValue")
	tags := getTagsFromContext(ctx)
	assert.Equal(t, "newValue", tags["newKey"], "Expected 'newKey' to be added to the context")
}

// TestAddInstanceAttrWithEmptyContext tests adding instance attributes with an empty context
func TestAddInstanceAttrWithEmptyContext(t *testing.T) {
	ctx := context.Background()
	ctx = tracing.AddInstanceAttr(ctx, "instance1", "invoker1", "callpath1", "workflow1")

	// Ensure the attributes were added correctly
	tags := getTagsFromContext(ctx)
	assert.Equal(t, "instance1", tags["instance"])
	assert.Equal(t, "invoker1", tags["invoker"])
	assert.Equal(t, "callpath1", tags["callpath"])
	assert.Equal(t, "workflow1", tags["workflow"])
}

// TestGetAttributesWithSpan simulates a context with a trace span and retrieves attributes
func TestGetAttributesWithSpan(t *testing.T) {
	// Simulate a trace span context
	ctx := context.Background()
	end := initTestWithMockTelemetry()
	defer end()
	ctx, end, err := tracing.NewSpan(ctx, "test")
	assert.NoError(t, err)
	// Retrieve attributes from context
	attributes := tracing.GetRawLogEntryWithStatus(ctx, tracing.LevelInfo, "Test message", core.LogStatus("ok"))
	assert.NotNil(t, attributes["trace"], "Expected 'trace' attribute in the log entry")
	assert.NotNil(t, attributes["span"], "Expected 'span' attribute in the log entry")
	end()
}

// TestGetRawLogEntryWithEmptyContext tests GetRawLogEntryWithStatus with an empty context
func TestGetRawLogEntryWithEmptyContext(t *testing.T) {
	ctx := context.Background()
	logEntry := tracing.GetRawLogEntryWithStatus(ctx, tracing.LevelWarn, "empty context", core.LogStatus("warning"))

	assert.Equal(t, "WARN", logEntry["level"].(tracing.LogLevel).String(), "Expected log level to be 'WARN'")
	assert.Equal(t, "empty context", logEntry["msg"], "Expected message to be 'empty context'")
	assert.Equal(t, core.LogStatus("warning"), logEntry["status"], "Expected status to be 'warning'")
}

// TestBuildInstanceTrackWithEmptyDescent checks building an instance track when no descendants exist
func TestBuildInstanceTrackWithEmptyDescent(t *testing.T) {
	id1 := uuid.New()
	instance := &engine.Instance{
		Instance: &instancestore.InstanceData{
			ID: id1,
		},
		DescentInfo: &engine.InstanceDescentInfo{
			Descent: []engine.ParentInfo{}, // No descendants
		},
	}

	track := tracing.BuildInstanceTrack(instance)
	expected := fmt.Sprintf("instance.%s", id1)
	assert.Equal(t, expected, track, "Expected track to be '"+expected+"' when there are no descendants")
}

// TestAddStateAttrWithEmptyContext tests adding state attributes with an empty context
func TestAddStateAttrWithEmptyContext(t *testing.T) {
	ctx := context.Background()
	ctx = tracing.AddStateAttr(ctx, "state1")

	tags := getTagsFromContext(ctx)
	assert.Equal(t, "state1", tags["state"], "Expected 'state' to be 'state1'")
}

// TestBuildInstanceTrackWithNilDescent checks building an instance track when descent info is nil
func TestBuildInstanceTrackWithNilDescent(t *testing.T) {
	id1 := uuid.New()
	instance := &engine.Instance{
		Instance: &instancestore.InstanceData{
			ID: id1,
		},
		DescentInfo: nil, // Nil descent info
	}

	track := tracing.BuildInstanceTrack(instance)
	expected := fmt.Sprintf("instance.%s", id1)
	assert.Equal(t, expected, track, "Expected track to be '"+expected+"' when descent info is nil")
}

// TestBuildInstanceTrackViaCallpathEmpty tests building instance track via empty callpath
func TestBuildInstanceTrackViaCallpathEmpty(t *testing.T) {
	track := tracing.BuildInstanceTrackViaCallpath("")
	assert.Equal(t, "instance.", track, "Expected track to be 'instance.' when callpath is empty")
}
