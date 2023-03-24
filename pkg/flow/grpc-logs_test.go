package flow

import (
	"context"
	_ "embed"
	"encoding/json"
	"strings"
	"testing"

	"github.com/direktiv/direktiv/pkg/flow/database/entwrapper"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/flow/mock"
)

//go:embed mockdata/entlog_loop.json
var loopjson string

//go:embed mockdata/entlog_loopFunctionNested.json
var loopnestedjson string

//go:embed mockdata/entlog_nestedLoop.json
var loopnestedloopjson string

var expectedLoopnestedjsonWFLooperlooperSTATESolveIt2 = []string{
	"Preparing workflow triggered by",
	"Sleeping until",
	"Starting workflow",
	"Running state logic",
	"Creating cookie",
	"Creating new",
	"Creating URL",
	"Sending request",
	"Running state logic",
	"returned",
	"returned",
}

func TestQueryMatchState(t *testing.T) {
	res := assertQueryMatchState(t, loopjson, "test", "", "", 21)
	assertsortedByTime(t, res)
	res = assertQueryMatchState(t, loopjson, "test", "solve", "", 16)
	assertsortedByTime(t, res)
	res = assertQueryMatchState(t, loopjson, "test", "solve", "1", 6)
	assertsortedByTime(t, res)
	res = assertQueryMatchState(t, loopnestedjson, "looperlooper", "solve", "", 12)
	assertsortedByTime(t, res)
	res = assertQueryMatchState(t, loopnestedjson, "looper", "", "3", 8)
	assertsortedByTime(t, res)
	res = assertQueryMatchState(t, loopnestedjson, "looper", "", "3", 8)
	assertsortedByTime(t, res)
	res = assertQueryMatchState(t, loopnestedloopjson, "looperlooper", "solve", "0", 51)
	assertsortedByTime(t, res)
	res = assertQueryMatchState(t, loopnestedjson, "looperlooper", "solve", "0", 10)
	assertLogmsgsContain(t, res, expectedLoopnestedjsonWFLooperlooperSTATESolveIt2)
	assertLogmsgsPresent(t, res, expectedLoopnestedjsonWFLooperlooperSTATESolveIt2)
	res = assertQueryMatchState(t, loopnestedloopjson, "looper", "solve", "0", 13)
	assertsortedByTime(t, res)
}

func assertQueryMatchState(t *testing.T, jsondump, wf, state, iterator string, resLen int) []*ent.LogMsg {
	t.Helper()
	logmsgs := make([]*ent.LogMsg, 0)
	err := json.Unmarshal([]byte(jsondump), &logmsgs)
	if err != nil {
		t.Error(err)
	}
	res := filterMatchByWfStateIterator(wf+"::"+state+"::"+iterator, logmsgs)
	if len(res) != resLen {
		t.Errorf("len off; was %d, want %d", len(res), resLen)
	}
	for _, v := range res {
		nestedloopPresent := checkNestedLoop(res)
		if iterator == "" && state != "" {
			assertTagsFiltered(t, v.Tags, "workflow", wf)
			assertTagsFiltered(t, v.Tags, "state-id", state)
		}
		if iterator == "" && state == "" {
			assertTagsFiltered(t, v.Tags, "workflow", wf)
		}
		if !nestedloopPresent && iterator != "" {
			assertTagsFiltered(t, v.Tags, "loop-index", iterator)
		}
	}
	resSecond := filterMatchByWfStateIterator(wf+"::"+state+"::"+iterator, res)
	if len(res) != len(resSecond) {
		t.Errorf("len off when runned second time; was first run %d, is %d, should %d", len(res), len(resSecond), resLen)
	}
	return res
}

func TestGetChildByIterator(t *testing.T) {
	logmsgs := make([]*ent.LogMsg, 0)
	err := json.Unmarshal([]byte(loopnestedjson), &logmsgs)
	if err != nil {
		t.Error(err)
		return
	}
	child := filterByIterrator("2", logmsgs)
	if child == nil {
		t.Errorf("did not found")
	}
}

