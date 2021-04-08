package direktiv

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

// FlowSync is the name of postgres pubsub channel
const FlowSync = "flowsync"

const (
	AddTimerSync = iota
	DeleteTimerSync
	EnableTimerSync
	DisableTimerSync
	CancelIsolate
	CancelSubflow
)

// SyncRequest sync maintenance requests between instances subscribed to FlowSync
type SyncRequest struct {
	Cmd    int
	Sender uuid.UUID
	ID     interface{}
}

func (s *WorkflowServer) startDatabaseListener() error {

	conninfo := s.config.Database.DB

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Error(err)
		}
	}

	minReconn := 10 * time.Second
	maxReconn := time.Minute
	listener := pq.NewListener(conninfo, minReconn, maxReconn, reportProblem)
	err := listener.Listen(FlowSync)
	if err != nil {
		return err
	}

	go func(l *pq.Listener) {

		defer l.UnlistenAll()

		for {

			notification, more := <-l.Notify
			if !more {
				log.Info("Database listener closed.")
				return
			}

			if notification == nil {
				continue
			}

			req := new(SyncRequest)
			err = json.Unmarshal([]byte(notification.Extra), req)
			if err != nil {
				log.Errorf("Unexpected notification on database listener: %v", err)
				continue
			}

			// only handle if not send by this server
			if s.id != req.Sender {
				log.Debugf("sync received: %v", req)

				switch req.Cmd {
				case AddTimerSync:
					s.tmManager.syncTimerAdd(int(req.ID.(float64)))
				case DeleteTimerSync:
					s.tmManager.syncTimerDelete(req.ID.(string))
				case EnableTimerSync:
					s.tmManager.syncTimerEnable(req.ID.(string))
				case DisableTimerSync:
					s.tmManager.syncTimerDisable(req.ID.(string))
				case CancelSubflow:
					s.engine.finishCancelSubflow(req.ID.(string))
				}

			}

		}

	}(listener)

	return nil

}

func syncServer(ctx context.Context, db *dbManager, sid *uuid.UUID, id interface{}, cmd int) error {

	var sr SyncRequest
	sr.Cmd = cmd

	if sid != nil {
		sr.Sender = *sid
	}

	sr.ID = id

	b, err := json.Marshal(sr)
	if err != nil {
		return err
	}

	conn, err := db.dbEnt.DB().Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, "SELECT pg_notify($1, $2)", FlowSync, string(b))
	if err, ok := err.(*pq.Error); ok {

		log.Debugf("db notification failed: %v", err)
		if err.Code == "57014" {
			return fmt.Errorf("canceled query")
		}

		return err

	}

	return err

}
