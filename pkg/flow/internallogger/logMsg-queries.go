package internallogger

import (
	"context"
	"fmt"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LogMsgQuery interface {
	whereWorkflow(workflowId uuid.UUID)
	whereNamespace(namespaceId uuid.UUID)
	whereInstance(instanceId uuid.UUID)
	whereRootInstanceIdEQ(rootId string)
	whereInstanceCallPathHasPrefix(prefix string)
	whereLogLevel(loglevel string)
	whereWorkflowIsNil()
	whereNamespaceIsNIl()
	whereInstanceIsNIl()
	whereMinumLogLevel(loglevel string)
	whereMirrorActivityID(id uuid.UUID)
	withLimit(limit int)
	withOffset(offset int)
	// GetServerLogs(ctx context.Context) []*LogMsgs
	getLimit() int
	getOffset() int
	getAll(ctx context.Context, db *gorm.DB) ([]*LogMsgs, error)
}

type LogMsgRepo interface {
	QueryLogs(ctx context.Context, ql LogMsgQuery) ([]*LogMsgs, error)
	create(l *logMsg) error
}

// type LogMsgStorer interface {
// 	// QueryLogs(ctx context.Context, db *gorm.DB) LogMsgQuery
// 	create(ctx context.Context, l logMessage)
// }

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

func (b *LogMsgQueryBuilder) whereWorkflow(workflowId uuid.UUID) {
	wEq := b.whereEQStatements
	wEq = append(wEq, fmt.Sprintf("workflow_id = '%s'", workflowId.String()))
	b.whereEQStatements = wEq
}

func (b *LogMsgQueryBuilder) whereWorkflowIsNil() {
	wEq := b.whereEQStatements
	wEq = append(wEq, "workflow_id IS NULL")
	b.whereEQStatements = wEq
}

func (b *LogMsgQueryBuilder) whereNamespaceIsNIl() {
	wEq := b.whereEQStatements
	wEq = append(wEq, "namespace_logs IS NULL")
	b.whereEQStatements = wEq
}

func (b *LogMsgQueryBuilder) whereInstanceIsNIl() {
	wEq := b.whereEQStatements
	wEq = append(wEq, "instance_logs IS NULL")
	b.whereEQStatements = wEq
}

func (b *LogMsgQueryBuilder) whereNamespace(namespaceId uuid.UUID) {
	wEq := b.whereEQStatements
	wEq = append(wEq, fmt.Sprintf("namespace_logs = '%s'", namespaceId.String()))
	b.whereEQStatements = wEq
}

func (b *LogMsgQueryBuilder) whereInstance(instanceId uuid.UUID) {
	wEq := b.whereEQStatements
	wEq = append(wEq, fmt.Sprintf("instance_logs = '%s'", instanceId.String()))
	b.whereEQStatements = wEq
}

func (b *LogMsgQueryBuilder) whereRootInstanceIdEQ(rootId string) {
	wEq := b.whereEQStatements
	wEq = append(wEq, fmt.Sprintf("root_instance_id = '%s'", rootId))
	b.whereEQStatements = wEq
}

func (b *LogMsgQueryBuilder) whereInstanceCallPathHasPrefix(prefix string) {
	wEq := b.whereEQStatements
	wEq = append(wEq, fmt.Sprintf("log_instance_call_path like '%s%%'", prefix))
	b.whereEQStatements = wEq
}

func (b *LogMsgQueryBuilder) whereLogLevel(loglevel string) {
	wEq := b.whereEQStatements
	wEq = append(wEq, fmt.Sprintf("level = '%s'", loglevel))
	b.whereEQStatements = wEq
}

func (b *LogMsgQueryBuilder) whereMirrorActivityID(id uuid.UUID) {
	wEq := b.whereEQStatements
	wEq = append(wEq, fmt.Sprintf("mirror_activity_id = '%s'", id.String()))
	b.whereEQStatements = wEq
}

func (b *LogMsgQueryBuilder) whereMinumLogLevel(loglevel string) {
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

func (b *LogMsgQueryBuilder) withLimit(limit int) {
	b.limit = limit
}

func (b *LogMsgQueryBuilder) withOffset(offset int) {
	b.offset = offset
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

func (b *LogMsgQueryBuilder) getAll(ctx context.Context, db *gorm.DB) ([]*LogMsgs, error) {
	query, err := b.build()
	if err != nil {
		return nil, err
	}
	resultList := make([]*LogMsgs, 0)
	if db == nil {
		return nil, fmt.Errorf("db was nil")
	}
	res := db.WithContext(ctx).Raw(query).Scan(&resultList)
	if res.Error != nil {
		return nil, res.Error
	}

	return resultList, nil
}

func (b *LogMsgQueryBuilder) getLimit() int {
	return b.limit
}

func (b *LogMsgQueryBuilder) getOffset() int {
	return b.offset
}


func buildPageInfo(lq LogMsgQuery) grpc.PageInfo {
	return grpc.PageInfo{
		Limit:  int32(lq.getLimit()),
		Offset: int32(lq.getOffset()),
	}
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
