package flow

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	"github.com/direktiv/direktiv/pkg/model"
	pkgevents "github.com/direktiv/direktiv/pkg/refactor/events"
	"github.com/dop251/goja"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	eventTypeString   = "type"
	filterPrefix      = "filter-"
	sendEventFunction = "sendEvent"
)

func init() {
	gob.Register(new(event.EventContextV1))
	gob.Register(new(event.EventContextV03))
}

type events struct {
	*server
}

type CacheObject struct {
	value sync.Map
}

var eventFilterCache = &CacheObject{}

func initEvents(srv *server) (*events, error) {
	events := new(events)

	events.server = srv

	return events, nil
}

func (events *events) Close() error {
	return nil
}

func (events *events) sendEvent(data []byte) {
	n := strings.SplitN(string(data), "/", 2)

	if len(n) != 2 {
		events.sugar.Errorf("namespace and id must be set for delayed events")
		return
	}

	id, err := uuid.Parse(n[1])
	if err != nil {
		events.sugar.Errorf("namespace id invalid")
		return
	}

	ctx := context.Background()

	ns, err := events.edb.Namespace(ctx, id)
	if err != nil {
		events.sugar.Error(err)
		return
	}

	err = events.flushEvent(ctx, n[0], ns, true)
	if err != nil {
		events.sugar.Errorf("can not flush delayed event: %v", err)
		return
	}
}

var syncMtx sync.Mutex

func (events *events) syncEventDelays() {
	// syncMtx.Lock()
	// defer syncMtx.Unlock()

	// // disable old timer
	// events.timers.mtx.Lock()
	// for i := range events.timers.timers {
	// 	ti := events.timers.timers[i]
	// 	if ti.name == "sendEventTimer" {
	// 		events.timers.disableTimer(ti)
	// 		break
	// 	}
	// }
	// events.timers.mtx.Unlock()
	// TODO:
	// ctx := context.Background()

	// for {
	// 	e, err := events.getEarliestEvent(ctx)
	// 	if err != nil {
	// 		if derrors.IsNotFound(err) {
	// 			return
	// 		}

	// 		events.sugar.Errorf("can not sync event delays: %v", err)
	// 		return
	// 	}

	//
	// 	err = events.database.Namespace(ctx, cached, e.Edges.Namespace.ID)
	// 	if err != nil {
	// 		return
	// 	}

	// 	if e.Fire.Before(time.Now()) {
	// 		err = events.flushEvent(ctx, e.EventId, cached.Namespace, false)
	// 		if err != nil {
	// 			events.sugar.Errorf("can not flush event %s: %v", e.ID, err)
	// 		}
	// 		continue
	// 	}

	// 	err = events.timers.addOneShot("sendEventTimer", sendEventFunction,
	// 		e.Fire, []byte(fmt.Sprintf("%s/%s", e.ID, e.Edges.Namespace.ID.String())))
	// 	if err != nil {
	// 		events.sugar.Errorf("can not reschedule event timer: %v", err)
	// 	}

	// 	break
	// }
}

func (events *events) flushEvent(ctx context.Context, eventID string, ns *database.Namespace, rearm bool) error {
	// tctx, tx, err := events.database.Tx(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer rollback(tx)

	// e, err := events.markEventAsProcessed(tctx, ns, eventID)
	// if err != nil {
	// 	return err
	// }

	// err = tx.Commit()
	// if err != nil {
	// 	return err
	// }
	// TODO is this needed?

	defer func(r bool) {
		if r {
			events.syncEventDelays()
		}
	}(rearm)

	// err = events.handleEvent(ns, e)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func (events *events) handleEvent(ns *database.Namespace, ce *cloudevents.Event) error {
	e := pkgevents.EventEngine{
		WorkflowStart: func(workflowID uuid.UUID, ev ...*cloudevents.Event) {
			// events.metrics.InsertRecord
			events.logger.Debugf(context.TODO(), ns.ID, events.flow.GetAttributes(), "invoking workflow")
			events.engine.EventsInvoke(wf, ev...)
		},
		WakeInstance: func(instanceID uuid.UUID, step int, ev []*cloudevents.Event) {
			// events.metrics.InsertRecord
			events.logger.Debugf(context.TODO(), ns.ID, events.flow.GetAttributes(), "invoking instance %v", instanceID)
			events.engine.wakeEventsWaiter(instanceID, step, ev) // TODO
		},
		GetListenersByTopic: func(ctx context.Context, s string) ([]*pkgevents.EventListener, error) {
			res := make([]*pkgevents.EventListener, 0)
			err := events.runSqlTx(ctx, func(tx *sqlTx) error {
				r, err := tx.DataStore().EventListenerTopics().GetListeners(ctx, s)
				if err != nil {
					return err
				}
				res = r
				return nil
			})
			if err != nil {
				return nil, err
			}
			return res, nil
		},
		UpdateListeners: func(ctx context.Context, listener []*pkgevents.EventListener) []error {
			events.logger.Debugf(context.TODO(), ns.ID, events.flow.GetAttributes(), "update listener")
			err := events.runSqlTx(ctx, func(tx *sqlTx) error {
				errs := tx.DataStore().EventListener().Update(ctx, listener)
				for _, err2 := range errs {
					if err2 != nil {
						return err2
					}
				}
				return nil
			})
			if err != nil {
				return nil
			}
			return nil
		},
	}
	e.ProcessEvents(context.TODO(), ns.ID, []event.Event{*ce})
	metricsCloudEventsCaptured.WithLabelValues(ns.Name, ce.Type(), ce.Source(), ns.Name).Inc()
	return nil
}

