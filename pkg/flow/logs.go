package flow

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type logMessage struct {
	ctx context.Context //nolint:containedctx
	t   time.Time
	msg string
	ns  *ent.Namespace
	wf  *wfData
	in  *ent.Instance
	tag map[string]string
}

type taggedInstance interface {
	tags() map[string]string
	instance() *ent.Instance
	//addTag(t, v string)
}

type taggedNamespace interface {
	tags() map[string]string
	ns() *ent.Namespace
	//addTag(t, v string)
}

type taggedWfData interface {
	tags() map[string]string
	wD() *wfData
	//addTag(t, v string)
}

func instanceTags(in *ent.Instance) map[string]string {
	tags := make(map[string]string)
	tags["ins-as"] = in.As
	tags["ins-invoker"] = in.Invoker
	if in.Edges.Namespace != nil {
		tags["ins-namespace-name"] = in.Edges.Namespace.Name
	}
	return tags
}

func namespaceTags(n *ent.Namespace) map[string]string {
	tags := make(map[string]string)
	tags["ns-name"] = n.Name
	return tags
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

		if l.in != nil {
			srv.workerLogToInstance(l)
		} else if l.wf != nil {
			srv.workerLogToWorkflow(l)
		} else if l.ns != nil {
			srv.workerLogToNamespace(l)
		} else {
			srv.workerLogToServer(l)
		}

	}

}

func (srv *server) closeLogWorkers() {
	close(srv.logQueue)
	srv.logWorkersWG.Wait()
}

func (srv *server) workerLogToServer(l *logMessage) {

	logc := srv.db.LogMsg

	util.Trace(l.ctx, l.msg)

	_, err := logc.Create().SetMsg(l.msg).SetT(l.t).Save(context.Background())
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

	err := srv.storeLogMsg(l)
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	span := trace.SpanFromContext(l.ctx)
	tid := span.SpanContext().TraceID()

	srv.sugar.Infow(l.msg, "trace", tid, "namespace", l.ns.Name, "namespace-id", l.ns.ID.String())

	srv.pubsub.NotifyNamespaceLogs(l.ns)

}

func (srv *server) workerLogToWorkflow(l *logMessage) {

	err := srv.storeLogMsg(l)
	if err != nil {
		srv.sugar.Error(err)
		return
	}
	span := trace.SpanFromContext(l.ctx)
	tid := span.SpanContext().TraceID()

	ns := l.wf.wf.Edges.Namespace
	srv.sugar.Infow(l.msg, "trace", tid, "namespace", ns.Name, "namespace-id", ns.ID.String(), "workflow-id", l.wf.wf.ID.String(), "workflow", GetInodePath(l.wf.path))

	srv.pubsub.NotifyWorkflowLogs(l.wf.wf)

}

