package flow_dbinit

import (
	"log"

	_ "github.com/lib/pq"
)

func RunApplication() {
	// TODO: rethink a db migration technique.
	log.Printf("db migration arrangements.\n")
}
