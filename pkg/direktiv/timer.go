package direktiv

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	cron "github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"

	"github.com/vorteil/direktiv/ent"
	"github.com/vorteil/direktiv/ent/workflowinstance"
)

const (
	timerCleanServer          = "cleanServer"
	timerSchedWorkflow        = "schedWorkflow"
	timerCleanInstanceRecords = "cleanInstanceRecords"
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
	name      string
	data      []byte

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

func (tm *timerManager) prepDisableTimer(ti *timerItem) (string, error) {

	switch ti.timerType {
	case timerTypeOneShot:
		// only if the timer had been setup
		if ti.oneshot.timer != nil {
			ti.oneshot.timer.Stop()
		}
	case timerTypeCron:
		tm.cron.Remove(ti.cron.cronID)
	default:
		return "", fmt.Errorf("unknown timer type")
	}

	return ti.name, nil
}

func (tm *timerManager) disableTimer(ti *timerItem) error {

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

	tm.mtx.Lock()
	defer tm.mtx.Unlock()
	delete(tm.timers, ti.name)

	return nil
}

func (tm *timerManager) executeFunction(ti *timerItem) {

	log.Debugf("execute timer %s", ti.name)

	err := ti.fn(ti.data)
	if err != nil {
		log.Errorf("can not run function for %s: %v", ti.name, err)
	}

	if ti.timerType == timerTypeOneShot {
		tm.disableTimer(ti)
	}

}

func (tm *timerManager) newTimerItem(name, fn string, data []byte, time *time.Time,
	pattern string) (*timerItem, error) {

	log.Debugf("adding new timer item %s", name)

	tm.mtx.Lock()
	defer tm.mtx.Unlock()

	var (
		exeFn func([]byte) error
		ok    bool
	)

	// check if the function had been registered
	if exeFn, ok = tm.fns[fn]; !ok {
		return nil, fmt.Errorf("can not add timer %s, invalid function %s", name, fn)
	}

	ti := new(timerItem)

	ti.timerType = timerTypeOneShot
	ti.oneshot.time = time
	ti.fn = exeFn
	ti.name = name
	ti.data = data

	if time == nil || time.IsZero() {
		ti.timerType = timerTypeCron
		ti.cron.pattern = pattern
	}

	tm.timers[name] = ti

	return ti, nil

}

func newTimerManager(s *WorkflowServer) (*timerManager, error) {

	tm := &timerManager{
		fns:    make(map[string]func([]byte) error),
		cron:   cron.New(),
		server: s,

		// timers can be key as name because it is unique
		timers: make(map[string]*timerItem),
	}

	// kick cron
	go tm.cron.Start()

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
		tm.disableTimer(ti)
	}

	log.Debugf("timers stopped")

}

func (tm *timerManager) addCron(name, fn, pattern string, data []byte) error {

	err := syncServer(tm.server.dbManager.ctx, tm.server.dbManager, &tm.server.id, map[string]interface{}{
		"name":    name,
		"fn":      fn,
		"pattern": pattern,
		"data":    data,
	}, AddCron)
	if err != nil {
		log.Error(err)
	}

	return tm.addCronNoBroadcast(name, fn, pattern, data)

}

func (tm *timerManager) addCronNoBroadcast(name, fn, pattern string, data []byte) error {

	// check if cron pattern matches
	c := cron.NewParser(cron.Minute | cron.Hour | cron.Dom |
		cron.Month | cron.DowOptional | cron.Descriptor)
	_, err := c.Parse(pattern)
	if err != nil {
		return err
	}

	ti, err := tm.newTimerItem(name, fn, data, nil, pattern)
	if err != nil {
		return err
	}

	err = func(ti *timerItem) error {
		id, err := tm.cron.AddFunc(ti.cron.pattern, func() {
			tm.executeFunction(ti)
		})
		if err != nil {
			log.Errorf("can not add cron function: %v", err)
			return fmt.Errorf("can not enable timer %s: %v", name, err)
		}
		ti.cron.cronID = id
		log.Debugf("added cron %s at %s", ti.name, ti.cron.pattern)
		return nil
	}(ti)

	tm.timers[name] = ti

	return err

}

