package direktiv

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vorteil/direktiv/ent/workflowinstance"

	hashstructure "github.com/mitchellh/hashstructure/v2"
	cron "github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/ent"
)

const (
	timerCleanServer          = "cleanServer"
	timerSchedWorkflow        = "schedWorkflow"
	timerCleanOneShot         = "cleanOneShot"
	timerCleanInstanceRecords = "cleanInstanceRecords"

	needsSyncRequest = true
	skipSyncRequest  = false
)

type timerManager struct {
	cron   *cron.Cron
	fns    map[string]func([]byte) error
	server *WorkflowServer

	timers map[string]*timerItem
	mtx    sync.Mutex
}

type timerItem struct {
	timerType int
	dbItem    *ent.Timer

	fn   func([]byte) error
	cron struct {
		pattern string
		cronID  cron.EntryID
	}
	oneshot struct {
		time  *time.Time
		timer *time.Timer
	}
}

const (
	timerTypeCron = iota
	timerTypeOneShot
)

func (tm *timerManager) disableTimer(ti *timerItem, remove, needsSync bool) error {

	var err error
	if ti == nil {
		return fmt.Errorf("timer item is nil")
	}

	log.Debugf("disable timer %s", ti.dbItem.Name)

	switch ti.timerType {
	case timerTypeOneShot:
		// only if the timer had been setup
		if ti.oneshot.timer != nil {
			ti.oneshot.timer.Stop()
		}
	case timerTypeCron:
		tm.cron.Remove(ti.cron.cronID)
	default:
		return fmt.Errorf("unknown timer type")
	}

	if remove {
		log.Debugf("delete timer: %s", ti.dbItem.Name)
		delete(tm.timers, ti.dbItem.Name)
		err = tm.server.dbManager.deleteTimer(ti.dbItem.Name)
	}

	// send delete or disable request across cluster
	if needsSync {

		action := DisableTimerSync
		if remove {
			action = DeleteTimerSync
		}

		err := syncServer(context.Background(), tm.server.dbManager,
			&tm.server.id, ti.dbItem.Name, action)
		if err != nil {
			log.Errorf("can not send time delete sync request: %v", err)
		}

	}

	return err

}

func (tm *timerManager) enableTimer(ti *timerItem, needsSync bool) error {

	var err error

	if ti == nil {
		return fmt.Errorf("timer item is nil")
	}

	// double check that the db item still exists
	_, err = tm.server.dbManager.getTimer(ti.dbItem.Name)
	if err != nil {
		return err
	}

	log.Debugf("enabling timer %s, %s %v", ti.dbItem.Name, ti.dbItem.One, time.Now().UTC())

	switch ti.timerType {
	case timerTypeOneShot:
		duration := ti.oneshot.time.UTC().Sub(time.Now().UTC())

		if duration < 0 {
			tm.disableTimer(ti, true, needsSyncRequest)
			return fmt.Errorf("one-shot %s is in the past", ti.dbItem.Name)
		}

		err = func(ti *timerItem, duration time.Duration) error {

			timer := time.AfterFunc(duration, func() {

				// check if the entry is still there or if any other server fired it
				_, err := tm.server.dbManager.getTimer(ti.dbItem.Name)
				if err == nil {
					tm.executeFunction(ti)
				} else {
					// only remove it from the map, it has been done already
					tm.disableTimer(ti, false, needsSyncRequest)
				}

			})
			ti.oneshot.timer = timer
			log.Debugf("firing one-shot in %v", duration)

			return nil
		}(ti, duration)

	case timerTypeCron:

		log.Debugf("enable cron %s", ti.dbItem.Name)

		err = func(ti *timerItem) error {
			id, err := tm.cron.AddFunc(ti.cron.pattern, func() {
				tm.executeFunction(ti)
			})
			if err != nil {
				return fmt.Errorf("can not enable timer %s: %v", ti.dbItem.Fn, err)
			}

			ti.cron.cronID = id
			return nil
		}(ti)

	default:
		return fmt.Errorf("unknown timer type")
	}

	// sync
	if needsSync {
		err = syncServer(context.Background(), tm.server.dbManager,
			&tm.server.id, ti.dbItem.Name, EnableTimerSync)
		if err != nil {
			log.Errorf("can not send time enable sync request: %v", err)
		}
	}

	return err
}

