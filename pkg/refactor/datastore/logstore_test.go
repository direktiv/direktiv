package datastore_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/datastore/datastoresql"
	"github.com/direktiv/direktiv/pkg/refactor/logengine"
	"github.com/direktiv/direktiv/pkg/refactor/utils"
	"github.com/google/uuid"
)

func Test_Add_Get(t *testing.T) {
	db, err := utils.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	ds := datastoresql.NewSQLStore(db, "some_secret_key_")
	logstore := ds.Logs()

	addRandomMsgs(t, logstore, "namespace_logs", uuid.New(), logengine.Debug)
	addRandomMsgs(t, logstore, "workflow_id", uuid.New(), logengine.Debug)
	addRandomMsgs(t, logstore, "root_instance_id", uuid.New(), logengine.Debug)
	addRandomMsgs(t, logstore, "mirror_activity_id", uuid.New(), logengine.Debug)
	addRandomMsgs(t, logstore, "mirror_activity_id", uuid.New(), logengine.Error)
	addRandomMsgs(t, logstore, "mirror_activity_id", uuid.New(), logengine.Error)
	addRandomMsgs(t, logstore, "mirror_activity_id", uuid.New(), logengine.Info)
	addRandomMsgs(t, logstore, "mirror_activity_id", uuid.New(), logengine.Debug)
	q := make(map[string]interface{}, 0)
	q["level"] = logengine.Info
	got, err := logstore.Get(context.Background(), q, -1, -1)
	if err != nil {
		t.Error(err)
	}
	foundInfoMsg := false
	foundErrorMsg := false

	if len(got) < 1 {
		t.Error("got no results")
	}
	for _, le := range got {
		if le.Fields["level"] == logengine.Debug {
			t.Errorf("query for info level should not contain debug msgs")
		}
		if le.Fields["level"] == logengine.Info {
			foundInfoMsg = true
		}
		if le.Fields["level"] == logengine.Error {
			foundErrorMsg = true
		}
	}
	if !foundInfoMsg {
		t.Errorf("query for info level should contain info msgs")
	}
	if !foundErrorMsg {
		t.Errorf("query for info level should contain error msgs")
	}
}

func addRandomMsgs(t *testing.T, logstore logengine.LogStore, col string, id uuid.UUID, level logengine.LogLevel) {
	t.Helper()
	want := []string{}
	for i := 0; i < rand.Intn(20)+1; i++ { //nolint:gosec
		want = append(want, fmt.Sprintf("test msg %d", rand.Intn(100)+1)) //nolint:gosec
	}
	in := map[string]interface{}{}
	in[col] = id
	for _, v := range want {
		err := logstore.Append(context.Background(), time.Now(), level, v, in)
		if err != nil {
			t.Error(err)
		}
	}
	q := map[string]interface{}{}
	q[col] = id
	got, err := logstore.Get(context.Background(), q, -1, -1)
	if err != nil {
		t.Error(err)
	}
	if len(got) != len(want) {
		t.Error("got wrong number of results")
	}
	for _, le := range got {
		ok := false
		for _, v := range want {
			ok = ok || v == le.Msg
		}
		if !ok {
			t.Errorf("log entry is not found %s", le.Msg)
		}
		res, ok := le.Fields["level"]
		if !ok {
			t.Error("missing level value")
		}
		v, ok := res.(int)
		if !ok {
			t.Errorf("got wrong type for level")
			t.Fail()
		}
		wantLevelValue := fmt.Sprintf("%d", level)
		gotLevelValue := fmt.Sprintf("%d", v)
		if wantLevelValue != gotLevelValue {
			t.Errorf("wanted level %d got %s", level, res)
		}
	}
}
