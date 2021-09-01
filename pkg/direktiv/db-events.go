package direktiv

import (
	"context"
	"fmt"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/uuid"
	"github.com/vorteil/direktiv/ent"
	ce "github.com/vorteil/direktiv/ent/cloudevents"
	"github.com/vorteil/direktiv/ent/workflow"
	"github.com/vorteil/direktiv/ent/workflowevents"
	"github.com/vorteil/direktiv/ent/workfloweventswait"
	"github.com/vorteil/direktiv/ent/workflowinstance"
	"github.com/vorteil/direktiv/pkg/model"
)

func (db *dbManager) markEventAsProcessed(eventID, namespace string) (*cloudevents.Event, error) {

	tx, err := db.dbEnt.Tx(db.ctx)
	if err != nil {
		return nil, err
	}

	e, err := db.dbEnt.CloudEvents.Get(db.ctx, eventID)
	if err != nil {
		return nil, rollback(tx, err)
	}

	if e.Processed {
		return nil, rollback(tx, fmt.Errorf("event already processed"))
	}

	updater := db.dbEnt.CloudEvents.UpdateOne(e)
	updater.SetProcessed(true)

	e, err = updater.Save(db.ctx)
	if err != nil {
		return nil, rollback(tx, err)
	}

	ev := cloudevents.Event(e.Event)

	return &ev, tx.Commit()

}

func (db *dbManager) deleteExpiredEvents() error {

	_, err := db.dbEnt.CloudEvents.Delete().
		Where(
			ce.And(
				ce.Processed(true),
				ce.FireLT(time.Now().Add(-1*time.Hour)),
			),
		).
		Exec(db.ctx)

	return err

}

func (db *dbManager) getEarliestEvent() (*ent.CloudEvents, error) {

	e, err := db.dbEnt.CloudEvents.
		Query().
		Where(
			ce.And(
				ce.Processed(false),
			),
		).
		Order(ent.Asc(ce.FieldFire)).
		First(context.Background())

	return e, err

}

func (db *dbManager) addEvent(eventin *cloudevents.Event, ns string, delay int64) error {

	// calculate fire time
	t := time.Now().Unix() + delay

	// processed
	processed := (delay == 0)

	ev := event.Event(*eventin)

	_, err := db.dbEnt.CloudEvents.
		Create().
		SetEvent(ev).
		SetNamespace(ns).
		SetFire(time.Unix(t, 0)).
		SetProcessed(processed).
		SetID(eventin.ID()).
		Save(db.ctx)

	return err

}

func (db *dbManager) deleteWorkflowEventWait(id int) error {

	_, err := db.dbEnt.WorkflowEventsWait.
		Delete().
		Where(workfloweventswait.IDEQ(id)).
		Exec(db.ctx)

	return err

}

func (db *dbManager) deleteWorkflowEventListener(id int) error {

	err := db.deleteWorkflowEventWaitByListenerID(id)
	if err != nil {
		appLog.Errorf("can not delete event listeners wait for event listener: %v", err)
	}

	_, err = db.dbEnt.WorkflowEvents.
		Delete().
		Where(workflowevents.IDEQ(id)).
		Exec(db.ctx)

	return err
}

func (db *dbManager) deleteWorkflowEventWaitByListenerID(id int) error {

	_, err := db.dbEnt.WorkflowEventsWait.
		Delete().
		Where(workfloweventswait.HasWorkfloweventWith(workflowevents.IDEQ(id))).
		Exec(db.ctx)

	return err

}

func (db *dbManager) deleteWorkflowEventListenerByInstanceID(id int) error {

	var err error

	tx, err := db.dbEnt.Tx(db.ctx)
	if err != nil {
		return err
	}
	defer rollback(tx, err)

	var el *ent.WorkflowEvents
	el, err = db.getWorkflowEventByInstanceID(id)
	if err != nil {
		return err
	}

	err = db.deleteWorkflowEventListener(el.ID)
	if err != nil {
		return err
	}

	return tx.Commit()

}

func (db *dbManager) addWorkflowEventWait(ev map[string]interface{}, count, id int) (*ent.WorkflowEventsWait, error) {

	ww, err := db.dbEnt.WorkflowEventsWait.
		Create().
		SetEvents(ev).
		SetWorkfloweventID(id).
		Save(db.ctx)

	if err != nil {
		return nil, err
	}

	return ww, nil

}

// called by add workflow, adds event listeners if required
func (db *dbManager) processWorkflowEvents(ctx context.Context, tx *ent.Tx,
	wf *ent.Workflow, startDefinition model.StartDefinition, active bool) error {

	var events []model.StartEventDefinition
	if startDefinition != nil {
		events = startDefinition.GetEvents()
	}

	if len(events) > 0 && active {

		// delete everything event related
		wfe, err := db.getWorkflowEventByWorkflowUID(wf.ID)
		if err == nil {
			db.deleteWorkflowEventListener(wfe.ID)
		}

		var ev []map[string]interface{}
		for _, e := range events {
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
				count = len(events)
			}
		}

		_, err = tx.WorkflowEvents.
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
func (db *dbManager) addWorkflowEventListener(wfid uuid.UUID, wfinstance int,
	events []*model.ConsumeEventDefinition,
	signature []byte, all bool) (*ent.WorkflowEvents, error) {

	var ev []map[string]interface{}
	for _, e := range events {
		em := make(map[string]interface{})
		em[eventTypeString] = e.Type

		for kf, vf := range e.Context {
			em[fmt.Sprintf("%s%s", filterPrefix, strings.ToLower(kf))] = vf
		}
		ev = append(ev, em)
	}

	count := 1
	if all {
		count = len(events)
	}

	return db.dbEnt.WorkflowEvents.
		Create().
		SetWorkflowID(wfid).
		SetEvents(ev).
		SetCorrelations([]string{}).
		SetSignature(signature).
		SetWorkflowinstanceID(wfinstance).
		SetCount(count).
		Save(db.ctx)

}

func (db *dbManager) getWorkflowEventByID(id int) (*ent.WorkflowEvents, error) {

	return db.dbEnt.WorkflowEvents.
		Query().
		Where(workflowevents.IDEQ(id)).
		WithWorkflow().
		Only(db.ctx)

}

func (db *dbManager) getWorkflowEventByWorkflowUID(id uuid.UUID) (*ent.WorkflowEvents, error) {

	return db.dbEnt.WorkflowEvents.
		Query().
		Where(workflowevents.HasWorkflowWith(
			workflow.IDEQ(id),
		)).
		WithWorkflow().
		Only(db.ctx)

}

func (db *dbManager) getWorkflowEventByInstanceID(id int) (*ent.WorkflowEvents, error) {

	return db.dbEnt.WorkflowEvents.
		Query().
		Where(workflowevents.HasWorkflowinstanceWith(
			workflowinstance.IDEQ(id),
		)).
		Only(db.ctx)

}
