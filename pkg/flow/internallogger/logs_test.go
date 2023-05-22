package internallogger

// import (
// 	"context"
// 	"testing"
// 	"time"

// 	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
// 	entlog "github.com/direktiv/direktiv/pkg/flow/ent/logmsg"
// 	"github.com/direktiv/direktiv/pkg/flow/internal/testutils"
// 	"github.com/google/uuid"
// )

// var _notifyLogsTriggeredWith notifyLogsTriggeredWith

// func TestStoringLogMsg(t *testing.T) {
// 	db, err := testutils.DatabaseWrapper()
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	defer db.StopDB()
// 	sugar, telemetrylogs := testutils.ObservedLogger()
// 	logger := InitLogger()
// 	logger.StartLogWorkers(1, &db.Entw, &LogNotifyMock{}, sugar)
// 	tags := make(map[string]string)
// 	tags["recipientType"] = "server".String()
// 	tags["testTag"] = "server".String()
// 	recipent := uuid.New()
// 	msg := "test2"
// 	logger.Infof(context.Background(), recipent, tags, msg)

// 	logger.CloseLogWorkers()
// 	logs, err := db.Entw.Client.LogMsg.Query().Where(entlog.LevelEQ(string(Info))).All(context.Background())
// 	if err != nil {
// 		t.Errorf("query logmsg failed %v", err)
// 		return
// 	}
// 	if len(logs) < 1 {
// 		t.Errorf("expected to get 1 log-msg; got %d", len(logs))
// 		return
// 	}
// 	found := false
// 	for _, v := range logs {
// 		found = found || v.Msg == msg
// 	}
// 	if !found {
// 		t.Errorf("expected logmsg to be 'test; got '%s'", logs[0])
// 	}
// 	if _notifyLogsTriggeredWith.Id != recipent {
// 		t.Errorf("expected NotifyLogs to called with recipent %s; got '%s'", recipent, _notifyLogsTriggeredWith.Id)
// 	}
// 	if _notifyLogsTriggeredWith.RecipientType != "server" {
// 		t.Errorf("expected NotifyLogs to called with recipentType %s; got '%s'", "server", _notifyLogsTriggeredWith.RecipientType)
// 	}
// 	if len(telemetrylogs.All()) < 1 {
// 		t.Errorf("len of telemetry logs wrong")
// 		return
// 	}
// 	if telemetrylogs.All()[0].Message != msg {
// 		t.Errorf("wrong logmsg want '%s'; got '%s'", msg, telemetrylogs.All()[0].Message)
// 	}
// }

// func TestTelemetry(t *testing.T) {
// 	sugar, telemetrylogs := testutils.ObservedLogger()
// 	logger := InitLogger()
// 	logger.sugar = sugar

// 	tags := make(map[string]string)
// 	tags["recipientType"] = "server".String()
// 	msg := "test3"
// 	logger.Telemetry(context.Background(), Info, nil, msg)
// 	if len(telemetrylogs.All()) < 1 {
// 		t.Errorf("len of telemetry logs wrong")
// 		return
// 	}
// 	if telemetrylogs.All()[0].Message != msg {
// 		t.Errorf("wrong logmsg want '%s'; got '%s'", msg, telemetrylogs.All()[0].Message)
// 	}
// }

// func TestTelemetryWithTags(t *testing.T) {
// 	sugar, telemetrylogs := testutils.ObservedLogger()
// 	logger := InitLogger()
// 	logger.sugar = sugar

// 	tags := make(map[string]string)
// 	tags["recipientType"] = "server".String()
// 	msg := "test4"
// 	logger.Telemetry(context.Background(), Info, tags, msg)
// 	if len(telemetrylogs.All()) < 1 {
// 		t.Errorf("len of telemetry logs wrong")
// 		return
// 	}
// 	if telemetrylogs.All()[0].Message != msg {
// 		t.Errorf("wrong logmsg want '%s'; got '%s'", msg, telemetrylogs.All()[0].Message)
// 	}
// }

// func TestSendLogMsgToDB(t *testing.T) {
// 	db, err := testutils.DatabaseWrapper()
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	defer db.StopDB()

// 	sugar, telemetrylogs := testutils.ObservedLogger()
// 	logger := InitLogger()
// 	logger.StartLogWorkers(1, &db.Entw, &LogNotifyMock{}, sugar)

// 	tags := make(map[string]string)
// 	tags["recipientType"] = "server".String()
// 	recipent := uuid.New()
// 	msg := "test1"
// 	err = logger.SendLogMsgToDB(&logMessage{time.Now(), msg, Info, recipent, tags})
// 	if err != nil {
// 		t.Errorf("database transaction failed %v", err)
// 	}
// 	logs, err := db.Entw.Client.LogMsg.Query().Where(entlog.LevelEQ(string(Info))).All(context.Background())
// 	if err != nil {
// 		t.Errorf("query logmsg failed %v", err)
// 	}
// 	if len(logs) < 1 {
// 		t.Errorf("expected to get 1 log-msg; got %d", len(logs))
// 		return
// 	}
// 	if logs[0].Msg != msg {
// 		t.Errorf("expected logmsg to be '%s'; got '%s'", msg, logs[0])
// 	}
// 	if _notifyLogsTriggeredWith.Id != recipent {
// 		t.Errorf("expected NotifyLogs to called with recipent %s; got '%s'", recipent, _notifyLogsTriggeredWith.Id)
// 	}
// 	if _notifyLogsTriggeredWith.RecipientType != "server" {
// 		t.Errorf("expected NotifyLogs to called with recipentType %s; got '%s'", "server", _notifyLogsTriggeredWith.RecipientType)
// 	}
// 	if len(telemetrylogs.All()) > 0 {
// 		t.Errorf("its not excepted to log telemetry logs in this method")
// 		return
// 	}
// 	if _notifyLogsTriggeredWith.Id != recipent {
// 		t.Errorf("expected NotifyLogs to called with recipent %s; got '%s'", recipent, _notifyLogsTriggeredWith.Id)
// 	}
// 	if _notifyLogsTriggeredWith.RecipientType != "server" {
// 		t.Errorf("expected NotifyLogs to called with recipentType %s; got '%s'", "server", _notifyLogsTriggeredWith.RecipientType)
// 	}
// }

// type LogNotifyMock struct{}

// func (ln *LogNotifyMock) NotifyLogs(recipientID uuid.UUID, recipientType recipient.RecipientType) {
// 	_notifyLogsTriggeredWith = notifyLogsTriggeredWith{
// 		recipientID,
// 		recipientType,
// 	}
// }

// type notifyLogsTriggeredWith struct {
// 	Id            uuid.UUID
// 	RecipientType recipient.RecipientType
// }
