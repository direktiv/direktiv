package datastore_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore/datastoresql"
	"github.com/direktiv/direktiv/pkg/refactor/logengine"
	"github.com/google/uuid"
)

func Test_Add_Get(t *testing.T) {
	db, err := database.NewMockGorm()
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
		err := logstore.Append(context.Background(), time.Now(), level, v, fmt.Sprintf("%v", id), in)
		if err != nil {
			t.Error(err)
		}
	}
	q := map[string]interface{}{}
	q[col] = id
	got, err := logstore.Get(context.Background(), -1, -1, fmt.Sprintf("%v", id), q)
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
		levels := []string{"debug", "info", "error"}
		wantLevelValue := levels[level]
		gotLevelValue := fmt.Sprintf("%v", res)
		if wantLevelValue != gotLevelValue {
			t.Errorf("wanted level %s got %s", wantLevelValue, res)
		}
	}
}
