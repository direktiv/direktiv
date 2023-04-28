package internallogger

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/internallogger/logstore"
	logquerybuilder "github.com/direktiv/direktiv/pkg/refactor/internallogger/logstore/log-querybuilder"
	"github.com/direktiv/direktiv/pkg/refactor/utils"
	"github.com/google/uuid"
)

//go:embed mockdata/entlog_loop.json
var loopjson string

//go:embed mockdata/entlog_loopFunctionNested.json
var loopnestedjson string

//go:embed mockdata/entlog_nestedLoop.json
var loopnestedloopjson string

var _notifyLogsTriggeredWith notifyLogsTriggeredWith

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

func TestStoringLogMsg(t *testing.T) {
	sugar, telemetrylogs := utils.ObservedLogger()
	logger := InitLogger()
	gorm, cleanup, err := utils.DatabaseGorm()
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err := cleanup()
		if err != nil {
			fmt.Println(err)
		}
	}()
	logger.StartLogWorkers(1, gorm, &LogNotifyMock{}, sugar)
	tags := make(map[string]string)
	tags["recipientType"] = recipient.Server.String()
	tags["testTag"] = recipient.Server.String()
	recipent := uuid.New()
	msg := "test2"
	ctx := context.TODO()
	logger.Infof(ctx, recipent, tags, msg)
	logger.CloseLogWorkers()
	store, err := logger.Store()
	if err != nil {
		t.Error(err)
	}
	logs, err := store(ctx, logquerybuilder.GetServerLogs(0, 0))
	if err != nil {
		t.Error(err)
	}
	if len(logs) < 1 {
		t.Errorf("expected to get 1 log-msg; got %d", len(logs))
		return
	}
	found := false
	for _, v := range logs {
		found = found || v.Msg == msg
	}
	if !found {
		t.Errorf("expected logmsg to be 'test; got '%s'", logs[0])
	}
	if _notifyLogsTriggeredWith.Id != recipent {
		t.Errorf("expected NotifyLogs to called with recipent %s; got '%s'", recipent, _notifyLogsTriggeredWith.Id)
	}
	if _notifyLogsTriggeredWith.RecipientType != recipient.Instance {
		t.Errorf("expected NotifyLogs to called with recipentType %s; got '%s'", recipient.Instance, _notifyLogsTriggeredWith.RecipientType)
	}
	if len(telemetrylogs.All()) < 1 {
		t.Errorf("len of telemetry logs wrong")
		return
	}
	if telemetrylogs.All()[0].Message != msg {
		t.Errorf("wrong logmsg want '%s'; got '%s'", msg, telemetrylogs.All()[0].Message)
	}
}

func TestTelemetry(t *testing.T) {
	sugar, telemetrylogs := utils.ObservedLogger()
	logger := InitLogger()
	logger.sugar = sugar

	tags := make(map[string]string)
	tags["recipientType"] = recipient.Server.String()
	msg := "test3"
	logger.telemetry(context.Background(), Info, nil, msg)
	if len(telemetrylogs.All()) < 1 {
		t.Errorf("len of telemetry logs wrong")
		return
	}
	if telemetrylogs.All()[0].Message != msg {
		t.Errorf("wrong logmsg want '%s'; got '%s'", msg, telemetrylogs.All()[0].Message)
	}
}

func TestTelemetryWithTags(t *testing.T) {
	sugar, telemetrylogs := utils.ObservedLogger()
	logger := InitLogger()
	logger.sugar = sugar

	tags := make(map[string]string)
	tags["recipientType"] = recipient.Server.String()
	msg := "test4"
	logger.telemetry(context.Background(), Info, tags, msg)
	if len(telemetrylogs.All()) < 1 {
		t.Errorf("len of telemetry logs wrong")
		return
	}
	if telemetrylogs.All()[0].Message != msg {
		t.Errorf("wrong logmsg want '%s'; got '%s'", msg, telemetrylogs.All()[0].Message)
	}
}

