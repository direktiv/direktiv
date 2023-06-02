package logengine_test

import (
	"context"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore/datastoresql"
	"github.com/direktiv/direktiv/pkg/refactor/logengine"
	"github.com/google/uuid"
)

func Test_Log(t *testing.T) {
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	sq := datastoresql.NewSQLStore(db, "some_secret_key_")
	logstore := sq.Logs()

	ds := logengine.DataStoreBetterLogger{Store: logstore, LogError: func(template string, args ...interface{}) { t.Errorf(template, args...) }}
	tags := make(map[string]interface{})
	id := uuid.New()
	tags["workflow_id"] = id
	tags["namespace"] = "someNsName"
	tags["workflow"] = "someWfName"
	ds.Errorf(context.Background(), id, tags, "test %s", "msg")
	keysNValues := make(map[string]interface{})
	got, err := ds.Store.Get(context.Background(), -1, -1, "someNsName/someWfName", keysNValues)
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
