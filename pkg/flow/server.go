package flow

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/flow/nohome"
	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	"github.com/direktiv/direktiv/pkg/mirror"
	pubsub2 "github.com/direktiv/direktiv/pkg/pubsub"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/nats-io/nats.go"
)

type server struct {
	ID uuid.UUID

	config *core.Config

	// db       *ent.Client
	pubsub *pubsub.Pubsub

	// the new pubsub bus
	Bus *pubsub2.Bus

	timers *timers
	Engine *engine

	rawDB *sql.DB

	db *database.DB

	MirrorManager *mirror.Manager

	ServiceManager core.ServiceManager

	flow   *flow
	events *events
	nats   *nats.Conn

	ConfigureWorkflow func(event *pubsub2.FileSystemChangeEvent) error
}

type mirrorCallbacks struct {
	store    datastore.MirrorStore
	fstore   filestore.FileStore
	varstore datastore.RuntimeVariablesStore
	wfconf   func(ctx context.Context, nsID uuid.UUID, nsName string, file *filestore.File) error
}

func (c *mirrorCallbacks) ConfigureWorkflowFunc(ctx context.Context, nsID uuid.UUID, nsName string, file *filestore.File) error {
	return c.wfconf(ctx, nsID, nsName, file)
}

// func (c *mirrorCallbacks) ProcessLogger() mirror.ProcessLogger {
// 	return c.logger
// }

func (c *mirrorCallbacks) Store() datastore.MirrorStore {
	return c.store
}

func (c *mirrorCallbacks) FileStore() filestore.FileStore {
	return c.fstore
}

func (c *mirrorCallbacks) VarStore() datastore.RuntimeVariablesStore {
	return c.varstore
}

var _ mirror.Callbacks = &mirrorCallbacks{}

func (srv *server) cleanup(closer func() error) {
	err := closer()
	if err != nil {
		slog.Error("server cleanup failed", "error", err)
	}
}

func (srv *server) NotifyCluster(msg string) error {
	ctx := context.Background()

	conn, err := srv.rawDB.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, "SELECT pg_notify($1, $2)", pubsub.FlowSync, msg)

	perr := new(pq.Error)

	if errors.As(err, &perr) {
		slog.Error("database notification to cluster failed", "error", perr)
		if perr.Code == "57014" {
			return fmt.Errorf("canceled query")
		}

		return err
	}

	return nil
}

func (srv *server) NotifyHostname(hostname, msg string) error {
	ctx := context.Background()

	conn, err := srv.rawDB.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	channel := fmt.Sprintf("hostname:%s", hostname)

	_, err = conn.ExecContext(ctx, "SELECT pg_notify($1, $2)", channel, msg)

	perr := new(pq.Error)

	if errors.As(err, &perr) {
		fmt.Fprintf(os.Stderr, "db notification failed: %v", perr)
		if perr.Code == "57014" {
			return fmt.Errorf("canceled query")
		}

		return err
	}

	return nil
}

func (srv *server) PublishToCluster(payload string) {
	srv.pubsub.Publish(&pubsub.PubsubUpdate{
		Handler: nohome.PubsubNotifyFunction,
		Key:     payload,
	})
}

func (srv *server) CacheNotify(req *pubsub.PubsubUpdate) {
	if srv.ID.String() == req.Sender {
		return
	}

	// TODO: Alan, needfix.
	// srv.database.HandleNotification(req.Key)
}