func TestSendLogMsgToDB(t *testing.T) {
	sugar, telemetrylogs := utils.ObservedLogger()
	logger := InitLogger()
	gorm, cleanup, err := utils.DatabaseGorm()
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err := cleanup()
		if err != nil {
			fmt.Println(err)
		}
	}()
	logger.StartLogWorkers(1, gorm, &LogNotifyMock{}, sugar)

	tags := make(map[string]interface{})
	tags["recipientType"] = recipient.Server.String()
	recipent := uuid.New()
	msg := "test1"
	err = logger.create(&queuedLogMsg{
		recipientID:   recipent,
		recipientType: recipient.Instance,
		LogMsg: logstore.LogMsg{
			T:     time.Now(),
			Msg:   msg,
			Level: "info",
			Tags:  tags,
		},
	})
	if err != nil {
		t.Errorf("database transaction failed %v", err)
	}
	store, err := logger.Store()
	if err != nil {
		t.Error(err)
	}

	logs, err := store(context.Background(), logquerybuilder.GetInstanceLogsNoInheritance(recipent, 0, 0))
	if err != nil {
		t.Error(err)
	}
	if len(logs) < 1 {
		t.Errorf("expected to get 1 log-msg; got %d", len(logs))
		return
	}
	if logs[0].Msg != msg {
		t.Errorf("expected logmsg to be '%s'; got '%s'", msg, logs[0])
	}
	if _notifyLogsTriggeredWith.Id != recipent {
		t.Errorf("expected NotifyLogs to called with recipent %s; got '%s'", recipent, _notifyLogsTriggeredWith.Id)
	}
	if _notifyLogsTriggeredWith.RecipientType != recipient.Instance {
		t.Errorf("expected NotifyLogs to called with recipentType %s; got '%s'", recipient.Instance, _notifyLogsTriggeredWith.RecipientType)
	}
	if len(telemetrylogs.All()) > 0 {
		t.Errorf("its not excepted to log telemetry logs in this method")
		return
	}
	if _notifyLogsTriggeredWith.Id != recipent {
		t.Errorf("expected NotifyLogs to called with recipent %s; got '%s'", recipent, _notifyLogsTriggeredWith.Id)
	}
	if _notifyLogsTriggeredWith.RecipientType != recipient.Instance {
		t.Errorf("expected NotifyLogs to called with recipentType %s; got '%s'", recipient.Instance, _notifyLogsTriggeredWith.RecipientType)
	}
}

type LogNotifyMock struct{}

func (ln *LogNotifyMock) NotifyLogs(recipientID uuid.UUID, recipientType recipient.RecipientType) {
	_notifyLogsTriggeredWith = notifyLogsTriggeredWith{
		recipientID,
		recipientType,
	}
}

type notifyLogsTriggeredWith struct {
	Id            uuid.UUID
	RecipientType recipient.RecipientType
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

func assertLogmsgsPresent(t *testing.T, msgs []*logstore.LogMsg, expectedMsg []string) {
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

func assertLogmsgsContain(t *testing.T, msgs []*logstore.LogMsg, expectedMsg []string) {
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

func assertsortedByTime(t *testing.T, in []*logstore.LogMsg) bool {
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

func assertQueryMatchState(t *testing.T, jsondump, wf, state, iterator string, resLen int) []*logstore.LogMsg {
	t.Helper()
	logmsgs := make([]*logstore.LogMsg, 0)
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

func checkNestedLoop(in []*logstore.LogMsg) bool {
	for _, v := range in {
		if v.Tags["state-type"] == "foreach" {
			return true
		}
	}
	return false
}

func assertTagsFiltered(t *testing.T, tags map[string]interface{}, tag, value string) {
	t.Helper()
	if tags[tag] != value {
		t.Errorf("found wrong tag-value for %s: %s want %s", tag, tags[tag], value)
	}
}

func TestFilterByIterrator(t *testing.T) {
	logmsgs := make([]*logstore.LogMsg, 0)
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
	logmsgs := make([]*logstore.LogMsg, 0)
	err := json.Unmarshal([]byte(loopnestedjson), &logmsgs)
	if err != nil {
		t.Error(err)
		return
	}
	res := FilterLogmsg(&grpc.PageFilter{
		Field: "TESTFIELD",
		Type:  "TESTTYPE",
	}, logmsgs)
	ok := assertEquals(t, logmsgs, res)
	if !ok {
		t.Error("input slice should not be modified if Pagefilter does not match")
	}
	res = FilterLogmsg(&grpc.PageFilter{
		Field: "QUERY",
		Type:  "MATCH",
		Val:   "looperlooper::solve",
	}, logmsgs)
	equals := assertEquals(t, logmsgs, res)
	if equals {
		t.Error("input slice should have been filtered")
	}
}

func assertEquals(t *testing.T, a []*logstore.LogMsg, b []*logstore.LogMsg) bool {
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
