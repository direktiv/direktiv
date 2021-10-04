package flow

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/vorteil/direktiv/pkg/flow/ent"
	"github.com/vorteil/direktiv/pkg/util"
	"go.opentelemetry.io/otel/trace"
)

func logsOrder(p *pagination) ent.LogMsgPaginateOption {

	field := ent.LogMsgOrderFieldT
	direction := ent.OrderDirectionAsc

	if p.order != nil {

		if x := p.order.Field; x != "" && x == "TIMESTAMP" {
			field = ent.LogMsgOrderFieldT
		}

		if x := p.order.Direction; x != "" && x == "DESC" {
			direction = ent.OrderDirectionDesc
		}

	}

	return ent.WithLogMsgOrder(&ent.LogMsgOrder{
		Direction: direction,
		Field:     field,
	})

}

func logsFilter(p *pagination) ent.LogMsgPaginateOption {

	if p.filter == nil {
		return nil
	}

	// TODO

	return ent.WithLogMsgFilter(func(query *ent.LogMsgQuery) (*ent.LogMsgQuery, error) {

		return query, nil

	})

}

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

func (srv *server) logToWorkflow(ctx context.Context, t time.Time, wf *ent.Workflow, msg string, a ...interface{}) {

	logc := srv.db.LogMsg

	msg = fmt.Sprintf(msg, a...)

	util.Trace(ctx, msg)

	_, err := logc.Create().SetMsg(msg).SetWorkflow(wf).SetT(t).Save(ctx)
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	span := trace.SpanFromContext(ctx)
	tid := span.SpanContext().TraceID()

	ns := wf.Edges.Namespace
	srv.sugar.Infow(msg, "trace", tid, "namespace", ns.Name, "namespace-id", ns.ID.String(), "workflow-id", wf.ID.String())

	srv.pubsub.NotifyWorkflowLogs(wf)

}

func (srv *server) logToInstance(ctx context.Context, t time.Time, in *ent.Instance, msg string, a ...interface{}) {

	logc := srv.db.LogMsg

	msg = fmt.Sprintf(msg, a...)

	util.Trace(ctx, msg)

	_, err := logc.Create().SetMsg(msg).SetInstance(in).SetT(t).Save(ctx)
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	span := trace.SpanFromContext(ctx)
	tid := span.SpanContext().TraceID()

	ns := in.Edges.Namespace
	wf := in.Edges.Workflow
	srv.sugar.Infow(msg, "trace", tid, "namespace", ns.Name, "namespace-id", ns.ID.String(), "workflow-id", wf.ID.String(), "workflow", in.As, "instance", in.ID.String())

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

	engine.sugar.Debugf("Running state logic -- %s:%v (%s)", im.ID().String(), im.Step(), im.logic.ID())
	if im.GetMemory() == nil && len(wakedata) == 0 && err == nil {
		engine.logToInstance(ctx, time.Now(), im.in, "Running state logic (step:%v) -- %s", im.Step(), im.logic.ID())
	}

}

func this() string {
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	elems := strings.Split(fn.Name(), ".")
	return elems[len(elems)-1]
}
