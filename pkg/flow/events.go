package flow

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	cevents "github.com/direktiv/direktiv/pkg/flow/ent/cloudevents"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	hash "github.com/mitchellh/hashstructure/v2"
	glob "github.com/ryanuber/go-glob"
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

func initEvents(srv *server) (*events, error) {

	events := new(events)

	events.server = srv

	return events, nil

}

func (events *events) Close() error {
	return nil
}

func matchesExtensions(eventMap, extensions map[string]interface{}) bool {

	for k, f := range eventMap {

		if strings.HasPrefix(k, filterPrefix) {
			kt := strings.TrimPrefix(k, filterPrefix)

			if v, ok := extensions[kt]; ok {

				fs, ok := f.(string)
				vs, ok2 := v.(string)

				// if both are strings we can glob
				if ok && ok2 && !glob.Glob(fs, vs) {
					return false
				}

			} else {
				return false
			}

		}
	}

	return true

}

func hasEventInList(ev *cloudevents.Event, evl []*cloudevents.Event) bool {

	for _, e := range evl {

		if ev.Context.GetID() == e.Context.GetID() &&
			ev.Context.GetSource() == e.Context.GetSource() {
			return true
		}

	}

	return false

}

func (events *events) sendEvent(data []byte) {

	n := strings.SplitN(string(data), "/", 2)

	if len(n) != 2 {
		events.sugar.Errorf("namespace and id must be set for delayed events")
		return
	}

	nsid, err := uuid.Parse(n[1])
	if err != nil {
		events.sugar.Errorf("namespace id invalid")
		return
	}

	ctx := context.Background()

	ns, err := events.db.Namespace.Get(ctx, nsid)
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

	syncMtx.Lock()
	defer syncMtx.Unlock()

	// disable old timer
	events.timers.mtx.Lock()
	for i := range events.timers.timers {
		ti := events.timers.timers[i]
		if ti.name == "sendEventTimer" {
			events.timers.disableTimer(ti)
			break
		}
	}
	events.timers.mtx.Unlock()

	ctx := context.Background()

	for {
		e, err := events.getEarliestEvent(ctx, events.db.CloudEvents)
		if err != nil {
			if IsNotFound(err) {
				return
			}

			events.sugar.Errorf("can not sync event delays: %v", err)
			return
		}

		if e.Fire.Before(time.Now()) {
			err = events.flushEvent(ctx, e.EventId, e.Edges.Namespace, false)
			if err != nil {
				events.sugar.Errorf("can not flush event %s: %v", e.ID, err)
			}
			continue
		}

		err = events.timers.addOneShot("sendEventTimer", sendEventFunction,
			e.Fire, []byte(fmt.Sprintf("%s/%s", e.ID, e.Edges.Namespace.ID.String())))
		if err != nil {
			events.sugar.Errorf("can not reschedule event timer: %v", err)
		}

		break

	}

}

