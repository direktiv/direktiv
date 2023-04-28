package logstore_test

import (
	"context"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/refactor/internallogger/logstore"
	"github.com/direktiv/direktiv/pkg/refactor/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestQueries(t *testing.T) {
	db := setup()
	id := uuid.New()
	lStore := logstore.NewLogStore(db)
	want := createLogmsg(id, lStore)
	lq := queryTestLogs{}
	got, err := lStore.QueryLogs(context.Background(), lq)
	if err != nil {
		t.Error(err)
	}
	firstItem := got[0]
	if want.Msg != firstItem.Msg {
		t.Errorf("got a wrong log msg. Want %s, got %s", want.Msg, firstItem.Msg)
	}

	if firstItem.Tags["test"] != "testvalue" {
		t.Error("missing tag")
	}
}

func setup() *gorm.DB {
	db, err := utils.NewMockGorm()
	if err != nil {
		panic(err)
	}
	if db == nil {
		panic("nil db happened")
	}
	return db
}

func createLogmsg(id uuid.UUID, lStore *logstore.LogStore) logstore.LogMsg {
	tags := make(map[string]interface{})
	tags["test"] = "testvalue"
	l := logstore.LogMsg{
		T:     time.Now(),
		Msg:   "test",
		Level: "error",
		Tags:  tags,
	}
	err := lStore.Create(id, recipient.Server, l)
	if err != nil {
		panic(err)
	}
	return l
}

//	resp := new(grpc.NamespaceLogsResponse)

type queryTestLogs struct{}

func (ql queryTestLogs) Build() (string, error) {
	return "SELECT * FROM log_msgs;", nil
}
