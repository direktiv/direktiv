package flow

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
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

func (locks *locks) lockDB(id uint64, wait int) (*sql.Conn, error) {
	var (
		err  error
		conn *sql.Conn
	)

	defer func() {
		if err != nil && conn != nil {
			conn.Close()
		}
	}()

	// ctx, cancel := context.WithTimeout(context.Background(),
	// 	time.Duration(wait)*5*time.Second)
	// defer cancel()

	conn, err = locks.db.Conn(context.Background())
	if err != nil {
		return nil, err
	}

	_, err = conn.ExecContext(context.Background(), "SELECT pg_advisory_lock($1)", int64(id))

	perr := new(pq.Error)

	if errors.As(err, &perr) {
		if perr.Code == "57014" {
			return conn, fmt.Errorf("canceled query")
		}
		return conn, err
	}

	return conn, err
}

func (locks *locks) unlockDB(id uint64, conn *sql.Conn) (err error) {
	defer func() {
		err = conn.Close()
		if err != nil {
			err = fmt.Errorf("can not close database connection %d: %w", id, err)
		}
	}()

	_, err = conn.ExecContext(context.Background(),
		"SELECT pg_advisory_unlock($1)", int64(id))
	if err != nil {
		err = fmt.Errorf("can not unlock lock %d: %w", id, err)
	}

	return err
}

func (engine *engine) lock(key string, timeout time.Duration) (context.Context, *sql.Conn, error) {
	hash, err := hashstructure.Hash(key, hashstructure.FormatV2, nil)
	if err != nil {
		return nil, nil, derrors.NewInternalError(err)
	}

	wait := int(timeout.Seconds())

	conn, err := engine.locks.lockDB(hash, wait)
	if err != nil {
		return nil, nil, derrors.NewInternalError(err)
	}

	st := debug.Stack()
	ch := make(chan bool, 1)

	ctx, cancel := context.WithCancel(context.Background())
	engine.cancellersLock.Lock()

	go func() {
		select {
		case <-time.After(time.Second * 30):
			fmt.Println("---- POTENTIAL LOCK LEAK ")
			fmt.Println(string(st))
			fmt.Println("----")
		case <-ch:
		}
	}()

	engine.cancellers[key] = func() {
		defer func() {
			r := recover()
			if r != nil {
				engine.sugar.Errorf("Recovered from an okay panic: %v.", err)
			}
		}()
		cancel()
		close(ch)
	}

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
		panic(err)
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
