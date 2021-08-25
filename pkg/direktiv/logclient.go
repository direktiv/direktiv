package direktiv

import (
	"time"

	"database/sql"
	"encoding/json"
)

type logClient struct {
	db *sql.DB
}

func newLogDBClient() *logClient {
	return &logClient{}
}

func (lc *logClient) name() string {
	return "logclient"
}

func runLogQuery(rows *sql.Rows) ([]map[string]interface{}, error) {

	var (
		data     string
		retArray []map[string]interface{}
	)

	for rows.Next() {
		err := rows.Scan(&data)
		if err != nil {
			appLog.Errorf("error scanning namespace log results: %v", err)
			return retArray, err
		}
		result := make(map[string]interface{})
		err = json.Unmarshal([]byte(data), &result)
		if err != nil {
			appLog.Errorf("error unmarshaling namespace logs: %v", err)
			return retArray, err
		}
		retArray = append(retArray, result)
	}

	return retArray, nil
}

func (lc *logClient) deleteInstanceLogs(id string) error {

	_, err := lc.db.Exec("delete from fluentbit where data->>'instance' = $1", id)

	return err

}

func (lc *logClient) deleteNamespaceLogs() error {

	delTime := time.Now().Add(time.Duration(-10) * time.Minute)

	_, err := lc.db.Exec("delete from fluentbit where data->>'instance' is null and time < $1",
		delTime.Format("2006-01-02 15:04:05"))

	return err

}

func (lc *logClient) logsForInstance(id string, offset, limit int32) ([]map[string]interface{}, error) {

	rows, err := lc.db.Query(`SELECT data FROM fluentbit WHERE
    data->>'instance' = $1 ORDER BY time asc LIMIT $2 OFFSET $3;`,
		id, limit, offset)
	if err != nil {
		appLog.Errorf("error querying namespace logs: %v", err)
		return nil, err
	}

	return runLogQuery(rows)

}

func (lc *logClient) logsForNamespace(ns string, offset, limit int32) ([]map[string]interface{}, error) {

	rows, err := lc.db.Query(`SELECT data FROM fluentbit WHERE data->>'namespace' = $1
    and data->>'instance' is null ORDER BY time desc LIMIT $2 OFFSET $3;`,
		ns, limit, offset)
	if err != nil {
		appLog.Errorf("error querying namespace logs: %v", err)
		return nil, err
	}

	return runLogQuery(rows)

}

func (lc *logClient) stop() {
	lc.db.Close()
}

func (lc *logClient) start(s *WorkflowServer) error {

	db, err := sql.Open("postgres", s.config.Database.DB)
	if err != nil {
		return err
	}

	lc.db = db

	appLog.Debug("ping database")
	err = lc.db.Ping()
	if err != nil {
		return err
	}

	return nil
}
