package flow

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type logMessage struct {
	ctx    context.Context //nolint:containedctx
	t      time.Time
	msg    string
	cached *database.CacheData
}

func (srv *server) startLogWorkers(n int) {
	srv.logWorkersWG.Add(n)
	for i := 0; i < n; i++ {
		go srv.logWorker()
	}
}

func (srv *server) logWorker() {

	defer srv.logWorkersWG.Done()

	for {

		l, more := <-srv.logQueue
		if !more {
			return
		}

		if l.cached == nil {
			srv.workerLogToServer(l)
		} else if l.cached.Instance != nil {
			srv.workerLogToInstance(l)
		} else if l.cached.Workflow != nil {
			srv.workerLogToWorkflow(l)
		} else if l.cached.Namespace != nil {
			srv.workerLogToNamespace(l)
		} else {
			panic("how?")
		}

	}

}

func (srv *server) closeLogWorkers() {
	close(srv.logQueue)
	srv.logWorkersWG.Wait()
}

func (srv *server) workerLogToServer(l *logMessage) {

	util.Trace(l.ctx, l.msg)

	clients := srv.edb.Clients(nil)

	_, err := clients.LogMsg.Create().SetMsg(l.msg).SetT(l.t).Save(context.Background())
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	span := trace.SpanFromContext(l.ctx)
	tid := span.SpanContext().TraceID()

	srv.sugar.Infow(l.msg, "trace", tid)

	srv.pubsub.NotifyServerLogs()

}

func (srv *server) workerLogToNamespace(l *logMessage) {

	util.Trace(l.ctx, l.msg)

	clients := srv.edb.Clients(nil)

	_, err := clients.LogMsg.Create().SetMsg(l.msg).SetNamespaceID(l.cached.Namespace.ID).SetT(l.t).Save(l.ctx)
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	span := trace.SpanFromContext(l.ctx)
	tid := span.SpanContext().TraceID()

	srv.sugar.Infow(l.msg, "trace", tid, "namespace", l.cached.Namespace.Name, "namespace-id", l.cached.Namespace.ID.String())

	srv.pubsub.NotifyNamespaceLogs(l.cached.Namespace)

}

func (srv *server) workerLogToWorkflow(l *logMessage) {

	util.Trace(l.ctx, l.msg)

	clients := srv.edb.Clients(nil)

	_, err := clients.LogMsg.Create().SetMsg(l.msg).SetWorkflowID(l.cached.Workflow.ID).SetT(l.t).Save(l.ctx)
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	span := trace.SpanFromContext(l.ctx)
	tid := span.SpanContext().TraceID()

	srv.sugar.Infow(l.msg, "trace", tid, "namespace", l.cached.Namespace.Name, "namespace-id", l.cached.Namespace.ID.String(), "workflow-id", l.cached.Workflow.ID.String(), "workflow", GetInodePath(l.cached.Path()))

	srv.pubsub.NotifyWorkflowLogs(l.cached.Workflow)

}