func (events *events) flushEvent(ctx context.Context, eventID string, ns *ent.Namespace, rearm bool) error {

	tx, err := events.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	e, err := events.markEventAsProcessed(ctx, tx.CloudEvents, eventID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	defer func(r bool) {
		if r {
			events.syncEventDelays()
		}
	}(rearm)

	err = events.handleEvent(ns, e)
	if err != nil {
		return err
	}

	return nil

}

func (events *events) updateMultipleEvents(ce *cloudevents.Event, id uuid.UUID,
	correlations []string) ([]*cloudevents.Event, error) {

	var retEvents []*cloudevents.Event
	db := events.db.DB()

	chash := generateCorrelationHash(ce, ce.Type(), correlations)

	data, err := eventToBytes(*ce)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(`update events_waits
	set events = jsonb_set(events, $1, $2, true)
	WHERE events::jsonb ? $3 and events_wfeventswait = $4
	returning *`, fmt.Sprintf("{%s}", chash), fmt.Sprintf(`"%s"`, base64.StdEncoding.EncodeToString(data)), chash, id)
	if err != nil {
		return retEvents, err
	}

	rc := 0

	ctx := context.Background()

	for rows.Next() {
		rc++

		var (
			id         uuid.UUID
			eventID    string
			eventsData []byte
		)

		err := rows.Scan(&id, &eventsData, &eventID)
		if err != nil {
			events.sugar.Errorf("can not scan result: %v", err)
			continue
		}

		var eventsIn map[string]interface{}
		err = json.Unmarshal(eventsData, &eventsIn)
		if err != nil {
			events.sugar.Errorf("can not unmarshal existing events")
			return retEvents, err
		}

		counter := len(eventsIn)

		for _, v := range eventsIn {
			if v != nil {
				counter--
			}
		}

		// all events have values so we can start the workflow
		if counter == 0 {

			err := events.db.EventsWait.DeleteOneID(id).Exec(ctx)
			if err != nil {
				events.sugar.Error(err)
				continue
			}

			// get data for all events
			var e map[string]string
			json.Unmarshal(eventsData, &e)

			for _, g := range e {
				b, err := base64.StdEncoding.DecodeString(g)
				if err != nil {
					events.sugar.Errorf("event data corrupt: %v", err)
					continue
				}

				ev, err := bytesToEvent(b)
				if err != nil {
					events.sugar.Errorf("event data corrupt: %v", err)
					continue
				}

				if !hasEventInList(ev, retEvents) {
					retEvents = append(retEvents, ev)
				}
			}

		}

	}

	return retEvents, nil

}

func (events *events) handleEvent(ns *ent.Namespace, ce *cloudevents.Event) error {

	var (
		id                                          uuid.UUID
		count                                       int
		singleEvent, corBytes, allEvents, signature []byte
		wf                                          string
		captured                                    bool
	)

	db := events.db.DB()

	rows, err := db.Query(`select
	we.oid, signature, count, correlations, we.events, workflow_wfevents, v
	from events we
	inner join workflows w
		on w.oid = workflow_wfevents
	inner join namespaces n
		on n.oid = w.namespace_workflows,
	jsonb_array_elements(events) as v
	where v::json->>'type' = $1
	and n.oid = $2`, ce.Type(), ns.ID.String())
	if err != nil {
		return err
	}
	defer rows.Close()

	ctx := context.Background()

	var conn *sql.Conn
	for rows.Next() {

		err := rows.Scan(&id, &signature, &count, &corBytes, &allEvents, &wf, &singleEvent)
		if err != nil {
			events.sugar.Errorf("process row error: %v", err)
			continue
		}

		hash, _ := hash.Hash(fmt.Sprintf("%d%v%v", id, allEvents, wf), hash.FormatV2, nil)

		unlock := func() {
			if conn != nil {
				events.locks.unlockDB(hash, conn)
			}
		}

		conn, err = events.locks.lockDB(hash, int(defaultLockWait.Seconds()))
		if err != nil {
			events.sugar.Errorf("can not lock event row: %d, %v", id, err)
			continue
		}

		events.sugar.Debugf("event listener %s is candidate", id.String())

		var eventMap map[string]interface{}
		err = json.Unmarshal(singleEvent, &eventMap)
		if err != nil {
			unlock()
			events.sugar.Errorf("can not marshall event map: %v", err)
			continue
		}

		// adding source for comparison
		m := ce.Context.GetExtensions()

		// if there is none, we need to create one for source
		if m == nil {
			m = make(map[string]interface{})
		}

		m["source"] = ce.Context.GetSource()

		// check filters
		if !matchesExtensions(eventMap, m) {
			events.sugar.Debugf("event listener %d does not match", id)
			unlock()
			continue
		}

		captured = true

		var ae []map[string]interface{}
		json.Unmarshal(allEvents, &ae)

		var retEvents []*cloudevents.Event

		if count == 1 {

			retEvents = append(retEvents, ce)

		} else {

			var correlations []string

			// get correlations
			if len(corBytes) > 0 {
				json.Unmarshal(corBytes, &correlations)
			}

			es, err := events.updateMultipleEvents(ce, id, correlations)
			if err != nil {
				events.sugar.Errorf("can not handle multi event: %v", err)
				unlock()
				continue
			}

			retEvents = append(retEvents, es...)

			// no update executed, means we have a candidate but no existing events
			// for a multi event workflow
			if len(retEvents) == 0 {

				events.sugar.Debugf("no events waiting")

				// get event types
				var eventTypes []string
				for _, g := range ae {
					eventTypes = append(eventTypes, g[eventTypeString].(string))
				}

				// only add if the correlation id exists
				if generateCorrelationHash(ce, ce.Type(), correlations) != "" {
					err := events.addEventListenerWait(ctx, ce, id, correlations, eventTypes)
					if err != nil {
						events.sugar.Errorf("can not create workflow event wait: %v", err)
						unlock()
						continue
					}
				}

			}

		}

		unlock()

		// if single or multiple added events we fire
		if len(retEvents) > 0 {

			if len(signature) == 0 {

				go events.engine.EventsInvoke(wf, retEvents...)

			} else {

				d, err := events.reverseTraverseToWorkflow(ctx, wf)
				if err != nil {
					events.engine.sugar.Error(err)
					return nil
				}

				err = events.deleteWorkflowEventListeners(ctx, events.db.Events, d.wf)
				if err != nil {
					events.engine.sugar.Error(err)
					return nil
				}

				go events.engine.wakeEventsWaiter(signature, retEvents)

			}

		}

	}

	if captured {
		metricsCloudEventsCaptured.WithLabelValues(ns.Name, ce.Type(), ce.Source(), ns.Name).Inc()
	}

	return nil

}

func generateCorrelationHash(cevent *cloudevents.Event,
	ets string, correlations []string) string {

	hashBase := make(map[string]interface{})

	// check if the correlation id exists and generate the struct for the correlation hash
	for _, k := range correlations {
		if cevent.Extensions()[strings.ToLower(k)] != nil {
			hashBase[k] = fmt.Sprintf("%v", cevent.Extensions()[strings.ToLower(k)])
		} else {
			return ""
		}
	}

	hashBase[eventTypeString] = ets
	h, _ := hash.Hash(hashBase, hash.FormatV2, nil)

	return fmt.Sprintf("%d", h)

}

func (events *events) addEventListenerWait(ctx context.Context, cevent *cloudevents.Event, id uuid.UUID,
	correlations, eventTypes []string) error {

	sevents := make(map[string]interface{})

	for _, v := range eventTypes {
		if v == cevent.Type() {
			data, err := eventToBytes(*cevent)
			if err != nil {
				return err
			}
			sevents[generateCorrelationHash(cevent, v, correlations)] = base64.StdEncoding.EncodeToString(data)
		} else {
			sevents[generateCorrelationHash(cevent, v, correlations)] = nil
		}
	}

	err := events.addWorkflowEventWait(ctx, events.db.EventsWait, sevents, 1, id)
	if err != nil {
		return err
	}

	return nil

}

func eventToBytes(cevent cloudevents.Event) ([]byte, error) {

	var ev bytes.Buffer

	enc := gob.NewEncoder(&ev)
	err := enc.Encode(cevent)
	if err != nil {
		return nil, fmt.Errorf("can not convert event to bytes: %v", err)
	}

	return ev.Bytes(), nil

}

func bytesToEvent(b []byte) (*cloudevents.Event, error) {

	ev := new(cloudevents.Event)

	enc := gob.NewDecoder(bytes.NewReader(b))
	err := enc.Decode(ev)
	if err != nil {
		return nil, fmt.Errorf("can not convert bytes to event: %v", err)
	}

	return ev, nil
}

func eventListenersOrder(p *pagination) ent.EventsPaginateOption {

	field := ent.EventsOrderFieldUpdatedAt
	direction := ent.OrderDirectionAsc

	if p.order != nil {

		if x := p.order.Field; x != "" && x == "UPDATED" {
			field = ent.EventsOrderFieldUpdatedAt
		}

	}

	return ent.WithEventsOrder(&ent.EventsOrder{
		Direction: direction,
		Field:     field,
	})

}

func eventListenersFilter(p *pagination) ent.EventsPaginateOption {

	if p.filter == nil {
		return nil
	}

	filter := p.filter.Val

	return ent.WithEventsFilter(func(query *ent.EventsQuery) (*ent.EventsQuery, error) {

		if filter == "" {
			return query, nil
		}

		field := p.filter.Field
		if field == "" {
			return query, nil
		}

		// switch field {
		// case "NAME":

		// 	ftype := p.filter.Type
		// 	if ftype == "" {
		// 		return query, nil
		// 	}

		// 	switch ftype {
		// 	case "CONTAINS":
		// 		return query.Where(entns.NameContains(filter)), nil
		// 	}
		// }

		return query, nil

	})

}

func (flow *flow) EventListeners(ctx context.Context, req *grpc.EventListenersRequest) (*grpc.EventListenersResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	namespace := req.GetNamespace()

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	opts := []ent.EventsPaginateOption{}
	opts = append(opts, eventListenersOrder(p))
	filter := eventListenersFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

	nsc := flow.db.Namespace

	ns, err := flow.getNamespace(ctx, nsc, namespace)
	if err != nil {
		return nil, err
	}

	query := ns.QueryNamespacelisteners()
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return nil, err
	}

	var resp grpc.EventListenersResponse

	err = atob(cx, &resp)
	if err != nil {
		return nil, err
	}

	resp.Namespace = namespace

	m := make(map[string]string)

	for idx, edge := range cx.Edges {

		// resp.Edges[idx].Node.UpdatedAt = edge.Node.UpdatedAt

		in, _ := edge.Node.Instance(ctx)
		if in != nil {
			resp.Edges[idx].Node.Instance = in.ID.String()
		}

		wf, err := edge.Node.Workflow(ctx)
		if err != nil {
			return nil, err
		}

		path, exists := m[wf.ID.String()]
		if !exists {
			wfd, err := flow.reverseTraverseToWorkflow(ctx, wf.ID.String())
			if err != nil {
				return nil, err
			}
			path = wfd.path
			m[wf.ID.String()] = path
		}

		resp.Edges[idx].Node.Workflow = path

		resp.Edges[idx].Node.Mode = "or"
		if edge.Node.Count > 1 {
			resp.Edges[idx].Node.Mode = "and"
		}

		edefs := make([]*grpc.EventDef, 0)
		for _, ev := range edge.Node.Events {

			var et string
			if v, ok := ev["type"]; ok {
				et, _ = v.(string)
			}

			delete(ev, "type")

			filters := make(map[string]string)

			for k, v := range ev {
				if !strings.HasPrefix(k, "filter-") {
					continue
				}
				k = strings.TrimPrefix(k, "filter-")
				if s, ok := v.(string); ok {
					filters[k] = s
				}
			}

			edefs = append(edefs, &grpc.EventDef{
				Type:    et,
				Filters: filters,
			})

		}

		resp.Edges[idx].Node.Events = edefs

	}

	return &resp, nil

}

