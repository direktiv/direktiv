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

func (srv *server) logToServer(ctx context.Context, t time.Time, msg string, a ...interface{}) {

	logc := srv.db.LogMsg

	msg = fmt.Sprintf(msg, a...)

	util.Trace(ctx, msg)

	_, err := logc.Create().SetMsg(msg).SetT(t).Save(ctx)
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	span := trace.SpanFromContext(ctx)
	tid := span.SpanContext().TraceID()

	srv.sugar.Infow(msg, "trace", tid)

	srv.pubsub.NotifyServerLogs()

}

func (srv *server) logToNamespace(ctx context.Context, t time.Time, ns *ent.Namespace, msg string, a ...interface{}) {

	logc := srv.db.LogMsg

	msg = fmt.Sprintf(msg, a...)

	util.Trace(ctx, msg)

	_, err := logc.Create().SetMsg(msg).SetNamespace(ns).SetT(t).Save(ctx)
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	span := trace.SpanFromContext(ctx)
	tid := span.SpanContext().TraceID()

	srv.sugar.Infow(msg, "trace", tid, "namespace", ns.Name, "namespace-id", ns.ID.String())

	srv.pubsub.NotifyNamespaceLogs(ns)

}

func (srv *server) logToWorkflow(ctx context.Context, t time.Time, d *wfData, msg string, a ...interface{}) {

	logc := srv.db.LogMsg

	msg = fmt.Sprintf(msg, a...)

	util.Trace(ctx, msg)

	_, err := logc.Create().SetMsg(msg).SetWorkflow(d.wf).SetT(t).Save(ctx)
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	span := trace.SpanFromContext(ctx)
	tid := span.SpanContext().TraceID()

	ns := d.wf.Edges.Namespace
	srv.sugar.Infow(msg, "trace", tid, "namespace", ns.Name, "namespace-id", ns.ID.String(), "workflow-id", d.wf.ID.String(), "workflow", GetInodePath(d.path))

	srv.pubsub.NotifyWorkflowLogs(d.wf)

}

// log To instance with string interpolation
func (srv *server) logToInstance(ctx context.Context, t time.Time, in *ent.Instance, msg string, a ...interface{}) {

	msg = fmt.Sprintf(msg, a...)

	srv.logToInstanceRaw(ctx, t, in, msg)
}

// log To instance with raw string
func (srv *server) logToInstanceRaw(ctx context.Context, t time.Time, in *ent.Instance, msg string) {
	logc := srv.db.LogMsg

	util.Trace(ctx, msg)

	_, err := logc.Create().SetMsg(msg).SetInstance(in).SetT(t).Save(ctx)
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	span := trace.SpanFromContext(ctx)
	tid := span.SpanContext().TraceID()

	nsid := ""
	nsname := ""
	if in.Edges.Namespace != nil {
		nsid = in.Edges.Namespace.ID.String()
		nsname = in.Edges.Namespace.Name
	}

	wfid := ""
	if in.Edges.Workflow != nil {
		wfid = in.Edges.Workflow.ID.String()
	}

	srv.sugar.Infow(msg, "trace", tid, "namespace", nsname, "namespace-id", nsid, "workflow-id", wfid, "workflow", GetInodePath(in.As), "instance", in.ID.String())

	srv.pubsub.NotifyInstanceLogs(in)
}

func (engine *engine) UserLog(ctx context.Context, im *instanceMemory, msg string, a ...interface{}) {

	engine.logToInstance(ctx, time.Now(), im.in, msg, a...)

	s := fmt.Sprintf(msg, a...)

	// TODO: detect content type and handle base64 data

	wf, err := engine.InstanceWorkflow(ctx, im)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

	if attr := wf.LogToEvents; attr != "" {
		event := cloudevents.NewEvent()
		event.SetID(uuid.New().String())
		event.SetSource(wf.ID.String()) // TODO: resolve to a human-readable path
		event.SetType("direktiv.instanceLog")
		event.SetExtension("logger", attr)
		event.SetDataContentType("application/json")
		event.SetData("application/json", s)
		err = engine.events.BroadcastCloudevent(ctx, im.in.Edges.Namespace, &event, 0)
		if err != nil {
			engine.sugar.Errorf("failed to broadcast cloudevent: %v", err)
			return
		}
	}

}

func (engine *engine) logRunState(ctx context.Context, im *instanceMemory, wakedata []byte, err error) {

	engine.sugar.Debugf("Running state logic -- %s:%v (%s)", im.ID().String(), im.Step(), im.logic.GetID())
	if im.GetMemory() == nil && len(wakedata) == 0 && err == nil {
		engine.logToInstance(ctx, time.Now(), im.in, "Running state logic (step:%v) -- %s", im.Step(), im.logic.GetID())
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
