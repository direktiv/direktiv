package internallogger

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/database/entwrapper"
	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Logger struct {
	logQueue     chan *logMessage
	logWorkersWG sync.WaitGroup
	sugar        *zap.SugaredLogger
	edb          *entwrapper.Database // TODO: remove
	pubsub       *pubsub.Pubsub
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
	level       string
	recipientID uuid.UUID
	tags        map[string]string
	ctx         context.Context
}

func (logger *Logger) StartLogWorkers(n int, db *entwrapper.Database, pubsub *pubsub.Pubsub, sugar *zap.SugaredLogger) {
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

func (logger *Logger) LogToServer(ctx context.Context, t time.Time, msg string, a ...interface{}) {
	defer func() {
		_ = recover()
	}()

	logger.logQueue <- &logMessage{
		t:   t,
		msg: fmt.Sprintf(msg, a...),
	}
}

func (logger *Logger) Debug(ctx context.Context, t time.Time, recipientID uuid.UUID, tags map[string]string, msg string) {
	logger.sendToWorker(ctx, t, recipientID, tags, "debug", msg)
}

func (logger *Logger) Debugf(ctx context.Context, t time.Time, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
	logger.sendToWorker(ctx, t, recipientID, tags, "debug", fmt.Sprintf(msg, a...))
}

func (logger *Logger) Info(ctx context.Context, t time.Time, recipientID uuid.UUID, tags map[string]string, msg string) {
	logger.sendToWorker(ctx, t, recipientID, tags, "info", msg)
}

func (logger *Logger) Infof(ctx context.Context, t time.Time, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
	logger.sendToWorker(ctx, t, recipientID, tags, "info", fmt.Sprintf(msg, a...))
}

func (logger *Logger) Error(ctx context.Context, t time.Time, recipientID uuid.UUID, tags map[string]string, msg string) {
	logger.sendToWorker(ctx, t, recipientID, tags, "error", msg)
}

func (logger *Logger) Errorf(ctx context.Context, t time.Time, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
	logger.sendToWorker(ctx, t, recipientID, tags, "error", fmt.Sprintf(msg, a...))
}

func (logger *Logger) Panic(ctx context.Context, t time.Time, recipientID uuid.UUID, tags map[string]string, msg string) {
	logger.sendToWorker(ctx, t, recipientID, tags, "panic", msg)
}

func (logger *Logger) Panicf(ctx context.Context, t time.Time, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
	logger.sendToWorker(ctx, t, recipientID, tags, "panic", fmt.Sprintf(msg, a...))
}

func (logger *Logger) sendToWorker(ctx context.Context, t time.Time, recipientID uuid.UUID, tags map[string]string, level string, msg string) {
	defer func() {
		_ = recover()
	}()

	logger.logQueue <- &logMessage{
		t:           t,
		msg:         msg,
		tags:        tags,
		level:       level,
		recipientID: recipientID,
		ctx:         ctx,
	}
}

func (logger *Logger) telemetry(ctx context.Context, msg string, tags map[string]string) {
	span := trace.SpanFromContext(ctx)
	tid := span.SpanContext().TraceID()
	if tags == nil {
		logger.sugar.Infow(msg, "trace", tid)
	} else {
		ar := make([]interface{}, 0, len(tags)*2+1)
		for k, v := range tags {
			ar = append(ar, k, v)
		}
		ar = append(ar, "trace", tid)
		switch tags["level"] {
		case "info":
			logger.sugar.Infow(msg, ar...)
		case "debug":
			logger.sugar.Debugw(msg, ar...)
		case "error":
			logger.sugar.Errorw(msg, ar...)
		case "panic":
			logger.sugar.DPanicw(msg, ar...)
		}
	}
}

func (logger *Logger) SendLogMsgToDB(l *logMessage) error {
	ctx := context.Background() // logs are often queued and stored after their originating requests have ended.
	clients := logger.edb.Clients(ctx)
	lc := clients.LogMsg.Create().SetMsg(l.msg).SetT(l.t).SetLevel(l.level).SetTags(l.tags)

	switch l.tags["recipientType"] {
	case "server":
	case "instance":
		callpath := AppendInstanceID(l.tags["callpath"], l.recipientID.String())
		rootInstance, err := GetRootinstanceID(callpath)
		if err != nil {
			return err
		}
		lc.SetInstanceID(l.recipientID).SetRootInstanceId(rootInstance).SetLogInstanceCallPath(callpath)
	case "namespace":
		lc.SetNamespaceID(l.recipientID)
	case "workflow":
		lc.SetWorkflowID(l.recipientID)
	case "mirror":
		lc.SetActivityID(l.recipientID)
	default:
		logger.sugar.Panicf("recipientType was not set", l.msg, l.tags)
		panic("how?")
	}
	_, err := lc.Save(ctx)
	if err != nil {
		logger.sugar.Panicf("error storing logmsg", err)
		return err
	}
	logger.telemetry(l.ctx, l.msg, l.tags)
	logger.pubsub.NotifyLogs(l.recipientID, l.tags["recipientType"])
	return nil
}
