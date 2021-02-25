package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/inconshreveable/log15"
	"github.com/vorteil/direktiv/pkg/dlog"
)

type Logger struct {
	db *sql.DB
}

func (l *Logger) Connect(database string) error {
	if l == nil {
		l = new(Logger)
	}

	var err error

	l.db, err = sql.Open("postgres", database)
	if err != nil {
		return fmt.Errorf("Failed to initialize server: %w", err)
	}

	return nil
}

func (l *Logger) CloseConnection() error {
	return l.db.Close()
}

func NewLogger(database string) (*Logger, error) {
	l := new(Logger)
	err := l.Connect(database)
	return l, err
}

type dbLogger struct {
	log15.Logger
	handler *Handler
}

func (dl *dbLogger) Close() error {
	return dl.handler.Close()
}

func (l *Logger) LoggerFunc(namespace, instance string) (dlog.Logger, error) {

	lg := new(dbLogger)
	lg.Logger = log15.New()

	h, err := NewHandler(&HandlerArgs{
		Driver:                      l.db,
		Namespace:                   namespace,
		InstanceID:                  instance,
		InsertFrequencyMilliSeconds: 500,
	})
	if err != nil {
		return nil, err
	}

	lg.handler = h
	lg.SetHandler(h)

	return lg, nil

}

func (l *Logger) QueryLogs(ctx context.Context, instance string, limit, offset int) (dlog.QueryReponse, error) {
	testLOG := dlog.QueryReponse{
		Limit:  limit,
		Offset: offset,
		// Data:   make([]map[string]interface{}, 0),
	}

	var Msg string
	var Ctx string
	var Lvl int
	var Time int64
	var err error

	sqlStatement := `SELECT msg, ctx, time, lvl FROM logs
	WHERE instance=$1
	LIMIT $2 OFFSET $3;`
	rows, err := l.db.Query(sqlStatement, instance, limit, offset)
	if err != nil {
		return testLOG, err
	}

	for rows.Next() {
		ctxMap := make(map[string]string)
		// dataMap := make(map[string]interface{})

		err = rows.Scan(&Msg, &Ctx, &Time, &Lvl)
		if err != nil {
			break
		}

		err := json.Unmarshal([]byte(Ctx), &ctxMap)
		if err != nil {
			break
		}

		// msg, _ := base64.StdEncoding.DecodeString(Msg)
		// dataMap["msg"] = Msg
		// dataMap["lvl"] = Lvl
		// dataMap["time"] = Time
		// dataMap["ctx"] = ctxMap

		testLOG.Logs = append(testLOG.Logs, dlog.LogEntry{
			// TODO: Level: ,
			Message:   Msg,
			Timestamp: Time,
			Context:   ctxMap,
		})
		// testLOG.Data = append(testLOG.Data, dataMap)
	}

	if err == nil {
		err = rows.Err()
	}

	testLOG.Count = len(testLOG.Logs)
	return testLOG, err
}

func (l *Logger) DeleteInstanceLogs(instance string) error {

	tx, err := l.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = l.db.Exec("DELETE FROM logs WHERE instance = $1", instance)
	if err != nil {
		return err
	}

	return tx.Commit()
}
