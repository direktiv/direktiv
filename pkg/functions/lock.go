package functions

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/mitchellh/hashstructure/v2"
)

type locks struct {
	db *sql.DB
}

var locksmgr *locks

func initLocks(conn string) error {

	var err error

	locks := new(locks)

	locks.db, err = sql.Open("postgres", conn)
	if err != nil {
		return err
	}

	locks.db.SetConnMaxIdleTime(-1)
	locks.db.SetConnMaxLifetime(-1)
	locks.db.SetMaxOpenConns(10)
	locks.db.SetMaxIdleConns(10)

	locksmgr = locks

	return nil

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

	err = conn.QueryRowContext(context.Background(), "SELECT pg_try_advisory_lock($1)", int64(id)).Scan(&gotLock)
	if err != nil {
		return false, nil, err
	}
	if !gotLock {
		err = conn.Close()
		if err != nil {
			fmt.Println("CLOSE LOCK CONN ERROR", err)
		}
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

func (locks *locks) lock(key string, blocking bool) (*sql.Conn, error) {

	hash, err := hashstructure.Hash(key, hashstructure.FormatV2, nil)
	if err != nil {
		return nil, err
	}

	wait := int(time.Second)

	if blocking {
		wait = int(time.Minute) * 15
	}

	logger.Debugf("locking %s", key)

	conn, err := locks.lockDB(hash, wait)
	if err != nil {
		return nil, err
	}

	return conn, nil

}

func (locks *locks) unlock(key string, conn *sql.Conn) {

	hash, err := hashstructure.Hash(key, hashstructure.FormatV2, nil)
	if err != nil {
		panic(err)
	}

	logger.Debugf("unlocking %s", key)

	err = locks.unlockDB(hash, conn)
	if err != nil {
		return
	}

}

/*
var kubernetesLock *distributed_locker.DistributedLocker

func initKubernetesLock() error {

	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	dc, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	logger.Debugf("lock for cm %s in namespace %s",
		os.Getenv("DIREKTIV_LOCK_CM"), os.Getenv(util.DirektivNamespace))

	kubernetesLock = distributed_locker.NewKubernetesLocker(
		dc, schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "configmaps",
		}, os.Getenv("DIREKTIV_LOCK_CM"), os.Getenv(util.DirektivNamespace),
	)

	logger.Infof("kubernetes lock created")

	return nil

}

func kubeLock(key string, blocking bool) (lockgate.LockHandle, error) {

	logger.Debugf("locking %s", key)

	acquired, lock, err := kubernetesLock.Acquire(key,
		lockgate.AcquireOptions{Shared: false, NonBlocking: blocking,
			Timeout: 30 * time.Second})

	if err != nil {
		return lockgate.LockHandle{}, err
	}

	if !acquired {
		return lockgate.LockHandle{}, fmt.Errorf("lock %s not aquired", key)
	}

	return lock, nil

}

func kubeUnlock(lock lockgate.LockHandle) {

	logger.Debugf("unlocking %s", lock.LockName)

	err := kubernetesLock.Release(lock)
	if err != nil {
		logger.Errorf("can not unlock %v: %v", lock.LockName, err)
	}

}

*/
