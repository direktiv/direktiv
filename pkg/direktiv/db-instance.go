package direktiv

import (
	"context"
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

func (db *dbManager) addWorkflowInstance(ns, workflowID, instanceID, input string) (*ent.WorkflowInstance, error) {

	count, err := db.dbEnt.WorkflowInstance.
		Query().
		Where(workflowinstance.HasWorkflowWith(workflow.HasNamespaceWith(namespace.IDEQ(ns)))).
		Where(workflowinstance.BeginTimeGT(time.Now().Add(-maxInstancesLimitInterval))).
		Count(db.ctx)
	if err != nil {
		return nil, err
	}

	// only limit if running in prod mode
	if log.GetLevel() != log.DebugLevel && count > maxInstancesPerInterval {
		return nil, NewCatchableError("direktiv.limits.instances", "new workflow instance rejected because it would exceed the maximum number of new workflow instances (%d) per time interval (%s) for the namespace", maxInstancesPerInterval, maxInstancesLimitInterval)
	}

	wf, err := db.getNamespaceWorkflow(workflowID, ns)
	if err != nil {
		return nil, err
	}

	wi, err := db.dbEnt.WorkflowInstance.
		Create().
		SetInstanceID(instanceID).
		SetInvokedBy("").
		SetRevision(wf.Revision).
		SetStatus("pending").
		SetBeginTime(time.Now()).
		SetInput(input).
		SetWorkflow(wf).
		Save(db.ctx)

	if err != nil {
		return nil, err
	}

	return wi, nil

}

func (db *dbManager) getWorkflowInstanceByID(ctx context.Context, id int) (*ent.WorkflowInstance, error) {

	return db.dbEnt.WorkflowInstance.
		Query().
		Where(workflowinstance.IDEQ(id)).
		Only(ctx)

}

func (db *dbManager) getWorkflowInstance(ctx context.Context, id string) (*ent.WorkflowInstance, error) {

	return db.dbEnt.WorkflowInstance.
		Query().
		Where(workflowinstance.InstanceIDEQ(id)).
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