func (flow *flow) EventListeners(ctx context.Context, req *grpc.EventListenersRequest) (*grpc.EventListenersResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	var resListeners []*pkgevents.EventListener
	totalListeners := 0
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		li, t, err := tx.DataStore().EventListener().Get(ctx, ns.ID, int(req.Pagination.Limit), int(req.Pagination.Offset))
		if err != nil {
			return err
		}
		resListeners = li
		totalListeners = t
		return nil
	})
	if err != nil {
		return nil, err
	}
	resp := new(grpc.EventListenersResponse)
	resp.Namespace = ns.Name
	resp.PageInfo = &grpc.PageInfo{Total: int32(totalListeners)}

	resp.Results = bytedata.ConvertEventListeners(resListeners)

	return resp, nil
}

func (flow *flow) EventListenersStream(req *grpc.EventListenersRequest, srv grpc.Flow_EventListenersStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeEventListeners(ns)
	defer flow.cleanup(sub.Close)
resend:
	var resListeners []*pkgevents.EventListener
	totalListeners := 0
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		li, t, err := tx.DataStore().EventListener().Get(ctx, ns.ID, int(req.Pagination.Limit), int(req.Pagination.Offset))
		if err != nil {
			return err
		}
		resListeners = li
		totalListeners = t
		return nil
	})
	if err != nil {
		return err
	}
	resp := new(grpc.EventListenersResponse)
	resp.Namespace = ns.Name
	resp.PageInfo = &grpc.PageInfo{Total: int32(totalListeners)}

	resp.Results = bytedata.ConvertEventListeners(resListeners)

	nhash = bytedata.Checksum(resp)
	if nhash != phash {
		err = srv.Send(resp)
		if err != nil {
			return err
		}
	}
	phash = nhash

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend
}

func (flow *flow) BroadcastCloudevent(ctx context.Context, in *grpc.BroadcastCloudeventRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	namespace := in.GetNamespace()
	rawevent := in.GetCloudevent()

	event := new(cloudevents.Event)
	err := event.UnmarshalJSON(rawevent)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid cloudevent: %v", err)
	}
	if event.SpecVersion() == "" {
		event.SetSpecVersion("1.0")
	}
	if event.ID() == "" {
		event.SetID(uuid.NewString())
	}
	// NOTE: this validate check added to sanitize Azure's dodgy cloudevents.
	err = event.Validate()
	if err != nil && strings.Contains(err.Error(), "dataschema") {
		event.SetDataSchema("")
		err = event.Validate()
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid cloudevent: %v", err)
		}
	}

	// NOTE: remarshal / unmarshal necessary to overcome issues with cloudevents library.
	data, err := json.Marshal(event)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid cloudevent: %v", err)
	}

	err = event.UnmarshalJSON(data)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid cloudevent: %v", err)
	}

	ns, err := flow.edb.NamespaceByName(ctx, namespace)
	if err != nil {
		return nil, err
	}

	timer := in.GetTimer()

	err = flow.events.BroadcastCloudevent(ctx, ns, event, timer)
	if err != nil {
		return nil, err
	}

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) HistoricalEvent(ctx context.Context, in *grpc.HistoricalEventRequest) (*grpc.HistoricalEventResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	eid := in.GetId()
	if eid == "" {
		eid = uuid.NewString()
	}

	ns, err := flow.edb.NamespaceByName(ctx, in.GetNamespace())
	if err != nil {
		return nil, err
	}
	var cevent *pkgevents.Event
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		evs, err := tx.DataStore().EventHistory().GetByID(ctx, eid)
		if err != nil {
			return err
		}
		cevent = evs
		return nil
	})
	if err != nil {
		return nil, err
	}
	var resp grpc.HistoricalEventResponse

	resp.Id = eid
	resp.Namespace = ns.Name
	resp.ReceivedAt = timestamppb.New(cevent.ReceivedAt)

	resp.Source = cevent.Event.Source()
	resp.Type = cevent.Event.Type()

	resp.Cloudevent = []byte(cevent.Event.String())

	return &resp, nil
}

