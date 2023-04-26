package internallogger

import (
	"time"

	"github.com/google/uuid"
)

type LogMsgs struct {
	Oid                 uuid.UUID `gorm:"primaryKey"`
	T                   time.Time
	Msg                 string
	Level               string
	RootInstanceId      string
	LogInstanceCallPath string
	Tags                map[string]string
	WorkflowId          uuid.UUID
	MirrorActivityId    uuid.UUID
	InstanceLogs        uuid.UUID
	NamespaceLogs       uuid.UUID
}