func (tm *timerManager) addOneShot(name, fn string, timeos time.Time, data []byte) error {

	utc := timeos.UTC()

	ti, err := tm.newTimerItem(name, fn, data, &utc, "")
	if err != nil {
		return err
	}

	duration := ti.oneshot.time.UTC().Sub(time.Now().UTC())
	if duration < 0 {
		return fmt.Errorf("one-shot %s is in the past", ti.name)
	}

	func(ti *timerItem, duration time.Duration) error {

		timer := time.AfterFunc(duration, func() {
			tm.executeFunction(ti)
		})
		ti.oneshot.timer = timer
		log.Debugf("firing one-shot in %v", duration)

		return nil
	}(ti, duration)

	return nil

}

func (tm *timerManager) deleteTimersForInstance(name string) error {

	log.Debugf("deleting timers for instance %s", name)

	err := syncServer(tm.server.dbManager.ctx, tm.server.dbManager, &tm.server.id, name, CancelInstanceTimers)
	if err != nil {
		log.Error(err)
	}

	return tm.deleteTimersForInstanceNoBroadcast(name)

}

func (tm *timerManager) deleteTimersForInstanceNoBroadcast(name string) error {

	log.Debugf("deleting timers for instance %s", name)

	var keys []string

	tm.mtx.Lock()
	defer tm.mtx.Unlock()

	delT := func(pattern, name string) error {
		for _, n := range tm.timers {
			if strings.HasPrefix(n.name, fmt.Sprintf(pattern, name)) {
				key, err := tm.prepDisableTimer(n)
				if err != nil {
					return err
				}
				keys = append(keys, key)
			}
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

	for _, key := range keys {
		delete(tm.timers, key)
	}

	return nil
}

func (tm *timerManager) deleteTimerByName(oldController, newController, name string) error {

	if oldController != newController && oldController != "" {
		// send delete to specific server
		var err error
		req := map[string]interface{}{"action": "deleteTimer"}
		req["timerId"] = name
		err = publishToHostname(tm.server.engine.db, oldController, req)
		if err != nil {
			log.Error(err)
		}
		return nil
	}

	// delete local timer
	var key string
	var err error

	tm.mtx.Lock()

	if ti, ok := tm.timers[name]; ok {
		key, err = tm.prepDisableTimer(ti)
		if err != nil {
			log.Error(err)
		}
	}

	delete(tm.timers, key)

	tm.mtx.Unlock()

	if newController == "" {
		// broadcast timer delete
		err := syncServer(tm.server.dbManager.ctx, tm.server.dbManager, &tm.server.id, name, CancelTimer)
		if err != nil {
			log.Error(err)
		}
	}

	return nil
}

// cron job to delete old instance records / logs
func (tm *timerManager) cleanInstanceRecords(data []byte) error {
	log.Debugf("deleting old instance records/logs")
	ctx := context.Background()

	// search db for instances where "endTime" > defined lifespan
	wfis, err := tm.server.dbManager.dbEnt.WorkflowInstance.Query().
		Where(workflowinstance.EndTimeLTE(time.Now().Add(time.Minute * -10))).All(ctx)
	if err != nil {
		return err
	}

	// for each result, delete instance logs and delete row from DB
	for _, wfi := range wfis {
		err = tm.server.instanceLogger.DeleteInstanceLogs(wfi.InstanceID)
		if err != nil {
			if !ent.IsNotFound(err) {
				return err
			}
		}

		err = tm.server.dbManager.deleteWorkflowInstance(wfi.ID)
		if err != nil {
			if !ent.IsNotFound(err) {
				return err
			}
		}
	}
	log.Debugf("deleted %d instance records", len(wfis))

	return nil
}

func (tm *timerManager) deleteCronForWorkflow(id string) error {
	return tm.deleteTimerByName("", "", fmt.Sprintf("cron:%s", id))
}
