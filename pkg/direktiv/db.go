package direktiv

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/ent"
	"github.com/vorteil/direktiv/ent/hook"

	// "github.com/vorteil/direktiv/ent/messageid"
	"github.com/vorteil/direktiv/ent/namespace"
	"github.com/vorteil/direktiv/ent/timer"
	"github.com/vorteil/direktiv/ent/workflow"
	"github.com/vorteil/direktiv/ent/workflowevents"
	"github.com/vorteil/direktiv/ent/workfloweventswait"
	"github.com/vorteil/direktiv/ent/workflowinstance"
	"github.com/vorteil/direktiv/pkg/model"
)

const (
	filterPrefix = "filter-"
)

// DBManager contains all database related information and functions
type dbManager struct {
	dbEnt *ent.Client
	ctx   context.Context
}

func newDBManager(ctx context.Context, conn string) (*dbManager, error) {

	var err error
	db := &dbManager{
		ctx: ctx,
	}

	log.Debugf("connecting db")

	db.dbEnt, err = ent.Open("postgres", conn)
	if err != nil {
		log.Errorf("can not connect to db: %v", err)
		return nil, err
	}

	// Run the auto migration tool.
	if err := db.dbEnt.Schema.Create(db.ctx); err != nil {
		log.Errorf("failed creating schema resources: %v", err)
		return nil, err
	}

	// increasing the version of the workflow by one
	// can be used to lookup which workflow uses which revision
	db.dbEnt.Workflow.Use(func(next ent.Mutator) ent.Mutator {
		return hook.WorkflowFunc(func(ctx context.Context, m *ent.WorkflowMutation) (ent.Value, error) {
			m.AddRevision(1)
			return next.Mutate(ctx, m)
		})
	})

	return db, nil

}

func rollback(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		err = fmt.Errorf("%v: %v", err, rerr)
	}
	return err
}

// func (db *dbManager) messageIDCleaner(data []byte) error {
//
// 	old := time.Now().Add(time.Minute * -5)
//
// 	i, err := db.dbEnt.MessageID.
// 		Delete().
// 		Where(messageid.AddedLT(old)).
// 		Exec(db.ctx)
//
// 	if err != nil {
// 		log.Errorf("can not delete old messages: %v", err)
// 		return err
// 	}
//
// 	log.Debugf("old messages deleted %d", i)
//
// 	return nil
//
// }

func (db *dbManager) tryLockDB(id uint64) (bool, *sql.Conn, error) {

	var gotLock bool
	conn, err := db.dbEnt.DB().Conn(db.ctx)
	if err != nil {
		return false, nil, err
	}

	conn.QueryRowContext(db.ctx, "SELECT pg_try_advisory_lock($1)", int64(id)).Scan(&gotLock)

	// close conn if we did not get the lock
	if !gotLock {
		conn.Close()
	}

	return gotLock, conn, nil

}

func (db *dbManager) lockDB(id uint64, wait int) (*sql.Conn, error) {

	ctx, cancel := context.WithTimeout(db.ctx, time.Duration(wait)*time.Second)
	defer cancel()

	conn, err := db.dbEnt.DB().Conn(db.ctx)
	if err != nil {
		return nil, err
	}

	_, err = conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", int64(id))

	if err, ok := err.(*pq.Error); ok {

		log.Debugf("db lock failed: %v", err)
		if err.Code == "57014" {
			conn.Close()
			return conn, fmt.Errorf("canceled query")
		}

		conn.Close()
		return conn, err

	}

	return conn, err

}

// func (db *dbManager) lockDB(ltype, id, wait int) (*sql.Conn, error) {
//
// 	ctx, cancel := context.WithTimeout(db.ctx, time.Duration(wait)*time.Second)
// 	defer cancel()
//
// 	conn, err := db.dbEnt.DB().Conn(db.ctx)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	_, err = conn.ExecContext(ctx, "SELECT pg_advisory_lock($1, $2)", ltype, id)
//
// 	if err, ok := err.(*pq.Error); ok {
//
// 		log.Debugf("db lock failed: %v", err)
// 		if err.Code == "57014" {
// 			return conn, fmt.Errorf("canceled query")
// 		}
//
// 		return conn, err
//
// 	}
//
// 	return conn, err
//
// }

func (db *dbManager) unlockDB(id uint64, conn *sql.Conn) error {

	_, err := conn.ExecContext(db.ctx,
		"SELECT pg_advisory_unlock($1)", int64(id))

	if err != nil {
		log.Errorf("can not unlock lock %d: %v", id, err)
	}
	conn.Close()

	return err

}

func (db *dbManager) getNamespace(name string) (*ent.Namespace, error) {

	ns, err := db.dbEnt.Namespace.
		Query().
		Where(namespace.IDEQ(name)).
		Only(db.ctx)

	if err != nil {
		return nil, err
	}

	return ns, nil

}

func (db *dbManager) addNamespace(ctx context.Context, name string) (*ent.Namespace, error) {

	key := make([]byte, 32)
	_, err := rand.Read(key)

	if err != nil {
		return nil, err
	}

	ns, err := db.dbEnt.Namespace.
		Create().
		SetID(name).
		SetKey(key).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return ns, nil

}

func (db *dbManager) deleteNamespace(ctx context.Context, name string) error {

	i, err := db.dbEnt.Namespace.
		Delete().
		Where(namespace.IDEQ(name)).
		Exec(ctx)

	if err != nil {
		return err
	}

	if i == 0 {
		return fmt.Errorf("namespace %s does not exist", name)
	}

	return nil

}

