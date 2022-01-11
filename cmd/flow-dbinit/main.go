package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Masterminds/semver"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var generations map[int]func(*sql.DB) error

func main() {

	log.Printf("Checking database for schema updates...")

	// get db connection
	conn := os.Getenv(util.DBConn)
	db, err := sql.Open("postgres", conn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// check database has been initialized
	qstr := `SELECT EXISTS (
		SELECT FROM pg_catalog.pg_class c
		JOIN   pg_catalog.pg_namespace n ON n.oid = c.relnamespace
		WHERE  n.nspname = 'public'
		AND    c.relname = 'workflows'
		AND    c.relkind = 'r'    -- only tables
	);`
	row := db.QueryRow(qstr)
	var initialized bool
	err = row.Scan(&initialized)
	if err != nil {
		log.Printf("error running sql: %v", err)
		os.Exit(1)
	}

	if !initialized {
		log.Printf("Database hasn't been initialized. Aborting.")
		return
	}

	// initialize generation table if not exists
	qstr = `CREATE TABLE IF NOT EXISTS db_generation (
		generation VARCHAR
	)`
	_, err = db.Exec(qstr)
	if err != nil {
		log.Printf("error running sql: %v", err)
		os.Exit(1)
	}

	// initialize upgrade transaction
	tx, err := db.Begin()
	if err != nil {
		log.Printf("error running sql: %v", err)
		os.Exit(1)
	}
	defer tx.Rollback()

	row = tx.QueryRow(`SELECT generation FROM db_generation`)
	var gen string
	err = row.Scan(&gen)
	if err != nil {
		if err == sql.ErrNoRows {
			gen = "0.5.10"
		} else {
			log.Printf("error running sql: %v", err)
			os.Exit(1)
		}
	}

	log.Printf("Current database generation: %v", gen)

	// perform upgrades
	upgraders := make([]generationUpgrader, 0)

	upgraders = append(upgraders, generationUpgrader{
		version: "0.6.0",
		logic:   updateGeneration_0_6_0,
	})

	for _, upgrader := range upgraders {

		// check if version needs upgrading
		v1, err := semver.NewVersion(gen)
		if err != nil {
			log.Printf("error parsing generation: %v", err)
			os.Exit(1)
		}

		v2, err := semver.NewVersion(upgrader.version)
		if err != nil {
			panic(err)
		}

		if !v2.GreaterThan(v1) {
			continue
		}

		// upgrade

		log.Printf("Updating to generation %s\n", upgrader.version)

		err = upgrader.logic(tx)
		if err != nil {
			log.Printf("error running sql: %v", err)
			os.Exit(1)
		}

		_, err = db.Exec(`DELETE FROM db_generation`)
		if err != nil {
			log.Printf("error running sql: %v", err)
			os.Exit(1)
		}

		_, err = db.Exec(fmt.Sprintf(`INSERT INTO db_generation(generation) VALUES('%s')`, upgrader.version))
		if err != nil {
			log.Printf("error running sql: %v", err)
			os.Exit(1)
		}

		log.Printf("Updating to generation %s finished\n", upgrader.version)

	}

	err = tx.Commit()
	if err != nil {
		log.Printf("error running sql: %v", err)
		os.Exit(1)
	}

}

type generationUpgrader struct {
	version string
	logic   func(tx *sql.Tx) error
}

func updateGeneration_0_6_0(db *sql.Tx) error {

	sqls := []string{
		fmt.Sprintf("ALTER TABLE refs ADD COLUMN created_at timestamp NOT NULL DEFAULT '%v';", time.Now().UTC().Format("2006-01-02T15:04:05-0700")),
		fmt.Sprintf("ALTER TABLE events ADD COLUMN created_at timestamp NOT NULL DEFAULT '%v';", time.Now().UTC().Format("2006-01-02T15:04:05-0700")),
		fmt.Sprintf("ALTER TABLE events ADD COLUMN updated_at timestamp NOT NULL DEFAULT '%v';", time.Now().UTC().Format("2006-01-02T15:04:05-0700")),
		"ALTER TABLE events ADD COLUMN namespace_namespacelisteners uuid;",
		"ALTER TABLE revisions ADD COLUMN metadata jsonb NOT NULL DEFAULT '{ \"hello\": \"world\"}'",
		"delete from metrics;", // we can not save metrics
	}

	for i := range sqls {
		sql := sqls[i]
		_, err := db.Exec(sql)
		if err != nil {
			return err
		}
	}

	rows, err := db.Query(`SELECT events.oid, workflows.namespace_workflows FROM events INNER JOIN workflows ON workflows.oid = events.workflow_wfevents WHERE events.namespace_namespacelisteners IS NULL`)
	if err != nil {
		if err != nil {
			return err
		}
		return nil
	}
	defer rows.Close()

	for rows.Next() {

		var oid, id uuid.UUID
		err = rows.Scan(&oid, &id)
		if err != nil {
			return err
		}

		_, err = db.Exec(fmt.Sprintf(`UPDATE events SET namespace_namespacelisteners = '%s' WHERE oid = '%s'`, id.String(), oid.String()))
		if err != nil {
			return err
		}

	}

	return nil

}
