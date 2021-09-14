package flow

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

const (
	timerCleanServer           = "cleanServer"
	timerSchedWorkflow         = "schedWorkflow"
	timerCleanInstanceRecords  = "cleanInstanceRecords"
	timerCleanNamespaceRecords = "cleanNamespaceRecords"
	timerCleanExpiredEvents    = "timerCleanExpiredEvents"
	wfCron                     = "wfcron"
)

const (
	timerTypeCron = iota
	timerTypeOneShot
)

type timers struct {
	mtx      sync.Mutex
	cron     *cron.Cron
	fns      map[string]func([]byte)
	timers   map[string]*timer
	pubsub   *pubsub
	hostname string
}

func initTimers(pubsub *pubsub) (*timers, error) {

	timers := new(timers)
	timers.fns = make(map[string]func([]byte))
	timers.cron = cron.New()
	timers.timers = make(map[string]*timer) // timers can be key as name because it is unique
	timers.pubsub = pubsub

	var err error

	timers.hostname, err = os.Hostname()
	if err != nil {
		return nil, err
	}

	go timers.cron.Start()

	return timers, nil

}

func (timers *timers) Close() error {

	timers.stopTimers()

	return nil

}

type timer struct {
	timerType int
	name      string
	data      []byte

	fn func([]byte)

	cron struct {
		pattern string
		cronID  cron.EntryID
	}

	oneshot struct {
		time  *time.Time
		timer *time.Timer
	}
}

// stopTimers stops crons and one-shots
func (timers *timers) stopTimers() {

	ctx := timers.cron.Stop()
	<-ctx.Done()

	for _, timer := range timers.timers {
		timers.disableTimer(timer)
	}

}

func (timers *timers) prepDisableTimer(timer *timer) string {

	switch timer.timerType {

	case timerTypeOneShot:
		// only if the timer had been setup
		if timer.oneshot.timer != nil {
			timer.oneshot.timer.Stop()
		}

	case timerTypeCron:
		timers.cron.Remove(timer.cron.cronID)

	default:
		fmt.Fprintf(os.Stderr, "%v\n", fmt.Errorf("unknown timer type"))

	}

	return timer.name

}

func (timers *timers) disableTimer(timer *timer) {

	name := timers.prepDisableTimer(timer)

	timers.mtx.Lock()
	defer timers.mtx.Unlock()

	delete(timers.timers, name)

}

func (timers *timers) executeFunction(timer *timer) {

	timer.fn(timer.data)

	if timer.timerType == timerTypeOneShot {
		timers.disableTimer(timer)
	}

}

func (timers *timers) newTimer(name, fn string, data []byte, time *time.Time, pattern string) (*timer, error) {

	timers.mtx.Lock()
	defer timers.mtx.Unlock()

	exeFn, ok := timers.fns[fn]
	if !ok {
		return nil, fmt.Errorf("can not add timer %s, invalid function %s", name, fn)
	}

	timer := new(timer)

	timer.timerType = timerTypeOneShot
	timer.oneshot.time = time
	timer.fn = exeFn
	timer.name = name
	timer.data = data

	if time == nil || time.IsZero() {
		timer.timerType = timerTypeCron
		timer.cron.pattern = pattern
	}

	timers.timers[name] = timer

	return timer, nil

}

// registerFunction adds functions which can be executed by one-shots or crons
func (timers *timers) registerFunction(name string, fn func([]byte)) {

	if _, ok := timers.fns[name]; ok {
		panic(fmt.Errorf("function already exists"))
	}

	timers.fns[name] = fn

}

/* TODO
func (timers *timers) addCron(name, fn, pattern string, data []byte) error {

	err := syncServer(ctx, timers.srv, &timers.srv.ID, map[string]interface{}{
		"name":    name,
		"fn":      fn,
		"pattern": pattern,
		"data":    data,
	}, AddCron)
	if err != nil {
		appLog.Error(err)
	}

	return timers.addCronNoBroadcast(name, fn, pattern, data)

}
*/

