package direktiv

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/ent"
	"github.com/vorteil/direktiv/ent/hook"
	secretsgrpc "github.com/vorteil/direktiv/pkg/secrets/grpc"
	"google.golang.org/grpc"
)

const (
	filterPrefix = "filter-"

	maxInstancesPerInterval   = 100
	maxInstancesLimitInterval = time.Minute
)

// DBManager contains all database related information and functions
type dbManager struct {
	dbEnt *ent.Client
	ctx   context.Context
	tm    *timerManager

	grpcConn      *grpc.ClientConn
	secretsClient secretsgrpc.SecretsServiceClient
}

func newDBManager(ctx context.Context, conn string, config *Config) (*dbManager, error) {

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

	kubeReq.mockup = true

	// setting the knative service template
	if config.MockupMode == 0 {

		st, err := ioutil.ReadFile("/etc/config/template")
		if err != nil {
			return nil, err
		}
		kubeReq.serviceTempl = string(st)

		kubeReq.sidecar = config.FlowAPI.Sidecar

		kubeReq.mockup = false

	}

	// get secrets client
	db.grpcConn, err = GetEndpointTLS(config, secretsComponent, config.SecretsAPI.Endpoint)
	if err != nil {
		return nil, err
	}
	db.secretsClient = secretsgrpc.NewSecretsServiceClient(db.grpcConn)

	return db, nil

}

func rollback(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		err = fmt.Errorf("%v: %v", err, rerr)
	}
	return err
}

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

func (db *dbManager) unlockDB(id uint64, conn *sql.Conn) error {

	_, err := conn.ExecContext(db.ctx,
		"SELECT pg_advisory_unlock($1)", int64(id))

	if err != nil {
		log.Errorf("can not unlock lock %d: %v", id, err)
	}
	conn.Close()

	return err

}
