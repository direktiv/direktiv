package logengine_test

import (
	"context"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/datastore/sql"
	"github.com/direktiv/direktiv/pkg/refactor/logengine"
	"github.com/direktiv/direktiv/pkg/refactor/utils"
)

func Test_Log(t *testing.T) {
	db, err := utils.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	sq := sql.NewSQLStore(db, "some_secret_key_")
	logstore := sq.Logs()

	ds := logengine.DataStoreBetterLogger{Store: logstore, LogError: func(template string, args ...interface{}) { t.Errorf(template, args...) }}
	tags := make(map[string]interface{})
	tags["workflow_id"] = "some-id"
	ds.Log(tags, "error", "test %s", "msg")
	keysNValues := make(map[string]interface{})
	keysNValues["workflow_id"] = "some-id"
	got, err := ds.Store.Get(context.Background(), keysNValues, -1, -1)
	if err != nil {
		t.Error(err)
	}
	if len(got) == 0 {
		t.Error("got no results")
		t.Fail()
	}
	if got[0].Msg != "test msg" {
		t.Fail()
	}
}
