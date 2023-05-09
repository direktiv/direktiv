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

	addRandomMsgs(t, logstore, "namespace_logs", uuid.New(), "")
	addRandomMsgs(t, logstore, "workflow_id", uuid.New(), "")
	addRandomMsgs(t, logstore, "root_instance_id", uuid.New(), "")
	addRandomMsgs(t, logstore, "mirror_activity_id", uuid.New(), "")
	addRandomMsgs(t, logstore, "mirror_activity_id", uuid.New(), "panic")
	addRandomMsgs(t, logstore, "mirror_activity_id", uuid.New(), "error")
	addRandomMsgs(t, logstore, "mirror_activity_id", uuid.New(), "info")
	addRandomMsgs(t, logstore, "mirror_activity_id", uuid.New(), "debug")
	q := make(map[string]interface{}, 0)
	q["level"] = "info"
	got, err := logstore.Get(context.Background(), q, -1, -1)
	if err != nil {
		t.Error(err)
	}
	foundInfoMsg := false
	foundErrorMsg := false
	foundPanicMsg := false

	if len(got) < 1 {
		t.Error("got no results")
	}
	for _, le := range got {
		if le.Fields["level"] == "debug" {
			t.Errorf("query for info level should not contain debug msgs")
		}
		if le.Fields["level"] == "info" {
			foundInfoMsg = true
		}
		if le.Fields["level"] == "error" {
			foundErrorMsg = true
		}
		if le.Fields["level"] == "panic" {
			foundPanicMsg = true
		}
	}
	if !foundInfoMsg {
		t.Errorf("query for info level should contain info msgs")
	}
	if !foundErrorMsg {
		t.Errorf("query for info level should contain error msgs")
	}
	if !foundPanicMsg {
		t.Errorf("query for info level should contain panic msgs")
	}
}

func addRandomMsgs(t *testing.T, logstore logengine.LogStore, col string, id uuid.UUID, level string) {
	t.Helper()
	want := []string{}
	for i := 0; i < rand.Intn(20)+1; i++ { //nolint:gosec
		want = append(want, fmt.Sprintf("test msg %d", rand.Intn(100)+1)) //nolint:gosec
	}
	in := map[string]interface{}{}
	in[col] = id
	for _, v := range want {
		if level != "" {
			in["level"] = level
		}
		err := logstore.Append(context.Background(), time.Now(), v, in)
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
		if level != "" {
			res, ok := le.Fields["level"]
			if !ok {
				t.Error("missing level value")
			}
			if res != level {
				t.Errorf("wanted level %s got %s", level, res)
			}
		}
	}
}
