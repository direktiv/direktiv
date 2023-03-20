package internallogger

import (
	"context"
	"testing"

	"entgo.io/ent/dialect"
	"github.com/direktiv/direktiv/pkg/flow/database/entwrapper"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entlog "github.com/direktiv/direktiv/pkg/flow/ent/logmsg"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

var (
	sugar                    *zap.SugaredLogger
	_notifyLogsTriggeredWith notifyLogsTriggeredWith
)

func databaseWrapper() (entwrapper.Database, error) {
	client, err := ent.Open(dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	if err != nil {
		return entwrapper.Database{}, err
	}
	ctx := context.Background()

	if err := client.Schema.Create(ctx); err != nil {
		return entwrapper.Database{}, err
	}
	entw := entwrapper.Database{
		Client: client,
		Sugar:  sugar,
	}
	return entw, nil
}

func TestStoringLogMsg(t *testing.T) {
	entw, err := databaseWrapper()
	if err != nil {
		t.Error("error initializing the db ", err)
	}
	defer entw.Close()

	sugar = zap.S()
	logger := InitLogger()
	logger.StartLogWorkers(1, &entw, &LogNotifyMock{}, sugar)
	tags := make(map[string]string)
	tags["recipientType"] = "server"
	recipent := uuid.New()
	logger.Infof(context.Background(), recipent, tags, "test")
	logger.CloseLogWorkers()
	logs, err := entw.Client.LogMsg.Query().Where(entlog.LevelEQ("info")).All(context.Background())
	if err != nil {
		t.Errorf("query logmsg failed %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("expected to get 1 log-msg; got %d", len(logs))
	}
	if logs[0].Msg != "test" {
		t.Errorf("expected logmsg to be 'test; got '%s'", logs[0])
	}
	if _notifyLogsTriggeredWith.Id != recipent {
		t.Errorf("expected NotifyLogs to called with recipent %s; got '%s'", recipent, _notifyLogsTriggeredWith.Id)
	}
	if _notifyLogsTriggeredWith.RecipientType != "server" {
		t.Errorf("expected NotifyLogs to called with recipentType %s; got '%s'", "server", _notifyLogsTriggeredWith.RecipientType)
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
