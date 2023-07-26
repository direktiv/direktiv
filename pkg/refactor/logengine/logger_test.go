package logengine_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/logengine"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func Test_ChachedSQLLogStore(t *testing.T) {
	dbMock := make(chan logengine.LogEntry, 1)
	logger, logWorker, closeLogWorkers := logengine.NewCachedLogger(1,
		func(
			ctx context.Context,
			timestamp time.Time,
			level logengine.LogLevel,
			msg string,
			keysAndValues map[string]interface{},
		) error {
			tagsCopy := map[string]interface{}{}
			for k, v := range keysAndValues {
				tagsCopy[k] = v
			}
			tagsCopy["level"] = level
			l := logengine.LogEntry{
				T:      time.Now(),
				Msg:    msg,
				Fields: tagsCopy,
			}
			dbMock <- l

			return nil
		},
		func(objectID uuid.UUID, objectType string) {},
		func(template string, args ...interface{}) {
			t.Errorf(template, args...)
		})

	go logWorker()
	defer closeLogWorkers()
	tags := map[string]string{}
	tags["recipientType"] = "server"
	tagsCopy := map[string]string{}
	for k, v := range tags {
		tagsCopy[k] = v
	}
	// simple test
	source := uuid.New()
	logger.Debugf(context.Background(), source, tagsCopy, "test msg")
	logs, err := waitForTrigger(t, dbMock)
	if err != nil {
		t.Error("expected to get logs but got none")
	}
	if logs.Msg != "test msg" {
		t.Error("got wrong log entry")
	}
	if logs.T.After(time.Now()) {
		t.Error("some thing is broken with the timestamp")
	}
	gotSource, ok := logs.Fields["source"]
	if !ok {
		t.Error("better logger was expected to add the source to the log tags")
	}
	if gotSource != source && gotSource != source.String() {
		t.Errorf("want source id %v got %v", source, gotSource)
	}
	source = uuid.New()
	// smart instance logs test
	tags = map[string]string{}
	tags["recipientType"] = "instance"
	tags["callpath"] = "/"
	tags["instance-id"] = source.String()

	tagsCopy = map[string]string{}
	for k, v := range tags {
		tagsCopy[k] = v
	}
	logger.Debugf(context.Background(), source, tagsCopy, "test msg2")
	logs, err = waitForTrigger(t, dbMock)
	if err != nil {
		t.Error("expected to get logs but got none")
	}
	if logs.Msg != "test msg2" {
		t.Error("got wrong log entry")
	}
	if logs.T.After(time.Now()) {
		t.Error("some thing is broken with the timestamp")
	}
	gotSource, ok = logs.Fields["source"]
	if !ok {
		t.Error("better logger was expected to add the source to the log tags")
	}
	if gotSource != source && gotSource != source.String() {
		t.Errorf("want source id %v got %v", source, gotSource)
	}
	gotCallpath, ok := logs.Fields["callpath"]
	if !ok {
		t.Error("better logger was expected to add the source to the log tags")
	}
	callpath := "/" + source.String()
	if gotCallpath != callpath {
		t.Errorf("want callpath %v got %v", callpath, gotCallpath)
	}
}

func Test_TracingLogger(t *testing.T) {
	traceTagsMock := false

	logger := logengine.SugarBetterJSONLogger{
		Sugar: zap.S(),
		AddTraceFrom: func(ctx context.Context, toTags map[string]string) map[string]string {
			traceTagsMock = true

			return toTags
		},
	}
	tags := map[string]string{}
	tags["recipientType"] = "server"
	tagsCopy := map[string]string{}
	for k, v := range tags {
		tagsCopy[k] = v
	}
	// simple test
	source := uuid.New()
	logger.Debugf(context.Background(), source, tagsCopy, "test msg")
	if !traceTagsMock {
		t.Error("mock was not called")
	}
}

func waitForTrigger(t *testing.T, c chan logengine.LogEntry) (*logengine.LogEntry, error) {
	t.Helper()
	var count int
	for {
		select {
		case startedAction := <-c:
			return &startedAction, nil
		default:
			if count > 3 {
				return nil, fmt.Errorf("timeout")
			}
			time.Sleep(1 * time.Millisecond)
			count++
		}
	}
}
