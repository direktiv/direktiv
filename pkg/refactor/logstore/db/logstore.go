package db

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/logstore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _ logstore.LogStore = &SQLLogStore{} // Ensures SQLLogStore struct conforms to logstore.LogStore interface.

func NewSQLLogStore(db *gorm.DB) logstore.LogStore {
	return &SQLLogStore{
		db: db,
	}
}

type SQLLogStore struct {
	db *gorm.DB
}

// Append implements logstore.LogStore.
func (sl *SQLLogStore) Append(ctx context.Context, timestamp time.Time, msg string, keysAndValues ...interface{}) error {
	fields, err := mapKeysAndValues(keysAndValues...)
	if err != nil {
		return err
	}
	lvl, ok := fields["level"]
	if !ok {
		return fmt.Errorf("level was missing as argument in the keysAndValues pair")
	}
	if lvl != "debug" && lvl != "info" && lvl != "error" && lvl != "panic" {
		return fmt.Errorf("field level provided in keysAndValues has an invalid value %s", lvl)
	}
	l := gormLogMsg{
		Oid:   uuid.New(),
		T:     timestamp,
		Msg:   msg,
		Level: fmt.Sprintf("%s", lvl),
	}
	recipientType, ok := fields["recipientType"]
	if !ok {
		return fmt.Errorf("recipientType was missing as argument in the keysAndValues pair")
	}
	var recipientID interface{}
	var id uuid.UUID
	if recipientType != "server" {
		recipientID, ok = fields["recipientID"]
		if !ok {
			return fmt.Errorf("recipientID was missing as argument in the keysAndValues pair")
		}
		id, err = uuid.Parse(fmt.Sprintf("%s", recipientID))
		if err != nil {
			return fmt.Errorf("recipientID was invalid %w", err)
		}
	}
	switch recipientType {
	case "instance":
		l.InstanceLogs = id
		logInstanceCallPath, ok := fields["logInstanceCallPath"]
		if !ok {
			return fmt.Errorf("logInstanceCallPath was missing as argument in the keysAndValues pair")
		}
		rootInstanceID, ok := fields["rootInstanceID"]
		if !ok {
			return fmt.Errorf("logInstanceCallPath was missing as argument in the keysAndValues pair")
		}
		l.LogInstanceCallPath = fmt.Sprintf("%s", logInstanceCallPath)
		l.RootInstanceID = fmt.Sprintf("%s", rootInstanceID)
	case "workflow":
		l.WorkflowID = id
	case "namespace":
		l.NamespaceLogs = id
	case "mirror":
		l.MirrorActivityID = id
	case "server":
		// do nothing
	}
	delete(fields, "level")
	delete(fields, "rootInstanceID")
	delete(fields, "logInstanceCallPath")
	delete(fields, "recipientType")
	l.Tags = fields
	tx := sl.db.Table("log_msgs").WithContext(ctx).Create(&l)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

// Get implements logstore.LogStore.
func (sl *SQLLogStore) Get(ctx context.Context, keysAndValues ...interface{}) ([]*logstore.LogEntry, error) {
	_ = ctx.Err()                           // linter
	_ = keysAndValues[len(keysAndValues)-1] // linter
	panic("unimplemented")
}

func mapKeysAndValues(keysAndValues ...interface{}) (map[string]interface{}, error) {
	fields := make(map[string]interface{})
	if len(keysAndValues) == 0 {
		return nil, fmt.Errorf("keysAndValues where not provided")
	}
	if len(keysAndValues)%2 != 0 {
		return nil, fmt.Errorf("keysAndValues was not a list of key value pairs")
	}
	for i := 0; i < len(keysAndValues); i += 2 {
		key := fmt.Sprintf("%s", keysAndValues[i])
		fields[key] = keysAndValues[i+1]
	}

	return fields, nil
}

type gormLogMsg struct {
	Oid                 uuid.UUID `gorm:"primaryKey"`
	T                   time.Time
	Msg                 string
	Level               string
	Tags                jsonb     `sql:"type:jsonb"`
	WorkflowID          uuid.UUID `gorm:"default:null"`
	MirrorActivityID    uuid.UUID `gorm:"default:null"`
	InstanceLogs        uuid.UUID `gorm:"default:null"`
	NamespaceLogs       uuid.UUID `gorm:"default:null"`
	RootInstanceID      string    `gorm:"default:null"`
	LogInstanceCallPath string
}

type jsonb map[string]interface{}

func (j jsonb) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)

	return string(valueString), err
}

func (j *jsonb) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		if err := json.Unmarshal([]byte(fmt.Sprint(v)), &j); err != nil {
			return err
		}
	default:
		b, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("type assertion failed")
		}
		if err := json.Unmarshal(b, &j); err != nil {
			return err
		}
	}

	return nil
}
