package dblogstore

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

const (
	ns  = "namespace"
	wf  = "workflow"
	srv = "server"
	ins = "instance"
	mir = "mirror"
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
	lvl, err := getLevel(fields)
	if err != nil {
		return fmt.Errorf("level was missing as argument in the keysAndValues pair")
	}
	l := gormLogMsg{
		Oid:   uuid.New(),
		T:     timestamp,
		Msg:   msg,
		Level: lvl,
	}
	recipientType, err := getRecipientType(fields)
	if err != nil {
		return err
	}
	switch recipientType {
	case ins:
		id, err := getRecipientID("instance-id", fields)
		if err != nil {
			return err
		}
		l.InstanceLogs = id
		logInstanceCallPath, ok := fields["callpath"]
		if !ok {
			return fmt.Errorf("callpath was missing as argument in the keysAndValues pair")
		}
		rootInstanceID, ok := fields["rootInstanceID"]
		if !ok {
			return fmt.Errorf("rootInstanceID was missing as argument in the keysAndValues pair")
		}
		l.LogInstanceCallPath = fmt.Sprintf("%s", logInstanceCallPath)
		l.RootInstanceID = fmt.Sprintf("%s", rootInstanceID)
	case wf:
		id, err := getRecipientID("workflow-id", fields)
		if err != nil {
			return err
		}
		l.WorkflowID = id
	case ns:
		id, err := getRecipientID("namespace-id", fields)
		if err != nil {
			return err
		}
		l.NamespaceLogs = id
	case mir:
		id, err := getRecipientID("mirror-id", fields)
		if err != nil {
			return err
		}
		l.MirrorActivityID = id
	case srv:
		// do nothing
	}
	delete(fields, "level")
	delete(fields, "rootInstanceID")
	delete(fields, "callpath")
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
	fields, err := mapKeysAndValues(keysAndValues...)
	if err != nil {
		return nil, err
	}
	query, err := buildQuery(fields)
	if err != nil {
		return nil, err
	}
	resultList := make([]*gormLogMsg, 0)
	tx := sl.db.WithContext(ctx).Raw(query).Scan(&resultList)
	if tx.Error != nil {
		return nil, tx.Error
	}
	convertedList := make([]*logstore.LogEntry, 0, len(resultList))
	for _, e := range resultList {
		m := make(map[string]interface{})
		for k, e := range e.Tags {
			m[k] = e
		}
		m["level"] = e.Level
		convertedList = append(convertedList, &logstore.LogEntry{
			T:      e.T,
			Msg:    e.Msg,
			Fields: m,
		})
	}

	return convertedList, nil
}

func buildQuery(fields map[string]interface{}) (string, error) {
	ql := newQueryBuilder()
	recipientType, err := getRecipientType(fields)
	if err != nil {
		return "", err
	}
	level, ok := fields["level"]
	if ok {
		levelValue, ok := level.(string)
		if !ok {
			return "", fmt.Errorf("level should be a string")
		}
		ql.setMinumLogLevel(levelValue)
	}

	if recipientType == srv {
		ql.setWorkflowIsNil()
		ql.setNamespaceIsNIl()
		ql.setInstanceIsNIl()
	}
	var id uuid.UUID

	if recipientType == wf {
		ql.setWorkflow(id)
	}
	if recipientType == ns {
		ql.setNamespace(id)
	}
	if recipientType == ins {
		var err error
		ql, err = addInstanceValuesToQuery(ql, fields)
		if err != nil {
			return "", err
		}
	}
	if recipientType == mir {
		ql.whereMirrorActivityID(id)
	}
	limit, ok := fields["limit"]
	var limitValue int
	if ok {
		limitValue, ok = limit.(int)
		if !ok {
			return "", fmt.Errorf("limit should be an int")
		}
	}
	offset, ok := fields["offset"]
	var offsetValue int
	if ok {
		offsetValue, ok = offset.(int)
		if !ok {
			return "", fmt.Errorf("offset should be an int")
		}
	}

	if limitValue > 0 {
		ql.setLimit(limitValue)
	}
	if offsetValue > 0 {
		ql.setOffset(offsetValue)
	}
	query, err := ql.build()
	if err != nil {
		return "", err
	}

	return query, nil
}

func getLevel(fields map[string]interface{}) (string, error) {
	lvl, ok := fields["level"]
	if !ok {
		return "", fmt.Errorf("level was missing as argument in the keysAndValues pair")
	}
	switch lvl {
	case "debug":
	case "info":
	case "error":
	case "panic":
	default:
		return "", fmt.Errorf("field level provided in keysAndValues has an invalid value %s", lvl)
	}
	value, ok := lvl.(string)
	if !ok {
		return "", fmt.Errorf("level should be a string")
	}

	return value, nil
}

func getRecipientID(fieldName string, fields map[string]interface{}) (uuid.UUID, error) {
	recipientID, ok := fields[fieldName]
	if !ok {
		return uuid.UUID{}, fmt.Errorf("%s was missing as argument in the keysAndValues pair", fieldName)
	}
	id, err := uuid.Parse(fmt.Sprintf("%s", recipientID))
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("recipientID was invalid %w", err)
	}

	return id, nil
}

func getRecipientType(fields map[string]interface{}) (string, error) {
	recipientType, ok := fields["recipientType"]
	if !ok {
		return "", fmt.Errorf("recipientType was missing as argument in the keysAndValues pair")
	}
	if recipientType != srv &&
		recipientType != ins &&
		recipientType != mir &&
		recipientType != ns &&
		recipientType != wf {
		return "", fmt.Errorf("invalid recipientType %s", recipientType)
	}
	value, ok := recipientType.(string)
	if !ok {
		return "", fmt.Errorf("recipientType should be a string")
	}

	return value, nil
}

