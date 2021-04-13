package direktiv

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/vorteil/direktiv/ent"
	"github.com/vorteil/direktiv/ent/namespace"
	"github.com/vorteil/direktiv/ent/workflow"
	"github.com/vorteil/direktiv/pkg/model"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func (db *dbManager) addWorkflow(ctx context.Context, ns, name, description string, active bool,
	logToEvents string, workflow []byte, startDefinition model.StartDefinition) (*ent.Workflow, error) {

	tx, err := db.dbEnt.Tx(ctx)
	if err != nil {
		return nil, err
	}

	wf, err := tx.Workflow.
		Create().
		SetName(name).
		SetActive(active).
		SetLogToEvents(logToEvents).
		SetWorkflow(workflow).
		SetDescription(description).
		SetNamespaceID(ns).
		Save(ctx)

	if err != nil {
		return nil, rollback(tx, err)
	}

	err = db.processWorkflowEvents(ctx, tx, wf, startDefinition, active)
	if err != nil {
		return nil, rollback(tx, err)
	}

	return wf, tx.Commit()

}

func (db *dbManager) updateWorkflow(ctx context.Context, id string, revision *int, name, description string,
	active *bool, logToEvents *string, workflow []byte, startDefinition model.StartDefinition) (*ent.Workflow, error) {

	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	tx, err := db.dbEnt.Tx(db.ctx)
	if err != nil {
		return nil, err
	}

	var updater *ent.WorkflowUpdateOne

	if revision == nil {

		updater = tx.Workflow.UpdateOneID(uid)

	} else {

		wf, err := tx.Workflow.Get(ctx, uid)
		if err != nil {
			return nil, rollback(tx, err)
		}

		if wf.Revision != *revision {
			return nil, rollback(tx, errors.New("the workflow has already been updated"))
		}

		updater = wf.Update()

	}

	updater = updater.
		SetName(name).
		SetDescription(description).
		SetWorkflow(workflow)

	if active != nil {
		updater = updater.SetActive(*active)
	}

	if logToEvents != nil {
		updater = updater.SetLogToEvents(*logToEvents)
	}

	wf, err := updater.Save(ctx)
	if err != nil {
		return nil, rollback(tx, err)
	}

	err = db.processWorkflowEvents(ctx, tx, wf, startDefinition, wf.Active)
	if err != nil {
		return nil, rollback(tx, err)
	}

	return wf, tx.Commit()

}

func (db *dbManager) deleteWorkflow(ctx context.Context, id string) error {

	u, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	// delete all event listeners and events
	uid, _ := uuid.Parse(id)
	wfe, err := db.getWorkflowEventByWorkflowUID(uid)
	if err == nil {
		db.deleteWorkflowEventListener(wfe.ID)
	}

	// delete all workflow instances
	err = db.deleteWorkflowInstancesByWorkflow(ctx, u)
	if err != nil {
		log.Errorf("can not delete workflow instances: %v", err)
	}

	// delete cron

	i, err := db.dbEnt.Workflow.Delete().
		Where(workflow.IDEQ(u)).
		Exec(ctx)

	if err != nil {
		return err
	}

	if i == 0 {
		return fmt.Errorf("workflow with id %s does not exist", id)
	}

	return nil

}

func (db *dbManager) getWorkflowByID(id uuid.UUID) (*ent.Workflow, error) {

	return db.dbEnt.Workflow.
		Query().
		Where(workflow.IDEQ(id)).
		Only(db.ctx)

}

func (db *dbManager) getWorkflowById(ctx context.Context, ns, id string) (*ent.Workflow, error) {

	return db.dbEnt.Workflow.
		Query().
		Where(workflow.NameEQ(id)).
		Where(workflow.HasNamespaceWith(namespace.IDEQ(ns))).
		Only(ctx)

}

func (db *dbManager) getWorkflowByUid(ctx context.Context, uid string) (*ent.Workflow, error) {

	u, err := uuid.Parse(uid)
	if err != nil {
		return nil, err
	}

	return db.dbEnt.Workflow.
		Query().
		Where(workflow.IDEQ(u)).
		WithNamespace().
		Only(ctx)

}

func (db *dbManager) getWorkflow(id string) (*ent.Workflow, error) {

	u, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	return db.dbEnt.Workflow.
		Query().
		Where(workflow.IDEQ(u)).
		WithNamespace().
		Only(db.ctx)

}

func (db *dbManager) getNamespaceWorkflow(n string, ns string) (*ent.Workflow, error) {

	return db.dbEnt.Workflow.
		Query().
		Where(workflow.HasNamespaceWith(namespace.IDEQ(ns))).
		Where(workflow.NameEQ(n)).
		WithNamespace().
		Only(db.ctx)

}

func (db *dbManager) getWorkflows(ctx context.Context, ns string, offset, limit int) ([]*ent.Workflow, error) {

	if limit == 0 {
		limit = math.MaxInt32
	}

	wfs, err := db.dbEnt.Workflow.
		Query().
		Limit(limit).
		Offset(offset).
		Select(workflow.FieldID, workflow.FieldName, workflow.FieldCreated, workflow.FieldDescription, workflow.FieldActive, workflow.FieldRevision, workflow.FieldLogToEvents).
		Where(workflow.HasNamespaceWith(namespace.IDEQ(ns))).
		Order(ent.Asc(namespace.FieldID)).
		All(ctx)

	if err != nil {
		return nil, err
	}

	return wfs, nil

}

func (db *dbManager) getWorkflowCount(ctx context.Context, ns string, offset, limit int) (int, error) {

	wfCount, err := db.dbEnt.Workflow.
		Query().
		Select(workflow.FieldID, workflow.FieldName, workflow.FieldCreated, workflow.FieldDescription, workflow.FieldActive, workflow.FieldRevision).
		Where(workflow.HasNamespaceWith(namespace.IDEQ(ns))).Count(ctx)

	return wfCount, err
}