func (flow *flow) EventListenersStream(req *grpc.EventListenersRequest, srv grpc.Flow_EventListenersStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	namespace := req.GetNamespace()

	p, err := getPagination(req.Pagination)
	if err != nil {
		return err
	}

	opts := []ent.EventsPaginateOption{}
	opts = append(opts, eventListenersOrder(p))
	filter := eventListenersFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

	nsc := flow.db.Namespace

	ns, err := flow.getNamespace(ctx, nsc, namespace)
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeEventListeners(ns)
	defer flow.cleanup(sub.Close)

resend:

	query := ns.QueryNamespacelisteners()
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return err
	}

	var resp = &grpc.EventListenersResponse{}

	err = atob(cx, &resp)
	if err != nil {
		return err
	}

	resp.Namespace = namespace

	m := make(map[string]string)

	for idx, edge := range cx.Edges {

		// resp.Edges[idx].Node.UpdatedAt = edge.Node.UpdatedAt

		in, _ := edge.Node.Instance(ctx)
		if in != nil {
			resp.Edges[idx].Node.Instance = in.ID.String()
		}

		wf, err := edge.Node.Workflow(ctx)
		if err != nil {
			return err
		}

		path, exists := m[wf.ID.String()]
		if !exists {
			wfd, err := flow.reverseTraverseToWorkflow(ctx, wf.ID.String())
			if err != nil {
				return err
			}
			path = wfd.path
			m[wf.ID.String()] = path
		}

		resp.Edges[idx].Node.Workflow = path

		resp.Edges[idx].Node.Mode = "or"
		if edge.Node.Count > 1 {
			resp.Edges[idx].Node.Mode = "and"
		}

		edefs := make([]*grpc.EventDef, 0)
		for _, ev := range edge.Node.Events {

			var et string
			if v, ok := ev["type"]; ok {
				et, _ = v.(string)
			}

			delete(ev, "type")

			filters := make(map[string]string)

			for k, v := range ev {
				if !strings.HasPrefix(k, "filter-") {
					continue
				}
				k = strings.TrimPrefix(k, "filter-")
				if s, ok := v.(string); ok {
					filters[k] = s
				}
			}

			edefs = append(edefs, &grpc.EventDef{
				Type:    et,
				Filters: filters,
			})

		}

		resp.Edges[idx].Node.Events = edefs

	}

	nhash = checksum(resp)
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

	ns, err := flow.getNamespace(ctx, flow.db.Namespace, namespace)
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

	namespace := in.GetNamespace()
	eid := in.GetId()

	ns, err := flow.getNamespace(ctx, flow.db.Namespace, namespace)
	if err != nil {
		return nil, err
	}

	cevent, err := ns.QueryCloudevents().Where(cevents.EventIdEQ(eid)).Only(ctx)
	if err != nil {
		return nil, err
	}

	var resp grpc.HistoricalEventResponse

	resp.Id = eid
	resp.Namespace = namespace
	resp.ReceivedAt = timestamppb.New(cevent.Created)

	resp.Source = cevent.Event.Source()
	resp.Type = cevent.Event.Type()

	resp.Cloudevent = []byte(cevent.Event.String())

	return &resp, nil

}

