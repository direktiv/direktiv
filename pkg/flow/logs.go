package flow

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/vorteil/direktiv/pkg/flow/ent"
)

func logsOrder(p *pagination) ent.LogMsgPaginateOption {

	field := ent.LogMsgOrderFieldT
	direction := ent.OrderDirectionAsc

	if p.order != nil {

		if x := p.order.Field; x != nil && *x == "TIMESTAMP" {
			field = ent.LogMsgOrderFieldT
		}

		if x := p.order.Direction; x != nil && *x == "DESC" {
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

	_, err := logc.Create().SetMsg(msg).SetT(t).Save(ctx)
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	srv.pubsub.NotifyServerLogs()

}

func (srv *server) logToNamespace(ctx context.Context, t time.Time, ns *ent.Namespace, msg string, a ...interface{}) {

	logc := srv.db.LogMsg

	msg = fmt.Sprintf(msg, a...)

	_, err := logc.Create().SetMsg(msg).SetNamespace(ns).SetT(t).Save(ctx)
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	srv.pubsub.NotifyNamespaceLogs(ns)

}

func (srv *server) logToWorkflow(ctx context.Context, t time.Time, wf *ent.Workflow, msg string, a ...interface{}) {

	logc := srv.db.LogMsg

	msg = fmt.Sprintf(msg, a...)

	_, err := logc.Create().SetMsg(msg).SetWorkflow(wf).SetT(t).Save(ctx)
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	srv.pubsub.NotifyWorkflowLogs(wf)

}

func (srv *server) logToInstance(ctx context.Context, t time.Time, in *ent.Instance, msg string, a ...interface{}) {

	logc := srv.db.LogMsg

	msg = fmt.Sprintf(msg, a...)

	_, err := logc.Create().SetMsg(msg).SetInstance(in).SetT(t).Save(ctx)
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	srv.pubsub.NotifyInstanceLogs(in)

}

func (engine *engine) UserLog(ctx context.Context, im *instanceMemory, msg string, a ...interface{}) {

	engine.logToInstance(ctx, time.Now(), im.in, msg, a...)

	// TODO

	/*
		s := fmt.Sprintf(msg, a...)

		// TODO: detect content type and handle base64 data

		if attr := im.LogToEvents(); attr != "" {
			event := cloudevents.NewEvent()
			event.SetID(uuid.New().String())
			event.SetSource(wli.wf.ID)
			event.SetType("direktiv.instanceLog")
			event.SetExtension("logger", attr)
			event.SetDataContentType("application/json")
			event.SetData(s)
			data, err := event.MarshalJSON()
			if err != nil {
				engine.sugar.Errorf("failed to marshal UserLog cloudevent: %v", err)
				return
			}
			_, err = engine.ingressClient.BroadcastEvent(ctx, &ingress.BroadcastEventRequest{
				Namespace:  &wli.namespace,
				Cloudevent: data,
			})
			if err != nil {
				engine.sugar.Errorf("failed to broadcast cloudevent: %v", err)
				return
			}
		}
	*/

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
