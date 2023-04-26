package internallogger

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LogMsgQuery interface {
	WhereWorkflow(workflowId uuid.UUID)
	WhereNamespace(namespaceId uuid.UUID)
	WhereInstance(instanceId uuid.UUID)
	WhereRootInstanceIdEQ(rootId string)
	WhereInstanceCallPathHasPrefix(prefix string)
	WhereLogLevel(loglevel string)
	WhereMinumLogLevel(loglevel string)
	WithLimit(limit int)
	WithOffset(offset int)
	// GetServerLogs(ctx context.Context) []*LogMsgs
	GetLimit() int
	GetOffset() int
	GetAll(ctx context.Context, db *gorm.DB) ([]*LogMsgs, error)
}

type LogMsgStorer interface {
	// QueryLogs(ctx context.Context, db *gorm.DB) LogMsgQuery
	create(ctx context.Context, l logMessage)
}

type LogMsgQueryBuilder struct {
	whereEQStatements []string
	limit             int
	offset            int
}

func QueryLogs() *LogMsgQueryBuilder {
	return &LogMsgQueryBuilder{
		whereEQStatements: []string{},
	}
}

func (b *LogMsgQueryBuilder) WhereWorkflow(workflowId uuid.UUID) {
	wEq := b.whereEQStatements
	wEq = append(wEq, fmt.Sprintf("workflow_id = '%s'", workflowId.String()))
	b.whereEQStatements = wEq
}

func (b *LogMsgQueryBuilder) WhereNamespace(namespaceId uuid.UUID) {
	wEq := b.whereEQStatements
	wEq = append(wEq, fmt.Sprintf("namespace_logs = '%s'", namespaceId.String()))
	b.whereEQStatements = wEq
}

func (b *LogMsgQueryBuilder) WhereInstance(instanceId uuid.UUID) {
	wEq := b.whereEQStatements
	wEq = append(wEq, fmt.Sprintf("instance_logs = '%s'", instanceId.String()))
	b.whereEQStatements = wEq
}

func (b *LogMsgQueryBuilder) WhereRootInstanceIdEQ(rootId string) {
	wEq := b.whereEQStatements
	wEq = append(wEq, fmt.Sprintf("root_instance_id = '%s'", rootId))
	b.whereEQStatements = wEq
}

func (b *LogMsgQueryBuilder) WhereInstanceCallPathHasPrefix(prefix string) {
	wEq := b.whereEQStatements
	wEq = append(wEq, fmt.Sprintf("log_instance_call_path like '%s%%'", prefix))
	b.whereEQStatements = wEq
}

func (b *LogMsgQueryBuilder) WhereLogLevel(loglevel string) {
	wEq := b.whereEQStatements
	wEq = append(wEq, fmt.Sprintf("level = '%s'", loglevel))
	b.whereEQStatements = wEq
}

func (b *LogMsgQueryBuilder) WhereMinumLogLevel(loglevel string) {
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
		q += fmt.Sprintf("level = %s", e)
		if i < len(levels) {
			q += " OR "
		}
	}
	q += ")"
	wEq = append(wEq, q)
	b.whereEQStatements = wEq
}

func (b *LogMsgQueryBuilder) WithLimit(limit int) {
	b.limit = limit
}

func (b *LogMsgQueryBuilder) WithOffset(offset int) {
	b.offset = offset
}

func (b *LogMsgQueryBuilder) GetServerLogs(loglevel string) []*logMessage {
	return nil
}

func (b *LogMsgQueryBuilder) build() (string, error) {
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

func (b *LogMsgQueryBuilder) GetAll(ctx context.Context, db *gorm.DB) ([]*LogMsgs, error) {
	query, err := b.build()
	if err != nil {
		return nil, err
	}
	var resultList []*LogMsgs
	if db == nil {
		panic("db was nil")
	}
	res := db.WithContext(ctx).Raw(query).Scan(&resultList)
	if res.Error != nil {
		return nil, res.Error
	}

	return resultList, nil
}

func (b *LogMsgQueryBuilder) GetLimit() int {
	return b.limit
}

func (b *LogMsgQueryBuilder) GetOffset() int {
	return b.offset
}

// func (b *LogMsgQueryBuilder) GetFirst() (*LogMsgs, error) {
// 	res, err := b.GetAll()
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(res) > 1 {
// 		return nil, fmt.Errorf("got more the one entries")
// 	}
// 	if len(res) == 0 {
// 		return nil, nil
// 	}
// 	return &res[0], nil
// }
