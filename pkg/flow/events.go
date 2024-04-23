package flow

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	pkgevents "github.com/direktiv/direktiv/pkg/refactor/events"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func init() {
	gob.Register(new(event.EventContextV1))
	gob.Register(new(event.EventContextV03))
}

type events struct {
	*server
	appendStagingEvent func(ctx context.Context, events ...*pkgevents.StagingEvent) ([]*pkgevents.StagingEvent, []error)
}

func initEvents(srv *server, appendStagingEvent func(ctx context.Context, events ...*pkgevents.StagingEvent) ([]*pkgevents.StagingEvent, []error)) (*events, error) {
	events := new(events)

	events.server = srv
	events.appendStagingEvent = appendStagingEvent
	return events, nil
}

func (events *events) Close() error {
	return nil
}

func (events *events) handleEvent(ctx context.Context, ns uuid.UUID, nsName string, ce *cloudevents.Event) error {
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID()
	spanID := span.SpanContext().SpanID()
	slog := *slog.With("trace", traceID, "span", spanID, "namespace", nsName)
	e := pkgevents.EventEngine{
		WorkflowStart: func(workflowID uuid.UUID, ev ...*cloudevents.Event) {
			// events.metrics.InsertRecord

			slog.Debug("Starting workflow via CloudEvent.")
			_, end := traceMessageTrigger(ctx, "wf: "+workflowID.String())
			defer end()
			events.engine.EventsInvoke(workflowID, ev...)
		},
		WakeInstance: func(instanceID uuid.UUID, step int, ev []*cloudevents.Event) {
			// events.metrics.InsertRecord
			slog.Debug("invoking instance via cloudevent", "instance", instanceID)
			_, end := traceMessageTrigger(ctx, "ins: "+instanceID.String()+" step: "+fmt.Sprint(step))
			defer end()
			events.engine.wakeEventsWaiter(instanceID, ev)
		},
		GetListenersByTopic: func(ctx context.Context, s string) ([]*pkgevents.EventListener, error) {
			ctx, end := traceGetListenersByTopic(ctx, s)
			defer end()
			res := make([]*pkgevents.EventListener, 0)
			err := events.runSqlTx(ctx, func(tx *database.SQLStore) error {
				r, err := tx.DataStore().EventListenerTopics().GetListeners(ctx, s)
				if err != nil {
					slog.Error("Error fetching event-listener-topics.", "error", err)
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
			slog.Debug("Updating listeners starting.")
			err := events.runSqlTx(ctx, func(tx *database.SQLStore) error {
				errs := tx.DataStore().EventListener().UpdateOrDelete(ctx, listener)
				for _, err2 := range errs {
					if err2 != nil {
						slog.Debug("Error updating listeners.", "error", err2)
						return err2
					}
				}
				return nil
			})
			if err != nil {
				return []error{fmt.Errorf("%w", err)}
			}
			slog.Debug("Updating listeners complete.")
			return nil
		},
	}
	ctx, end := traceProcessingMessage(ctx)
	defer end()
	// tx, err := events.beginSqlTx(ctx)
	// if err != nil {
	// 	return err
	// }
	e.ProcessEvents(ctx, ns, []event.Event{*ce}, func(template string, args ...interface{}) {
		slog.Error(fmt.Sprintf(template, args...))
	})
	// tx.Commit(ctx)
	metricsCloudEventsCaptured.WithLabelValues(nsName, ce.Type(), ce.Source(), nsName).Inc()
	return nil
}

func (flow *flow) EventListeners(ctx context.Context, req *grpc.EventListenersRequest) (*grpc.EventListenersResponse, error) {
	slog.Debug("Handling gRPC request", "this", this())

	var resListeners []*pkgevents.EventListener
	var ns *datastore.Namespace
	var err error

	totalListeners := 0
	err = flow.runSqlTx(ctx, func(tx *database.SQLStore) error {
		ns, err = tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
		if err != nil {
			return err
		}

		var t int
		var li []*pkgevents.EventListener
		li, t, err = tx.DataStore().EventListener().Get(ctx, ns.ID, int(req.Pagination.Limit), int(req.Pagination.Offset))
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
	slog.Debug("Handling gRPC request", "this", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	var ns *datastore.Namespace
	var err error
	err = flow.runSqlTx(ctx, func(tx *database.SQLStore) error {
		ns, err = tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
		return err
	})
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeEventListeners(ns)
	defer flow.cleanup(sub.Close)
resend:
	var resListeners []*pkgevents.EventListener
	totalListeners := 0

	err = flow.runSqlTx(ctx, func(tx *database.SQLStore) error {
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
		err := srv.Send(resp)
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
	slog.Debug("Handling gRPC request", "this", this())
	ctx, end := startIncomingEvent(ctx, "flow")
	defer end()

	namespace := in.GetNamespace()
	rawevent := in.GetCloudevent()

	event := new(cloudevents.Event)
	ctx, endValidation := traceValidatingEvent(ctx)

	err := event.UnmarshalJSON(rawevent)
	if err != nil {
		endValidation()
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
			endValidation()
			return nil, status.Errorf(codes.InvalidArgument, "invalid cloudevent: %v", err)
		}
	}

	// NOTE: remarshal / unmarshal necessary to overcome issues with cloudevents library.
	data, err := json.Marshal(event)
	if err != nil {
		endValidation()
		return nil, status.Errorf(codes.InvalidArgument, "invalid cloudevent: %v", err)
	}

	err = event.UnmarshalJSON(data)
	if err != nil {
		endValidation()
		return nil, status.Errorf(codes.InvalidArgument, "invalid cloudevent: %v", err)
	}

	var ns *datastore.Namespace
	err = flow.runSqlTx(ctx, func(tx *database.SQLStore) error {
		ns, err = tx.DataStore().Namespaces().GetByName(ctx, namespace)
		endValidation()
		return err
	})
	if err != nil {
		endValidation()
		return nil, err
	}

	timer := in.GetTimer()
	endValidation()
	err = flow.events.BroadcastCloudevent(ctx, ns, event, timer)
	if err != nil {
		endValidation()
		return nil, status.Errorf(codes.Aborted, "cloudevent was not accepted: %v", err)
	}

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) HistoricalEvent(ctx context.Context, in *grpc.HistoricalEventRequest) (*grpc.HistoricalEventResponse, error) {
	slog.Debug("Handling gRPC request", "this", this())

	eid := in.GetId()
	if eid == "" {
		eid = uuid.NewString()
	}

	var cevent *pkgevents.Event
	var ns *datastore.Namespace
	var err error
	err = flow.runSqlTx(ctx, func(tx *database.SQLStore) error {
		ns, err = tx.DataStore().Namespaces().GetByName(ctx, in.GetNamespace())
		if err != nil {
			return err
		}
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

	resp.Cloudevent, err = cevent.Event.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// var cloudeventsOrderings = []*orderingInfo{
// 	{
// 		db:           "ReceivedAt",
// 		req:          "RECEIVED",
// 		defaultOrder: ent.Desc,
// 	},
// 	{
// 		db:           "id",
// 		req:          "ID",
// 		defaultOrder: ent.Asc,
// 	},
// }

const (
	contains = "CONTAINS"
	cr       = "CREATED"
	after    = "AFTER"
	before   = "BEFORE"
)

func (flow *flow) EventHistory(ctx context.Context, req *grpc.EventHistoryRequest) (*grpc.EventHistoryResponse, error) {
	slog.Debug("Handling gRPC request", "this", this())

	count := 0
	var res []*pkgevents.Event
	var err error
	var ns *datastore.Namespace
	err = flow.runSqlTx(ctx, func(tx *database.SQLStore) error {
		ns, err = tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
		if err != nil {
			return err
		}
		queryParams := []string{}
		for _, f := range req.Pagination.Filter {
			if f.Field == cr && f.Type == after {
				queryParams = append(queryParams, "received_after", f.Val)
			}
			if f.Field == cr && f.Type == before {
				queryParams = append(queryParams, "received_before", f.Val)
			}
			if f.Field == "TYPE" && f.Type == contains {
				queryParams = append(queryParams, "type_contains", f.Val)
			}
			if f.Field == "TEXT" && f.Type == contains {
				queryParams = append(queryParams, "event_contains", f.Val)
			}
		}
		re, t, err := tx.DataStore().EventHistory().Get(ctx, int(req.Pagination.Limit), int(req.Pagination.Offset), ns.ID, queryParams...)
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
		x := &grpc.Event{
			ReceivedAt: timestamppb.New(e.ReceivedAt),
			Id:         e.Event.ID(),
			Source:     e.Event.Source(),
			Type:       e.Event.Type(),
		}

		x.Cloudevent, err = e.Event.MarshalJSON()
		if err != nil {
			return nil, err
		}

		finalResults = append(finalResults, x)
	}
	resp.Events.Results = finalResults
	resp.Events.PageInfo = &grpc.PageInfo{Total: int32(count), Limit: req.Pagination.Limit, Offset: req.Pagination.Offset}

	return resp, nil
}

func (flow *flow) EventHistoryStream(req *grpc.EventHistoryRequest, srv grpc.Flow_EventHistoryStreamServer) error {
	slog.Debug("Handling gRPC request", "this", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	var err error
	var ns *datastore.Namespace
	err = flow.runSqlTx(ctx, func(tx *database.SQLStore) error {
		ns, err = tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
		return err
	})
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeEvents(ns)
	defer flow.cleanup(sub.Close)

resend:

	count := 0
	var res []*pkgevents.Event
	queryParams := []string{}
	for _, f := range req.Pagination.Filter {
		if f.Field == cr && f.Type == after {
			queryParams = append(queryParams, "received_after", f.Val)
		}
		if f.Field == cr && f.Type == before {
			queryParams = append(queryParams, "received_before", f.Val)
		}
		if f.Field == "TYPE" && f.Type == contains {
			queryParams = append(queryParams, "type_contains", f.Val)
		}
		if f.Field == "TEXT" && f.Type == contains {
			queryParams = append(queryParams, "event_contains", f.Val)
		}
	}
	err = flow.runSqlTx(ctx, func(tx *database.SQLStore) error {
		re, t, err := tx.DataStore().EventHistory().Get(ctx, int(req.Pagination.Limit), int(req.Pagination.Offset), ns.ID, queryParams...)
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
		x := &grpc.Event{
			ReceivedAt: timestamppb.New(e.ReceivedAt),
			Id:         e.Event.ID(),
			Source:     e.Event.Source(),
			Type:       e.Event.Type(),
		}

		x.Cloudevent, err = e.Event.MarshalJSON()
		if err != nil {
			return err
		}

		finalResults = append(finalResults, x)
	}
	resp.Events.Results = finalResults
	resp.Events.PageInfo = &grpc.PageInfo{Total: int32(count), Limit: req.Pagination.Limit, Offset: req.Pagination.Offset}

	nhash = bytedata.Checksum(resp)
	if nhash != phash {
		err := srv.Send(resp)
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
	slog.Debug("Handling gRPC request", "this", this())

	eid := req.GetId()
	if eid == "" {
		eid = uuid.NewString()
	}

	var cevent *pkgevents.Event
	var err error
	var ns *datastore.Namespace
	err = flow.runSqlTx(ctx, func(tx *database.SQLStore) error {
		ns, err = tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
		if err != nil {
			return err
		}

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

func (events *events) ReplayCloudevent(ctx context.Context, ns *datastore.Namespace, cevent *pkgevents.Event) error {
	event := cevent.Event
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID()
	spanID := span.SpanContext().SpanID()

	slog.Debug("Replaying event", "trace", traceID, "span", spanID, "namespace", ns.Name, "event", event.ID(), "event_type", event.Type(), "event_source", event.Source())

	err := events.handleEvent(ctx, ns.ID, ns.Name, event)
	if err != nil {
		return err
	}

	// Todo: Yassir, delete this entirely in cycle6.
	// if eventing is configured, event goes to knative event service
	// if it is from knative sink not
	// if events.server.config.EnableEventing && ctx.Value(EventingCtxKeySource) == nil {
	// 	PublishKnativeEvent(event)
	// }

	return nil
}

func (events *events) BroadcastCloudevent(ctx context.Context, ns *datastore.Namespace, event *cloudevents.Event, timer int64) error {
	// events.logger.Infof(ctx, ns.ID, database.GetAttributes(recipient.Namespace, ns), "Event received: %s (%s / %s) target time: %v", event.ID(), event.Type(), event.Source(), time.Unix(timer, 0))
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID()
	spanID := span.SpanContext().SpanID()

	slog.With("trace", traceID, "span", spanID, "namespace", "event", event.ID(), "event_type", event.Type(), "event_source", event.Source())

	metricsCloudEventsReceived.WithLabelValues(ns.Name, event.Type(), event.Source(), ns.Name).Inc()
	ctx, end := traceBrokerMessage(ctx, *event)
	defer end()

	err := events.addEvent(ctx, event, ns)
	if err != nil {
		return err
	}

	events.pubsub.NotifyEvents(ns)

	// handle event
	if timer == 0 {
		err = events.handleEvent(ctx, ns.ID, ns.Name, event)
		if err != nil {
			return err
		}
	} else {
		_, errs := events.appendStagingEvent(ctx, &pkgevents.StagingEvent{
			Event: &pkgevents.Event{
				Namespace:     ns.ID,
				Event:         event,
				ReceivedAt:    time.Now().UTC(),
				NamespaceName: ns.Name,
			},
			DatabaseID:   uuid.New(),
			DelayedUntil: time.Unix(timer, 0),
		})
		for _, err2 := range errs {
			if err2 != nil {
				slog.Error("failed to create delayed event", "error", err2)
			}
		}
	}

	// Todo: Yassir, delete this entirely in cycle6.
	// if eventing is configured, event goes to knative event service
	// if it is from knative sink not
	// if events.server.config.EnableEventing && ctx.Value(EventingCtxKeySource) == nil {
	// 	PublishKnativeEvent(event)
	// }

	return nil
}

func (events *events) listenForEvents(ctx context.Context, im *instanceMemory, ceds []*model.ConsumeEventDefinition, all bool) error {
	var transformedEvents []*model.ConsumeEventDefinition
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID()
	spanID := span.SpanContext().SpanID()

	slog.With("trace", traceID, "span", spanID, "namespace", im.Namespace(), "track", "namespace."+im.Namespace().Name, "instance", im.ID())

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

	err := events.addInstanceEventListener(ctx, im.Namespace().ID, im.Namespace().Name, im.GetInstanceID(), transformedEvents, im.Step(), all)
	if err != nil {
		return err
	}

	slog.Debug("Registered to receive events.")

	return nil
}