func (timers *timers) addCron(name, fn, pattern string, data []byte) error {

	// check if cron pattern matches
	c := cron.NewParser(cron.Minute | cron.Hour | cron.Dom |
		cron.Month | cron.DowOptional | cron.Descriptor)
	_, err := c.Parse(pattern)
	if err != nil {
		return err
	}

	t, err := timers.newTimer(name, fn, data, nil, pattern)
	if err != nil {
		return err
	}

	id, err := timers.cron.AddFunc(t.cron.pattern, func() {
		timers.executeFunction(t)
	})
	if err != nil {
		return fmt.Errorf("can not enable timer %s: %v", name, err)
	}

	t.cron.cronID = id

	timers.timers[name] = t

	return nil

}

func (timers *timers) addOneShot(name, fn string, timeos time.Time, data []byte) error {

	utc := timeos.UTC()

	t, err := timers.newTimer(name, fn, data, &utc, "")
	if err != nil {
		return err
	}

	duration := t.oneshot.time.UTC().Sub(time.Now().UTC())
	if duration < 0 {
		return fmt.Errorf("one-shot %s is in the past", t.name)
	}

	func(timer *timer, duration time.Duration) error {

		clock := time.AfterFunc(duration, func() {
			timers.executeFunction(timer)
		})

		timer.oneshot.timer = clock

		return nil

	}(t, duration)

	return nil

}

func (timers *timers) deleteTimersForInstance(name string) {

	timers.pubsub.ClusterDeleteInstanceTimers(name)

}

func (timers *timers) deleteTimersForInstanceNoBroadcast(name string) {

	var keys []string

	timers.mtx.Lock()
	defer timers.mtx.Unlock()

	delT := func(pattern, name string) {
		for _, n := range timers.timers {
			if strings.HasPrefix(n.name, fmt.Sprintf(pattern, name)) {
				key := timers.prepDisableTimer(n)
				keys = append(keys, key)
			}
		}
	}

	patterns := []string{
		"timeout:%s",
		"%s",
	}

	for _, p := range patterns {
		delT(p, name)
	}

	for _, key := range keys {
		delete(timers.timers, key)
	}

}

func (timers *timers) deleteInstanceTimersHandler(req *pubsubUpdate) {

	timers.deleteTimersForInstanceNoBroadcast(req.Key)

}

func (timers *timers) deleteTimerHandler(req *pubsubUpdate) {
	timers.deleteTimerByName("", timers.pubsub.hostname, req.Key)

}

func (timers *timers) deleteTimerByName(oldController, newController, name string) {

	if oldController != newController && oldController != "" {

		// send delete to specific server

		timers.pubsub.HostnameDeleteTimer(oldController, name)

		return

	}

	// delete local timer
	var key string

	timers.mtx.Lock()

	if timer, ok := timers.timers[name]; ok {
		key = timers.prepDisableTimer(timer)
	}

	delete(timers.timers, key)

	timers.mtx.Unlock()

	if newController == "" {
		timers.pubsub.ClusterDeleteTimer(name)
	}

}

func (timers *timers) deleteCronForWorkflow(id string) {
	timers.deleteTimerByName("", timers.hostname, fmt.Sprintf("cron:%s", id))
}

/*
func (timers *timers) cleanExpiredEvents(data []byte) error {

	return timers.srv.deleteExpiredEvents()

}

// cron job delete old namespace logs every 2 hrs
func (timers *timers) cleanNamespaceRecords(data []byte) error {

	lc := timers.srv.components[util.LogComponent].(*logClient)
	return lc.deleteNamespaceLogs()

}

// cron job to delete old instance records / logs
func (timers *timers) cleanInstanceRecords(data []byte) error {
	ctx := context.Background()

	// search db for instances where "endTime" > defined lifespan
	wfis, err := timers.server.dbManager.dbEnt.WorkflowInstance.Query().
		Where(workflowinstance.EndTimeLTE(time.Now().Add(time.Minute * -10))).All(ctx)
	if err != nil {
		return err
	}

	// for each result, delete instance logs and delete row from DB
	for _, wfi := range wfis {

		lc := timers.server.components[util.LogComponent].(*logClient)
		err = lc.deleteInstanceLogs(wfi.InstanceID)
		if err != nil {
			return err
		}

		err = timers.server.dbManager.deleteWorkflowInstance(wfi.ID)
		if err != nil {
			if !ent.IsNotFound(err) {
				return err
			}
		}
	}
	appLog.Debugf("deleted %d instance records", len(wfis))

	return nil
}
*/