var cloudeventsOrderings = []*orderingInfo{
	{
		db:           "ReceivedAt",
		req:          "RECEIVED",
		defaultOrder: ent.Desc,
	},
	{
		db:           "id",
		req:          "ID",
		defaultOrder: ent.Asc,
	},
}

const (
	contains = "CONTAINS"
	cr       = "CREATED"
	after    = "AFTER"
	before   = "BEFORE"
)

func (flow *flow) EventHistory(ctx context.Context, req *grpc.EventHistoryRequest) (*grpc.EventHistoryResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	count := 0
	var res []*pkgevents.Event
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		re, t, err := tx.DataStore().EventHistory().Get(ctx, int(req.Pagination.Limit), int(req.Pagination.Offset), ns.ID)
		if err != nil {
			return err
		}
		count = t
		res = re
		return nil
	})
	if err != nil {
		return nil, err
	}
	resp := new(grpc.EventHistoryResponse)
	resp.Namespace = ns.Name
	resp.Events = new(grpc.Events)
	finalResults := make([]*grpc.Event, 0, len(res))
	for _, e := range res {
		finalResults = append(finalResults, &grpc.Event{
			ReceivedAt: timestamppb.New(e.ReceivedAt),
			Id:         e.Event.ID(),
			Source:     e.Event.Source(),
			Type:       e.Event.Type(),
			Cloudevent: []byte(e.Event.String()),
		})
	}
	resp.Events.Results = finalResults
	resp.Events.PageInfo = &grpc.PageInfo{Total: int32(count), Limit: req.Pagination.Limit, Offset: req.Pagination.Offset}

	return resp, nil
}

func (flow *flow) EventHistoryStream(req *grpc.EventHistoryRequest, srv grpc.Flow_EventHistoryStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeEvents(ns)
	defer flow.cleanup(sub.Close)

resend:

	count := 0
	var res []*pkgevents.Event
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		re, t, err := tx.DataStore().EventHistory().Get(ctx, int(req.Pagination.Limit), int(req.Pagination.Offset), ns.ID)
		if err != nil {
			return err
		}
		count = t
		res = re
		return nil
	})
	if err != nil {
		return err
	}
	resp := new(grpc.EventHistoryResponse)
	resp.Namespace = ns.Name
	resp.Events = new(grpc.Events)
	finalResults := make([]*grpc.Event, 0, len(res))
	for _, e := range res {
		finalResults = append(finalResults, &grpc.Event{
			ReceivedAt: timestamppb.New(e.ReceivedAt),
			Id:         e.Event.ID(),
			Source:     e.Event.Source(),
			Type:       e.Event.Type(),
		})
	}
	resp.Events.Results = finalResults
	resp.Events.PageInfo = &grpc.PageInfo{Total: int32(count), Limit: req.Pagination.Limit, Offset: req.Pagination.Offset}

	nhash = bytedata.Checksum(resp)
	if nhash != phash {
		err = srv.Send(resp)
		if err != nil {
			return err
		}
	}
	phash = nhash

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend
}

func (flow *flow) ReplayEvent(ctx context.Context, req *grpc.ReplayEventRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	eid := req.GetId()
	if eid == "" {
		eid = uuid.NewString()
	}
	if err != nil {
		return nil, err
	}
	var cevent *pkgevents.Event
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		evs, err := tx.DataStore().EventHistory().GetByID(ctx, eid)
		if err != nil {
			return err
		}
		cevent = evs
		return nil
	})
	if err != nil {
		return &emptypb.Empty{}, err
	}
	err = flow.events.ReplayCloudevent(ctx, ns, cevent)
	if err != nil {
		return nil, err
	}

	var resp emptypb.Empty

	return &resp, nil
}

