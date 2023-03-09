package internallogger

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/entwrapper"
	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	"github.com/direktiv/direktiv/pkg/util"
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
	ctx    context.Context //nolint:containedctx
	t      time.Time
	msg    string
	cached *database.CacheData
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

		if l.cached == nil {
			logger.workerLogToServer(l)
		} else if l.cached.Instance != nil {
			logger.workerLogToInstance(l)
		} else if l.cached.Workflow != nil {
			logger.workerLogToWorkflow(l)
		} else if l.cached.Namespace != nil {
			logger.workerLogToNamespace(l)
		} else {
			panic("how?")
		}

	}
}

func (logger *Logger) CloseLogWorkers() {
	close(logger.logQueue)
	logger.logWorkersWG.Wait()
}

func (logger *Logger) workerLogToServer(l *logMessage) {
	util.Trace(l.ctx, l.msg)

	clients := logger.edb.Clients(context.Background())

	_, err := clients.LogMsg.Create().SetMsg(l.msg).SetT(l.t).Save(context.Background())
	if err != nil {
		logger.sugar.Error(err)
		return
	}

	span := trace.SpanFromContext(l.ctx)
	tid := span.SpanContext().TraceID()

	logger.sugar.Infow(l.msg, "trace", tid)

	logger.pubsub.NotifyServerLogs()
}

func (logger *Logger) workerLogToNamespace(l *logMessage) {
	util.Trace(l.ctx, l.msg)

	clients := logger.edb.Clients(context.Background())

	_, err := clients.LogMsg.Create().SetMsg(l.msg).SetNamespaceID(l.cached.Namespace.ID).SetT(l.t).Save(l.ctx)
	if err != nil {
		logger.sugar.Error(err)
		return
	}

	span := trace.SpanFromContext(l.ctx)
	tid := span.SpanContext().TraceID()

	logger.sugar.Infow(l.msg, "trace", tid, "namespace", l.cached.Namespace.Name, "namespace-id", l.cached.Namespace.ID.String())

	logger.pubsub.NotifyNamespaceLogs(l.cached.Namespace)
}

func (logger *Logger) workerLogToWorkflow(l *logMessage) {
	util.Trace(l.ctx, l.msg)

	clients := logger.edb.Clients(context.Background())

	_, err := clients.LogMsg.Create().SetMsg(l.msg).SetWorkflowID(l.cached.Workflow.ID).SetT(l.t).Save(l.ctx)
	if err != nil {
		logger.sugar.Error(err)
		return
	}

	span := trace.SpanFromContext(l.ctx)
	tid := span.SpanContext().TraceID()

	logger.sugar.Infow(l.msg, "trace", tid, "namespace", l.cached.Namespace.Name, "namespace-id", l.cached.Namespace.ID.String(), "workflow-id", l.cached.Workflow.ID.String() /*"workflow", GetInodePath(l.cached.Path())*/)

	logger.pubsub.NotifyWorkflowLogs(l.cached.Workflow)
}

func (logger *Logger) workerLogToInstance(l *logMessage) {
	util.Trace(l.ctx, l.msg)

	ctx := context.Background() // logs are often queued and stored after their originating requests have ended.

	clients := logger.edb.Clients(ctx)

	callpath := AppendInstanceID(l.cached.Instance.CallPath, l.cached.Instance.ID.String())
	rootInstance, err := GetRootinstanceID(callpath)
	if err != nil {
		logger.sugar.Error(err)
		return
	}
	_, err = clients.LogMsg.Create().SetMsg(l.msg).SetInstanceID(l.cached.Instance.ID).SetT(l.t).SetRootInstanceId(rootInstance).SetLogInstanceCallPath(callpath).Save(ctx)
	if err != nil {
		logger.sugar.Error(err)
		return
	}

	span := trace.SpanFromContext(l.ctx)
	tid := span.SpanContext().TraceID()

	nsid := ""
	nsname := ""
	if l.cached.Namespace != nil {
		nsid = l.cached.Namespace.ID.String()
		nsname = l.cached.Namespace.Name
	}

	wfid := ""
	if l.cached.Workflow != nil {
		wfid = l.cached.Workflow.ID.String()
	}

	logger.sugar.Infow(l.msg, "trace", tid, "namespace", nsname, "namespace-id", nsid, "workflow-id", wfid, "workflow" /*GetInodePath(l.cached.Instance.As),*/, "instance", l.cached.Instance.ID.String())

	logger.pubsub.NotifyInstanceLogs(l.cached.Instance)
}

// Extracts the rootInstanceID from a callpath.
// Forexpl. /c1d87df6-56fb-4b03-a9e9-00e5122e4884/105cbf37-76b9-452a-b67d-5c9a8cd54ecc.
// The callpath has to contain a rootInstanceID as first element. In this case the rootInstanceID would be
// c1d87df6-56fb-4b03-a9e9-00e5122e4884.
func GetRootinstanceID(callpath string) (string, error) {
	path := strings.Split(callpath, "/")
	if len(path) < 2 {
		return "", errors.New("Instance Callpath is malformed")
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
		ctx: ctx,
		t:   t,
		msg: fmt.Sprintf(msg, a...),
	}
}

func (logger *Logger) LogToNamespace(ctx context.Context, t time.Time, cached *database.CacheData, msg string, a ...interface{}) {
	defer func() {
		_ = recover()
	}()

	cd := *cached // We do this to zero some fields without modifying the argument.
	cd.Workflow = nil
	cd.Instance = nil

	logger.logQueue <- &logMessage{
		ctx:    ctx,
		t:      t,
		msg:    fmt.Sprintf(msg, a...),
		cached: &cd,
	}
}

func (logger *Logger) LogToWorkflow(ctx context.Context, t time.Time, cached *database.CacheData, msg string, a ...interface{}) {
	defer func() {
		_ = recover()
	}()

	cd := *cached // We do this to zero some fields without modifying the argument.
	cd.Workflow = nil

	logger.logQueue <- &logMessage{
		ctx:    ctx,
		t:      t,
		msg:    fmt.Sprintf(msg, a...),
		cached: &cd,
	}
}

// log To instance with string interpolation.
func (logger *Logger) LogToInstance(ctx context.Context, t time.Time, cached *database.CacheData, msg string, a ...interface{}) {
	msg = fmt.Sprintf(msg, a...)

	logger.LogToInstanceRaw(ctx, t, cached, msg)
}

// log To instance with raw string.
func (logger *Logger) LogToInstanceRaw(ctx context.Context, t time.Time, cached *database.CacheData, msg string) {
	defer func() {
		_ = recover()
	}()

	logger.logQueue <- &logMessage{
		ctx:    ctx,
		t:      t,
		msg:    msg,
		cached: cached,
	}
}

func (logger *Logger) LogToMirrorActivity(ctx context.Context, t time.Time, ns *database.Namespace, mirror *database.Mirror, act *database.MirrorActivity, msg string, a ...interface{}) {
	msg = fmt.Sprintf(msg, a...)

	util.Trace(ctx, msg)

	clients := logger.edb.Clients(ctx)

	_, err := clients.LogMsg.Create().SetMsg(msg).SetActivityID(act.ID).SetT(t).Save(ctx)
	if err != nil {
		logger.sugar.Error(err)
		return
	}

	span := trace.SpanFromContext(ctx)
	tid := span.SpanContext().TraceID()

	logger.sugar.Infow(msg, "trace", tid, "namespace", ns.Name, "namespace-id", ns.ID.String(), "mirror-id", mirror.ID.String())

	logger.pubsub.NotifyMirrorActivityLogs(act)
}
