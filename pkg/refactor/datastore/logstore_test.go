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
	id := uuid.New()
	addRandomMsgs(t, logstore, "source", id, logengine.Info)
	q := make(map[string]interface{}, 0)
	q["level"] = logengine.Info
	q["source"] = id
	got, _, err := logstore.Get(context.Background(), q, -1, -1)
	if err != nil {
		t.Error(err)
	}

	if len(got) < 1 {
		t.Error("got no results")
	}
}

func addRandomMsgs(t *testing.T, logstore logengine.LogStore, col string, id uuid.UUID, level logengine.LogLevel) {
	t.Helper()
	want := []string{}
	c := rand.Intn(20) + 1 //nolint:gosec
	for i := 0; i < c; i++ {
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
	got, count, err := logstore.Get(context.Background(), q, -1, -1)
	if err != nil {
		t.Error(err)
	}
	if count != c {
		t.Errorf("got wrong total count Want %v got %v", c, count)
	}
	if len(got) != len(want) {
		t.Error("got wrong number of results.")
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