func (events *events) ReplayCloudevent(ctx context.Context, ns *database.Namespace, cevent *pkgevents.Event) error {
	event := cevent.Event

	events.logger.Infof(ctx, ns.ID, ns.GetAttributes(), "Replaying event: %s (%s / %s)", event.ID(), event.Type(), event.Source())

	err := events.handleEvent(ns, event)
	if err != nil {
		return err
	}

	// if eventing is configured, event goes to knative event service
	// if it is from knative sink not
	if events.server.conf.Eventing && ctx.Value(EventingCtxKeySource) == nil {
		PublishKnativeEvent(event)
	}

	return nil
}

func (events *events) BroadcastCloudevent(ctx context.Context, ns *database.Namespace, event *cloudevents.Event, timer int64) error {
	events.logger.Infof(ctx, ns.ID, database.GetAttributes(recipient.Namespace, ns), "Event received: %s (%s / %s)", event.ID(), event.Type(), event.Source())

	metricsCloudEventsReceived.WithLabelValues(ns.Name, event.Type(), event.Source(), ns.Name).Inc()

	// add event to db
	err := events.addEvent(ctx, event, ns, timer)
	if err != nil {
		return err
	}

	events.pubsub.NotifyEvents(ns)

	// handle event
	if timer == 0 {
		err = events.handleEvent(ns, event)
		if err != nil {
			return err
		}
	} else {
		// if we have a delay we need to update event delay
		// sending nil as server id so all instances calling it
		events.pubsub.UpdateEventDelays()
	}

	// if eventing is configured, event goes to knative event service
	// if it is from knative sink not
	if events.server.conf.Eventing && ctx.Value(EventingCtxKeySource) == nil {
		PublishKnativeEvent(event)
	}

	return nil
}

func (events *events) updateEventDelaysHandler(req *pubsub.PubsubUpdate) {
	events.syncEventDelays()
}

func (events *events) listenForEvents(ctx context.Context, im *instanceMemory, ceds []*model.ConsumeEventDefinition, all bool) error {
	var transformedEvents []*model.ConsumeEventDefinition

	for i := range ceds {
		ev := new(model.ConsumeEventDefinition)
		ev.Context = make(map[string]interface{})

		err := copier.Copy(ev, ceds[i])
		if err != nil {
			return err
		}

		for k, v := range ceds[i].Context {
			ev.Context[k], err = jqOne(im.data, v)
			if err != nil {
				return fmt.Errorf("failed to execute jq query for key '%s' on event definition %d: %w", k, i, err)
			}
		}

		transformedEvents = append(transformedEvents, ev)
	}

	err := events.addInstanceEventListener(ctx, im.Namespace().ID, im.GetInstanceID(), transformedEvents, im.Step(), all)
	if err != nil {
		return err
	}

	events.logger.Infof(ctx, im.GetInstanceID(), im.GetAttributes(), "Registered to receive events.")

	return nil
}

