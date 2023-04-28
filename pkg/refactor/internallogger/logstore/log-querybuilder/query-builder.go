package logquerybuilder

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type LogMsgQuery interface {
	Build() (string, error)
}

type LogMsgQueryBuilder struct {
	whereEQStatements []string
	limit             int
	offset            int
}

func newQueryBuilder() *LogMsgQueryBuilder {
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

func (b *LogMsgQueryBuilder) Build() (string, error) {
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

// func (b *LogMsgQueryBuilder) getAll(ctx context.Context, db *gorm.DB) ([]*LogMsgs, error) {
// 	query, err := b.build()
// 	if err != nil {
// 		return nil, err
// 	}
// 	resultList := make([]*LogMsgs, 0)
// 	if db == nil {
// 		return nil, fmt.Errorf("db was nil")
// 	}
// 	res := db.WithContext(ctx).Raw(query).Scan(&resultList)
// 	if res.Error != nil {
// 		return nil, res.Error
// 	}

// 	return resultList, nil
// }

func (b *LogMsgQueryBuilder) getLimit() int {
	return b.limit
}

func (b *LogMsgQueryBuilder) getOffset() int {
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

func GetInstanceLogsNoInheritance(instanceID uuid.UUID, limit, offset int) LogMsgQuery {
	ql := newQueryBuilder()

	ql.whereInstance(instanceID)

	if limit > 0 {
		ql.withLimit(limit)
	}
	if offset > 0 {
		ql.withOffset(offset)
	}
	// l, err := ql.getAll(ctx, ls.db)
	// if err != nil {
	// 	return nil, err
	// }
	return ql
}

func GetServerLogs(limit, offset int) LogMsgQuery {
	ql := newQueryBuilder()
	ql.whereWorkflowIsNil()
	ql.whereNamespaceIsNIl()
	ql.whereInstanceIsNIl()
	if limit > 0 {
		ql.withLimit(limit)
	}
	if offset > 0 {
		ql.withOffset(offset)
	}
	// logs, err := ql.getAll(ctx, ls.db)
	// if err != nil {
	// 	return nil, err
	// }
	return ql
}

func GetNamespaceLogs(namespaceID uuid.UUID, limit, offset int) LogMsgQuery {
	ql := newQueryBuilder()
	id := namespaceID
	ql.whereNamespace(id)
	if limit > 0 {
		ql.withLimit(limit)
	}
	if offset > 0 {
		ql.withOffset(offset)
	}
	// logs, err := ql.getAll(ctx, ls.db)
	// if err != nil {
	// 	return nil, err
	// }
	return ql
}

func GetWorkflowLogs(workflowID uuid.UUID, limit, offset int) LogMsgQuery {
	ql := newQueryBuilder()
	id := workflowID
	ql.whereWorkflow(id)
	if limit > 0 {
		ql.withLimit(limit)
	}
	if offset > 0 {
		ql.withOffset(offset)
	}
	// logs, err := ql.getAll(ctx, ls.db)
	// if err != nil {
	// 	return nil, err
	// }
	return ql
}

func GetMirrorActivityLogs(mirror uuid.UUID, limit, offset int) LogMsgQuery {
	ql := newQueryBuilder()
	id := mirror
	ql.whereMirrorActivityID(id)
	if limit > 0 {
		ql.withLimit(limit)
	}
	if offset > 0 {
		ql.withOffset(offset)
	}
	// logs, err := ql.getAll(ctx, ls.db)
	// if err != nil {
	// 	return nil, err
	// }
	return ql
}

func GetInstanceLogs(callpath, instanceID string, limit, offset int) (LogMsgQuery, error) {
	prefix := AppendInstanceID(callpath, instanceID)
	root, err := GetRootinstanceID(prefix)
	if err != nil {
		return nil, err
	}
	callerIsRoot, err := IsCallerRoot(callpath, instanceID)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	ql := newQueryBuilder()

	ql.whereRootInstanceIdEQ(root)
	if !callerIsRoot {
		ql.whereInstanceCallPathHasPrefix(prefix)
	}

	if limit > 0 {
		ql.withLimit(limit)
	}
	if offset > 0 {
		ql.withOffset(offset)
	}
	return ql, err
}

func IsCallerRoot(callpath, instanceID string) (bool, error) {
	prefix := AppendInstanceID(callpath, instanceID)
	root, err := GetRootinstanceID(prefix)
	if err != nil {
		return false, err
	}
	return root == instanceID, nil
}

// Appends a InstanceID to the InstanceCallPath.
func AppendInstanceID(callpath, instanceID string) string {
	if callpath == "/" {
		return "/" + instanceID
	}
	return callpath + "/" + instanceID
}

// Extracts the rootInstanceID from a callpath.
// Forexpl. /c1d87df6-56fb-4b03-a9e9-00e5122e4884/105cbf37-76b9-452a-b67d-5c9a8cd54ecc.
// The callpath has to contain a rootInstanceID as first element. In this case the rootInstanceID would be
// c1d87df6-56fb-4b03-a9e9-00e5122e4884.
func GetRootinstanceID(callpath string) (string, error) {
	path := strings.Split(callpath, "/")
	if len(path) < 2 {
		return "", errors.New("instance Callpath is malformed")
	}
	_, err := uuid.Parse(path[1])
	if err != nil {
		return "", err
	}
	return path[1], nil
}
