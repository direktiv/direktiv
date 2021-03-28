package direktiv

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/uuid"
	"github.com/mitchellh/hashstructure/v2"
	hash "github.com/mitchellh/hashstructure/v2"
	glob "github.com/ryanuber/go-glob"
	log "github.com/sirupsen/logrus"
)

const (
	eventTypeString = "type"
)

func init() {
	gob.Register(new(event.EventContextV1))
	gob.Register(new(event.EventContextV03))
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
					log.Debugf("%s does not match %s", vs, fs)
					return false
				}

				if v == f {
					return true
				}

			} else {
				log.Debugf("event does not contain %v", kt)
				return false
			}

		}
	}

	return true
}

func (s *WorkflowServer) updateMultipleEvents(ce *cloudevents.Event, id int,
	correlations []string) ([]*cloudevents.Event, error) {

	var retEvents []*cloudevents.Event
	db := s.dbManager.dbEnt.DB()

	chash := generateCorrelationHash(ce, ce.Type(), correlations)

	// sql := fmt.Sprintf(`update workflow_events_waits
	// 	set events = jsonb_set(events, '{%s}', '"%s"', true)
	// 	WHERE events::jsonb ? '%s' and workflow_events_wfeventswait = %d
	// 	returning *`, chash,
	// 	base64.StdEncoding.EncodeToString(eventToBytes(*ce)), chash, id)

	rows, err := db.Query(`update workflow_events_waits
	set events = jsonb_set(events, $1, $2, true)
	WHERE events::jsonb ? $3 and workflow_events_wfeventswait = $4
	returning *`, fmt.Sprintf("{%s}", chash), fmt.Sprintf(`"%s"`, base64.StdEncoding.EncodeToString(eventToBytes(*ce))), chash, id)
	if err != nil {
		return retEvents, err
	}

	rc := 0

	for rows.Next() {
		rc++

		var (
			id, eventID int
			events      []byte
		)

		err := rows.Scan(&id, &events, &eventID)
		if err != nil {
			log.Errorf("can not scan result: %v", err)
			continue
		}

		var eventsIn map[string]interface{}
		err = json.Unmarshal(events, &eventsIn)
		if err != nil {
			log.Errorf("can not unmarshal existing events")
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

			s.dbManager.deleteWorkflowEventWait(id)

			// get data for all events
			var e map[string]string
			json.Unmarshal(events, &e)

			for _, g := range e {
				b, err := base64.StdEncoding.DecodeString(g)
				if err != nil {
					log.Errorf("event data corrupt: %v", err)
					continue
				}
				ev := bytesToEvent(b)
				retEvents = append(retEvents, ev)
			}

		}

	}

	return retEvents, nil

}

func (s *WorkflowServer) handleEvent(namespace string, ce *cloudevents.Event) error {

	log.Debugf("handle event %s", ce.Type())

	var (
		id, count                                   int
		singleEvent, corBytes, allEvents, signature []byte
		wf                                          string
	)

	db := s.dbManager.dbEnt.DB()

	// get candidates for event
	// query := fmt.Sprintf(`select
	// 	id, signature, count, correlations, events, workflow_wfevents,
	// 	v from workflow_events etl,
	// 	jsonb_array_elements(etl.events) as v
	// 	where v::json->>'type' = '%s'`,
	// 	ce.Type())

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
			log.Errorf("process row error: %v", err)
			continue
		}

		hash, _ := hashstructure.Hash(fmt.Sprintf("%d%v%v", id, allEvents, wf), hashstructure.FormatV2, nil)

		unlock := func() {
			if conn != nil {
				s.dbManager.unlockDB(hash, conn)
			}
		}

		conn, err = s.dbManager.lockDB(hash, 5)

		if err != nil {
			log.Errorf("can not lock event row: %d, %v", id, err)
			continue
		}

		log.Debugf("event listener %d is candidate", id)

		var eventMap map[string]interface{}
		err = json.Unmarshal(singleEvent, &eventMap)
		if err != nil {
			unlock()
			log.Errorf("can not marshall event map: %v", err)
			continue
		}

		// check filters
		if !matchesExtensions(eventMap, ce.Context.GetExtensions()) {
			log.Debugf("event listener %d does not match", id)
			unlock()
			continue
		}

		var ae []map[string]interface{}
		json.Unmarshal(allEvents, &ae)

		var retEvents []*cloudevents.Event

		if count == 1 {

			log.Debugf("single event workflow")
			retEvents = append(retEvents, ce)

		} else {

			var correlations []string

			// get correlations
			if len(corBytes) > 0 {
				json.Unmarshal(corBytes, &correlations)
			}

			es, err := s.updateMultipleEvents(ce, id, correlations)
			if err != nil {
				log.Errorf("can not handle multi event: %v", err)
				unlock()
				continue
			}

			retEvents = append(retEvents, es...)

			// no update executed, means we have a candidate but no existing events
			// for a multi event workflow
			if len(retEvents) == 0 {

				log.Debugf("no events waiting")

				// get event types
				var eventTypes []string
				for _, g := range ae {
					eventTypes = append(eventTypes, g[eventTypeString].(string))
				}

				// only add if the correlation id exists
				if generateCorrelationHash(ce, ce.Type(), correlations) != "" {
					err := s.addEventListenerWait(ce, id, correlations, eventTypes)
					if err != nil {
						log.Errorf("can not create workflow event wait: %v", err)
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
			log.Debugf("run workflow %v with %d events", uid, len(retEvents))
			if len(signature) == 0 {
				go s.engine.EventsInvoke(uid, retEvents...)
			} else {
				log.Debugf("calling with signature %v", string(signature))
				s.dbManager.deleteWorkflowEventListener(id)
				go s.engine.wakeEventsWaiter(signature, retEvents)
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
			log.Debugf("event does not contain correlation id: %s", k)
			return ""
		}
	}

	hashBase[eventTypeString] = ets
	h, _ := hash.Hash(hashBase, hash.FormatV2, nil)

	return fmt.Sprintf("%d", h)

}

func (s *WorkflowServer) addEventListenerWait(cevent *cloudevents.Event, id int,
	correlations, eventTypes []string) error {

	events := make(map[string]interface{})

	for _, v := range eventTypes {
		if v == cevent.Type() {
			events[generateCorrelationHash(cevent, v, correlations)] = base64.StdEncoding.EncodeToString(eventToBytes(*cevent))
		} else {
			events[generateCorrelationHash(cevent, v, correlations)] = nil
		}
	}

	_, err := s.dbManager.addWorkflowEventWait(events, 1, id)
	return err

}

func eventToBytes(cevent cloudevents.Event) []byte {

	var ev bytes.Buffer
	enc := gob.NewEncoder(&ev)
	err := enc.Encode(cevent)
	if err != nil {
		log.Errorf("can not convert event to bytes: %v", err)
	}
	return ev.Bytes()
}

func bytesToEvent(b []byte) *cloudevents.Event {

	gob.Register(new(event.EventContextV1))

	ev := new(cloudevents.Event)
	enc := gob.NewDecoder(bytes.NewReader(b))
	err := enc.Decode(ev)
	if err != nil {
		log.Errorf("can not convert bytes to event: %v", err)
	}
	return ev
}