func cloudeventsOrder(p *pagination) ent.CloudEventsPaginateOption {

	field := ent.CloudEventsOrderFieldCreated
	direction := ent.OrderDirectionDesc

	if p.order != nil {

		if x := p.order.Field; x != "" && x == "ID" {
			field = ent.CloudEventsOrderFieldID
		}

		if x := p.order.Field; x != "" && x == "RECEIVED" {
			field = ent.CloudEventsOrderFieldCreated
		}

		if x := p.order.Direction; x != "" && x == "DESC" {
			direction = ent.OrderDirectionDesc
		}

		if x := p.order.Direction; x != "" && x == "ASC" {
			direction = ent.OrderDirectionAsc
		}

	}

	return ent.WithCloudEventsOrder(&ent.CloudEventsOrder{
		Direction: direction,
		Field:     field,
	})

}

func cloudeventsFilter(p *pagination) ent.CloudEventsPaginateOption {

	if p.filter == nil {
		return nil
	}

	filter := p.filter.Val

	return ent.WithCloudEventsFilter(func(query *ent.CloudEventsQuery) (*ent.CloudEventsQuery, error) {

		if filter == "" {
			return query, nil
		}

		field := p.filter.Field
		if field == "" {
			return query, nil
		}

		// switch field {
		// case "AS":

		// 	ftype := p.filter.Type
		// 	if ftype == "" {
		// 		return query, nil
		// 	}

		// 	switch ftype {
		// 	case "CONTAINS":
		// 		return query.Where(entcevents.AsContains(filter)), nil
		// 	}
		// }

		return query, nil

	})

}

