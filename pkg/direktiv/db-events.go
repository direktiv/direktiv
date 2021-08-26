package direktiv

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/vorteil/direktiv/ent"
	"github.com/vorteil/direktiv/ent/workflow"
	"github.com/vorteil/direktiv/ent/workflowevents"
	"github.com/vorteil/direktiv/ent/workfloweventswait"
	"github.com/vorteil/direktiv/ent/workflowinstance"
	"github.com/vorteil/direktiv/pkg/model"
)

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
