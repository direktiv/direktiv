package direktiv

import (
	"context"
	"math"
	"time"

	"github.com/vorteil/direktiv/ent"
	"github.com/vorteil/direktiv/ent/namespace"
	"github.com/vorteil/direktiv/ent/workflow"
	"github.com/vorteil/direktiv/ent/workflowinstance"

	log "github.com/sirupsen/logrus"
)

func (db *dbManager) deleteWorkflowInstance(id int) error {



	err := db.dbEnt.WorkflowInstance.DeleteOneID(id).Exec(db.ctx)
	if err != nil {
		return err
	}

	// delete all events attached to this instance
	

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