func (srv *server) workerLogToInstance(l *logMessage) {

	util.Trace(l.ctx, l.msg)

	clients := srv.edb.Clients(nil)

	_, err := clients.LogMsg.Create().SetMsg(l.msg).SetInstanceID(l.cached.Instance.ID).SetT(l.t).Save(l.ctx)
	if err != nil {
		srv.sugar.Error(err)
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

	srv.sugar.Infow(l.msg, "trace", tid, "namespace", nsname, "namespace-id", nsid, "workflow-id", wfid, "workflow", GetInodePath(l.cached.Instance.As), "instance", l.cached.Instance.ID.String())

	srv.pubsub.NotifyInstanceLogs(l.cached.Instance)

}

func (srv *server) logToServer(ctx context.Context, t time.Time, msg string, a ...interface{}) {

	defer func() {
		_ = recover()
	}()

	srv.logQueue <- &logMessage{
		ctx: ctx,
		t:   t,
		msg: fmt.Sprintf(msg, a...),
	}

}

func (srv *server) logToNamespace(ctx context.Context, t time.Time, cached *database.CacheData, msg string, a ...interface{}) {

	defer func() {
		_ = recover()
	}()

	var cd database.CacheData // We do this to zero some fields without modifying the argument.
	cd = *cached
	cd.Workflow = nil
	cd.Instance = nil

	srv.logQueue <- &logMessage{
		ctx:    ctx,
		t:      t,
		msg:    fmt.Sprintf(msg, a...),
		cached: &cd,
	}

}

func (srv *server) logToWorkflow(ctx context.Context, t time.Time, cached *database.CacheData, msg string, a ...interface{}) {

	defer func() {
		_ = recover()
	}()

	var cd database.CacheData // We do this to zero some fields without modifying the argument.
	cd = *cached
	cd.Workflow = nil

	srv.logQueue <- &logMessage{
		ctx:    ctx,
		t:      t,
		msg:    fmt.Sprintf(msg, a...),
		cached: &cd,
	}

}

// log To instance with string interpolation.
func (srv *server) logToInstance(ctx context.Context, t time.Time, cached *database.CacheData, msg string, a ...interface{}) {

	msg = fmt.Sprintf(msg, a...)

	srv.logToInstanceRaw(ctx, t, cached, msg)

}

// log To instance with raw string.
func (srv *server) logToInstanceRaw(ctx context.Context, t time.Time, cached *database.CacheData, msg string) {

	defer func() {
		_ = recover()
	}()

	srv.logQueue <- &logMessage{
		ctx:    ctx,
		t:      t,
		msg:    msg,
		cached: cached,
	}

}

func (engine *engine) UserLog(ctx context.Context, im *instanceMemory, msg string, a ...interface{}) {

	engine.logToInstance(ctx, time.Now(), im.cached, msg, a...)

	s := fmt.Sprintf(msg, a...)

	if attr := im.cached.Workflow.LogToEvents; attr != "" {
		event := cloudevents.NewEvent()
		event.SetID(uuid.New().String())
		event.SetSource(im.cached.Workflow.ID.String())
		event.SetType("direktiv.instanceLog")
		event.SetExtension("logger", attr)
		event.SetDataContentType("application/json")
		err := event.SetData("application/json", s)
		if err != nil {
			engine.sugar.Errorf("Failed to create CloudEvent: %v.", err)
		}

		err = engine.events.BroadcastCloudevent(ctx, im.cached, &event, 0)
		if err != nil {
			engine.sugar.Errorf("Failed to broadcast CloudEvent: %v.", err)
			return
		}
	}

}

func (engine *engine) logRunState(ctx context.Context, im *instanceMemory, wakedata []byte, err error) {

	engine.sugar.Debugf("Running state logic -- %s:%v (%s) (%v)", im.ID().String(), im.Step(), im.logic.GetID(), time.Now())
	if im.GetMemory() == nil && len(wakedata) == 0 && err == nil {
		engine.logToInstance(ctx, time.Now(), im.cached, "Running state logic (step:%v) -- %s", im.Step(), im.logic.GetID())
	}

}

func this() string {
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	elems := strings.Split(fn.Name(), ".")
	return elems[len(elems)-1]
}

func parent() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return ""
	}
	fn := runtime.FuncForPC(pc)
	elems := strings.Split(fn.Name(), ".")
	return elems[len(elems)-1]
}

func (srv *server) logToMirrorActivity(ctx context.Context, t time.Time, ns *database.Namespace, mirror *database.Mirror, act *database.MirrorActivity, msg string, a ...interface{}) {

	msg = fmt.Sprintf(msg, a...)

	util.Trace(ctx, msg)

	clients := srv.edb.Clients(nil)

	_, err := clients.LogMsg.Create().SetMsg(msg).SetActivityID(act.ID).SetT(t).Save(ctx)
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	span := trace.SpanFromContext(ctx)
	tid := span.SpanContext().TraceID()

	srv.sugar.Infow(msg, "trace", tid, "namespace", ns.Name, "namespace-id", ns.ID.String(), "mirror-id", mirror.ID.String())

	srv.pubsub.NotifyMirrorActivityLogs(act)

}