func addInstanceValuesToQuery(ql *logMsgQueryBuilder, fields map[string]interface{}) (*logMsgQueryBuilder, error) {
	prefix, ok := fields["callpath"]
	if !ok {
		return nil, fmt.Errorf("callpath was missing as argument in the keysAndValues pair")
	}
	rootInstanceID, ok := fields["rootInstanceID"]
	if !ok {
		return nil, fmt.Errorf("rootInstanceID was missing as argument in the keysAndValues pair")
	}
	callerIsRoot, ok := fields["isCallerRoot"]
	if !ok {
		return nil, fmt.Errorf("isCallerRoot was missing as argument in the keysAndValues pair")
	}
	callerIsRootValue, ok := callerIsRoot.(bool)
	if !ok {
		return nil, fmt.Errorf("callerIsRoot should be an bool")
	}
	ql.setRootInstanceIDEQ(fmt.Sprintf("%s", rootInstanceID))
	if !callerIsRootValue {
		ql.setInstanceCallPathHasPrefix(fmt.Sprintf("%s", prefix))
	}

	return ql, nil
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

type logMsgQueryBuilder struct {
	whereEQStatements []string
	limit             int
	offset            int
}

func newQueryBuilder() *logMsgQueryBuilder {
	return &logMsgQueryBuilder{
		whereEQStatements: []string{},
	}
}

func (b *logMsgQueryBuilder) setWorkflow(workflowID uuid.UUID) {
	wEq := b.whereEQStatements
	wEq = append(wEq, fmt.Sprintf("workflow_id = '%s'", workflowID.String()))
	b.whereEQStatements = wEq
}

func (b *logMsgQueryBuilder) setWorkflowIsNil() {
	wEq := b.whereEQStatements
	wEq = append(wEq, "workflow_id IS NULL")
	b.whereEQStatements = wEq
}

func (b *logMsgQueryBuilder) setNamespaceIsNIl() {
	wEq := b.whereEQStatements
	wEq = append(wEq, "namespace_logs IS NULL")
	b.whereEQStatements = wEq
}

func (b *logMsgQueryBuilder) setInstanceIsNIl() {
	wEq := b.whereEQStatements
	wEq = append(wEq, "instance_logs IS NULL")
	b.whereEQStatements = wEq
}

func (b *logMsgQueryBuilder) setNamespace(namespaceID uuid.UUID) {
	wEq := b.whereEQStatements
	wEq = append(wEq, fmt.Sprintf("namespace_logs = '%s'", namespaceID.String()))
	b.whereEQStatements = wEq
}

// func (b *logMsgQueryBuilder) whereInstance(instanceID uuid.UUID) {
// 	wEq := b.whereEQStatements
// 	wEq = append(wEq, fmt.Sprintf("instance_logs = '%s'", instanceID.String()))
// 	b.whereEQStatements = wEq
// }

func (b *logMsgQueryBuilder) setRootInstanceIDEQ(rootID string) {
	wEq := b.whereEQStatements
	wEq = append(wEq, fmt.Sprintf("root_instance_id = '%s'", rootID))
	b.whereEQStatements = wEq
}

func (b *logMsgQueryBuilder) setInstanceCallPathHasPrefix(prefix string) {
	wEq := b.whereEQStatements
	wEq = append(wEq, fmt.Sprintf("log_instance_call_path like '%s%%'", prefix))
	b.whereEQStatements = wEq
}

// func (b *logMsgQueryBuilder) whereLogLevel(loglevel string) {
// 	wEq := b.whereEQStatements
// 	wEq = append(wEq, fmt.Sprintf("level = '%s'", loglevel))
// 	b.whereEQStatements = wEq
// }

func (b *logMsgQueryBuilder) whereMirrorActivityID(id uuid.UUID) {
	wEq := b.whereEQStatements
	wEq = append(wEq, fmt.Sprintf("mirror_activity_id = '%s'", id.String()))
	b.whereEQStatements = wEq
}

func (b *logMsgQueryBuilder) setMinumLogLevel(loglevel string) {
	wEq := b.whereEQStatements
	levels := []string{"debug", "info", "error", "panic"}
	switch loglevel {
	case "debug":
	case "info":
		levels = levels[1:]
	case "error":
		levels = levels[2:]
	case "panic":
		levels = levels[3:]
	}
	q := "( "
	for i, e := range levels {
		q += fmt.Sprintf("level = '%s' ", e)
		if i < len(levels)-1 {
			q += " OR "
		}
	}
	q += " )"
	wEq = append(wEq, q)
	b.whereEQStatements = wEq
}

func (b *logMsgQueryBuilder) setLimit(limit int) {
	b.limit = limit
}

func (b *logMsgQueryBuilder) setOffset(offset int) {
	b.offset = offset
}

func (b *logMsgQueryBuilder) build() (string, error) {
	if len(b.whereEQStatements) < 1 {
		return "", fmt.Errorf("no Where statements where provided")
	}
	q := `SELECT oid, t, msg, level, root_instance_id, log_instance_call_path, tags, workflow_id, mirror_activity_id, instance_logs, namespace_logs
	FROM log_msgs `
	q += "WHERE "
	for i, e := range b.whereEQStatements {
		q += e
		if i+1 < len(b.whereEQStatements) {
			q += " AND "
		}
	}
	q += " ORDER BY t ASC"
	if b.limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", b.limit)
	}
	if b.offset > 0 {
		q += fmt.Sprintf(" OFFSET %d", b.offset)
	}

	return q + ";", nil
}
