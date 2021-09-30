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
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	hash "github.com/mitchellh/hashstructure/v2"
	glob "github.com/ryanuber/go-glob"
	"github.com/vorteil/direktiv/pkg/flow/ent"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
	"github.com/vorteil/direktiv/pkg/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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
	for i := range events.timers.timers {
		ti := events.timers.timers[i]
		if ti.name == "sendEventTimer" {
			events.timers.disableTimer(ti)
			break
		}
	}

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

				err = events.deleteWorkflowEventListeners(ctx, d.wf)
				if err != nil {
					events.engine.sugar.Error(err)
					return nil
				}

				go events.engine.wakeEventsWaiter(signature, retEvents)

			}

		}

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

func (flow *flow) BroadcastCloudevent(ctx context.Context, in *grpc.BroadcastCloudeventRequest) (*emptypb.Empty, error) {

	namespace := in.GetNamespace()
	rawevent := in.GetCloudevent()

	event := new(cloudevents.Event)
	err := event.UnmarshalJSON(rawevent)
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

func (events *events) BroadcastCloudevent(ctx context.Context, ns *ent.Namespace, event *cloudevents.Event, timer int64) error {

	events.logToNamespace(ctx, time.Now(), ns, "Event received: %s (%s / %s)", event.ID(), event.Type(), event.Source())

	// add event to db
	err := events.addEvent(ctx, events.db.CloudEvents, event, ns, timer)
	if err != nil {
		return err
	}

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