func TestBuildInstanceLogResp(t *testing.T) {
	jsondump := loopjson
	ctx := context.Background()

	db, err := mock.DatabaseWrapper()
	if err != nil {
		t.Error(err)
		return
	}
	defer db.StopDB()
	logmsgs := make([]*ent.LogMsg, 0)
	err = json.Unmarshal([]byte(jsondump), &logmsgs)
	if err != nil {
		t.Error(err)
	}
	in := make(map[string]*ent.LogMsg)
	inLen := 0
	for _, v := range logmsgs {
		e, err := storeLogmsg(ctx, &db.Entw, v)
		if err != nil {
			t.Error(err)
		}
		in[e.Msg] = e
		inLen++
	}
	query := buildInstanceLogsQuery(ctx, &db.Entw, "", "", false)
	page := grpc.Pagination{}

	logmsgs, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, &page, query, logsOrderings, logsFilters)
	if err != nil {
		t.Error(err)
		return
	}
	if len(logmsgs) != inLen {
		t.Errorf("Missing Results len was %d should %d", len(logmsgs), inLen)
	}
	res, err := buildInstanceLogResp(ctx, logmsgs, pi, &page, "ns", "1a0d5909-223f-4f44-86d7-1833ab4d21c8")
	if err != nil {
		t.Error(err)
		return
	}
	if res.Instance != "1a0d5909-223f-4f44-86d7-1833ab4d21c8" {
		t.Errorf("Responde with wrong instid got %s want %s", res.Instance, "1a0d5909-223f-4f44-86d7-1833ab4d21c8")
	}
	if res.Namespace != "ns" {
		t.Errorf("Responde with wrong namespace got %s want %s", res.Namespace, "ns")
	}
	assertNoMissingLogs(t, res, in)
	res, err = buildInstanceLogResp(ctx, logmsgs, pi, &page, "ns", "1a0d5909-223f-4f44-86d7-1833ab4d21c8")
	if err != nil {
		t.Error(err)
		return
	}
	if res.Instance != "1a0d5909-223f-4f44-86d7-1833ab4d21c8" {
		t.Errorf("Response had with wrong instanceId got %s want %s", res.Instance, "1a0d5909-223f-4f44-86d7-1833ab4d21c8")
	}
	if res.Namespace != "ns" {
		t.Errorf("Responde with wrong namespace got %s want %s", res.Namespace, "ns")
	}
	assertNoMissingLogs(t, res, in)
}

func assertNoMissingLogs(t *testing.T, res *grpc.InstanceLogsResponse, in map[string]*ent.LogMsg) {
	t.Helper()
	if len(res.Results) == 0 {
		t.Errorf("missing results")
	}
	for _, l := range res.Results {
		original, ok := in[l.Msg]
		if !ok {
			t.Errorf("missing log entry %s", original)
		}
	}
}

func assertLogmsgsContain(t *testing.T, msgs []*ent.LogMsg, expectedMsg []string) {
	t.Helper()
	for _, v := range msgs {
		ok := false
		for _, e := range expectedMsg {
			ok = ok || strings.Contains(v.Msg, e)
		}
		if !ok {
			t.Errorf("logmsg %s was not expected", v.Msg)
		}
	}
}

func assertLogmsgsPresent(t *testing.T, msgs []*ent.LogMsg, expectedMsg []string) {
	t.Helper()

	for _, e := range expectedMsg {
		ok := false
		for _, v := range msgs {
			ok = ok || strings.Contains(v.Msg, e)
		}
		if !ok {
			t.Errorf("logmsg %s was missing", e)
		}
	}
}

func assertTagsFiltered(t *testing.T, tags map[string]string, tag, value string) {
	t.Helper()
	if tags[tag] != value {
		t.Errorf("found wrong tag-value for %s: %s want %s", tag, tags[tag], value)
	}
}

func assertsortedByTime(t *testing.T, in []*ent.LogMsg) bool {
	t.Helper()
	if len(in) < 2 {
		return true
	}

	for i := 2; i < len(in); i = i + 2 {
		a := in[i]
		b := in[i]
		if i+1 < len(in) {
			b = in[i+1]
		}
		if a.T.After(b.T) {
			t.Errorf("Array not sorted")
			return false
		}
	}
	return true
}

func checkNestedLoop(in []*ent.LogMsg) bool {
	for _, v := range in {
		if v.Tags["state-type"] == "foreach" {
			return true
		}
	}
	return false
}

func storeLogmsg(ctx context.Context, entw *entwrapper.Database, l *ent.LogMsg) (*ent.LogMsg, error) {
	clients := entw.Clients(ctx)
	return clients.LogMsg.Create().SetMsg(l.Msg).SetT(l.T).SetLevel(l.Level).SetTags(l.Tags).SetRootInstanceId(l.RootInstanceId).SetLogInstanceCallPath(l.LogInstanceCallPath).Save(ctx)
}
