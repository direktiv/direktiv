package direktiv

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/vorteil/direktiv/ent"
	"github.com/vorteil/direktiv/ent/namespace"
	"github.com/vorteil/direktiv/ent/workflow"
	"github.com/vorteil/direktiv/ent/workflowinstance"

	log "github.com/sirupsen/logrus"
)

func (db *dbManager) deleteWorkflowInstance(id int) error {

	wfi, err := db.getWorkflowInstanceByID(db.ctx, id)
	if err != nil {
		return err
	}

	wf := wfi.Edges.Workflow
	ns := wf.Edges.Namespace

	if db.tm != nil {
		err := db.tm.deleteTimersForInstance(wfi.InstanceID)
		if err != nil {
			log.Errorf("can not delete timers for instance %s", wfi.InstanceID)
		}
	}

	// delete all events attached to this instance
	err = db.deleteWorkflowEventListenerByInstanceID(id)
	if err != nil && !ent.IsNotFound(err) {
		log.Errorf("can not delete event listeners for instance: %v", err)
	}

	err = db.dbEnt.WorkflowInstance.DeleteOneID(id).Exec(db.ctx)
	if err != nil {
		return err
	}

	err = (*db.varStorage).DeleteAllInScope(db.ctx, ns.ID, wf.ID.String(), wfi.InstanceID)
	if err != nil {
		return err
	}

	return nil
}

func (db *dbManager) deleteWorkflowInstancesByWorkflow(ctx context.Context, wf uuid.UUID) error {

	instances, err := db.getWorkflowInstancesByWFID(ctx, wf, 0, 0)
	if err != nil {
		return err
	}

	for _, i := range instances {
		err := db.deleteWorkflowInstance(i.ID)
		if err != nil {
			log.Errorf("can not delete workflow instance %s", i.InstanceID)
		}
	}

	return nil
}

func (db *dbManager) addWorkflowInstance(ctx context.Context, ns, workflowID, instanceID, input string, cronCheck bool) (*ent.WorkflowInstance, error) {

	tx, err := db.dbEnt.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if cronCheck {

		t := time.Now().Add(time.Second * 30 * -1)

		wf, err := tx.WorkflowInstance.
			Query().
			Limit(1).
			Where(workflowinstance.BeginTimeGT(t)).
			Order(ent.Desc(workflowinstance.FieldBeginTime)).All(ctx)
		if err != nil {
			return nil, err
		}

		if len(wf) > 0 {
			return nil, errors.New("cron already invoked")
		}

	}

	wf, err := tx.Workflow.
		Query().
		Where(workflow.HasNamespaceWith(namespace.IDEQ(ns))).
		Where(workflow.NameEQ(workflowID)).
		WithNamespace().
		Only(ctx)
	if err != nil {
		return nil, err
	}

	wi, err := tx.WorkflowInstance.
		Create().
		SetInstanceID(instanceID).
		SetInvokedBy("").
		SetRevision(wf.Revision).
		SetStatus("pending").
		SetBeginTime(time.Now()).
		SetInput(input).
		SetWorkflow(wf).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	wi, err = db.dbEnt.WorkflowInstance.Get(ctx, wi.ID)
	if err != nil {
		return nil, err
	}

	return wi, nil

}

func (db *dbManager) getWorkflowInstanceByID(ctx context.Context, id int) (*ent.WorkflowInstance, error) {

	return db.dbEnt.WorkflowInstance.
		Query().
		Where(workflowinstance.IDEQ(id)).
		WithWorkflow(func(q *ent.WorkflowQuery) {
			q.WithNamespace()
		}).
		Only(ctx)

}

func (db *dbManager) getWorkflowInstanceExpired(ctx context.Context) ([]*ent.WorkflowInstance, error) {

	t := time.Now().Add(-1 * time.Minute)

	return db.dbEnt.WorkflowInstance.
		Query().
		Select(workflowinstance.FieldInstanceID, workflowinstance.FieldStatus,
			workflowinstance.FieldDeadline, workflowinstance.FieldFlow).
		Where(
			workflowinstance.And(
				workflowinstance.DeadlineLT(t),
				workflowinstance.StatusEQ("pending"),
			),
		).
		All(ctx)

}

func (db *dbManager) getWorkflowInstance(ctx context.Context, id string) (*ent.WorkflowInstance, error) {

	return db.dbEnt.WorkflowInstance.
		Query().
		Where(workflowinstance.InstanceIDEQ(id)).
		WithWorkflow(func(q *ent.WorkflowQuery) {
			q.WithNamespace()
		}).
		Only(ctx)

}

func (db *dbManager) getWorkflowInstances(ctx context.Context, ns string, offset, limit int) ([]*ent.WorkflowInstance, error) {

	if limit == 0 {
		limit = math.MaxInt32
	}

	wfs, err := db.dbEnt.WorkflowInstance.
		Query().
		Limit(limit).
		Offset(offset).
		Select(workflowinstance.FieldInstanceID, workflowinstance.FieldStatus, workflowinstance.FieldBeginTime).
		Where(workflowinstance.HasWorkflowWith(workflow.HasNamespaceWith(namespace.IDEQ(ns)))).
		Order(ent.Desc(workflowinstance.FieldBeginTime)).
		All(ctx)

	if err != nil {
		return nil, err
	}

	return wfs, nil

}

func (db *dbManager) getWorkflowInstancesByWFID(ctx context.Context, wf uuid.UUID, offset, limit int) ([]*ent.WorkflowInstance, error) {

	wfs, err := db.dbEnt.WorkflowInstance.
		Query().
		Select(workflowinstance.FieldInstanceID, workflowinstance.FieldStatus, workflowinstance.FieldBeginTime).
		Where(workflowinstance.HasWorkflowWith(workflow.IDEQ(wf))).
		Limit(limit).
		Offset(offset).
		Order(ent.Desc(workflowinstance.FieldBeginTime)).
		All(ctx)

	if err != nil {
		return nil, err
	}

	return wfs, nil

}
