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
	hash "github.com/mitchellh/hashstructure/v2"
	glob "github.com/ryanuber/go-glob"
	"github.com/vorteil/direktiv/pkg/flow/ent"
	entcev "github.com/vorteil/direktiv/pkg/flow/ent/cloudevents"
	entev "github.com/vorteil/direktiv/pkg/flow/ent/events"
	entevw "github.com/vorteil/direktiv/pkg/flow/ent/eventswait"
	entinst "github.com/vorteil/direktiv/pkg/flow/ent/instance"
	"github.com/vorteil/direktiv/pkg/flow/ent/workflow"
	"github.com/vorteil/direktiv/pkg/model"
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

func (events *events) markEventAsProcessed(eventID, namespace string) (*cloudevents.Event, error) {

	ctx := context.Background()

	id, err := uuid.Parse(eventID)
	if err != nil {
		return nil, err
	}

	tx, err := events.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	e, err := events.db.CloudEvents.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if e.Processed {
		return nil, fmt.Errorf("event already processed")
	}

	updater := events.db.CloudEvents.UpdateOne(e)
	updater.SetProcessed(true)

	e, err = updater.Save(ctx)
	if err != nil {
		return nil, err
	}

	ev := cloudevents.Event(e.Event)

	return &ev, tx.Commit()

}

func (events *events) deleteExpiredEvents() error {

	ctx := context.Background()

	_, err := events.db.CloudEvents.Delete().
		Where(
			entcev.And(
				entcev.Processed(true),
				entcev.FireLT(time.Now().Add(-1*time.Hour)),
			),
		).
		Exec(ctx)

	return err

}

func (events *events) getEarliestEvent() (*ent.CloudEvents, error) {

	ctx := context.Background()

	e, err := events.db.CloudEvents.
		Query().
		Where(
			entcev.And(
				entcev.Processed(false),
			),
		).
		Order(ent.Asc(entcev.FieldFire)).
		First(ctx)

	return e, err

}

func (events *events) addEvent(eventin *cloudevents.Event, ns string, delay int64) error {

	ctx := context.Background()

	// calculate fire time
	t := time.Now().Unix() + delay

	// processed
	processed := (delay == 0)

	ev := event.Event(*eventin)

	_, err := events.db.CloudEvents.
		Create().
		SetEvent(ev).
		SetNamespace(ns).
		SetFire(time.Unix(t, 0)).
		SetProcessed(processed).
		SetEventId(eventin.ID()).
		Save(ctx)

	return err

}

func (events *events) deleteWorkflowEventWait(id uuid.UUID) error {

	ctx := context.Background()

	_, err := events.db.EventsWait.
		Delete().
		Where(entevw.IDEQ(id)).
		Exec(ctx)

	return err

}

func (events *events) deleteWorkflowEventListener(id uuid.UUID) error {

	ctx := context.Background()

	err := events.deleteWorkflowEventWaitByListenerID(id)
	if err != nil {
		events.sugar.Errorf("can not delete event listeners wait for event listener: %v", err)
	}

	_, err = events.db.Events.
		Delete().
		Where(entev.IDEQ(id)).
		Exec(ctx)

	return err
}

func (events *events) deleteWorkflowEventWaitByListenerID(id uuid.UUID) error {

	ctx := context.Background()

	_, err := events.db.EventsWait.
		Delete().
		Where(entevw.HasWorkfloweventWith(entev.IDEQ(id))).
		Exec(ctx)

	return err

}

