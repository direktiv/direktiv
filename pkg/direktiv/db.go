package direktiv

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/ent"
	"github.com/vorteil/direktiv/ent/hook"
	secretsgrpc "github.com/vorteil/direktiv/pkg/secrets/grpc"
	"github.com/vorteil/direktiv/pkg/util"
	"github.com/vorteil/direktiv/pkg/varstore"
	"google.golang.org/grpc"
)

const (
	filterPrefix = "filter-"
)

// DBManager contains all database related information and functions
type dbManager struct {
	dbEnt      *ent.Client
	ctx        context.Context
	tm         *timerManager
	varStorage *varstore.VarStorage

	grpcConn      *grpc.ClientConn
	secretsClient secretsgrpc.SecretsServiceClient

	dbForLock *sql.DB
}

func prepLockDB(conn string) (*sql.DB, error) {

	db, err := sql.Open("postgres", conn)

	db.SetConnMaxIdleTime(-1)
	db.SetConnMaxLifetime(-1)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, err

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

	udb := db.dbEnt.DB()
	udb.SetMaxIdleConns(10)
	udb.SetMaxOpenConns(10)

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

	// get secrets client
	db.grpcConn, err = util.GetEndpointTLS(util.TLSSecretsComponent)
	if err != nil {
		return nil, err
	}
	db.secretsClient = secretsgrpc.NewSecretsServiceClient(db.grpcConn)

	db.dbForLock, err = prepLockDB(conn)
	if err != nil {
		return nil, err
	}

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

	conn, err := db.dbForLock.Conn(context.Background())
	if err != nil {
		return false, nil, err
	}

	conn.QueryRowContext(context.Background(), "SELECT pg_try_advisory_lock($1)", int64(id)).Scan(&gotLock)
	if !gotLock {
		conn.Close()
	}

	return gotLock, conn, nil

}

func (db *dbManager) lockDB(id uint64, wait int) (*sql.Conn, error) {

	var err error

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(wait)*time.Second)
	defer cancel()

	conn, err := db.dbForLock.Conn(ctx)
	if err != nil {
		return nil, err
	}

	_, err = conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", int64(id))

	if err, ok := err.(*pq.Error); ok {

		log.Debugf("db lock failed: %v", err)
		if err.Code == "57014" {
			return conn, fmt.Errorf("canceled query")
		}
		return conn, err

	}

	return conn, err

}

func (db *dbManager) unlockDB(id uint64, conn *sql.Conn) {

	_, err := conn.ExecContext(context.Background(),
		"SELECT pg_advisory_unlock($1)", int64(id))

	if err != nil {
		log.Errorf("can not unlock lock %d: %v", id, err)
	}

	err = conn.Close()

	if err != nil {
		log.Errorf("can not close database connection %d: %v", id, err)
	}

}