func (flow *flow) execFilter(ctx context.Context, namespace, filterName string, cloudevent []byte) ([]byte, error) {
	var script string
	var newBytesEvent []byte

	key := fmt.Sprintf("%s-%s", namespace, filterName)

	ns, err := flow.edb.NamespaceByName(ctx, namespace)
	if err != nil {
		return newBytesEvent, err
	}

	if jsCode, ok := eventFilterCache.get(key); ok {
		script = fmt.Sprintf("function filter() {\n %s \n}", jsCode)
	} else {
		var filters []*pkgevents.NamespaceCloudEventFilter
		err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
			f, _, err := tx.DataStore().EventFilter().Get(ctx, ns.ID, 0, 0)
			if err != nil {
				return err
			}
			filters = f
			return nil
		})
		if err != nil {
			return nil, err
		}
		var ceventfilter pkgevents.NamespaceCloudEventFilter
		for _, ncef := range filters {
			if ncef.Name == filterName {
				ceventfilter = *ncef
			}
		}

		script = fmt.Sprintf("function filter() {\n %s \n}", ceventfilter.JSCode)

		flow.sugar.Debugf("adding filter cache key: %v\n", key)
		eventFilterCache.put(key, ceventfilter.JSCode)
	}

	var mapEvent map[string]interface{}
	err = json.Unmarshal(cloudevent, &mapEvent)
	if err != nil {
		return newBytesEvent, err
	}

	// create js runtime
	vm := goja.New()
	time.AfterFunc(1*time.Second, func() {
		vm.Interrupt("block event filter")
	})

	err = vm.Set("event", mapEvent)
	if err != nil {
		return newBytesEvent, fmt.Errorf("failed to initialize js runtime: %w", err)
	}

	// add logging function
	err = vm.Set("nslog", func(txt interface{}) {
		flow.logger.Infof(ctx, ns.ID, ns.GetAttributes(), fmt.Sprintf("%v", txt))
	})
	if err != nil {
		return newBytesEvent, fmt.Errorf("failed to initialize js runtime: %w", err)
	}

	_, err = vm.RunString(script)
	if err != nil {
		flow.logger.Errorf(ctx, ns.ID, ns.GetAttributes(), "CloudEvent filter '%s' produced an error (1): %v", filterName, err)
		return newBytesEvent, err
	}

	f, ok := goja.AssertFunction(vm.Get("filter"))
	if !ok {
		flow.logger.Errorf(ctx, ns.ID, ns.GetAttributes(), "cloudEvent filter '%s' error: %v", filterName, err)
		return newBytesEvent, err
	}

	newEventMap, err := f(goja.Undefined())
	if err != nil {
		flow.logger.Errorf(ctx, ns.ID, ns.GetAttributes(), "CloudEvent filter '%s' produced an error (2): %v", filterName, err)
		return newBytesEvent, err
	}

	retValue := newEventMap.Export()

	// event dropped
	if retValue == nil {
		return newBytesEvent, nil
	}

	newBytesEvent, err = json.Marshal(newEventMap)
	if err != nil {
		flow.logger.Errorf(ctx, ns.ID, ns.GetAttributes(), "CloudEvent filter '%s' produced an error (3): %v", filterName, err)
		return newBytesEvent, err
	}

	return newBytesEvent, nil
}

func (flow *flow) ApplyCloudEventFilter(ctx context.Context, in *grpc.ApplyCloudEventFilterRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	resp := new(emptypb.Empty)

	namespace := in.GetNamespace()
	filterName := in.GetFilterName()
	cloudevent := in.GetCloudevent()

	ns, err := flow.edb.NamespaceByName(ctx, namespace)
	if err != nil {
		return nil, err
	}

	b, err := flow.execFilter(ctx, namespace, filterName, cloudevent)
	if err != nil {
		flow.logger.Errorf(ctx, ns.ID, ns.GetAttributes(),
			"executing filter failed: %s", err.Error())
		return resp, err
	}

	// dropped event
	if len(b) == 0 {
		flow.logger.Debugf(ctx, ns.ID, ns.GetAttributes(),
			"dropping event %s", string(cloudevent))
		return resp, nil
	}

	flow.sugar.Debugf("event after script is %v", string(b))

	br := &grpc.BroadcastCloudeventRequest{
		Namespace:  namespace,
		Cloudevent: b,
		Timer:      0,
	}

	resp, err = flow.BroadcastCloudevent(ctx, br)

	return resp, err
}

func (flow *flow) DeleteCloudEventFilter(ctx context.Context, in *grpc.DeleteCloudEventFilterRequest) (*emptypb.Empty, error) {
	var resp emptypb.Empty

	namespace := in.GetNamespace()
	filterName := in.GetFilterName()

	ns, err := flow.edb.NamespaceByName(ctx, namespace)
	if err != nil {
		return nil, err
	}
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		return tx.DataStore().EventFilter().Delete(ctx, ns.ID, filterName)
	})
	if err != nil {
		return &resp, err
	}

	key := fmt.Sprintf("%s-%s", namespace, filterName)
	eventFilterCache.delete(key)
	flow.server.pubsub.Publish(&pubsub.PubsubUpdate{
		Handler: deleteFilterCache,
		Key:     key,
	})

	return &resp, err
}

const (
	deleteFilterCache          = "deleteFilterCache"
	deleteFilterCacheNamespace = "deleteFilterCacheNamespace"
)

