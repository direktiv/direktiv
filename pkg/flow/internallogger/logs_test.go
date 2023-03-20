package internallogger

import (
	"context"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"github.com/direktiv/direktiv/pkg/flow/database/entwrapper"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entlog "github.com/direktiv/direktiv/pkg/flow/ent/logmsg"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

var _notifyLogsTriggeredWith notifyLogsTriggeredWith

func databaseWrapper() (entwrapper.Database, error) {
	client, err := ent.Open(dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	if err != nil {
		return entwrapper.Database{}, err
	}
	ctx := context.Background()

	if err := client.Schema.Create(ctx); err != nil {
		return entwrapper.Database{}, err
	}
	sugar := zap.S()
	entw := entwrapper.Database{
		Client: client,
		Sugar:  sugar,
	}
	return entw, nil
}

func observedLogger() (*zap.SugaredLogger, *observer.ObservedLogs) {
	observed, telemetrylogs := observer.New(zapcore.DebugLevel)
	sugar := zap.New(observed).Sugar()
	return sugar, telemetrylogs
}

func TestStoringLogMsg(t *testing.T) {
	entw, err := databaseWrapper()
	if err != nil {
		t.Error("error initializing the db ", err)
	}
	defer entw.Close()

	sugar, telemetrylogs := observedLogger()
	logger := InitLogger()
	logger.StartLogWorkers(1, &entw, &LogNotifyMock{}, sugar)
	tags := make(map[string]string)
	tags["recipientType"] = util.Server
	tags["testTag"] = util.Server
	recipent := uuid.New()
	msg := "test2"
	logger.Infof(context.Background(), recipent, tags, msg)

	logger.CloseLogWorkers()
	logs, err := entw.Client.LogMsg.Query().Where(entlog.LevelEQ(util.Info)).All(context.Background())
	if err != nil {
		t.Errorf("query logmsg failed %v", err)
		return
	}
	if len(logs) != 1 {
		t.Errorf("expected to get 1 log-msg; got %d", len(logs))
		return
	}
	if logs[0].Msg != msg {
		t.Errorf("expected logmsg to be 'test; got '%s'", logs[0])
	}
	if _notifyLogsTriggeredWith.Id != recipent {
		t.Errorf("expected NotifyLogs to called with recipent %s; got '%s'", recipent, _notifyLogsTriggeredWith.Id)
	}
	if _notifyLogsTriggeredWith.RecipientType != util.Server {
		t.Errorf("expected NotifyLogs to called with recipentType %s; got '%s'", util.Server, _notifyLogsTriggeredWith.RecipientType)
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
	sugar, telemetrylogs := observedLogger()
	logger := InitLogger()
	logger.sugar = sugar

	tags := make(map[string]string)
	tags["recipientType"] = util.Server
	msg := "test3"
	logger.Telemetry(context.Background(), util.Info, nil, msg)
	if len(telemetrylogs.All()) < 1 {
		t.Errorf("len of telemetry logs wrong")
		return
	}
	if telemetrylogs.All()[0].Message != msg {
		t.Errorf("wrong logmsg want '%s'; got '%s'", msg, telemetrylogs.All()[0].Message)
	}
}

func TestTelemetryWithTags(t *testing.T) {
	sugar, telemetrylogs := observedLogger()
	logger := InitLogger()
	logger.sugar = sugar

	tags := make(map[string]string)
	tags["recipientType"] = util.Server
	msg := "test4"
	logger.Telemetry(context.Background(), util.Info, tags, msg)
	if len(telemetrylogs.All()) < 1 {
		t.Errorf("len of telemetry logs wrong")
		return
	}
	if telemetrylogs.All()[0].Message != msg {
		t.Errorf("wrong logmsg want '%s'; got '%s'", msg, telemetrylogs.All()[0].Message)
	}
}

func TestSendLogMsgToDB(t *testing.T) {
	entw, err := databaseWrapper()
	if err != nil {
		t.Error("error initializing the db ", err)
	}
	defer entw.Close()

	sugar, telemetrylogs := observedLogger()
	logger := InitLogger()
	logger.StartLogWorkers(1, &entw, &LogNotifyMock{}, sugar)

	tags := make(map[string]string)
	tags["recipientType"] = util.Server
	recipent := uuid.New()
	msg := "test1"
	err = logger.SendLogMsgToDB(&logMessage{time.Now(), msg, util.Info, recipent, tags})
	if err != nil {
		t.Errorf("database trancaction failed %v", err)
	}
	logs, err := entw.Client.LogMsg.Query().Where(entlog.LevelEQ(util.Info)).All(context.Background())
	if err != nil {
		t.Errorf("query logmsg failed %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("expected to get 1 log-msg; got %d", len(logs))
		return
	}
	if logs[0].Msg != msg {
		t.Errorf("expected logmsg to be '%s'; got '%s'", msg, logs[0])
	}
	if _notifyLogsTriggeredWith.Id != recipent {
		t.Errorf("expected NotifyLogs to called with recipent %s; got '%s'", recipent, _notifyLogsTriggeredWith.Id)
	}
	if _notifyLogsTriggeredWith.RecipientType != util.Server {
		t.Errorf("expected NotifyLogs to called with recipentType %s; got '%s'", util.Server, _notifyLogsTriggeredWith.RecipientType)
	}
	if len(telemetrylogs.All()) > 0 {
		t.Errorf("its not excpected to log telemetry logs in this method")
		return
	}
	if _notifyLogsTriggeredWith.Id != recipent {
		t.Errorf("expected NotifyLogs to called with recipent %s; got '%s'", recipent, _notifyLogsTriggeredWith.Id)
	}
	if _notifyLogsTriggeredWith.RecipientType != util.Server {
		t.Errorf("expected NotifyLogs to called with recipentType %s; got '%s'", util.Server, _notifyLogsTriggeredWith.RecipientType)
	}
}

type LogNotifyMock struct{}

func (ln *LogNotifyMock) NotifyLogs(recipientID uuid.UUID, recipientType string) {
	_notifyLogsTriggeredWith = notifyLogsTriggeredWith{
		recipientID,
		recipientType,
	}
}

type notifyLogsTriggeredWith struct {
	Id            uuid.UUID
	RecipientType string
}
