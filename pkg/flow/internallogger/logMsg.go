package internallogger

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *JSONB) Scan(value interface{}) error {
	switch v := value.(type) {
	default:
		if err := json.Unmarshal(value.([]byte), &j); err != nil {
			return err
		}
	case string:
		if err := json.Unmarshal([]byte(fmt.Sprint(v)), &j); err != nil {
			return err
		}
	}

	return nil
}

type LogMsgs struct {
	Oid                 uuid.UUID `gorm:"primaryKey"`
	T                   time.Time
	Msg                 string
	Level               string
	RootInstanceId      string `gorm:"default:null"`
	LogInstanceCallPath string
	Tags                JSONB     `sql:"type:jsonb"`
	WorkflowId          uuid.UUID `gorm:"default:null"`
	MirrorActivityId    uuid.UUID `gorm:"default:null"`
	InstanceLogs        uuid.UUID `gorm:"default:null"`
	NamespaceLogs       uuid.UUID `gorm:"default:null"`
}
