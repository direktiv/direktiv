package dblogstore_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/logstore"
	"github.com/direktiv/direktiv/pkg/refactor/logstore/dblogstore"
	"github.com/direktiv/direktiv/pkg/refactor/utils"
	"github.com/google/uuid"
)

//go:embed mockdata/entlog_loop.json
var loopjson string

//go:embed mockdata/entlog_nestedLoop.json
var loopjsonnested string

type rawlogmsgs struct {
	T     time.Time
	Level string
	Msg   string
	Tags  map[string]string
}

func (r rawlogmsgs) String() string {
	s := ""
	for k, v := range r.Tags {
		s += k + ": " + v + " "
	}

	return fmt.Sprintf("time: %s level: %s msg: %s %s", r.T, r.Level, r.Msg, s)
}

func Test_Add_Get(t *testing.T) {
	db, err := utils.NewMockGorm()
	if err != nil {
		t.Fatal(err)
	}
	logstore := dblogstore.NewSQLLogStore(db)
	col := "namespace-id"
	id := uuid.New()
	wantMsg := fmt.Sprintf("test msg %d", rand.Intn(100)) //nolint:gosec
	add(t, logstore, wantMsg, col, id)
	got, err := logstore.Get(context.Background(), col, id)
	if err != nil {
		t.Error(err)
	}
	if len(got) != 1 {
		t.Error("got wrong number of results")
	}
	if wantMsg != got[0].Msg {
		t.Errorf("error: got %s want %s", got[0], wantMsg)
	}
	col = "namespace-id"
	id = uuid.New()
	addRandomMsgs(t, logstore, col, id)
	col = "workflow-id"
	id = uuid.New()
	addRandomMsgs(t, logstore, col, id)
	col = "root-instance-id"
	id = uuid.New()
	addRandomMsgs(t, logstore, col, id)
	col = "mirror-id"
	id = uuid.New()
	addRandomMsgs(t, logstore, col, id)
}

func addRandomMsgs(t *testing.T, logstore logstore.LogStore, col string, id uuid.UUID) {
	t.Helper()
	want := []string{}
	for i := 0; i < rand.Intn(20); i++ { //nolint:gosec
		want = append(want, fmt.Sprintf("test msg %d", rand.Intn(100))) //nolint:gosec
	}
	for _, v := range want {
		add(t, logstore, v, col, id)
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
	}
}

func add(t *testing.T, logstore logstore.LogStore, wantMsg string, nsName string, nsID uuid.UUID) {
	t.Helper()
	err := logstore.Append(context.Background(), time.Now(), wantMsg, nsName, nsID)
	if err != nil {
		t.Error(err)
	}
}

func Test_Get(t *testing.T) {
	db, err := utils.NewMockGorm()
	if err != nil {
		t.Fatal(err)
	}
	logstore := dblogstore.NewSQLLogStore(db)
	o := loadJSONToDB(t, logstore, loopjson)
	logs, err := logstore.Get(context.Background(), "callpath", "/", "rootInstanceID", "1a0d5909-223f-4f44-86d7-1833ab4d21c8")
	if err != nil {
		t.Error(err)
	}
	if len(logs) < 1 {
		t.Error("got no results")
	}
	if len(logs) != len(o) {
		t.Error("some result are missing")
	}
}

