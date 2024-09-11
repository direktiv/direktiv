package flow

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/pubsub"
	"github.com/robfig/cron/v3"
)

const (
	wfCron = "wfcron"
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
	pubsub   *pubsub.Bus
	hostname string
}

func initTimers(pubsub *pubsub.Bus) (*timers, error) {
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

	timers.pubsub.Subscribe(clusterDeleteInstanceTimersMessage{}, timers.deleteTimersForInstanceHandler)

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

// stopTimers stops crons and one-shots.
func (timers *timers) stopTimers() {
	ctx := timers.cron.Stop()
	<-ctx.Done()

	timers.mtx.Lock()
	defer timers.mtx.Unlock()
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

// must be locked before calling.
func (timers *timers) disableTimer(timer *timer) {
	name := timers.prepDisableTimer(timer)

	delete(timers.timers, name)
}

func (timers *timers) executeFunction(timer *timer) {
	timer.fn(timer.data)

	if timer.timerType == timerTypeOneShot {
		timers.mtx.Lock()
		defer timers.mtx.Unlock()

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

// registerFunction adds functions which can be executed by one-shots or crons.
func (timers *timers) registerFunction(name string, fn func([]byte)) {
	timers.mtx.Lock()
	defer timers.mtx.Unlock()

	if _, ok := timers.fns[name]; ok {
		panic(fmt.Errorf("function already exists"))
	}

	timers.fns[name] = fn
}

func (timers *timers) addCron(name, fn, pattern string, data []byte) error {
	timers.deleteCronForWorkflow(name)
	name = fmt.Sprintf("cron:%s", name)

	slog.Debug("Scheduling new cron job.", "cron_name", name, "cron_pattern", pattern)

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
		return fmt.Errorf("can not enable timer %s: %w", name, err)
	}

	t.cron.cronID = id

	timers.mtx.Lock()
	defer timers.mtx.Unlock()

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

	err = func(timer *timer, duration time.Duration) error {
		clock := time.AfterFunc(duration, func() {
			timers.executeFunction(timer)
		})

		timer.oneshot.timer = clock

		return nil
	}(t, duration)
	if err != nil {
		return err
	}

	return nil
}

type clusterDeleteInstanceTimersMessage struct {
	Name string
}

func (timers *timers) deleteTimersForInstance(name string) {
	_ = timers.pubsub.Publish(clusterDeleteInstanceTimersMessage{
		Name: name,
	})
}

func (timers *timers) deleteTimersForInstanceHandler(data string) {
	var msg clusterDeleteInstanceTimersMessage
	_ = json.Unmarshal([]byte(data), &msg)
	timers.deleteTimersForInstanceNoBroadcast(msg.Name)
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

type deleteTimerMessage struct {
	// timers.deleteTimerByName("", timers.pubsub.Hostname, req.Key)
	Name string
}

func (timers *timers) deleteTimer(name string) {
	_ = timers.pubsub.Publish(clusterDeleteInstanceTimersMessage{
		Name: name,
	})
}

func (timers *timers) deleteTimerHandler(data string) {
	var msg deleteTimerMessage
	_ = json.Unmarshal([]byte(data), &msg)
	timers.deleteTimerByName(msg.Name)
}

func (timers *timers) deleteTimerByName(name string) {
	// delete local timer
	var key string

	timers.mtx.Lock()

	if timer, ok := timers.timers[name]; ok {
		key = timers.prepDisableTimer(timer)
	}

	delete(timers.timers, key)

	timers.mtx.Unlock()
}

func (timers *timers) deleteCronForWorkflow(id string) {
	timers.deleteTimerByName(fmt.Sprintf("cron:%s", id))
}