func (db *dbManager) getNamespaces(ctx context.Context, offset, limit int) ([]*ent.Namespace, error) {

	if limit == 0 {
		limit = math.MaxInt32
	}

	ns, err := db.dbEnt.Namespace.
		Query().
		Limit(limit).
		Offset(offset).
		Order(ent.Asc(namespace.FieldID)).
		All(ctx)

	if err != nil {
		return nil, err
	}

	return ns, nil

}

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
			db.deleteWorkflowEventWaitByListenerID(wfe.ID)
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

// func (db *dbManager) addWorkflow(cmd CmdAddWorkflow) (*ent.Workflow, error) {
func (db *dbManager) addWorkflow(ctx context.Context, ns, name, description string, active bool,
	workflow []byte, startDefinition model.StartDefinition) (*ent.Workflow, error) {

	tx, err := db.dbEnt.Tx(ctx)
	if err != nil {
		return nil, err
	}

	wf, err := tx.Workflow.
		Create().
		SetName(name).
		SetActive(active).
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

func (db *dbManager) deleteWorkflowEventListener(id int) error {

	_, err := db.dbEnt.WorkflowEvents.
		Delete().
		Where(workflowevents.IDEQ(id)).
		Exec(db.ctx)

	return err
}

func (db *dbManager) addWorkflowEventListener(wfid uuid.UUID,
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
		SetCount(count).
		Save(db.ctx)

}

const maxInstancesPerInterval = 100
const maxInstancesLimitInterval = time.Minute

func (db *dbManager) addWorkflowInstance(ns, workflowId, instanceId, input string) (*ent.WorkflowInstance, error) {

	count, err := db.dbEnt.WorkflowInstance.
		Query().
		Where(workflowinstance.HasWorkflowWith(workflow.HasNamespaceWith(namespace.IDEQ(ns)))).
		Where(workflowinstance.BeginTimeGT(time.Now().Add(-maxInstancesLimitInterval))).
		Count(db.ctx)
	if err != nil {
		return nil, err
	}

	if count > maxInstancesPerInterval {
		return nil, NewCatchableError("direktiv.limits.instances", "new workflow instance rejected because it would exceed the maximum number of new workflow instances (%d) per time interval (%s) for the namespace", maxInstancesPerInterval, maxInstancesLimitInterval)
	}

	wf, err := db.getNamespaceWorkflow(workflowId, ns)
	if err != nil {
		return nil, err
	}

	wi, err := db.dbEnt.WorkflowInstance.
		Create().
		SetInstanceID(instanceId).
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

func (db *dbManager) updateWorkflow(ctx context.Context, id string, revision *int, name, description string,
	active *bool, workflow []byte, startDefinition model.StartDefinition) (*ent.Workflow, error) {

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

	wf, err := updater.Save(ctx)
	if err != nil {
		return nil, rollback(tx, err)
	}

	err = db.processWorkflowEvents(ctx, tx, wf, startDefinition, *active)
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

	i, err := db.dbEnt.Workflow.Delete().
		Where(workflow.IDEQ(u)).
		Exec(ctx)

	if err != nil {
		return err
	}

	if i == 0 {
		return fmt.Errorf("workflow with id %s does not exist", id)
	}

	// delete all event listeners and events
	uid, _ := uuid.Parse(id)
	wfe, err := db.getWorkflowEventByWorkflowUID(uid)
	if err == nil {
		db.deleteWorkflowEventWaitByListenerID(wfe.ID)
		db.deleteWorkflowEventListener(wfe.ID)
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
		Select(workflow.FieldID, workflow.FieldName, workflow.FieldCreated, workflow.FieldDescription, workflow.FieldActive, workflow.FieldRevision).
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

func (db *dbManager) addTimer(name, fn, pattern string, t *time.Time, data []byte) (*ent.Timer, error) {

	tc := db.dbEnt.Timer.
		Create().
		SetFn(fn).
		SetName(name).
		SetData(data)

	if t == nil {
		tc.SetCron(pattern)
	} else {
		tc.SetOne(*t)
	}

	return tc.Save(db.ctx)

}

func (db *dbManager) getTimerByID(id int) (*ent.Timer, error) {

	return db.dbEnt.Timer.
		Query().
		Where(timer.IDEQ(id)).
		Only(db.ctx)

}

func (db *dbManager) getTimersWithPrefix(name string) ([]*ent.Timer, error) {

	return db.dbEnt.Timer.
		Query().
		Where(timer.NameHasPrefix(name)).
		All(db.ctx)

}

func (db *dbManager) getTimer(name string) (*ent.Timer, error) {

	return db.dbEnt.Timer.
		Query().
		Where(timer.NameEQ(name)).
		Only(db.ctx)

}

func (db *dbManager) deleteExpiredOneshots() (int, error) {

	compTime := time.Now().UTC().Add(-2 * time.Minute)

	log.Debugf("cleaning expired one-shot events")

	d, err := db.dbEnt.Timer.
		Delete().
		Where(
			timer.And(
				timer.OneNotNil(),
				timer.OneLT(compTime),
			)).
		Exec(db.ctx)

	if err != nil {
		return 0, err
	}

	return d, nil

}

func (db *dbManager) getTimers() ([]*ent.Timer, error) {

	timers, err := db.dbEnt.Timer.
		Query().
		All(db.ctx)

	if err != nil {
		return nil, err
	}

	return timers, nil

}

func (db *dbManager) deleteTimer(name string) error {

	_, err := db.dbEnt.Timer.
		Delete().
		Where(timer.NameEQ(name)).
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

func (db *dbManager) deleteWorkflowEventWait(id int) error {

	_, err := db.dbEnt.WorkflowEventsWait.
		Delete().
		Where(workfloweventswait.IDEQ(id)).
		Exec(db.ctx)

	return err

}

// func (db *dbManager) addWorkflowEventWait(ev map[string]map[string]interface{}, count, id int) (*ent.WorkflowEventsWait, error) {
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