func (tm *timerManager) executeFunction(ti *timerItem) {

	if ti == nil {
		log.Errorf("timer item is nil")
		return
	}

	log.Debugf("execute timer %s", ti.dbItem.Name)

	// get lock
	hash, _ := hashstructure.Hash(fmt.Sprintf("%d%s", ti.dbItem.ID, ti.dbItem.Name),
		hashstructure.FormatV2, nil)
	hasLock, conn, err := tm.server.dbManager.tryLockDB(hash)
	if err != nil {
		log.Debugf("can not get lock %d", ti.dbItem.ID)
		return
	}

	if hasLock {

		unlock := func(hashin uint64) {
			// delay the unlock to make sure minimal time offsets accross a cluster
			// does not make that fire a second time if the executin is fast
			tm.server.dbManager.unlockDB(hashin, conn)
		}
		defer unlock(hash)

		ct, err := tm.server.dbManager.getTimerByID(ti.dbItem.ID)
		if err != nil {
			log.Errorf("can not get timer: %v", err)
			return
		}

		// if one shot was deleted, it has already failed fetching it
		// if it is around +/- 30s we stop here. the timer execution was so short
		// that another server got the lock as well. Can happen if they are millis off.
		last := ct.Last
		secs := time.Now().Sub(last).Seconds()
		if secs > -30 || secs < 30 {
			log.Debugf("double call, not executing")
		}

		if ti.timerType == timerTypeOneShot {
			log.Debugf("%s is one shot, disable (execute)", ti.dbItem.Name)
			tm.disableTimer(ti, true, needsSyncRequest)
		} else {
			// update last run time
			ti.dbItem, err = tm.server.dbManager.updateRunTime(ti.dbItem, time.Now())
			if err != nil {
				log.Debugf("can not set update time: %v", err)
				return
			}
		}

		err = ti.fn(ti.dbItem.Data)
		if err != nil {
			log.Errorf("can not run function for %s: %v", ti.dbItem.Name, err)
		}

	} else {
		log.Debugf("timer already locked %s", ti.dbItem.Name)
	}

}

func (tm *timerManager) newTimerItem(name, fn string, data []byte, time *time.Time,
	pattern string, dbItem *ent.Timer, needsSync bool) (*timerItem, error) {

	log.Debugf("adding new timer item %s", name)

	tm.mtx.Lock()
	defer tm.mtx.Unlock()

	var (
		exeFn func([]byte) error
		ok    bool
		err   error
	)

	// check if the function had been registered
	if exeFn, ok = tm.fns[fn]; !ok {
		return nil, fmt.Errorf("can not add timer %s, invalid function %s", name, fn)
	}

	if dbItem == nil {
		dbItem, err = tm.server.dbManager.addTimer(name, fn, pattern, time, data)
		if err != nil {
			return nil, err
		}
	}

	ti := new(timerItem)

	ti.timerType = timerTypeOneShot
	ti.oneshot.time = time
	ti.dbItem = dbItem
	ti.fn = exeFn

	if time == nil || time.IsZero() {
		ti.timerType = timerTypeCron
		ti.cron.pattern = pattern
	}

	tm.timers[dbItem.Name] = ti

	if needsSync {
		err = syncServer(context.Background(), tm.server.dbManager,
			&tm.server.id, ti.dbItem.ID, AddTimerSync)
		if err != nil {
			log.Errorf("can not send time add sync request: %v", err)
		}
	}

	return ti, tm.enableTimer(ti, false)
}

func newTimerManager(s *WorkflowServer) (*timerManager, error) {

	tm := &timerManager{
		fns:    make(map[string]func([]byte) error),
		cron:   cron.New(),
		server: s,

		// timers can be key as name because it is unique
		timers: make(map[string]*timerItem),
	}

	return tm, nil

}

// registerFunction adds functions which can be executed by one-shots or crons
func (tm *timerManager) registerFunction(name string, fn func([]byte) error) error {

	log.Debugf("adding timer function %s", name)
	if _, ok := tm.fns[name]; ok {
		return fmt.Errorf("function already exists")
	}

	tm.fns[name] = fn

	return nil
}

// stopTimers stops crons and one-shots
func (tm *timerManager) stopTimers() {

	log.Debugf("stopping timers")

	// stop all crons and clean
	ctx := tm.cron.Stop()
	<-ctx.Done()

	for _, ti := range tm.timers {
		log.Debugf("%s is one shot, disable (stopTimers)", ti.dbItem.Name)
		tm.disableTimer(ti, false, skipSyncRequest)
	}

	log.Debugf("timers stopped")

}

func (tm *timerManager) startTimers() error {

	log.Debugf("starting timers")
	ts, err := tm.server.dbManager.getTimers()
	if err != nil {
		return err
	}

	for _, t := range ts {

		_, err := tm.newTimerItem(t.Name, t.Fn, t.Data, &t.One,
			t.Cron, t, skipSyncRequest)
		if err != nil {
			log.Errorf("can not add timer: %v", err)
			continue
		}

	}

	log.Debugf("timers started")
	go tm.cron.Start()

	return nil

}

func (tm *timerManager) syncTimerAdd(id int) {

	log.Debugf("got sync request")
	t, err := tm.server.dbManager.getTimerByID(id)
	if err != nil {
		return
	}

	tm.mtx.Lock()
	if _, ok := tm.timers[t.Name]; ok {
		log.Debugf("timer already available")
	}
	tm.mtx.Unlock()

	tm.newTimerItem(t.Name, t.Fn, t.Data, &t.One, t.Cron, t, skipSyncRequest)

}

