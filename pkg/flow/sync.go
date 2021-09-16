package flow

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

const CancelActionMessage = "cancelAction"

// SyncSubscribeTo subscribes to direktiv interna postgres pub/sub
func SyncSubscribeTo(log *zap.Logger, dbConnString, topic string, fn func(interface{})) error {

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Error(err.Error())
		}
	}

	listener := pq.NewListener(dbConnString, 10*time.Second,
		time.Minute, reportProblem)
	err := listener.Listen(flowSync)
	if err != nil {
		return err
	}

	go func(l *pq.Listener) {

		defer func() {
			l.UnlistenAll()
			l.Close()
		}()

		for {

			notification, more := <-l.Notify
			if !more {
				log.Info("Database listener closed.")
				return
			}

			if notification == nil {
				continue
			}

			req := new(PubsubUpdate)
			err = json.Unmarshal([]byte(notification.Extra), req)
			if err != nil {
				log.Error(fmt.Sprintf("Unexpected notification on database listener: %v", err))
				continue
			}

			if req.Handler == topic {
				fn(req)
			}

		}

	}(listener)

	return nil

}
