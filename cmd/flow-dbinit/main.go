package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/direktiv/direktiv/pkg/util"
	_ "github.com/lib/pq"
)

var generations map[int]func(*sql.DB) error

func main() {

	log.Printf("checking database for schema updates")

	// get db connection
	conn := os.Getenv(util.DBConn)
	db, err := sql.Open("postgres", conn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// TODO: fetch generation
	// if no generation detected, we assume it is < 0.6.0
	generations := make(map[int]func(*sql.DB) error)
	generations[0] = updateGeneration0

	startingGeneration := 0

	for {
		genFunc, ok := generations[startingGeneration]
		if !ok {
			break
		}
		log.Printf("updating to generation %d\n", startingGeneration)
		err := genFunc(db)
		if err != nil {
			log.Printf("error updating to generation %d: %v\n", startingGeneration, err)
			panic(err)
		}
		log.Printf("updating to generation %d finished\n", startingGeneration)
		startingGeneration++
	}

}

func updateGeneration0(db *sql.DB) error {

	sqls := []string{
		fmt.Sprintf("ALTER TABLE refs ADD COLUMN created_at timestamp NOT NULL DEFAULT '%v';", time.Now().UTC().Format("2006-01-02T15:04:05-0700")),
		"ALTER TABLE revisions ADD COLUMN metadata jsonb NOT NULL DEFAULT '{ \"hello\": \"world\"}'",
		"delete from metrics;", // we can not save metrics
	}

	for i := range sqls {
		sql := sqls[i]
		_, err := db.Exec(sql)
		if err != nil {
			log.Printf("error running sql: %v", err)
		}
	}

	return nil

}
