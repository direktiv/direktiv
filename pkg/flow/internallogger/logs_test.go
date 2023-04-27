package internallogger

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/flow/internal/testutils"
	"github.com/google/uuid"
)

var _notifyLogsTriggeredWith notifyLogsTriggeredWith

func TestStoringLogMsg(t *testing.T) {
	sugar, telemetrylogs := testutils.ObservedLogger()
	logger := InitLogger()
	gorm, cleanup, err := testutils.DatabaseGorm()
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
	tags["recipientType"] = recipient.Instance.String()
	tags["testTag"] = recipient.Server.String()
	recipent := uuid.New()
	msg := "test2"
	ctx := context.TODO()
	logger.Infof(ctx, recipent, tags, msg)
	logger.CloseLogWorkers()
	lq := QueryLogs()
	lq.WhereInstance(recipent)
	logs, err := lq.GetAll(context.Background(), gorm)
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
	sugar, telemetrylogs := testutils.ObservedLogger()
	logger := InitLogger()
	logger.sugar = sugar

	tags := make(map[string]string)
	tags["recipientType"] = recipient.Server.String()
	msg := "test3"
	logger.Telemetry(context.Background(), Info, nil, msg)
	if len(telemetrylogs.All()) < 1 {
		t.Errorf("len of telemetry logs wrong")
		return
	}
	if telemetrylogs.All()[0].Message != msg {
		t.Errorf("wrong logmsg want '%s'; got '%s'", msg, telemetrylogs.All()[0].Message)
	}
}

func TestTelemetryWithTags(t *testing.T) {
	sugar, telemetrylogs := testutils.ObservedLogger()
	logger := InitLogger()
	logger.sugar = sugar

	tags := make(map[string]string)
	tags["recipientType"] = recipient.Server.String()
	msg := "test4"
	logger.Telemetry(context.Background(), Info, tags, msg)
	if len(telemetrylogs.All()) < 1 {
		t.Errorf("len of telemetry logs wrong")
		return
	}
	if telemetrylogs.All()[0].Message != msg {
		t.Errorf("wrong logmsg want '%s'; got '%s'", msg, telemetrylogs.All()[0].Message)
	}
}

func TestSendLogMsgToDB(t *testing.T) {
	sugar, telemetrylogs := testutils.ObservedLogger()
	logger := InitLogger()
	gorm, cleanup, err := testutils.DatabaseGorm()
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
	err = logger.SendLogMsgToDB(&logMsg{
		recipientID:   recipent,
		recipientType: recipient.Instance,
		LogMsgs: &LogMsgs{
			T:            time.Now(),
			Msg:          msg,
			Level:        "info",
			Tags:         tags,
			InstanceLogs: recipent,
		},
	})
	if err != nil {
		t.Errorf("database transaction failed %v", err)
	}
	lq := QueryLogs()
	lq.WhereInstance(recipent)
	logs, err := lq.GetAll(context.TODO(), gorm)
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
