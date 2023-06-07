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
	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	enteventsfilter "github.com/direktiv/direktiv/pkg/flow/ent/cloudeventfilters"
	cevents "github.com/direktiv/direktiv/pkg/flow/ent/cloudevents"
	entevents "github.com/direktiv/direktiv/pkg/flow/ent/events"
	entinst "github.com/direktiv/direktiv/pkg/flow/ent/instance"
	entns "github.com/direktiv/direktiv/pkg/flow/ent/namespace"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/dop251/goja"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	hash "github.com/mitchellh/hashstructure/v2"
	"github.com/ryanuber/go-glob"
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

	cached := new(database.CacheData)

	err = events.database.Namespace(ctx, cached, id)
	if err != nil {
		events.sugar.Error(err)
		return
	}

	err = events.flushEvent(ctx, n[0], cached.Namespace, true)
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
		e, err := events.getEarliestEvent(ctx)
		if err != nil {
			if derrors.IsNotFound(err) {
				return
			}

			events.sugar.Errorf("can not sync event delays: %v", err)
			return
		}

		cached := new(database.CacheData)
		err = events.database.Namespace(ctx, cached, e.Edges.Namespace.ID)
		if err != nil {
			return
		}

		if e.Fire.Before(time.Now()) {
			err = events.flushEvent(ctx, e.EventId, cached.Namespace, false)
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

func (events *events) flushEvent(ctx context.Context, eventID string, ns *database.Namespace, rearm bool) error {
	tctx, tx, err := events.database.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	e, err := events.markEventAsProcessed(tctx, ns, eventID)
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

func (events *events) handleEventLoopLogic(ctx context.Context, rows *sql.Rows, ce *cloudevents.Event) {
	var (
		id                                uuid.UUID
		count                             int
		singleEvent, allEvents, signature []byte
		wf                                string
	)

	err := rows.Scan(&id, &signature, &count, &allEvents, &wf, &singleEvent)
	if err != nil {
		events.sugar.Errorf("process row error: %v", err)
		return
	}

	hash, err := hash.Hash(fmt.Sprintf("%d%v%v", id, allEvents, wf), hash.FormatV2, nil)
	if err != nil {
		events.sugar.Errorf("failed to generate hash: %v", err)
		return
	}

	conn, err := events.locks.lockDB(hash, int(defaultLockWait.Seconds()))
	if err != nil {
		events.sugar.Errorf("can not lock event row: %d, %v", id, err)
		return
	}

	unlock := func(conn *sql.Conn, hash uint64) {
		err = events.locks.unlockDB(hash, conn)
		if err != nil {
			events.sugar.Errorf("events mutex unlock error: %v", err)
		}
	}
	defer unlock(conn, hash)

	events.sugar.Debugf("event listener %s is candidate", id.String())

	var eventMap map[string]interface{}
	err = json.Unmarshal(singleEvent, &eventMap)
	if err != nil {
		events.sugar.Errorf("can not marshall event map: %v", err)
		return
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
		events.sugar.Debugf("event listener %s does not match", id.String())
		return
	}

	// deleteEventListener = append(deleteEventListener, id)

	var ae []map[string]interface{}
	err = json.Unmarshal(allEvents, &ae)
	if err != nil {
		events.sugar.Errorf("failed to unmarshal events: %v", err)
		return
	}

	var retEvents []*cloudevents.Event

	if count == 1 {
		retEvents = append(retEvents, ce)
	} else {
		var eventMapAll []map[string]interface{}
		err = json.Unmarshal(allEvents, &eventMapAll) // why are we doing this again?
		if err != nil {
			events.sugar.Errorf("failed to unmarshal events: %v", err)
			return
		}

		// set value
		updateItem := eventMapAll[int(eventMap["idx"].(float64))]

		data, err := eventToBytes(*ce)
		if err != nil {
			events.sugar.Errorf("can not update convert event: %v", err)
			return
		}

		updateItem["time"] = time.Now().Unix()
		updateItem["value"] = base64.StdEncoding.EncodeToString(data)

		needsUpdate := false
		for _, v := range eventMapAll {
			// if there is one entry without value we can skip this instance
			// won't fire anyways
			if v["value"] == "" {
				needsUpdate = true
				break
			} else {
				d, err := base64.StdEncoding.DecodeString(v["value"].(string))
				if err != nil {
					events.sugar.Errorf("cannot decode eventmap base64: %v", err)
					// continue // suspicious
					return
				}

				ce, err := bytesToEvent(d)
				if err != nil {
					events.sugar.Errorf("cannot unmarshal bytes to event: %v", err)
					// continue // suspicious
					return
				}

				retEvents = append(retEvents, ce)
			}
		}

		if needsUpdate {
			err = events.updateInstanceEventListener(ctx, id, eventMapAll)
			if err != nil {
				events.sugar.Errorf("can not update multi event: %v", err)
			}
			return
		}
	}

	// if single or multiple added events we fire
	if len(retEvents) > 0 {
		if len(signature) == 0 {
			go events.engine.EventsInvoke(wf, retEvents...)
		} else {
			id, err := uuid.Parse(wf)
			if err != nil {
				events.engine.sugar.Error(err)
				return
			}

			fStore, _, _, rollback, err := events.flow.beginSqlTx(ctx)
			if err != nil {
				events.engine.sugar.Error(err)
				return
			}
			defer rollback()

			file, err := fStore.GetFile(ctx, id)
			if err != nil {
				events.engine.sugar.Error(err)
				return
			}
			rollback()

			err = events.deleteEventListeners(ctx, file.RootID, id)
			if err != nil {
				events.engine.sugar.Error(err)
				return
			}

			go events.engine.wakeEventsWaiter(signature, retEvents)
		}
	}
}

func (events *events) handleEvent(ns *database.Namespace, ce *cloudevents.Event) error {
	db := events.edb.DB()

	// we have to select first because of the glob feature
	// this gives a basic list of eligible workflow instances waiting
	// we get all
	rows, err := db.Query(`select
	we.oid, signature, count, we.events, workflow_id, v
	from events we
	inner join filesystem_files w
		on w.id = workflow_id
	inner join namespaces n
		on n.oid = w.root_id,
	jsonb_array_elements(events) as v
	where v::json->>'type' = $1 and v::json->>'value' = ''
	and n.oid = $2`, ce.Type(), ns.ID.String())
	if err != nil {
		return err
	}
	defer rows.Close()

	ctx := context.Background()

	for rows.Next() {
		events.handleEventLoopLogic(ctx, rows, ce)
	}

	metricsCloudEventsCaptured.WithLabelValues(ns.Name, ce.Type(), ce.Source(), ns.Name).Inc()

	return nil
}

func eventToBytes(cevent cloudevents.Event) ([]byte, error) {
	var ev bytes.Buffer

	enc := gob.NewEncoder(&ev)
	err := enc.Encode(cevent)
	if err != nil {
		return nil, fmt.Errorf("can not convert event to bytes: %w", err)
	}

	return ev.Bytes(), nil
}

func bytesToEvent(b []byte) (*cloudevents.Event, error) {
	ev := new(cloudevents.Event)

	enc := gob.NewDecoder(bytes.NewReader(b))
	err := enc.Decode(ev)
	if err != nil {
		return nil, fmt.Errorf("can not convert bytes to event: %w", err)
	}

	return ev, nil
}

var eventListenersOrderings = []*orderingInfo{
	{
		db:           entevents.FieldUpdatedAt,
		req:          "UPDATED",
		defaultOrder: ent.Asc,
	},
}

var eventListenersFilters = map[*filteringInfo]func(query *ent.EventsQuery, v string) (*ent.EventsQuery, error){
	{
		field: "CREATED",
		ftype: "BEFORE",
	}: func(query *ent.EventsQuery, v string) (*ent.EventsQuery, error) {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, err
		}
		return query.Where(entevents.CreatedAtGTE(t)), nil
	},
	{
		field: "CREATED",
		ftype: "AFTER",
	}: func(query *ent.EventsQuery, v string) (*ent.EventsQuery, error) {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, err
		}
		return query.Where(entevents.CreatedAtLTE(t)), nil
	},
}

