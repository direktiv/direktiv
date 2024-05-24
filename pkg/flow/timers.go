package flow

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	gormlock "github.com/go-co-op/gocron-gorm-lock"
	"github.com/go-co-op/gocron/v2"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

const (
	wfCron = "wfcron"
)

type timers struct {
	mtx      sync.Mutex
	fns      map[string]func([]byte)
	pubsub   *pubsub.Pubsub
	hostname string

	scheduler gocron.Scheduler
}

func initTimers(pubsub *pubsub.Pubsub, db *gorm.DB) (*timers, error) {
	timers := new(timers)
	timers.fns = make(map[string]func([]byte))
	timers.pubsub = pubsub

	var err error

	timers.hostname, err = os.Hostname()
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&gormlock.CronJobLock{})
	if err != nil {
		return nil, err
	}

	locker, err := gormlock.NewGormLocker(db, timers.hostname)
	if err != nil {
		return nil, err
	}

	// set to UTC and use db locker
	cronScheduler, err := gocron.NewScheduler(
		gocron.WithDistributedLocker(locker),
		gocron.WithLocation(time.UTC),
	)
	if err != nil {
		return nil, err
	}

	cronScheduler.Start()
	timers.scheduler = cronScheduler

	return timers, nil
}

func (timers *timers) Close() error {
	return timers.scheduler.Shutdown()
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

func (timers *timers) addTimer(name, fn string, definition gocron.JobDefinition, data []byte) error {
	exeFn, ok := timers.fns[fn]
	if !ok {
		return fmt.Errorf("can not add timer %s, invalid function %s", name, fn)
	}

	_, err := timers.scheduler.NewJob(
		definition,
		gocron.NewTask(
			exeFn, data,
		),
		gocron.WithName(name),
	)

	return err
}

func (timers *timers) addCron(name, fn, pattern string, data []byte) error {
	name = fmt.Sprintf("cron:%s", name)

	slog.Debug("cron timer creating", slog.String("name", name))
	c := cron.NewParser(cron.Minute | cron.Hour | cron.Dom |
		cron.Month | cron.DowOptional | cron.Descriptor)
	_, err := c.Parse(pattern)
	if err != nil {
		return err
	}

	definition := gocron.CronJob(
		pattern,
		false,
	)

	return timers.addTimer(name, fn, definition, data)
}

func (timers *timers) addOneShot(name, fn string, timeos time.Time, data []byte) error {
	slog.Debug("one shot timer creating", slog.String("name", name))

	duration := timeos.UTC().Sub(time.Now().UTC())
	if duration < 0 {
		return fmt.Errorf("one-shot %s is in the past", name)
	}

	definition := gocron.OneTimeJob(
		gocron.OneTimeJobStartDateTime(timeos),
	)

	return timers.addTimer(name, fn, definition, data)
}

func (timers *timers) deleteTimersForInstance(name string) {
	timers.pubsub.ClusterDeleteInstanceTimers(name)
}

func (timers *timers) deleteTimersForInstanceNoBroadcast(name string) {
	slog.Debug("deleting timer for instance", slog.String("name", name))
	patterns := []string{
		"timeout:%s",
		"%s",
	}

	jobs := timers.scheduler.Jobs()
	for i := range jobs {
		job := jobs[i]
		for a := range patterns {
			pattern := patterns[a]
			if strings.HasPrefix(job.Name(), fmt.Sprintf(pattern, name)) {
				slog.Debug("deleting job", slog.String("name", job.Name()))
				err := timers.scheduler.RemoveJob(job.ID())
				if err != nil {
					slog.Warn("can not remove timer", slog.String("timer", job.Name()))
				}
			}
		}
	}
}

func (timers *timers) deleteInstanceTimersHandler(req *pubsub.PubsubUpdate) {
	timers.deleteTimersForInstanceNoBroadcast(req.Key)
}

func (timers *timers) deleteTimerHandler(req *pubsub.PubsubUpdate) {
	timers.deleteTimerByName("", timers.pubsub.Hostname, req.Key)
}

func (timers *timers) deleteTimerByName(oldController, newController, name string) {
	if oldController != newController && oldController != "" {
		// send delete to specific server
		timers.pubsub.HostnameDeleteTimer(oldController, name)

		return
	}

	jobs := timers.scheduler.Jobs()
	for i := range jobs {
		job := jobs[i]
		if job.Name() == name {
			err := timers.scheduler.RemoveJob(job.ID())
			if err != nil {
				slog.Warn("can not remove timer", slog.String("timer", name))
			}
		}
	}

	if newController == "" {
		timers.pubsub.ClusterDeleteTimer(name)
	}
}

func (timers *timers) deleteCronForWorkflow(id string) {
	timers.deleteTimerByName("", timers.hostname, fmt.Sprintf("cron:%s", id))
}
