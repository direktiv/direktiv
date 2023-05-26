package logengine_test

import (
	"context"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/datastore/datastoresql"
	"github.com/direktiv/direktiv/pkg/refactor/logengine"
	"github.com/direktiv/direktiv/pkg/refactor/utils"
	"github.com/google/uuid"
)

func Test_Log(t *testing.T) {
	db, err := utils.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	sq := datastoresql.NewSQLStore(db, "some_secret_key_")
	logstore := sq.Logs()

	ds := logengine.DataStoreBetterLogger{Store: logstore, LogError: func(template string, args ...interface{}) { t.Errorf(template, args...) }}
	tags := make(map[string]interface{})
	id := uuid.New()
	tags["workflow_id"] = id
	ds.Errorf(context.Background(), id, tags, "test %s", "msg")
	keysNValues := make(map[string]interface{})
	keysNValues["workflow_id"] = id
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
