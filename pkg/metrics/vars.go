package metrics

import (
	"os"
)

var (
	dbHost     = "127.0.0.1"
	dbPort     = "5432"
	dbUser     = "sisatech"
	dbPass     = "sisatech"
	dbSSLMode  = "disable"
	dbDatabase = "metrics"
)

const (
	DB_HOST     = "METRICS_DB_HOST"
	DB_PORT     = "METRICS_DB_PORT"
	DB_USER     = "METRICS_DB_USER"
	DB_PASS     = "METRICS_DB_PASS"
	DB_SSLMODE  = "METRICS_DB_SSLMODE"
	DB_DATABASE = "METRICS_DB_DATABASE"
)

func init() {

	varMap := make(map[string]*string)
	varMap[DB_HOST] = &dbHost
	varMap[DB_PORT] = &dbPort
	varMap[DB_USER] = &dbUser
	varMap[DB_PASS] = &dbPass
	varMap[DB_SSLMODE] = &dbSSLMode
	varMap[DB_DATABASE] = &dbDatabase

	for k, v := range varMap {
		setVar(k, v)
	}

}

func setVar(key string, obj *string) {
	// set obj to value of env var specified by key
	x := os.Getenv(key)
	if x != "" {
		obj = &x
	}
}