func (flow *flow) EventHistory(ctx context.Context, req *grpc.EventHistoryRequest) (*grpc.EventHistoryResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	opts := []ent.CloudEventsPaginateOption{}
	opts = append(opts, cloudeventsOrder(p))
	filter := cloudeventsFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

	nsc := flow.db.Namespace
	ns, err := flow.getNamespace(ctx, nsc, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	query := ns.QueryCloudevents()
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return nil, err
	}

	var resp grpc.EventHistoryResponse
	resp.Events = new(grpc.Events)
	resp.Events.PageInfo = new(grpc.PageInfo)
	resp.Namespace = ns.Name

	err = atob(cx.PageInfo, resp.Events.PageInfo)
	if err != nil {
		return nil, err
	}

	resp.Events.TotalCount = int32(cx.TotalCount)

	for _, x := range cx.Edges {

		edge := new(grpc.EventsEdge)
		resp.Events.Edges = append(resp.Events.Edges, edge)

		err = atob(x.Cursor, &edge.Cursor)
		if err != nil {
			return nil, err
		}

		edge.Node = new(grpc.Event)
		edge.Node.Id = x.Node.EventId
		edge.Node.ReceivedAt = timestamppb.New(x.Node.Created)
		edge.Node.Source = x.Node.Event.Source()
		edge.Node.Type = x.Node.Event.Type()
		edge.Node.Cloudevent = []byte(x.Node.Event.String())

	}

	return &resp, nil

}