func (srv *server) workerLogToInstance(l *logMessage) {

	err := srv.storeLogMsg(l)
	if err != nil {
		srv.sugar.Error(err)
		return
	}
	span := trace.SpanFromContext(l.ctx)
	tid := span.SpanContext().TraceID()

	nsid := ""
	nsname := ""
	if l.in.Edges.Namespace != nil {
		nsid = l.in.Edges.Namespace.ID.String()
		nsname = l.in.Edges.Namespace.Name
	}

	wfid := ""
	if l.in.Edges.Workflow != nil {
		wfid = l.in.Edges.Workflow.ID.String()
	}

	srv.sugar.Infow(l.msg, "trace", tid, "namespace", nsname, "namespace-id", nsid, "workflow-id", wfid, "workflow", GetInodePath(l.in.As), "instance", l.in.ID.String())

	srv.pubsub.NotifyInstanceLogs(l.in)

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

func (srv *server) tagLogToNamespace(ctx context.Context, t time.Time, tn taggedNamespace, msg string, a ...interface{}) {

	defer func() {
		_ = recover()
	}()

	srv.logQueue <- &logMessage{
		ctx: ctx,
		t:   t,
		msg: fmt.Sprintf(msg, a...),
		ns:  tn.ns(),
		tag: tn.tags(),
	}
}

func (srv *server) logToNamespace(ctx context.Context, t time.Time, ns *ent.Namespace, msg string, a ...interface{}) {

	defer func() {
		_ = recover()
	}()
	tag := namespaceTags(ns)
	srv.logQueue <- &logMessage{
		ctx: ctx,
		t:   t,
		msg: fmt.Sprintf(msg, a...),
		ns:  ns,
		tag: tag,
	}

}

func (srv *server) tagLogToWorkflow(ctx context.Context, t time.Time, td taggedWfData, msg string, a ...interface{}) {

	defer func() {
		_ = recover()
	}()

	srv.logQueue <- &logMessage{
		ctx: ctx,
		t:   t,
		msg: fmt.Sprintf(msg, a...),
		wf:  td.wD(),
		tag: td.tags(),
	}

}

func (srv *server) logToInstance(ctx context.Context, t time.Time, ti taggedInstance, msg string, a ...interface{}) {

	msg = fmt.Sprintf(msg, a...)
	srv.logToInstanceRaw(ctx, t, ti.instance(), ti.tags(), msg)

}

// log To instance with raw string.
func (srv *server) logToInstanceRaw(ctx context.Context, t time.Time, in *ent.Instance, tag map[string]string, msg string) {

	defer func() {
		_ = recover()
	}()

	srv.logQueue <- &logMessage{
		ctx: ctx,
		t:   t,
		msg: msg,
		in:  in,
		tag: tag,
	}

}

func (engine *engine) UserLog(ctx context.Context, im *instanceMemory, msg string, a ...interface{}) {

	engine.logToInstance(ctx, time.Now(), im, msg, a...)

	s := fmt.Sprintf(msg, a...)

	wf, err := engine.InstanceWorkflow(ctx, im)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

	if attr := wf.LogToEvents; attr != "" {
		event := cloudevents.NewEvent()
		event.SetID(uuid.New().String())
		event.SetSource(wf.ID.String())
		event.SetType("direktiv.instanceLog")
		event.SetExtension("logger", attr)
		event.SetDataContentType("application/json")
		err = event.SetData("application/json", s)
		if err != nil {
			engine.sugar.Errorf("Failed to create CloudEvent: %v.", err)
		}

		err = engine.events.BroadcastCloudevent(ctx, im.in.Edges.Namespace, &event, 0)
		if err != nil {
			engine.sugar.Errorf("Failed to broadcast CloudEvent: %v.", err)
			return
		}
	}

}

func (engine *engine) logRunState(ctx context.Context, im *instanceMemory, wakedata []byte, err error) {

	engine.sugar.Debugf("Running state logic -- %s:%v (%s)", im.ID().String(), im.Step(), im.logic.GetID())
	if im.GetMemory() == nil && len(wakedata) == 0 && err == nil {
		engine.logToInstance(ctx, time.Now(), im, "Running state logic (step:%v) -- %s", im.Step(), im.logic.GetID())
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

func (srv *server) logToMirrorActivity(ctx context.Context, t time.Time, act *ent.MirrorActivity, msg string, a ...interface{}) {

	logc := srv.db.LogMsg

	msg = fmt.Sprintf(msg, a...)

	util.Trace(ctx, msg)

	_, err := logc.Create().SetMsg(msg).SetActivity(act).SetT(t).Save(ctx)
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	span := trace.SpanFromContext(ctx)
	tid := span.SpanContext().TraceID()

	ns := act.Edges.Namespace
	srv.sugar.Infow(msg, "trace", tid, "namespace", ns.Name, "namespace-id", ns.ID.String(), "mirror-id", act.Edges.Mirror.ID.String())

	srv.pubsub.NotifyMirrorActivityLogs(act)

}

func (srv *server) storeLogMsg(l *logMessage) error {
	util.Trace(l.ctx, l.msg)
	ctx := context.Background()
	tx, err := srv.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("starting a transaction: %w", err)
	}
	defer rollback(tx)
	t := ""
	for k, v := range l.tag {
		t += fmt.Sprintf("[%s:%s]", k, v)
	}
	//lc := tx.LogMsg.Create().SetMsg(t + " -> " + l.msg).SetT(l.t) //TODO
	lc := tx.LogMsg.Create().SetMsg(t).SetT(l.t)
	if l.in != nil {
		lc.SetInstance(l.in)
	}
	if l.ns != nil {
		lc.SetNamespace(l.ns)
	}
	if l.wf != nil {
		lc.SetWorkflow(l.wf.wf)
	}
	lDB, err := lc.Save(ctx)
	if err != nil {
		srv.sugar.Error(err)
		return fmt.Errorf("failed creating the Logmsg: %w", err)
	}
	if l.tag == nil {
		return tx.Commit()
	}
	var bt []*ent.LogTagCreate
	for k, v := range l.tag {
		bt = append(bt, tx.LogTag.Create().SetType(k).SetValue(v).SetLogmsgID(lDB.ID))
	}
	_, err = tx.LogTag.CreateBulk(bt...).Save(ctx)
	if err != nil {
		return fmt.Errorf("failed creating the Logtags: %w", err)
	}
	return tx.Commit()
}