func Test_GetNestedLogs(t *testing.T) {
	db, err := utils.NewMockGorm()
	if err != nil {
		t.Fatal(err)
	}
	logstore := dblogstore.NewSQLLogStore(db)
	o := loadJSONToDB(t, logstore, loopjsonnested)
	logs, err := logstore.Get(context.Background(), "callpath", "/", "root-instance-id", "c8a13535-b5ae-469b-98f2-e64a831067f9")
	if err != nil {
		t.Error(err)
	}
	if len(logs) < 1 {
		t.Error("got no results")
	}
	if len(logs) != len(o) {
		t.Error("some result are missing")
	}
	logs, err = logstore.Get(context.Background(), "callpath", "/c8a13535-b5ae-469b-98f2-e64a831067f9", "root-instance-id", "c8a13535-b5ae-469b-98f2-e64a831067f9")
	if err != nil {
		t.Error(err)
	}
	if len(logs) < 1 {
		t.Error("got no results")
	}
	if len(logs) == len(o) {
		t.Error("no filtering happened")
	}
	want := make([]*rawlogmsgs, 0)
	for _, e := range o {
		if e.Tags["callpath"] == "/c8a13535-b5ae-469b-98f2-e64a831067f9" {
			want = append(want, e)
		}
	}
	for _, le := range logs {
		if le.Fields["callpath"] == "" {
			t.Error("Field was missing")
		}
	}
	if len(logs) < len(want) {
		t.Error("some result are missing")
	}
	logs, err = logstore.Get(context.Background(), "callpath", "/", "root-instance-id", "c8a13535-b5ae-469b-98f2-e64a831067f9", "level", "debug")
	if err != nil {
		t.Error(err)
	}
	if len(logs) < 1 {
		t.Error("got no results")
	}
	if len(logs) != len(o) {
		t.Error("debug level should filter out any log-entry")
	}
	logs, err = logstore.Get(context.Background(), "root-instance-id", "c8a13535-b5ae-469b-98f2-e64a831067f9", "offset", 10, "limit", 1000)
	if err != nil {
		t.Error(err)
	}
	if len(logs) < 1 {
		t.Error("got no results")
	}
	if len(logs) != len(o)-10 {
		t.Error("offset was not applied")
	}
	logs, err = logstore.Get(context.Background(), "callpath", "/", "root-instance-id", "c8a13535-b5ae-469b-98f2-e64a831067f9", "limit", 5)
	if err != nil {
		t.Error(err)
	}
	if len(logs) != 5 {
		t.Error("results should be limited to 5")
	}
	logs, err = logstore.Get(context.Background(), "callpath", "/", "root-instance-id", "c8a13535-b5ae-469b-98f2-e64a831067f9", "level", "info")
	if err != nil {
		t.Error(err)
	}
	if len(logs) < 1 {
		t.Error("got no results")
	}
	want = make([]*rawlogmsgs, 0)
	for _, e := range o {
		if e.Level != "debug" {
			want = append(want, e)
		}
	}
	if len(logs) != len(want) {
		t.Error("some result are missing")
	}
}

func loadJSONToDB(t *testing.T, logstore logstore.LogStore, jsondump string) []*rawlogmsgs {
	t.Helper()
	in := loadJSON(t, jsondump)
	for _, e := range in {
		keyValue := make([]interface{}, 0)
		val, ok := e.Tags["recipientType"]
		if !ok {
			panic("json was has invalid state, recipientType was missing in at least one entry")
		}
		if val == "instance" {
			root := getRoot(e.Tags["callpath"], e.Tags["instance-id"])
			keyValue = append(keyValue, "root-instance-id")
			keyValue = append(keyValue, root)
		}
		keyValue = append(keyValue, "level")
		keyValue = append(keyValue, e.Level)
		for k, v := range e.Tags {
			keyValue = append(keyValue, k)
			keyValue = append(keyValue, v)
		}
		err := logstore.Append(context.Background(), e.T, e.Msg, keyValue...)
		if err != nil {
			t.Errorf("Error occurred on append log entry %v : %v", e, err)
		}
	}

	return in
}

func loadJSON(t *testing.T, jsondump string) []*rawlogmsgs {
	t.Helper()
	logmsgs := make([]*rawlogmsgs, 0)
	err := json.Unmarshal([]byte(jsondump), &logmsgs)
	if err != nil {
		t.Fatal(err)
	}

	return logmsgs
}

func getRoot(callpath, ins string) string {
	if ins == "" {
		panic("json is in invaild state instanceId was missing")
	}
	if callpath == "" {
		panic("json was in invalid state callpath was missing")
	}
	tree := strings.Split(callpath, "/")
	if len(tree) == 0 {
		panic("json was in invalid state callpath must start with /")
	}
	if tree[1] == "" {
		return ins
	}

	return tree[1]
}
