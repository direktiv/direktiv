package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
)

type Handler struct {
	db         *sql.DB
	args       *HandlerArgs
	queueMutex sync.Mutex
	logQueue   chan *log15.Record
	queuedLogs []log15.Record
	closed     bool
}

type HandlerArgs struct {
	Driver                      *sql.DB
	InsertFrequencyMilliSeconds int
	Namespace                   string
	InstanceID                  string
}

func NewHandler(args *HandlerArgs) (*Handler, error) {

	out := new(Handler)

	out.args = args
	out.db = args.Driver

	return out.init()
}

func (h *Handler) init() (*Handler, error) {
	h.queuedLogs = make([]log15.Record, 0)
	h.logQueue = make(chan *log15.Record, 10)

	err := func() error {
		tx, err := h.db.BeginTx(context.Background(), nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		_, err = tx.Exec(`create table if not exists logs (
			id serial primary key,
			namespace text,
			instance text,
			time int,
			lvl int,
			msg text,
			ctx bytea
		)`)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`create index if not exists "idx_log_namespace_instance" on logs (namespace, instance)`)
		if err != nil {
			return err
		}

		return tx.Commit()
	}()
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			// ensure logs are inserted in order
			r, more := <-h.logQueue

			h.queueMutex.Lock()

			if !more {
				h.closed = true
				h.queueMutex.Unlock()
				return
			}

			h.queuedLogs = append(h.queuedLogs, *r)

			h.queueMutex.Unlock()
		}
	}()

	go func() {

		var err error
		var vals []interface{}
		var rowValues []string

		for {
			time.Sleep(time.Millisecond * time.Duration(h.args.InsertFrequencyMilliSeconds))

			h.queueMutex.Lock()

			if len(h.queuedLogs) == 0 {
				goto nextIter
			}

			rowValues = make([]string, 0)
			vals = make([]interface{}, 0)
			for i, msg := range h.queuedLogs {

				ctxMap := make(map[string]interface{}, 0)
				for i, c := range msg.Ctx {
					if i%2 == 1 {
						ctxMap[fmt.Sprintf("%s", msg.Ctx[i-1])] = fmt.Sprintf("%v", c)
					}
				}

				b, err := json.Marshal(ctxMap)
				if err != nil {
					// TODO handle err
					fmt.Printf("(todo: improve this log!) %s", err.Error())
				}

				idx := i * 6
				rowValues = append(rowValues, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)\n", idx+1, idx+2, idx+3, idx+4, idx+5, idx+6))

				// sEnc := base64.StdEncoding.EncodeToString([]byte(msg.Msg))
				vals = append(vals, h.args.Namespace, h.args.InstanceID, msg.Time.Unix(), msg.Lvl, msg.Msg, fmt.Sprintf("%s", b))

			}

			_, err = h.db.Exec(fmt.Sprintf("insert into logs (namespace, instance, time, lvl, msg, ctx) values %s", strings.Join(rowValues, ", ")), vals...)
			if err != nil {
				// TODO handle err
				fmt.Printf("(todo: improve this log!) %s", err.Error())
			}

		nextIter:

			h.queuedLogs = h.queuedLogs[:0]

			if h.closed {
				return
			}

			h.queueMutex.Unlock()

		}
	}()

	return h, nil
}

func (h *Handler) Log(r *log15.Record) error {
	h.logQueue <- r
	return nil
}

func (h *Handler) Close() error {

	defer func() {
		_ = recover()
	}()

	close(h.logQueue)
	return nil

}
