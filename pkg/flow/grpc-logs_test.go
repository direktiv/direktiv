package flow

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/flow/internal/testutils"
	"github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

//go:embed mockdata/entlog_loop.json
var loopjson string

//go:embed mockdata/entlog_loopFunctionNested.json
var loopnestedjson string

//go:embed mockdata/entlog_nestedLoop.json
var loopnestedloopjson string

const loopJsonValidInstanceID = "1a0d5909-223f-4f44-86d7-1833ab4d21c8"

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
	res = assertQueryMatchState(t, loopjson, "test", "", "100", 0)
	assertsortedByTime(t, res)
}

func assertQueryMatchState(t *testing.T, jsondump, wf, state, iterator string, resLen int) []*internallogger.LogMsgs {
	t.Helper()
	logmsgs := make([]*internallogger.LogMsgs, 0)
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

func TestWhiteboxTestServerLogs(t *testing.T) {
	srv := server{}
	flowSrv := flow{}

	flowSrv.server = &srv
	logs, logobserver := testutils.ObservedLogger()
	gdb, cleanup, err := testutils.DatabaseGorm()
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err := cleanup()
		if err != nil {
			fmt.Println(err)
		}
	}()
	srv.gormDB = gdb
	srv.sugar = logs
	reqSrvLogs := grpc.ServerLogsRequest{
		Pagination: &grpc.Pagination{},
	}
	resSrvLogs := requestServerLogs(t, flowSrv, &reqSrvLogs)
	if len(logobserver.All()) <= 0 {
		t.Error("some logmsg should heve been printed")
	}
	if int(resSrvLogs.PageInfo.Limit) > len(resSrvLogs.Results) {
		t.Errorf("got more results then specified in pageinfo")
	}
}

func TestFilterByIterrator(t *testing.T) {
	logmsgs := make([]*internallogger.LogMsgs, 0)
	err := json.Unmarshal([]byte(loopnestedjson), &logmsgs)
	if err != nil {
		t.Error(err)
		return
	}
	child := filterByIterrator("2", logmsgs)
	if child == nil {
		t.Errorf("did not found")
	}
	res := filterByIterrator("", logmsgs)
	if len(res) != 0 {
		t.Errorf("calling filterByIterrator with empty iterator is a invalid operation and should result a empty slice")
	}
}

func TestFilterLogmsg(t *testing.T) {
	logmsgs := make([]*internallogger.LogMsgs, 0)
	err := json.Unmarshal([]byte(loopnestedjson), &logmsgs)
	if err != nil {
		t.Error(err)
		return
	}
	res := filterLogmsg(&grpc.PageFilter{
		Field: "TESTFIELD",
		Type:  "TESTTYPE",
	}, logmsgs)
	ok := assertEquals(t, logmsgs, res)
	if !ok {
		t.Error("input slice should not be modified if Pagefilter does not match")
	}
	res = filterLogmsg(&grpc.PageFilter{
		Field: "QUERY",
		Type:  "MATCH",
		Val:   "looperlooper::solve",
	}, logmsgs)
	equals := assertEquals(t, logmsgs, res)
	if equals {
		t.Error("input slice should have been filtered")
	}
}

func requestServerLogs(t *testing.T, flowSrv flow, req *grpc.ServerLogsRequest) *grpc.ServerLogsResponse {
	t.Helper()

	ctx := context.Background()

	res, err := flowSrv.ServerLogs(ctx, req)
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}
	return res
}