func (flow *flow) deleteCache(req *pubsub.PubsubUpdate) {
	flow.sugar.Debugf("deleting filter cache key: %v\n", req.Key)
	eventFilterCache.delete(req.Key)
}

func deleteCacheNamespaceSync(delkey string) {
	eventFilterCache.value.Range(func(key, value any) bool {
		if strings.HasPrefix(key.(string), fmt.Sprintf("%s-", delkey)) {
			eventFilterCache.value.Delete(key.(string))
		}

		return true
	})
}

func (flow *flow) deleteCacheNamespace(req *pubsub.PubsubUpdate) {
	flow.sugar.Debugf("deleting filter cache for namespace: %v\n", req.Key)
	deleteCacheNamespaceSync(req.Key)
}

func (flow *flow) CreateCloudEventFilter(ctx context.Context, in *grpc.CreateCloudEventFilterRequest) (*emptypb.Empty, error) {
	var resp emptypb.Empty

	namespace := in.GetNamespace()
	filterName := in.GetFiltername()
	script := in.GetJsCode()

	fullScript := fmt.Sprintf("function filter() {\n %s \n}", script)

	// compiling js code is needed
	_, err := goja.Compile("filter", fullScript, false)
	if err != nil {
		err = status.Error(codes.FailedPrecondition, err.Error()) // precondition -> executable js script
		return &resp, err
	}

	ns, err := flow.edb.NamespaceByName(ctx, namespace)
	if err != nil {
		return nil, err
	}
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		return tx.DataStore().EventFilter().Create(ctx, ns.ID, filterName, script)
	})
	if err != nil {
		return &resp, err
	}

	key := fmt.Sprintf("%s-%s", namespace, filterName)
	flow.sugar.Debugf("adding filter cache key: %v\n", key)
	eventFilterCache.put(key, script)

	return &resp, err
}

func (flow *flow) GetCloudEventFilters(ctx context.Context, in *grpc.GetCloudEventFiltersRequest) (*grpc.GetCloudEventFiltersResponse, error) {
	var ls []*grpc.GetCloudEventFiltersResponse_EventFilter
	resp := new(grpc.GetCloudEventFiltersResponse)

	namespace := in.GetNamespace()

	ns, err := flow.edb.NamespaceByName(ctx, namespace)
	if err != nil {
		return nil, err
	}
	var res []*pkgevents.NamespaceCloudEventFilter

	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		le, _, err := tx.DataStore().EventFilter().Get(ctx, ns.ID, 0, 0)
		if err != nil {
			return err
		}
		res = le
		return nil
	})
	if err != nil {
		return nil, err
	}

	for _, s := range res {
		name := s.Name
		ls = append(ls, &grpc.GetCloudEventFiltersResponse_EventFilter{
			Name: name,
		})
	}

	resp.EventFilter = ls
	return resp, err
}

func (flow *flow) GetCloudEventFilterScript(ctx context.Context, in *grpc.GetCloudEventFilterScriptRequest) (*grpc.GetCloudEventFilterScriptResponse, error) {
	resp := new(grpc.GetCloudEventFilterScriptResponse)

	namespace := in.GetNamespace()
	filterName := in.GetName()

	ns, err := flow.edb.NamespaceByName(ctx, namespace)
	if err != nil {
		return nil, err
	}

	var filters []*pkgevents.NamespaceCloudEventFilter
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		f, _, err := tx.DataStore().EventFilter().Get(ctx, ns.ID, 0, 0)
		if err != nil {
			return err
		}
		filters = f
		return nil
	})
	if err != nil {
		return nil, err
	}
	var ceventfilter pkgevents.NamespaceCloudEventFilter
	for _, ncef := range filters {
		if ncef.Name == filterName {
			ceventfilter = *ncef
		}
	}

	resp.JsCode = ceventfilter.JSCode

	return resp, err
}

// func EventByteToCloudevent(byteEvent []byte) (event.Event, error) {
// 	ev := &event.Event{}
// 	err := json.Unmarshal(byteEvent, ev)
// 	return *ev, err

// }

func (c *CacheObject) get(key string) (string, bool) {
	v, ok := c.value.Load(key)
	var s string
	if ok {
		s, ok = v.(string)
		if ok {
			return s, true
		}
	}
	return "", false
}

func (c *CacheObject) put(key string, value string) {
	c.value.Store(key, value)
}

func (c *CacheObject) delete(key string) {
	c.value.Delete(key)
}