func (flow *flow) EventHistoryStream(req *grpc.EventHistoryRequest, srv grpc.Flow_EventHistoryStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	p, err := getPagination(req.Pagination)
	if err != nil {
		return err
	}

	opts := []ent.CloudEventsPaginateOption{}
	opts = append(opts, cloudeventsOrder(p))
	filter := cloudeventsFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

	nsc := flow.db.Namespace
	ns, err := flow.getNamespace(ctx, nsc, req.GetNamespace())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeEvents(ns)
	defer flow.cleanup(sub.Close)

resend:

	query := ns.QueryCloudevents()
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return err
	}

	resp := new(grpc.EventHistoryResponse)
	resp.Events = new(grpc.Events)
	resp.Events.PageInfo = new(grpc.PageInfo)
	resp.Namespace = ns.Name

	err = atob(cx.PageInfo, resp.Events.PageInfo)
	if err != nil {
		return err
	}

	resp.Events.TotalCount = int32(cx.TotalCount)

	for _, x := range cx.Edges {

		edge := new(grpc.EventsEdge)
		resp.Events.Edges = append(resp.Events.Edges, edge)

		err = atob(x.Cursor, &edge.Cursor)
		if err != nil {
			return err
		}

		edge.Node = new(grpc.Event)
		edge.Node.Id = x.Node.EventId
		edge.Node.ReceivedAt = timestamppb.New(x.Node.Created)
		edge.Node.Source = x.Node.Event.Source()
		edge.Node.Type = x.Node.Event.Type()
		edge.Node.Cloudevent = []byte(x.Node.Event.String())

	}

	nhash = checksum(resp)
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

	nsc := flow.db.Namespace
	ns, err := flow.getNamespace(ctx, nsc, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	eid := req.GetId()

	cevent, err := ns.QueryCloudevents().Where(cevents.EventIdEQ(eid)).Only(ctx)
	if err != nil {
		return nil, err
	}

	err = flow.events.ReplayCloudevent(ctx, ns, cevent)
	if err != nil {
		return nil, err
	}

	var resp emptypb.Empty

	return &resp, nil

}

func (events *events) ReplayCloudevent(ctx context.Context, ns *ent.Namespace, cevent *ent.CloudEvents) error {

	event := cevent.Event

	events.logToNamespace(ctx, time.Now(), ns, "Replaying event: %s (%s / %s)", event.ID(), event.Type(), event.Source())

	err := events.handleEvent(ns, &event)
	if err != nil {
		return err
	}

	// if eventing is configured, event goes to knative event service
	// if it is from knative sink not
	if events.server.conf.Eventing && ctx.Value(EventingCtxKeySource) == nil {
		PublishKnativeEvent(&event)
	}

	return nil

}

func (events *events) BroadcastCloudevent(ctx context.Context, ns *ent.Namespace, event *cloudevents.Event, timer int64) error {

	events.logToNamespace(ctx, time.Now(), ns, "Event received: %s (%s / %s)", event.ID(), event.Type(), event.Source())

	metricsCloudEventsReceived.WithLabelValues(ns.Name, event.Type(), event.Source(), ns.Name).Inc()

	// add event to db
	err := events.addEvent(ctx, events.db.CloudEvents, event, ns, timer)
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

const pubsubUpdateEventDelays = "updateEventDelays"

func (events *events) updateEventDelaysHandler(req *PubsubUpdate) {

	events.syncEventDelays()

}

type eventsWaiterSignature struct {
	InstanceID string
	Step       int
}

type eventsResultMessage struct {
	InstanceID string
	State      string
	Step       int
	Payloads   []*cloudevents.Event
}

func (events *events) listenForEvents(ctx context.Context, im *instanceMemory, ceds []*model.ConsumeEventDefinition, all bool) error {

	signature, err := json.Marshal(&eventsWaiterSignature{
		InstanceID: im.in.ID.String(),
		Step:       im.Step(),
	})
	if err != nil {
		return err
	}

	var transformedEvents []*model.ConsumeEventDefinition

	for i := range ceds {

		ev := new(model.ConsumeEventDefinition)
		ev.Context = make(map[string]interface{})

		err = copier.Copy(ev, ceds[i])
		if err != nil {
			return err
		}

		for k, v := range ceds[i].Context {

			ev.Context[k], err = jqOne(im.data, v)
			if err != nil {
				return fmt.Errorf("failed to execute jq query for key '%s' on event definition %d: %v", k, i, err)
			}

		}

		transformedEvents = append(transformedEvents, ev)

	}

	wf, err := events.engine.InstanceWorkflow(ctx, im)
	if err != nil {
		return err
	}

	err = events.addInstanceEventListener(ctx, events.db.Events, wf, im.in,
		transformedEvents, signature, all)
	if err != nil {
		return err
	}

	events.logToInstance(ctx, time.Now(), im.in, "Registered to receive events.")

	return nil

}
