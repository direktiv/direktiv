package internallogger

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/refactor/internallogger/logstore"
	logquerybuilder "github.com/direktiv/direktiv/pkg/refactor/internallogger/logstore/log-querybuilder"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Logger struct {
	logQueue     chan *queuedLogMsg
	logWorkersWG sync.WaitGroup
	sugar        *zap.SugaredLogger
	pubsub       LogNotify
	db           *gorm.DB
}

type queuedLogMsg struct {
	logstore.LogMsg
	recipientID   uuid.UUID
	recipientType recipient.RecipientType
}

type LogNotify interface {
	NotifyLogs(recipientID uuid.UUID, recipientType recipient.RecipientType)
}

func InitLogger() *Logger {
	logQueue := make(chan *queuedLogMsg, 1000)
	return &Logger{
		logQueue: logQueue,
	}
}

func (logger *Logger) StartLogWorkers(n int, db *gorm.DB, pubsub LogNotify, sugar *zap.SugaredLogger) {
	logger.db = db
	logger.pubsub = pubsub
	logger.sugar = sugar
	logger.logWorkersWG.Add(n)
	for i := 0; i < n; i++ {
		go logger.logWorker()
	}
}

func (logger *Logger) logWorker() {
	defer logger.logWorkersWG.Done()

	for {
		l, more := <-logger.logQueue
		if !more {
			return
		}
		_ = logger.create(l)
	}
}

func (logger *Logger) CloseLogWorkers() {
	close(logger.logQueue)
	logger.logWorkersWG.Wait()
}

// Extracts the rootInstanceID from a callpath.
// Forexpl. /c1d87df6-56fb-4b03-a9e9-00e5122e4884/105cbf37-76b9-452a-b67d-5c9a8cd54ecc.
// The callpath has to contain a rootInstanceID as first element. In this case the rootInstanceID would be
// c1d87df6-56fb-4b03-a9e9-00e5122e4884.
func getRootinstanceID(callpath string) (string, error) {
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

// Appends a InstanceID to the InstanceCallPath.
func AppendInstanceID(callpath, instanceID string) string {
	if callpath == "/" {
		return "/" + instanceID
	}
	return callpath + "/" + instanceID
}

func IsCallerRoot(callpath, instanceID string) (bool, error) {
	prefix := AppendInstanceID(callpath, instanceID)
	root, err := getRootinstanceID(prefix)
	if err != nil {
		return false, err
	}
	return root == instanceID, nil
}

func (logger *Logger) Debug(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string) {
	logger.telemetry(ctx, Debug, tags, msg)
	logger.sendToWorker(recipientID, tags, Debug, msg)
}

func (logger *Logger) Debugf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
	msg = fmt.Sprintf(msg, a...)
	logger.telemetry(ctx, Debug, tags, msg)
	logger.sendToWorker(recipientID, tags, Debug, msg)
}

func (logger *Logger) Info(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string) {
	logger.telemetry(ctx, Info, tags, msg)
	logger.sendToWorker(recipientID, tags, Info, msg)
}

func (logger *Logger) Infof(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
	msg = fmt.Sprintf(msg, a...)
	logger.telemetry(ctx, Info, tags, msg)
	logger.sendToWorker(recipientID, tags, Info, msg)
}

func (logger *Logger) Error(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string) {
	logger.telemetry(ctx, Error, tags, msg)
	logger.sendToWorker(recipientID, tags, Error, msg)
}

func (logger *Logger) Errorf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
	msg = fmt.Sprintf(msg, a...)
	logger.telemetry(ctx, Error, tags, msg)
	logger.sendToWorker(recipientID, tags, Error, msg)
}

func (logger *Logger) Panic(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string) {
	logger.telemetry(ctx, Panic, tags, msg)
	logger.sendToWorker(recipientID, tags, Panic, msg)
}

func (logger *Logger) Panicf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
	msg = fmt.Sprintf(msg, a...)
	logger.telemetry(ctx, Panic, tags, msg)
	logger.sendToWorker(recipientID, tags, Panic, msg)
}

func (logger *Logger) sendToWorker(recipientID uuid.UUID, tags map[string]string, level Level, msg string) {
	defer func() {
		_ = recover()
	}()
	recipientType, ok := recipient.Convert(tags["recipientType"])
	if !ok {
		panic(fmt.Errorf("unexpected recipientType %s", recipientType))
	}
	lTags := make(map[string]interface{})
	for k, v := range tags {
		lTags[k] = v
	}
	l := &queuedLogMsg{
		LogMsg: logstore.LogMsg{
			T:     time.Now(),
			Msg:   msg,
			Tags:  lTags,
			Level: string(level),
		},
		recipientType: recipientType,
		recipientID:   recipientID,
	}

	logger.logQueue <- l
}

func (logger *Logger) telemetry(ctx context.Context, level Level, tags map[string]string, msg string) {
	span := trace.SpanFromContext(ctx)
	tid := span.SpanContext().TraceID()
	if len(tags) == 0 {
		logger.sugar.Infow(msg, "trace", tid)
	} else {
		ar := make([]interface{}, len(tags)*2)
		i := 0
		for k, v := range tags {
			ar[i] = k
			ar[i+1] = v
			i = i + 2
		}
		ar = append(ar, "trace", tid)
		switch level {
		case Info:
			logger.sugar.Infow(msg, ar...)
		case Debug:
			logger.sugar.Debugw(msg, ar...)
		case Error:
			logger.sugar.Errorw(msg, ar...)
		case Panic:
			logger.sugar.Panicw(msg, ar...)
		default:
			logger.sugar.Debugw(msg, ar...) // this should never happen
		}
	}
}

func (logger *Logger) create(l *queuedLogMsg) error {
	// TODO: create
	logger.pubsub.NotifyLogs(l.recipientID, l.recipientType)
	return logstore.NewLogStore(logger.db).Create(l.recipientID, l.recipientType, l.LogMsg)
}

func (logger *Logger) Store() (func(ctx context.Context, ql logquerybuilder.LogMsgQuery) ([]*logstore.LogMsg, error), error) {
	if logger.db == nil {
		return nil, fmt.Errorf("Logger was not properly initialized")
	}
	return logstore.NewLogStore(logger.db).QueryLogs, nil
}