func (srv *server) registerFunctions() {
	srv.pubsub.RegisterFunction(nohome.PubsubNotifyFunction, srv.CacheNotify)

	srv.pubsub.RegisterFunction(pubsub.PubsubNotifyFunction, srv.pubsub.Notify)
	srv.pubsub.RegisterFunction(pubsub.PubsubDisconnectFunction, srv.pubsub.Disconnect)
	srv.pubsub.RegisterFunction(pubsub.PubsubDeleteTimerFunction, srv.timers.deleteTimerHandler)
	srv.pubsub.RegisterFunction(pubsub.PubsubDeleteInstanceTimersFunction, srv.timers.deleteInstanceTimersHandler)
	srv.pubsub.RegisterFunction(pubsub.PubsubCancelWorkflowFunction, srv.Engine.finishCancelWorkflow)
	srv.pubsub.RegisterFunction(pubsub.PubsubCancelMirrorProcessFunction, srv.Engine.finishCancelMirrorProcess)
	srv.pubsub.RegisterFunction(pubsub.PubsubConfigureRouterFunction, srv.flow.configureRouterHandler)

	srv.timers.registerFunction(timeoutFunction, srv.Engine.timeoutHandler)
	srv.timers.registerFunction(wfCron, srv.flow.cronHandler)
	srv.timers.registerFunction(retryWakeupFunction, srv.flow.Engine.retryWakeup)

	srv.pubsub.RegisterFunction(pubsub.PubsubDeleteActivityTimersFunction, srv.timers.deleteActivityTimersHandler)
}

func (srv *server) cronPoller() {
	for {
		srv.cronPoll()
		time.Sleep(time.Minute * 15)
	}
}

func (srv *server) cronPoll() {
	ctx := context.Background()
	tx, err := srv.flow.beginSQLTx(ctx)
	if err != nil {
		slog.Error("cronPoll executing transaction", "error", err)
		return
	}
	defer tx.Rollback()

	roots, err := tx.FileStore().GetAllRoots(ctx)
	if err != nil {
		slog.Error("cronPoll fetching all Roots from db", "error", err)
		return
	}

	for _, root := range roots {
		files, err := tx.FileStore().ForRootID(root.ID).ListAllFiles(ctx)
		if err != nil {
			slog.Error("cronPoll fetching RootID from db", "error", err)
			return
		}

		for _, file := range files {
			if file.Typ != filestore.FileTypeWorkflow {
				continue
			}

			srv.cronPollerWorkflow(ctx, tx, file)
		}
	}
}

func (srv *server) cronPollerWorkflow(ctx context.Context, tx *database.DB, file *filestore.File) {
	ms, err := validateRouter(ctx, tx, file)
	if err != nil {
		slog.Error("failed to validate Routing for a cron schedule", "error", err)
		return
	}

	if ms.Cron != "" {
		srv.timers.deleteCronForWorkflow(file.ID.String())
	}

	if ms.Cron != "" {
		err := srv.timers.addCron(file.ID.String(), wfCron, ms.Cron, []byte(file.ID.String()))
		if err != nil {
			slog.Error("failed to add cron schedule for workflow", "error", err, "cron_expression", ms.Cron)
			return
		}

		slog.Debug("loaded cron schedule for workflow", "workflow", file.Path, "cron_expression", ms.Cron)
	}
}

func this() string {
	pc, _, _, _ := runtime.Caller(1) //nolint:dogsled
	fn := runtime.FuncForPC(pc)
	elems := strings.Split(fn.Name(), ".")

	return elems[len(elems)-1]
}

func (srv *server) beginSQLTx(ctx context.Context, opts ...*sql.TxOptions) (*database.DB, error) {
	return srv.db.BeginTx(ctx, opts...)
}

func (srv *server) runSQLTx(ctx context.Context, fun func(tx *database.DB) error) error {
	tx, err := srv.beginSQLTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fun(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Initialize NATS connection using environment variables.
func (srv *server) initNATS(config *core.Config) error {
	natsURL := config.NatsHost
	if natsURL == "" {
		return fmt.Errorf("environment for NATS URL")
	}

	// Connect to NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS, err: %w", err)
	}

	// Store the NATS connection in the server struct
	srv.nats = nc
	slog.Info("successfully connected to NATS")

	return nil
}

// Example of publishing a message to NATS.
func (srv *server) publishDemoMessage(subject, msg string) error {
	err := srv.nats.Publish(subject, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to publish message, err: %w", err)
	}
	slog.Info("message published", "subject", subject, "msg", msg)

	return nil
}
