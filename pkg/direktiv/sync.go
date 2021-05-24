package direktiv

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	hash "github.com/mitchellh/hashstructure/v2"
	log "github.com/sirupsen/logrus"
)

// FlowSync is the name of postgres pubsub channel
const FlowSync = "flowsync"

// direktiv pub/sub items
const (
	CancelIsolate = iota
	CancelSubflow
	CancelTimer
	CancelInstanceTimers
	AddCron
)

// SyncRequest sync maintenance requests between instances subscribed to FlowSync
type SyncRequest struct {
	Cmd    int
	Sender uuid.UUID
	ID     interface{}
}

func syncAPIWait(dbConnString string, channel string, w chan bool) error {

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Error(err)
		}
	}

	listener := pq.NewListener(dbConnString, 10*time.Second,
		time.Minute, reportProblem)
	err := listener.Listen(channel)
	if err != nil {
		return err
	}

	w <- true

	defer listener.UnlistenAll()

	for {

		notification, more := <-listener.Notify
		if !more {
			log.Errorf("database listener closed")
			return fmt.Errorf("database listener closed")
		}

		if notification == nil {
			continue
		}

		w <- true

		return nil

	}

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

	err = listener.Listen(fmt.Sprintf("hostname:%s", s.hostname))
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

			if notification.Channel == FlowSync {
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
					case CancelSubflow:
						s.engine.finishCancelSubflow(req.ID.(string))
					case CancelTimer:
						s.tmManager.deleteTimerByName(s.hostname, s.hostname, req.ID.(string))
					case CancelInstanceTimers:
						s.tmManager.deleteTimersForInstanceNoBroadcast(req.ID.(string))
					case AddCron:
						m, ok := req.ID.(map[string]interface{})
						if ok {
							var name, fn, pattern string
							var data []byte
							if x, exists := m["name"]; exists {
								if str, ok := x.(string); ok {
									name = str
								}
							}
							if x, exists := m["fn"]; exists {
								if str, ok := x.(string); ok {
									fn = str
								}
							}
							if x, exists := m["pattern"]; exists {
								if str, ok := x.(string); ok {
									pattern = str
								}
							}
							if x, exists := m["data"]; exists {
								if b, ok := x.([]byte); ok {
									data = b
								}
							}
							err = s.tmManager.addCronNoBroadcast(name, fn, pattern, data)
							if err != nil {
								log.Error(err)
							}
						}
					}

				}
			} else {
				m := make(map[string]interface{})
				err = json.Unmarshal([]byte(notification.Extra), &m)
				if err != nil {
					log.Errorf("Unexpected notification on database listener: %v", err)
					continue
				}

				timerId, _ := m["timerId"]
				str, _ := timerId.(string)
				if str == "" {
					log.Errorf("Unexpected notification on database listener: %v", m)
					continue
				}

				err = s.tmManager.deleteTimerByName(s.hostname, s.hostname, str)
				if err != nil {
					log.Error(err)
					continue
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

func publishToHostname(db *dbManager, hostname string, req interface{}) error {

	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	conn, err := db.dbEnt.DB().Conn(db.ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	channel := fmt.Sprintf("hostname:%s", hostname)

	_, err = conn.ExecContext(db.ctx, "SELECT pg_notify($1, $2)", channel, string(b))
	if err, ok := err.(*pq.Error); ok {

		log.Debugf("db notification failed: %v", err)
		if err.Code == "57014" {
			return fmt.Errorf("canceled query")
		}

		return err

	}

	return err

}

func publishToAPI(db *dbManager, id string) error {

	conn, err := db.dbEnt.DB().Conn(db.ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	h, _ := hash.Hash(fmt.Sprintf("%s", id), hash.FormatV2, nil)
	channel := fmt.Sprintf("api:%d", h)

	_, err = conn.ExecContext(db.ctx, "SELECT pg_notify($1, $2)", channel, id)
	if err, ok := err.(*pq.Error); ok {

		log.Debugf("db notification failed: %v", err)
		if err.Code == "57014" {
			return fmt.Errorf("canceled query")
		}

		return err

	}

	return err

}
