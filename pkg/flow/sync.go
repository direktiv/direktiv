package flow

import (
	"encoding/json"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
)

const CancelActionMessage = "cancelAction"

// SyncSubscribeTo subscribes to direktiv interna postgres pub/sub
func SyncSubscribeTo(log *zap.Logger, addr, topic string, fn func(interface{})) error {

	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", addr)
		},
	}

	conn := pool.Get()

	_, err := conn.Do("PING")
	if err != nil {
		return fmt.Errorf("can't connect to redis, got error:\n%v", err)
	}

	go func() {

		rc := pool.Get()

		psc := redis.PubSubConn{Conn: rc}
		if err := psc.PSubscribe(topic); err != nil {
			log.Error(err.Error())
		}

		for {
			switch v := psc.Receive().(type) {
			default:
				data, _ := json.Marshal(v)
				log.Debug(string(data))
			case redis.Message:
				req := new(PubsubUpdate)
				err = json.Unmarshal(v.Data, req)
				if err != nil {
					log.Error(fmt.Sprintf("Unexpected notification on database listener: %v", err))
				} else {
					fn(req)
				}
			}
		}

	}()

	// reportProblem := func(ev pq.ListenerEventType, err error) {
	// 	if err != nil {
	// 		log.Error(err.Error())
	// 	}
	// }

	/*
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

	*/

	return nil

}
