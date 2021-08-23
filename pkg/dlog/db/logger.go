package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"
	"github.com/vorteil/direktiv/pkg/api"
	"github.com/vorteil/direktiv/pkg/dlog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Logger struct {
	db      *sql.DB
	router  *mux.Router
	server  *http.Server
	brokers map[string]*Broker
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

type logEntry struct {
	Message   interface{}       `json:"message"`
	Level     interface{}       `json:"level"`
	Timestamp interface{}       `json:"timestamp"`
	Ctx       map[string]string `json:"context"`
}

func (b *Broker) Publish(msg, level interface{}, timestamp int64, ctx map[string]string) error {
	data, err := json.Marshal(logEntry{
		Message:   msg,
		Level:     level,
		Timestamp: timestamppb.New(time.Unix(0, timestamp)),
		Ctx:       ctx,
	})

	if err != nil {
		return err
	}

	b.publishCh <- data
	return nil
}

func NewLogger(database string) (*Logger, error) {
	l := new(Logger)
	err := l.Connect(database)
	l.router = mux.NewRouter()
	l.router.HandleFunc("/logging/{namespace}/{workflowTarget}/{id}", l.dispatchLogs).Methods(http.MethodGet, http.MethodHead)
	l.server = &http.Server{
		Handler:      l.router,
		Addr:         ":7979",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	l.brokers = make(map[string]*Broker)

	go l.server.ListenAndServe()
	return l, err
}

func (l *Logger) dispatchLogs(w http.ResponseWriter, r *http.Request) {
	ns := mux.Vars(r)["namespace"]
	wf := mux.Vars(r)["workflowTarget"]
	id := mux.Vars(r)["id"]
	instance := fmt.Sprintf("%s/%s/%s", ns, wf, id)

	broker, ok := l.brokers[instance]
	if !ok {
		// FIXME: ERROR NOT WORKING!!!@!@
		err := fmt.Errorf("instance '%s' not found", instance)
		fmt.Printf("dispatch logs api failed: %s\n", err.Error())
		api.ErrHEADResponse(w, err)
		return
	}

	previousLogs, err := l.QueryLogs(context.Background(), instance, 10000, 0)
	if err != nil {
		// TODO
		api.ErrHEADResponse(w, err)
	}

	flusher, err := api.SetupSEEWriter(w)
	if err != nil {
		api.ErrHEADResponse(w, err)
		return
	}

	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}

	msgCh := broker.Subscribe()
	defer func() {
		// Make sure broker hasnt been deleted
		if b, ok := l.brokers[instance]; ok {
			b.Unsubscribe(msgCh)
		}
	}()

	// Send Previous logs
	for _, pLog := range previousLogs.Logs {
		fmt.Printf("pLog = %s\n", pLog.Message)
		pData, err := json.Marshal(logEntry{
			Message:   pLog.Message,
			Level:     pLog.Context,
			Timestamp: timestamppb.New(time.Unix(0, pLog.Timestamp)),
			Ctx:       pLog.Context,
		})
		if err != nil {
			api.ErrSSEResponse(w, flusher, fmt.Errorf("could not process previous log: %w", err))
			continue
		}

		_, err = w.Write([]byte(fmt.Sprintf("data: %s\n\n", pData)))
		if err != nil {
			api.ErrSSEResponse(w, flusher, fmt.Errorf("could not write previous log: %w", err))
		}
	}
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case msgData := <-msgCh:
			// fmt.Printf("WRITING DATA !!!  = %s\n", fmt.Sprintf("data: %s\n\n", msgData))
			_, err := w.Write([]byte(fmt.Sprintf("data: %s\n\n", msgData)))
			if err != nil {
				api.ErrSSEResponse(w, flusher, fmt.Errorf("could not write log: %w", err))
				return
			}

			flusher.Flush()
		}
	}
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

	// Create Broker if it doest exist
	if _, ok := l.brokers[instance]; !ok {
		l.brokers[instance] = NewBroker()
		go l.brokers[instance].Start()
	}

	h, err := NewHandler(&HandlerArgs{
		Driver:                      l.db,
		Namespace:                   namespace,
		InstanceID:                  instance,
		InsertFrequencyMilliSeconds: 250,
		Broker:                      l.brokers[instance],
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
	l.brokers[instance].Stop()
	delete(l.brokers, instance)

	return tx.Commit()
}