func TestBuildInstanceLogResp(t *testing.T) {
	jsondump := loopjson
	ctx := context.Background()

	gdb, cleanup, err := testutils.DatabaseGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	defer func() {
		err := cleanup()
		if err != nil {
			fmt.Println(err)
		}
	}()
	logmsgs := make([]*internallogger.LogMsgs, 0)
	err = json.Unmarshal([]byte(jsondump), &logmsgs)
	if err != nil {
		t.Error(err)
	}
	in := make(map[string]*internallogger.LogMsgs)
	inLen := 0
	id, e := uuid.Parse("1a0d5909-223f-4f44-86d7-1833ab4d21c8")
	if e != nil {
		t.Error(e)
	}
	for _, v := range logmsgs {
		v.InstanceLogs = id
		e, err := storeLogmsg(ctx, gdb, v)
		if err != nil {
			t.Error(err)
		}
		in[e.Msg] = e
		inLen++
	}

	if len(logmsgs) != inLen {
		t.Errorf("Missing Results len was %d should %d", len(logmsgs), inLen)
	}
	resultList := make([]*internallogger.LogMsgs, 0)
	gdb.Raw("SELECT * FROM log_msgs").Scan(&resultList)
	if len(resultList) == 0 {
		t.Error("insert failed")
	}
	if len(resultList) != inLen {
		t.Error("insert failed")
	}
	lq := internallogger.QueryLogs()
	lq.WhereInstance(id)
	page := grpc.Pagination{}
	pi := grpc.PageInfo{}
	ctx = context.Background()
	logs, err := lq.GetAll(ctx, gdb)
	if err != nil {
		t.Errorf("got an err: %v", err)
	}
	if len(logs) == 0 {
		t.Error("got zero logs")
	}
	if _, ok := logs[0].Tags["instance-id"]; !ok {
		t.Error(logs[0].Tags)
	}
	// buildPageInfo()
	res, err := buildInstanceLogResp(ctx, logs, &pi, &page, "ns", loopJsonValidInstanceID)
	if err != nil {
		t.Error(err)
		return
	}
	if res.Instance != loopJsonValidInstanceID {
		t.Errorf("Responded with wrong instance-id got %s want %s", res.Instance, loopJsonValidInstanceID)
	}
	if res.Namespace != "ns" {
		t.Errorf("Responded with wrong namespace got %s want %s", res.Namespace, "ns")
	}
	assertNoMissingLogs(t, res, in)

	res, err = buildInstanceLogResp(ctx, logs, &pi, &page, "ns", loopJsonValidInstanceID)
	if err != nil {
		t.Error(err)
		return
	}
	if res.Instance != loopJsonValidInstanceID {
		t.Errorf("Response had with wrong instance-Id got %s want %s", res.Instance, loopJsonValidInstanceID)
	}
	if res.Namespace != "ns" {
		t.Errorf("Responded with wrong namespace got %s want %s", res.Namespace, "ns")
	}
	assertNoMissingLogs(t, res, in)
	validFilter := &grpc.PageFilter{
		Field: "QUERY",
		Type:  "MATCH",
		Val:   "looperlooper::solve",
	}
	filters := make([]*grpc.PageFilter, 0)
	filters = append(filters, validFilter)
	page.Filter = filters
	res, err = buildInstanceLogResp(ctx, logs, &pi, &page, "ns", loopJsonValidInstanceID)
	if err != nil {
		t.Error(err)
		return
	}
	if res.Instance != loopJsonValidInstanceID {
		t.Errorf("Response had with wrong instance-Id got %s want %s", res.Instance, loopJsonValidInstanceID)
	}
	if res.Namespace != "ns" {
		t.Errorf("Responded with wrong namespace got %s want %s", res.Namespace, "ns")
	}
	if len(logmsgs) <= len(res.GetResults()) {
		t.Error("Logs should be filtered")
	}
}

func assertNoMissingLogs(t *testing.T, res *grpc.InstanceLogsResponse, in map[string]*internallogger.LogMsgs) {
	t.Helper()
	if len(res.Results) == 0 {
		t.Errorf("missing results")
	}
	for _, l := range res.Results {
		original, ok := in[l.Msg]
		if !ok {
			t.Errorf("missing log entry %s", original.Msg)
		}
	}
}

func assertLogmsgsContain(t *testing.T, msgs []*internallogger.LogMsgs, expectedMsg []string) {
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

func assertEquals(t *testing.T, a []*internallogger.LogMsgs, b []*internallogger.LogMsgs) bool {
	t.Helper()
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func assertLogmsgsPresent(t *testing.T, msgs []*internallogger.LogMsgs, expectedMsg []string) {
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

func assertTagsFiltered(t *testing.T, tags map[string]interface{}, tag, value string) {
	t.Helper()
	if tags[tag] != value {
		t.Errorf("found wrong tag-value for %s: %s want %s", tag, tags[tag], value)
	}
}

func assertsortedByTime(t *testing.T, in []*internallogger.LogMsgs) bool {
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

func checkNestedLoop(in []*internallogger.LogMsgs) bool {
	for _, v := range in {
		if v.Tags["state-type"] == "foreach" {
			return true
		}
	}
	return false
}

func storeLogmsg(ctx context.Context, db *gorm.DB, l *internallogger.LogMsgs) (*internallogger.LogMsgs, error) {
	l.Oid = uuid.New()
	db = db.Debug()
	t := db.Table("log_msgs").Create(l)
	if t.Error != nil {
		panic(t.Error)
		// fmt.Println(t.Error)
	}
	return l, nil
}
