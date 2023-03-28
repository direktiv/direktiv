package flow

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/direktiv/direktiv/pkg/flow/ent"
)

//go:embed mockdata/entlog_loop.json
var loopjson string

//go:embed mockdata/entlog_loopFunctionNested.json
var loopnestedjson string

func TestQueryMatchState(t *testing.T) {
	assertQueryMatchState(t, loopjson, "test", "solve", "", 16)
	assertQueryMatchState(t, loopjson, "test", "solve", "1", 6)
	assertQueryMatchState(t, loopnestedjson, "looperlooper", "solve", "", 12)
}

func assertQueryMatchState(t *testing.T, jsondump, wf, state, iterator string, resLen int) {
	t.Helper()
	logmsgs := make([]*ent.LogMsg, 0)
	err := json.Unmarshal([]byte(jsondump), &logmsgs)
	if err != nil {
		t.Error(err)
	}
	res := queryMatchState(wf+"::"+state+"::"+iterator, logmsgs)
	if len(res) != resLen {
		t.Errorf("len off; was %d, want %d", len(res), resLen)
	}
	for _, v := range res {
		assertTagsFiltered(t, v.Tags, "workflow", wf)
		assertTagsFiltered(t, v.Tags, "state-id", state)
		if iterator != "" {
			assertTagsFiltered(t, v.Tags, "loop-index", iterator)
		}
	}
	resSecond := queryMatchState(wf+"::"+state+"::"+iterator, res)
	if len(res) != len(resSecond) {
		t.Errorf("len off; was %d, want %d", len(res), len(resSecond))
	}
}

func TestGetChildByIterator(t *testing.T) {
	logmsgs := make([]*ent.LogMsg, 0)
	err := json.Unmarshal([]byte(loopnestedjson), &logmsgs)
	if err != nil {
		t.Error(err)
		return
	}
	child := queryMatchIterrator("2", logmsgs)
	if child == nil {
		t.Errorf("did not found")
	}
}

func assertTagsFiltered(t *testing.T, tags map[string]string, tag, value string) {
	t.Helper()
	if tags[tag] != value {
		t.Errorf("found wrong tag: %s want %s", tags[tag], value)
	}
}
