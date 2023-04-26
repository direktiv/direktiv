package internallogger_test

import (
	"context"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/direktiv/direktiv/pkg/refactor/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestQueries(t *testing.T) {
	db := setup()
	id := uuid.New()
	want := createLogmsg(id, db)
	q := internallogger.QueryLogs()
	q.WhereLogLevel("error")
	got, err := q.GetAll(context.TODO(), db)
	if err != nil {
		t.Error(err)
	}
	firstItem := got[0]
	if want.Oid != firstItem.Oid {
		t.Errorf("got a wrong id. Want %s, got %s", want.Oid, firstItem.Oid)
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

func createLogmsg(id uuid.UUID, db *gorm.DB) internallogger.LogMsgs {
	tags := make(map[string]string)
	tags["test"] = "testvalue"
	l := internallogger.LogMsgs{
		Oid:                 uuid.New(),
		T:                   time.Now(),
		Tags:                tags,
		Msg:                 "test",
		Level:               "error",
		WorkflowId:          id,
		RootInstanceId:      "",
		LogInstanceCallPath: "",
		MirrorActivityId:    uuid.New(),
		InstanceLogs:        uuid.New(),
		NamespaceLogs:       uuid.New(),
	}
	t := db.Table("log_msgs").Create(l)
	if t.Error != nil {
		panic(t.Error)
	}
	return l
}

//	resp := new(grpc.NamespaceLogsResponse)
