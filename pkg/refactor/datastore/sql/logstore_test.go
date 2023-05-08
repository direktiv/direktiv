package sql_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/datastore/sql"
	"github.com/direktiv/direktiv/pkg/refactor/logengine"
	"github.com/direktiv/direktiv/pkg/refactor/utils"
	"github.com/google/uuid"
)

func Test_Add_Get(t *testing.T) {
	db, err := utils.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	ds := sql.NewSQLStore(db, "some_secret_key_")
	logstore := ds.Logs()
	// basic
	addRandomMsgs(t, logstore, "namespace-id", uuid.New(), "")
	addRandomMsgs(t, logstore, "workflow-id", uuid.New(), "")
	addRandomMsgs(t, logstore, "root-instance-id", uuid.New(), "")
	addRandomMsgs(t, logstore, "mirror-id", uuid.New(), "")
	addRandomMsgs(t, logstore, "mirror-id", uuid.New(), "panic")
	addRandomMsgs(t, logstore, "mirror-id", uuid.New(), "error")
	addRandomMsgs(t, logstore, "mirror-id", uuid.New(), "info")
	addRandomMsgs(t, logstore, "mirror-id", uuid.New(), "debug")
	got, err := logstore.Get(context.Background(), "level", "info")
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

func add(t *testing.T, logstore logengine.LogStore, wantMsg string, col string, id uuid.UUID, keyValues ...interface{}) {
	t.Helper()
	in := make([]interface{}, 0)
	in = append(in, col, id)
	in = append(in, keyValues...)
	err := logstore.Append(context.Background(), time.Now(), wantMsg, in...)
	if err != nil {
		t.Error(err)
	}
}

func addRandomMsgs(t *testing.T, logstore logengine.LogStore, col string, id uuid.UUID, level string) {
	t.Helper()
	want := []string{}
	for i := 0; i < rand.Intn(20)+1; i++ { //nolint:gosec
		want = append(want, fmt.Sprintf("test msg %d", rand.Intn(100)+1)) //nolint:gosec
	}
	for _, v := range want {
		if level != "" {
			add(t, logstore, v, col, id, "level", level)
		} else {
			add(t, logstore, v, col, id)
		}
	}
	got, err := logstore.Get(context.Background(), col, id)
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
