package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/vorteil/direktiv/pkg/dlog"
)

type Logger struct {
	db            *sql.DB
	brokerManager *dlog.BrokerManager
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

	err = l.initDB()
	if err != nil {
		return err
	}

	return nil
}

func (l *Logger) CloseConnection() error {
	return l.db.Close()
}

// Testing !!!!!!!!!!! //

func NewLogger(database string) (*Logger, error) {
	l := new(Logger)
	err := l.Connect(database)
	l.brokerManager = dlog.NewBrokerManager()
	return l, err
}

func (l *Logger) StreamLogs(ctx context.Context, instance string) (chan interface{}, error) {
	broker, ok := l.brokerManager.GetBroker(instance)
	if !ok {
		return nil, fmt.Errorf("instance '%s' not found", instance)
	}

	ch := broker.Subscribe()

	// Unsubscribe when context is done
	go func(ch chan interface{}) {
		<-ctx.Done()
		broker.Unsubscribe(ch)
	}(ch)

	return ch, nil
}

// Testing !!!!!!!!!!! //

type dbLogger struct {
	log15.Logger
	handler *Handler
}

func (dl *dbLogger) Close() error {
	return dl.handler.Close()
}

func (l *Logger) NamespaceLogger(namespace string) (dlog.Logger, error) {
	lg := new(dbLogger)
	lg.Logger = log15.New()

	h, err := NewHandler(&HandlerArgs{
		Driver:                      l.db,
		Namespace:                   namespace,
		InsertFrequencyMilliSeconds: 500,
	})
	if err != nil {
		return nil, err
	}

	lg.handler = h
	lg.SetHandler(h)
	return lg, nil
}

func (l *Logger) LoggerFunc(namespace, instance string) (dlog.Logger, error) {

	lg := new(dbLogger)
	lg.Logger = log15.New()

	h, err := NewHandler(&HandlerArgs{
		Driver:                      l.db,
		Namespace:                   namespace,
		InstanceID:                  instance,
		InsertFrequencyMilliSeconds: 250,
		BrokerManager:               l.brokerManager,
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
	WHERE (instance is null or instance = '') AND namespace = $1
	ORDER BY time ASC
	LIMIT $2 OFFSET $3`
	if strings.Contains(instance, "/") {
		sqlStatement = `SELECT msg, ctx, time, lvl FROM logs
		WHERE instance=$1
		ORDER BY time ASC
		LIMIT $2 OFFSET $3;`
	}

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

func (l *Logger) DeleteNamespaceLogs(namespace string) error {
	tx, err := l.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = l.db.Exec("DELETE FROM logs WHERE namespace = $1 AND (instance is null or instance = '')", namespace)
	if err != nil {
		return err
	}

	return tx.Commit()
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

	fmt.Printf("!!!!!!!!!!!!!!!!!!!@!@ DELETING INSTANCE ID instance = %s \n ", instance)
	l.brokerManager.DeleteBroker(instance)

	return tx.Commit()
}
