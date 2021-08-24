package direktiv

import (
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

func (lc *logClient) logsForNamespace(ns string, offset, limit int32) ([]map[string]interface{}, error) {
	appLog.Infof("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")

	var retArray []map[string]interface{}

	// rows, err := lc.db.Query(`SELECT * FROM fluentbit WHERE data -> 'namespace' = '$1' LIMIT 10 OFFSET 0;`, ns)
	// SELECT * FROM fluentbit WHERE data->>'namespace' = 'jens'
	rows, err := lc.db.Query(`SELECT data FROM fluentbit WHERE data->>'namespace' = $1 ORDER BY time desc LIMIT $2 OFFSET $3;`,
		ns, limit, offset)
	if err != nil {
		appLog.Errorf("error querying namespace logs: %v", err)
		return retArray, err
	}

	var data string
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
