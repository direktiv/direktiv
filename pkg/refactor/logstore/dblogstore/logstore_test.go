package dblogstore_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/logstore"
	"github.com/direktiv/direktiv/pkg/refactor/logstore/dblogstore"
	"github.com/direktiv/direktiv/pkg/refactor/utils"
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

func Test_Add(t *testing.T) {
	db, err := utils.NewMockGorm()
	if err != nil {
		t.Fatal(err)
	}
	logstore := dblogstore.NewSQLLogStore(db)
	loadJSONToDB(t, logstore, loopjson)
}

func Test_Get(t *testing.T) {
	db, err := utils.NewMockGorm()
	if err != nil {
		t.Fatal(err)
	}
	logstore := dblogstore.NewSQLLogStore(db)
	o := loadJSONToDB(t, logstore, loopjson)
	logs, err := logstore.Get(context.Background(), "recipientType", "instance", "callpath", "/", "rootInstanceID", "1a0d5909-223f-4f44-86d7-1833ab4d21c8", "isCallerRoot", true)
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
	logs, err := logstore.Get(context.Background(), "recipientType", "instance", "callpath", "/", "rootInstanceID", "c8a13535-b5ae-469b-98f2-e64a831067f9", "isCallerRoot", true)
	if err != nil {
		t.Error(err)
	}
	if len(logs) < 1 {
		t.Error("got no results")
	}
	if len(logs) != len(o) {
		t.Error("some result are missing")
	}
	logs, err = logstore.Get(context.Background(), "recipientType", "instance", "callpath", "/c8a13535-b5ae-469b-98f2-e64a831067f9", "rootInstanceID", "c8a13535-b5ae-469b-98f2-e64a831067f9", "isCallerRoot", false)
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
	if len(logs) < len(want) {
		t.Error("some result are missing")
	}
	logs, err = logstore.Get(context.Background(), "recipientType", "instance", "callpath", "/", "rootInstanceID", "c8a13535-b5ae-469b-98f2-e64a831067f9", "level", "debug", "isCallerRoot", true)
	if err != nil {
		t.Error(err)
	}
	if len(logs) < 1 {
		t.Error("got no results")
	}
	if len(logs) != len(o) {
		t.Error("debug level should filter out any log-entry")
	}
	logs, err = logstore.Get(context.Background(), "recipientType", "instance", "callpath", "/", "rootInstanceID", "c8a13535-b5ae-469b-98f2-e64a831067f9", "level", "info", "isCallerRoot", true)
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
			keyValue = append(keyValue, "rootInstanceID")
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
