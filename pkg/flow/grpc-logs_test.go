package flow

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/internallogger"
	"github.com/direktiv/direktiv/pkg/refactor/internallogger/logstore"
	logquerybuilder "github.com/direktiv/direktiv/pkg/refactor/internallogger/logstore/log-querybuilder"
	"github.com/direktiv/direktiv/pkg/refactor/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

//go:embed mockdata/entlog_loop.json
var loopjson string

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

	gdb, cleanup, err := utils.DatabaseGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	defer func() {
		err := cleanup()
		if err != nil {
			fmt.Println(err)
		}
	}()
	logmsgs := make([]*logstore.LogMsg, 0)
	err = json.Unmarshal([]byte(jsondump), &logmsgs)
	if err != nil {
		t.Error(err)
	}
	in := make(map[string]*logstore.LogMsg)
	inLen := 0
	id, e := uuid.Parse("1a0d5909-223f-4f44-86d7-1833ab4d21c8")
	if e != nil {
		t.Error(e)
	}
	for _, v := range logmsgs {

		e, err := storeLogmsg(ctx, gdb, id, recipient.Instance, *v)
		if err != nil {
			t.Error(err)
		}
		in[e.Msg] = e
		inLen++
	}

	if len(logmsgs) != inLen {
		t.Errorf("Missing Results len was %d should %d", len(logmsgs), inLen)
	}
	resultList := make([]*logstore.LogMsg, 0)
	gdb.Raw("SELECT * FROM log_msgs").Scan(&resultList)
	if len(resultList) == 0 {
		t.Error("insert failed")
	}
	if len(resultList) != inLen {
		t.Error("insert failed")
	}
	page := grpc.Pagination{}
	pi := grpc.PageInfo{}
	ctx = context.Background()
	logger, _ := utils.ObservedLogger()
	ilogger := internallogger.InitLogger()
	ilogger.StartLogWorkers(1, gdb, &LogNotifyMock{}, logger)
	store, err := ilogger.Store()
	if err != nil {
		t.Error()
	}
	logs, err := store(ctx, logquerybuilder.GetInstanceLogsNoInheritance(id, 0, 0))
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

func assertNoMissingLogs(t *testing.T, res *grpc.InstanceLogsResponse, in map[string]*logstore.LogMsg) {
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

func storeLogmsg(ctx context.Context, db *gorm.DB, recipientID uuid.UUID, recipientType recipient.RecipientType, log logstore.LogMsg) (*logstore.LogMsg, error) {
	err := logstore.NewLogStore(db).Create(recipientID, recipientType, log)
	if err != nil {
		panic(err)
		// fmt.Println(t.Error)
	}
	return &log, nil
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

// func TestWhiteboxTestServerLogs(t *testing.T) {
// 	srv := server{}
// 	flowSrv := flow{}

// 	flowSrv.server = &srv
// 	logs, logobserver := utils.ObservedLogger()
// 	gdb, cleanup, err := utils.DatabaseGorm()
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	defer func() {
// 		err := cleanup()
// 		if err != nil {
// 			fmt.Println(err)
// 		}
// 	}()
// 	srv.gormDB = gdb
// 	srv.sugar = logs
// 	reqSrvLogs := grpc.ServerLogsRequest{
// 		Pagination: &grpc.Pagination{},
// 	}
// 	resSrvLogs := requestServerLogs(t, flowSrv, &reqSrvLogs)
// 	if len(logobserver.All()) <= 0 {
// 		t.Error("some logmsg should heve been printed")
// 	}
// 	if int(resSrvLogs.PageInfo.Limit) > len(resSrvLogs.Results) {
// 		t.Errorf("got more results then specified in pageinfo")
// 	}
// }
