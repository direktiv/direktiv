package dblogstore_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/logstore/dblogstore"
	"github.com/direktiv/direktiv/pkg/refactor/utils"
)

//go:embed mockdata/entlog_loop.json
var loopjson string

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
	in := loadJSON(t, loopjson)
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
