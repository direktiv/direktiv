package internallogger

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/database/entwrapper"
	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Logger struct {
	logQueue     chan *logMessage
	logWorkersWG sync.WaitGroup
	sugar        *zap.SugaredLogger
	edb          *entwrapper.Database
	pubsub       LogNotify
}

type LogNotify interface {
	NotifyLogs(recipientID uuid.UUID, recipientType recipient.RecipientType)
}

func InitLogger() *Logger {
	logQueue := make(chan *logMessage, 1000)
	return &Logger{
		logQueue: logQueue,
	}
}

type logMessage struct {
	t           time.Time
	msg         string
	level       Level
	recipientID uuid.UUID
	tags        map[string]string
}

func (logger *Logger) StartLogWorkers(n int, db *entwrapper.Database, pubsub LogNotify, sugar *zap.SugaredLogger) {
	logger.edb = db
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
		_ = logger.SendLogMsgToDB(l)
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

// Appends a InstanceID to the InstanceCallPath.
func AppendInstanceID(callpath, instanceID string) string {
	if callpath == "/" {
		return "/" + instanceID
	}
	return callpath + "/" + instanceID
}

func (logger *Logger) Debug(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string) {
	logger.Telemetry(ctx, Debug, tags, msg)
	logger.sendToWorker(recipientID, tags, Debug, msg)
}

func (logger *Logger) Debugf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
	msg = fmt.Sprintf(msg, a...)
	logger.Telemetry(ctx, Debug, tags, msg)
	logger.sendToWorker(recipientID, tags, Debug, msg)
}

func (logger *Logger) Info(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string) {
	logger.Telemetry(ctx, Info, tags, msg)
	logger.sendToWorker(recipientID, tags, Info, msg)
}

func (logger *Logger) Infof(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
	msg = fmt.Sprintf(msg, a...)
	logger.Telemetry(ctx, Info, tags, msg)
	logger.sendToWorker(recipientID, tags, Info, msg)
}

func (logger *Logger) Error(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string) {
	logger.Telemetry(ctx, Error, tags, msg)
	logger.sendToWorker(recipientID, tags, Error, msg)
}

func (logger *Logger) Errorf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
	msg = fmt.Sprintf(msg, a...)
	logger.Telemetry(ctx, Error, tags, msg)
	logger.sendToWorker(recipientID, tags, Error, msg)
}

func (logger *Logger) Panic(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string) {
	logger.Telemetry(ctx, Panic, tags, msg)
	logger.sendToWorker(recipientID, tags, Panic, msg)
}

func (logger *Logger) Panicf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
	msg = fmt.Sprintf(msg, a...)
	logger.Telemetry(ctx, Panic, tags, msg)
	logger.sendToWorker(recipientID, tags, Panic, msg)
}

func (logger *Logger) sendToWorker(recipientID uuid.UUID, tags map[string]string, level Level, msg string) {
	defer func() {
		_ = recover()
	}()

	logger.logQueue <- &logMessage{
		t:           time.Now(),
		msg:         msg,
		tags:        tags,
		level:       level,
		recipientID: recipientID,
	}
}

func (logger *Logger) Telemetry(ctx context.Context, level Level, tags map[string]string, msg string) {
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

func (logger *Logger) SendLogMsgToDB(l *logMessage) error {
	ctx := context.Background() // logs are often queued and stored after their originating requests have ended.
	clients := logger.edb.Clients(ctx)
	msg := strings.ReplaceAll(l.msg, "\u0000", "")
	lc := clients.LogMsg.Create().SetMsg(msg).SetT(l.t).SetLevel(string(l.level)).SetTags(l.tags)

	recipientType, ok := recipient.Convert(l.tags["recipientType"])
	if !ok {
		return fmt.Errorf("unexpected recipientType %s", recipientType)
	}
	switch recipientType {
	case recipient.Server:
	case recipient.Instance:
		callpath := AppendInstanceID(l.tags["callpath"], l.recipientID.String())
		rootInstance, err := GetRootinstanceID(callpath)
		if err != nil {
			return err
		}
		lc.SetInstanceID(l.recipientID).SetRootInstanceId(rootInstance).SetLogInstanceCallPath(callpath)
	case recipient.Namespace:
		lc.SetNamespaceID(l.recipientID)
	case recipient.Workflow:
		lc.SetWorkflowID(l.recipientID)
	case recipient.Mirror:
		lc.SetMirrorActivityID(l.recipientID)
	default:
		logger.sugar.Panicf("recipientType was not set: %s %v", l.msg, l.tags)
		return fmt.Errorf("recipientType was not set %s %v", l.msg, l.tags)
	}
	_, err := lc.Save(ctx)
	if err != nil {
		logger.sugar.Panicf("error storing logmsg: %v", err)
		return err
	}
	logger.pubsub.NotifyLogs(l.recipientID, recipientType)
	return nil
}
