package logstore

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	logquerybuilder "github.com/direktiv/direktiv/pkg/refactor/internallogger/logstore/log-querybuilder"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LogMsgRepo interface {
	QueryLogs(ctx context.Context, ql logquerybuilder.LogMsgQuery) ([]*LogMsg, error)
	Create(recipientID uuid.UUID, recipientType recipient.RecipientType, l LogMsg) error
}

type LogMsg struct {
	T     time.Time
	Msg   string
	Level string
	Tags  JSONB
}

type gormLogMsg struct {
	Oid                 uuid.UUID `gorm:"primaryKey"`
	T                   time.Time
	Msg                 string
	Level               string
	Tags                JSONB     `sql:"type:jsonb"`
	WorkflowId          uuid.UUID `gorm:"default:null"`
	MirrorActivityId    uuid.UUID `gorm:"default:null"`
	InstanceLogs        uuid.UUID `gorm:"default:null"`
	NamespaceLogs       uuid.UUID `gorm:"default:null"`
	RootInstanceId      string    `gorm:"default:null"`
	LogInstanceCallPath string
}

type LogStore struct {
	db *gorm.DB
}

func NewLogStore(db *gorm.DB) *LogStore {
	return &LogStore{
		db: db,
	}
}

func (ls *LogStore) QueryLogs(ctx context.Context, ql logquerybuilder.LogMsgQuery) ([]*LogMsg, error) {
	query, err := ql.Build()
	if err != nil {
		return nil, err
	}
	res, err := ls.getAll(ctx, ls.db, query)
	if err != nil {
		return nil, err
	}
	resultList := make([]*LogMsg, 0)
	for _, l := range res {
		resultList = append(resultList,
			&LogMsg{
				T:     l.T,
				Msg:   l.Msg,
				Level: l.Level,
				Tags:  l.Tags,
			})
	}
	return resultList, nil
}

func (ls *LogStore) getAll(ctx context.Context, db *gorm.DB, query string) ([]*gormLogMsg, error) {
	resultList := make([]*gormLogMsg, 0)
	if db == nil {
		return nil, fmt.Errorf("db was nil")
	}
	res := db.WithContext(ctx).Raw(query).Scan(&resultList)
	if res.Error != nil {
		return nil, res.Error
	}

	return resultList, nil
}

func (ls *LogStore) Create(recipientID uuid.UUID, recipientType recipient.RecipientType, log LogMsg) error {
	l := gormLogMsg{
		Oid:   uuid.New(),
		T:     log.T,
		Msg:   log.Msg,
		Level: log.Level,
		Tags:  log.Tags,
	}
	switch recipientType {
	case recipient.Server:
	case recipient.Instance:
		callpath := AppendInstanceID(fmt.Sprintf("%s", log.Tags["callpath"]), recipientID.String())
		rootInstance, err := GetRootinstanceID(callpath)
		if err != nil {
			panic(err)
		}
		l.InstanceLogs = recipientID
		l.RootInstanceId = rootInstance
		l.LogInstanceCallPath = callpath
	case recipient.Namespace:
		l.NamespaceLogs = recipientID
	case recipient.Workflow:
		l.WorkflowId = recipientID
	case recipient.Mirror:
		l.MirrorActivityId = recipientID
	default:
		return fmt.Errorf("recipientType %s is not implemented", recipientType)
	}
	tx := ls.db.Table("log_msgs").Create(&l)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
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