// sync function across cluster
func (tm *timerManager) syncTimerDelete(name string) {

	tm.mtx.Lock()
	defer tm.mtx.Unlock()

	if ti, ok := tm.timers[name]; ok {
		err := tm.disableTimer(ti, true, skipSyncRequest)
		if err != nil {
			log.Errorf("error executing sync command: %v", err)
		}
	}

}

func (tm *timerManager) syncTimerEnable(name string) {

	tm.mtx.Lock()
	defer tm.mtx.Unlock()

	if ti, ok := tm.timers[name]; ok {
		err := tm.enableTimer(ti, skipSyncRequest)
		if err != nil {
			log.Errorf("error executing sync enable command: %v", err)
		}
	}

}

func (tm *timerManager) syncTimerDisable(name string) {

	tm.mtx.Lock()
	defer tm.mtx.Unlock()

	if ti, ok := tm.timers[name]; ok {
		err := tm.disableTimer(ti, false, skipSyncRequest)
		if err != nil {
			log.Errorf("error executing sync disable command: %v", err)
		}
	}

}

func (tm *timerManager) addCron(name, fn, pattern string, data []byte) (*timerItem, error) {

	// cehck if cron pattern matches
	c := cron.NewParser(cron.Minute | cron.Hour | cron.Dom |
		cron.Month | cron.DowOptional | cron.Descriptor)
	_, err := c.Parse(pattern)

	if err != nil {
		return nil, err
	}

	return tm.newTimerItem(name, fn, data, nil, pattern, nil, needsSyncRequest)

}

func (tm *timerManager) addOneShot(name, fn string, timeos time.Time, data []byte) (*timerItem, error) {

	utc := timeos.UTC()
	return tm.newTimerItem(name, fn, data, &utc, "", nil, needsSyncRequest)

}

func (tm *timerManager) deleteTimersForInstance(name string) error {

	log.Debugf("deleting timers for instance %s", name)

	delT := func(pattern, name string) error {
		timers, err := tm.server.dbManager.getTimersWithPrefix(fmt.Sprintf(pattern, name))
		if err != nil {
			return err
		}

		for _, t := range timers {
			tm.actionTimerByName(t.Name, deleteTimerAction)
		}
		return nil
	}

	patterns := []string{
		"timeout:%s",
		"%s",
	}

	for _, p := range patterns {
		err := delT(p, name)
		if err != nil {
			return err
		}
	}

	return nil
}

const (
	deleteTimerAction = iota
	enableTimerAction
	disableTimerAction
)

func (tm *timerManager) actionTimerByName(name string, action int) error {

	tm.mtx.Lock()
	defer tm.mtx.Unlock()

	if ti, ok := tm.timers[name]; ok {

		switch action {
		case deleteTimerAction:
			return tm.disableTimer(ti, true, needsSyncRequest)
		case enableTimerAction:
			return tm.enableTimer(ti, needsSyncRequest)
		case disableTimerAction:
			return tm.disableTimer(ti, false, needsSyncRequest)
		default:
			return fmt.Errorf("unknown action %d", action)
		}

	}

	return fmt.Errorf("timer %s does not exist", name)

}

// cron job to delete orphaned one-shot timers
func (tm *timerManager) cleanOneShot(data []byte) error {

	d, err := tm.server.dbManager.deleteExpiredOneshots()
	if err != nil {
		return err
	}

	log.Debugf("%d old one-shots deleted", d)

	return nil
}

// cron job to delete old instance records / logs
func (tm *timerManager) cleanInstanceRecords(data []byte) error {
	log.Debugf("deleting old instance records/logs")
	ctx := context.Background()

	// search db for instances where "endTime" > defined lifespan
	wfis, err := tm.server.dbManager.dbEnt.WorkflowInstance.Query().Where(workflowinstance.EndTimeLTE(time.Now().Add(time.Minute * -10))).All(ctx)
	if err != nil {
		return err
	}

	// for each result, delete instance logs and delete row from DB
	for _, wfi := range wfis {
		err = tm.server.instanceLogger.DeleteInstanceLogs(wfi.InstanceID)
		if err != nil {
			return err
		}

		err = tm.server.dbManager.deleteWorkflowInstance(wfi.ID)
		if err != nil {
			return err
		}
	}
	log.Debugf("deleted %d instance records", len(wfis))

	return nil
}

func (tm *timerManager) deleteCronForWorkflow(id string) error {

	// get name
	name := fmt.Sprintf("cron:%s", id)

	// delete from database
	tm.server.dbManager.deleteTimer(name)

	// delete from local sync
	tm.syncTimerDelete(name)

	// send sync request
	return syncServer(context.Background(), tm.server.dbManager,
		&tm.server.id, name, DeleteTimerSync)

}