func (events *events) deleteWorkflowEventListenerByInstanceID(id uuid.UUID) error {

	var err error

	ctx := context.Background()

	tx, err := events.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	var el *ent.Events
	el, err = events.getWorkflowEventByInstanceID(id)
	if err != nil {
		return err
	}

	err = events.deleteWorkflowEventListener(el.ID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil

}

func (events *events) addWorkflowEventWait(ev map[string]interface{}, count int, id uuid.UUID) (*ent.EventsWait, error) {

	ctx := context.Background()

	ww, err := events.db.EventsWait.
		Create().
		SetEvents(ev).
		SetWorkfloweventID(id).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return ww, nil

}

// called by add workflow, adds event listeners if required
func (events *events) processWorkflowEvents(ctx context.Context, tx *ent.Tx,
	wf *ent.Workflow, startDefinition model.StartDefinition, active bool) error {

	var sevents []model.StartEventDefinition
	if startDefinition != nil {
		sevents = startDefinition.GetEvents()
	}

	if len(sevents) > 0 && active {

		// delete everything event related
		wfe, err := events.getWorkflowEventByWorkflowUID(wf.ID)
		if err == nil {
			events.deleteWorkflowEventListener(wfe.ID)
		}

		var ev []map[string]interface{}
		for _, e := range sevents {
			em := make(map[string]interface{})
			em[eventTypeString] = e.Type

			for kf, vf := range e.Filters {
				em[fmt.Sprintf("%s%s", filterPrefix, strings.ToLower(kf))] = vf
			}
			ev = append(ev, em)
		}

		correlations := []string{}
		count := 1

		switch d := startDefinition.(type) {
		case *model.EventsAndStart:
			{
				correlations = append(correlations, d.Correlate...)
				count = len(sevents)
			}
		}

		_, err = tx.Events.
			Create().
			SetWorkflow(wf).
			SetEvents(ev).
			SetCorrelations(correlations).
			SetCount(count).
			Save(ctx)

		if err != nil {
			return err
		}

	}

	return nil

}

// called from workflow instances to create event listeners
func (events *events) addWorkflowEventListener(wfid uuid.UUID, wfinstance uuid.UUID,
	sevents []*model.ConsumeEventDefinition,
	signature []byte, all bool) (*ent.Events, error) {

	ctx := context.Background()

	var ev []map[string]interface{}
	for _, e := range sevents {
		em := make(map[string]interface{})
		em[eventTypeString] = e.Type

		for kf, vf := range e.Context {
			em[fmt.Sprintf("%s%s", filterPrefix, strings.ToLower(kf))] = vf
		}
		ev = append(ev, em)
	}

	count := 1
	if all {
		count = len(sevents)
	}

	return events.db.Events.
		Create().
		SetWorkflowID(wfid).
		SetEvents(ev).
		SetCorrelations([]string{}).
		SetSignature(signature).
		SetWorkflowinstanceID(wfinstance).
		SetCount(count).
		Save(ctx)

}

func (events *events) getWorkflowEventByID(id uuid.UUID) (*ent.Events, error) {

	ctx := context.Background()

	return events.db.Events.
		Query().
		Where(entev.IDEQ(id)).
		WithWorkflow().
		Only(ctx)

}

func (events *events) getWorkflowEventByWorkflowUID(id uuid.UUID) (*ent.Events, error) {

	ctx := context.Background()

	return events.db.Events.
		Query().
		Where(entev.HasWorkflowWith(
			workflow.IDEQ(id),
		)).
		WithWorkflow().
		Only(ctx)

}

func (events *events) getWorkflowEventByInstanceID(id uuid.UUID) (*ent.Events, error) {

	ctx := context.Background()

	return events.db.Events.
		Query().
		Where(entev.HasWorkflowinstanceWith(
			entinst.IDEQ(id),
		)).
		Only(ctx)

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

	err := events.flushEvent(n[0], n[1], true)
	if err != nil {
		events.sugar.Errorf("can not flush delayed event: %v", err)
		return
	}

}

var syncMtx sync.Mutex

func (events *events) syncEventDelays() {

	syncMtx.Lock()
	defer syncMtx.Unlock()

	// sync with other servers
	events.sugar.Debugf("update event timeout")

	// disable old timer
	for i := range events.timers.timers {
		ti := events.timers.timers[i]
		if ti.name == "sendEventTimer" {
			events.timers.disableTimer(ti)
			break
		}
	}

	for {
		e, err := events.getEarliestEvent()
		if err != nil {
			events.sugar.Errorf("can not sync event delays: %v", err)
			return
		}

		if e.Fire.Before(time.Now()) {
			events.sugar.Debugf("flushing old event %s", e.ID)
			err = events.flushEvent(e.EventId, e.Namespace, false)
			if err != nil {
				events.sugar.Errorf("can not flush event %s: %v", e.ID, err)
			}
			continue
		}

		err = events.timers.addOneShot("sendEventTimer", sendEventFunction,
			e.Fire, []byte(fmt.Sprintf("%s/%s", e.ID, e.Namespace)))
		if err != nil {
			events.sugar.Errorf("can not reschedule event timer: %v", err)
		}

		break

	}

}

func (events *events) flushEvent(eventID, namespace string, rearm bool) error {

	events.sugar.Debugf("flushing cloud event %s (%s)", eventID, namespace)

	e, err := events.markEventAsProcessed(eventID, namespace)
	if err != nil {
		return err
	}

	defer func(r bool) {
		if r {
			events.syncEventDelays()
		}
	}(rearm)

	return events.handleEvent(namespace, e)

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

	rows, err := db.Query(`update workflow_events_waits
	set events = jsonb_set(events, $1, $2, true)
	WHERE events::jsonb ? $3 and workflow_events_wfeventswait = $4
	returning *`, fmt.Sprintf("{%s}", chash), fmt.Sprintf(`"%s"`, base64.StdEncoding.EncodeToString(data)), chash, id)
	if err != nil {
		return retEvents, err
	}

	rc := 0

	for rows.Next() {
		rc++

		var (
			id         uuid.UUID
			eventID    string
			eventsData []byte
		)

		err := rows.Scan(&id, &events, &eventID)
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

			err = events.deleteWorkflowEventWait(id)
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

func (events *events) handleEvent(namespace string, ce *cloudevents.Event) error {

	events.sugar.Debugf("handle event %s", ce.Type())

	var (
		id                                          uuid.UUID
		count                                       int
		singleEvent, corBytes, allEvents, signature []byte
		wf                                          string
	)

	db := events.db.DB()

	rows, err := db.Query(`select
	we.id, signature, count, correlations, events, workflow_wfevents, v
	from workflow_events we
	inner join workflows w
		on w.id = workflow_wfevents
	inner join namespaces n
		on n.id = w.namespace_workflows,
	jsonb_array_elements(events) as v
	where v::json->>'type' = $1
	and n.id = $2`, ce.Type(), namespace)
	if err != nil {
		return err
	}
	defer rows.Close()

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

		conn, err = events.locks.lockDB(hash, defaultLockWait)

		if err != nil {
			events.sugar.Errorf("can not lock event row: %d, %v", id, err)
			continue
		}

		events.sugar.Debugf("event listener %d is candidate", id)

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

			events.sugar.Debugf("single event workflow")
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
					err := events.addEventListenerWait(ce, id, correlations, eventTypes)
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
			uid, _ := uuid.Parse(wf)
			events.sugar.Debugf("run workflow %v with %d events", uid, len(retEvents))
			if len(signature) == 0 {
				// TODO
				// go events.engine.EventsInvoke(uid, retEvents...)
			} else {
				events.sugar.Debugf("calling with signature %v", string(signature))
				events.deleteWorkflowEventListener(id)
				// TODO
				// go events.wakeEventsWaiter(signature, retEvents)
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

func (events *events) addEventListenerWait(cevent *cloudevents.Event, id uuid.UUID,
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

	_, err := events.addWorkflowEventWait(sevents, 1, id)
	return err

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