func (flow *flow) EventListeners(ctx context.Context, req *grpc.EventListenersRequest) (*grpc.EventListenersResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(ctx)
	query := clients.Events.Query().Where(entevents.HasNamespaceWith(entns.ID(cached.Namespace.ID))).WithInstance(func(q *ent.InstanceQuery) {
		q.Select(entinst.FieldID)
	})

	results, pi, err := paginate[*ent.EventsQuery, *ent.Events](ctx, req.Pagination, query, eventListenersOrderings, eventListenersFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.EventListenersResponse)
	resp.Namespace = cached.Namespace.Name
	resp.PageInfo = pi

	err = bytedata.ConvertDataForOutput(results, &resp.Results)
	if err != nil {
		return nil, err
	}

	m := make(map[string]string)

	for idx, result := range results {
		if result.Edges.Instance != nil {
			resp.Results[idx].Instance = result.Edges.Instance.ID.String()
		} else {
			wfID := result.WorkflowID
			path, exists := m[wfID.String()]
			if !exists {
				fStore, _, _, rollback, err := flow.beginSqlTx(ctx)
				if err != nil {
					return nil, err
				}
				defer rollback()

				file, err := fStore.GetFile(ctx, wfID)
				if err != nil {
					return nil, err
				}
				rollback()

				path = file.Path
				m[wfID.String()] = path
			}

			resp.Results[idx].Workflow = path
		}

		resp.Results[idx].Mode = "or"
		if result.Count > 1 {
			resp.Results[idx].Mode = "and"
		}

		edefs := make([]*grpc.EventDef, 0)
		for _, ev := range result.Events {
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

		resp.Results[idx].Events = edefs
	}

	return resp, nil
}

func (flow *flow) EventListenersStream(req *grpc.EventListenersRequest, srv grpc.Flow_EventListenersStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeEventListeners(cached.Namespace)
	defer flow.cleanup(sub.Close)

	clients := flow.edb.Clients(ctx)

resend:

	query := clients.Events.Query().Where(entevents.HasNamespaceWith(entns.ID(cached.Namespace.ID))).WithInstance(func(q *ent.InstanceQuery) {
		q.Select(entinst.FieldID)
	})

	results, pi, err := paginate[*ent.EventsQuery, *ent.Events](ctx, req.Pagination, query, eventListenersOrderings, eventListenersFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.EventListenersResponse)
	resp.Namespace = cached.Namespace.Name
	resp.PageInfo = pi

	err = bytedata.ConvertDataForOutput(results, &resp.Results)
	if err != nil {
		return err
	}

	m := make(map[string]string)

	for idx, result := range results {
		if result.Edges.Instance != nil {
			resp.Results[idx].Instance = result.Edges.Instance.ID.String()
		} else {
			wfID := result.WorkflowID
			path, exists := m[wfID.String()]
			if !exists {
				fStore, _, _, rollback, err := flow.beginSqlTx(ctx)
				if err != nil {
					return err
				}
				defer rollback()

				file, err := fStore.GetFile(ctx, wfID)
				if err != nil {
					return err
				}
				rollback()

				path = file.Path
				m[wfID.String()] = path
			}

			resp.Results[idx].Workflow = path
		}

		resp.Results[idx].Mode = "or"
		if result.Count > 1 {
			resp.Results[idx].Mode = "and"
		}

		edefs := make([]*grpc.EventDef, 0)
		for _, ev := range result.Events {
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

		resp.Results[idx].Events = edefs
	}

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

	cached := new(database.CacheData)

	err = flow.database.NamespaceByName(ctx, cached, namespace)
	if err != nil {
		return nil, err
	}

	timer := in.GetTimer()

	err = flow.events.BroadcastCloudevent(ctx, cached.Namespace, event, timer)
	if err != nil {
		return nil, err
	}

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) HistoricalEvent(ctx context.Context, in *grpc.HistoricalEventRequest) (*grpc.HistoricalEventResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	eid := in.GetId()

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, in.GetNamespace())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(ctx)

	cevent, err := clients.CloudEvents.Query().Where(cevents.HasNamespaceWith(entns.ID(cached.Namespace.ID))).Where(cevents.EventIdEQ(eid)).Only(ctx)
	if err != nil {
		return nil, err
	}

	var resp grpc.HistoricalEventResponse

	resp.Id = eid
	resp.Namespace = cached.Namespace.Name
	resp.ReceivedAt = timestamppb.New(cevent.Created)

	resp.Source = cevent.Event.Source()
	resp.Type = cevent.Event.Type()

	resp.Cloudevent = []byte(cevent.Event.String())

	return &resp, nil
}

var cloudeventsOrderings = []*orderingInfo{
	{
		db:           cevents.FieldCreated,
		req:          "RECEIVED",
		defaultOrder: ent.Desc,
	},
	{
		db:           cevents.FieldEventId,
		req:          "ID",
		defaultOrder: ent.Asc,
	},
}

var cloudeventsFilters = map[*filteringInfo]func(query *ent.CloudEventsQuery, v string) (*ent.CloudEventsQuery, error){}

func (flow *flow) EventHistory(ctx context.Context, req *grpc.EventHistoryRequest) (*grpc.EventHistoryResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(ctx)
	query := clients.CloudEvents.Query().Where(cevents.HasNamespaceWith(entns.ID(cached.Namespace.ID)))

	results, pi, err := paginate[*ent.CloudEventsQuery, *ent.CloudEvents](ctx, req.Pagination, query, cloudeventsOrderings, cloudeventsFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.EventHistoryResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Events = new(grpc.Events)
	resp.Events.PageInfo = pi

	for _, x := range results {
		e := new(grpc.Event)
		resp.Events.Results = append(resp.Events.Results, e)

		e.Id = x.EventId
		e.ReceivedAt = timestamppb.New(x.Created)
		e.Source = x.Event.Source()
		e.Type = x.Event.Type()
		e.Cloudevent = []byte(x.Event.String())
	}

	for _, e := range req.Pagination.Filter {
		f := e.Field
		t := e.Type
		v := e.Val
		events := make([]*grpc.Event, 0)

		if t == "MATCH" && f == "TYPE" {
			for _, ev := range resp.Events.Results {
				if ev.Type == v {
					events = append(events, ev)
				}
			}
			resp.Events.Results = events
			resp.Events.PageInfo.Total = int32(len(events))
		}
	}

	return resp, nil
}

func (flow *flow) EventHistoryStream(req *grpc.EventHistoryRequest, srv grpc.Flow_EventHistoryStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeEvents(cached.Namespace)
	defer flow.cleanup(sub.Close)

resend:

	clients := flow.edb.Clients(ctx)

	query := clients.CloudEvents.Query().Where(cevents.HasNamespaceWith(entns.ID(cached.Namespace.ID)))

	results, pi, err := paginate[*ent.CloudEventsQuery, *ent.CloudEvents](ctx, req.Pagination, query, cloudeventsOrderings, cloudeventsFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.EventHistoryResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Events = new(grpc.Events)
	resp.Events.PageInfo = pi

	for _, x := range results {
		e := new(grpc.Event)
		resp.Events.Results = append(resp.Events.Results, e)

		e.Id = x.EventId
		e.ReceivedAt = timestamppb.New(x.Created)
		e.Source = x.Event.Source()
		e.Type = x.Event.Type()
		e.Cloudevent = []byte(x.Event.String())
	}

	for _, e := range req.Pagination.Filter {
		f := e.Field
		t := e.Type
		v := e.Val
		events := make([]*grpc.Event, 0)

		if t == "MATCH" && f == "TYPE" {
			for _, ev := range resp.Events.Results {
				if ev.Type == v {
					events = append(events, ev)
				}
			}
			resp.Events.Results = events
			resp.Events.PageInfo.Total = int32(len(events))
		}
	}

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

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	eid := req.GetId()

	clients := flow.edb.Clients(ctx)

	cevent, err := clients.CloudEvents.Query().Where(cevents.HasNamespaceWith(entns.ID(cached.Namespace.ID))).Where(cevents.EventIdEQ(eid)).Only(ctx)
	if err != nil {
		return nil, err
	}

	err = flow.events.ReplayCloudevent(ctx, cached, cevent)
	if err != nil {
		return nil, err
	}

	var resp emptypb.Empty

	return &resp, nil
}

func (events *events) ReplayCloudevent(ctx context.Context, cached *database.CacheData, cevent *ent.CloudEvents) error {
	event := cevent.Event

	events.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Replaying event: %s (%s / %s)", event.ID(), event.Type(), event.Source())

	err := events.handleEvent(cached.Namespace, &event)
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

type eventsWaiterSignature struct {
	InstanceID string
	Step       int
}

func (events *events) listenForEvents(ctx context.Context, im *instanceMemory, ceds []*model.ConsumeEventDefinition, all bool) error {
	signature, err := json.Marshal(&eventsWaiterSignature{
		InstanceID: im.cached.Instance.ID.String(),
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
				return fmt.Errorf("failed to execute jq query for key '%s' on event definition %d: %w", k, i, err)
			}
		}

		transformedEvents = append(transformedEvents, ev)
	}

	err = events.addInstanceEventListener(ctx, im.cached, transformedEvents, signature, all)
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

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, namespace)
	if err != nil {
		return newBytesEvent, err
	}

	if jsCode, ok := eventFilterCache.get(key); ok {
		script = fmt.Sprintf("function filter() {\n %s \n}", jsCode)
	} else {
		clients := flow.edb.Clients(ctx)

		ceventfilter, err := clients.CloudEventFilters.Query().Where(enteventsfilter.HasNamespaceWith(entns.ID(cached.Namespace.ID))).Where(enteventsfilter.NameEQ(filterName)).Only(ctx)
		if err != nil {
			err = status.Error(codes.NotFound, fmt.Sprintf("cloudEvent filter %s does not exist", filterName))
			return newBytesEvent, err
		}

		script = fmt.Sprintf("function filter() {\n %s \n}", ceventfilter.Jscode)

		flow.sugar.Debugf("adding filter cache key: %v\n", key)
		eventFilterCache.put(key, ceventfilter.Jscode)
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
		flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), fmt.Sprintf("%v", txt))
	})
	if err != nil {
		return newBytesEvent, fmt.Errorf("failed to initialize js runtime: %w", err)
	}

	_, err = vm.RunString(script)
	if err != nil {
		flow.logger.Errorf(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "CloudEvent filter '%s' produced an error (1): %v", filterName, err)
		return newBytesEvent, err
	}

	f, ok := goja.AssertFunction(vm.Get("filter"))
	if !ok {
		flow.logger.Errorf(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "cloudEvent filter '%s' error: %v", filterName, err)
		return newBytesEvent, err
	}

	newEventMap, err := f(goja.Undefined())
	if err != nil {
		flow.logger.Errorf(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "CloudEvent filter '%s' produced an error (2): %v", filterName, err)
		return newBytesEvent, err
	}

	retValue := newEventMap.Export()

	// event dropped
	if retValue == nil {
		return newBytesEvent, nil
	}

	newBytesEvent, err = json.Marshal(newEventMap)
	if err != nil {
		flow.logger.Errorf(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "CloudEvent filter '%s' produced an error (3): %v", filterName, err)
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

	cached := new(database.CacheData)
	err := flow.database.NamespaceByName(ctx, cached, namespace)
	if err != nil {
		return nil, err
	}

	b, err := flow.execFilter(ctx, namespace, filterName, cloudevent)
	if err != nil {
		flow.logger.Errorf(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace),
			"executing filter failed: %s", err.Error())
		return resp, err
	}

	// dropped event
	if len(b) == 0 {
		flow.logger.Debugf(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace),
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

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, namespace)
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(ctx)

	_, err = clients.CloudEventFilters.Query().Where(enteventsfilter.HasNamespaceWith(entns.ID(cached.Namespace.ID))).Where(enteventsfilter.NameEQ(filterName)).Only(ctx)
	if err != nil {
		err = status.Error(codes.NotFound, fmt.Sprintf("cloudEvent filter %s does not exist", filterName))
		return &resp, err
	}

	_, err = clients.CloudEventFilters.
		Delete().
		Where(
			enteventsfilter.And(
				enteventsfilter.NameEQ(filterName),
			)).
		Exec(ctx)

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

	cached := new(database.CacheData)

	err = flow.database.NamespaceByName(ctx, cached, namespace)
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(ctx)

	k, err := clients.CloudEventFilters.Query().Where(enteventsfilter.HasNamespaceWith(entns.ID(cached.Namespace.ID))).Where(enteventsfilter.NameEQ(filterName)).Count(ctx)
	if err != nil {
		return &resp, err
	}

	if k != 0 {
		err = status.Error(codes.AlreadyExists, fmt.Sprintf("CloudEvent filter %s already exists", filterName))
		return &resp, err
	}

	_, err = clients.CloudEventFilters.Create().SetName(filterName).SetNamespaceID(cached.Namespace.ID).SetJscode(script).Save(ctx)
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

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, namespace)
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(ctx)

	dbs, err := clients.CloudEventFilters.Query().Where(enteventsfilter.HasNamespaceWith(entns.ID(cached.Namespace.ID))).All(ctx)
	if err != nil {
		return resp, err
	}

	for _, s := range dbs {
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

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, namespace)
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(ctx)

	script, err := clients.CloudEventFilters.Query().Where(enteventsfilter.HasNamespaceWith(entns.ID(cached.Namespace.ID))).Where(enteventsfilter.NameEQ(filterName)).Only(ctx)
	if err != nil {
		err = status.Error(codes.NotFound, fmt.Sprintf("cloudEvent filter %s does not exist", filterName))
		return resp, err
	}

	resp.JsCode = script.Jscode

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
