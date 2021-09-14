package flow

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/mitchellh/hashstructure/v2"
)

const defaultLockWait = time.Second * 10

type locks struct {
	db *sql.DB
}

func initLocks(conn string) (*locks, error) {

	var err error

	locks := new(locks)

	locks.db, err = sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}

	locks.db.SetConnMaxIdleTime(-1)
	locks.db.SetConnMaxLifetime(-1)
	locks.db.SetMaxOpenConns(10)
	locks.db.SetMaxIdleConns(10)

	return locks, nil

}

func (locks *locks) Close() error {

	if locks.db != nil {

		err := locks.db.Close()
		if err != nil {
			return err
		}

		locks.db = nil

		return nil

	}

	return nil

}

func (locks *locks) tryLockDB(id uint64) (bool, *sql.Conn, error) {

	var gotLock bool

	conn, err := locks.db.Conn(context.Background())
	if err != nil {
		return false, nil, err
	}

	conn.QueryRowContext(context.Background(), "SELECT pg_try_advisory_lock($1)", int64(id)).Scan(&gotLock)
	if !gotLock {
		conn.Close()
	}

	return gotLock, conn, nil

}

func (locks *locks) lockDB(id uint64, wait int) (*sql.Conn, error) {

	var err error

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(wait)*time.Second)
	defer cancel()

	conn, err := locks.db.Conn(ctx)
	if err != nil {
		return nil, err
	}

	_, err = conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", int64(id))

	if err, ok := err.(*pq.Error); ok {

		if err.Code == "57014" {
			return conn, fmt.Errorf("canceled query")
		}
		return conn, err

	}

	return conn, err

}

func (locks *locks) unlockDB(id uint64, conn *sql.Conn) error {

	_, err := conn.ExecContext(context.Background(),
		"SELECT pg_advisory_unlock($1)", int64(id))

	if err != nil {
		return fmt.Errorf("can not unlock lock %d: %v", id, err)
	}

	err = conn.Close()

	if err != nil {
		return fmt.Errorf("can not close database connection %d: %v", id, err)
	}

	return nil

}

func (engine *engine) lock(key string, timeout time.Duration) (context.Context, *sql.Conn, error) {

	hash, err := hashstructure.Hash(key, hashstructure.FormatV2, nil)
	if err != nil {
		return nil, nil, NewInternalError(err)
	}

	wait := int(timeout.Seconds())

	conn, err := engine.locks.lockDB(hash, wait)
	if err != nil {
		return nil, nil, NewInternalError(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	engine.cancellersLock.Lock()
	engine.cancellers[key] = cancel
	engine.cancellersLock.Unlock()

	return ctx, conn, nil

}

func (engine *engine) InstanceLock(im *instanceMemory, timeout time.Duration) (context.Context, error) {

	key := im.ID().String()

	ctx, conn, err := engine.lock(key, timeout)
	if err != nil {
		return nil, err
	}

	im.lock = conn

	return ctx, nil

}

func (engine *engine) unlock(key string, conn *sql.Conn) {

	hash, err := hashstructure.Hash(key, hashstructure.FormatV2, nil)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

	engine.cancellersLock.Lock()
	defer engine.cancellersLock.Unlock()

	cancel := engine.cancellers[key]
	delete(engine.cancellers, key)
	cancel()

	err = engine.locks.unlockDB(hash, conn)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

}

func (engine *engine) InstanceUnlock(im *instanceMemory) {

	if im.lock == nil {
		return
	}

	engine.unlock(im.ID().String(), im.lock)
	im.lock = nil

}
