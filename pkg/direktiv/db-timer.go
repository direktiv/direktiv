package direktiv

import (
	"time"

	"github.com/vorteil/direktiv/ent"
	"github.com/vorteil/direktiv/ent/timer"

	log "github.com/sirupsen/logrus"
)

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
