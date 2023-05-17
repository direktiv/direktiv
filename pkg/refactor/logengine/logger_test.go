package logengine_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/logengine"
)

var _ logengine.LogStore = &mockstore{}

type mockstore struct {
	logs map[string]*logengine.LogEntry
}

func (m *mockstore) Append(ctx context.Context, timestamp time.Time, level string, msg string, keysAndValues map[string]interface{}) error {
	_ = ctx
	_ = level
	_ = timestamp
	if _, ok := keysAndValues["some_DB_Col"]; !ok {
		return fmt.Errorf("provide some_DB_Col")
	}
	if len(m.logs) == 0 {
		m.logs = make(map[string]*logengine.LogEntry)
	}
	m.logs[fmt.Sprintf("%s", keysAndValues["some_DB_Col"])] = &logengine.LogEntry{T: time.Now(), Msg: msg, Fields: keysAndValues}

	return nil
}

func (m *mockstore) Get(ctx context.Context, keysAndValues map[string]interface{}, limit int, offset int) ([]*logengine.LogEntry, error) {
	_ = ctx
	_ = limit
	_ = offset
	if len(m.logs) == 0 {
		m.logs = make(map[string]*logengine.LogEntry)
	}
	if _, ok := keysAndValues["some_DB_Col"]; !ok {
		return nil, fmt.Errorf("provide some_DB_Col")
	}
	res := make([]*logengine.LogEntry, 0)
	res = append(res, m.logs[fmt.Sprintf("%s", keysAndValues["some_DB_Col"])])

	return res, nil
}

func Test_Log(t *testing.T) {
	ds := logengine.DataStoreBetterLogger{Store: &mockstore{}, LogError: func(template string, args ...interface{}) { t.Errorf(template, args...) }}
	tags := make(map[string]interface{})
	tags["some_DB_Col"] = "value1"
	ds.Log(tags, "error", "test")
	keysNValues := make(map[string]interface{})
	keysNValues["some_DB_Col"] = "value1"
	got, err := ds.Store.Get(context.Background(), keysNValues, -1, -1)
	if err != nil {
		t.Error(err)
	}
	if len(got) == 0 {
		t.Error("got no results")
		t.Fail()
	}
	if got[0].Msg != "test" {
		t.Fail()
	}
}
