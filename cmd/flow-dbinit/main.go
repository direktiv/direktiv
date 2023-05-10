package flow_dbinit

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Masterminds/semver"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func RunApplication() {
	log.Printf("checking database for schema updates...\n")

	// get db connection.
	conn := os.Getenv(util.DBConn)
	db, err := sql.Open("postgres", conn)
	if err != nil {
		log.Printf("open sql error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// check if database has been initialized.
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
		log.Printf("sql error: %v\n", err)
		os.Exit(1)
	}
	if !initialized {
		log.Printf("database hasn't been initialized by ent automigrate.\n")
		os.Exit(0)
	}

	// initialize generation table if not exists.
	qstr = `CREATE TABLE IF NOT EXISTS db_generation (generation VARCHAR)`
	_, err = db.Exec(qstr)
	if err != nil {
		log.Printf("create db_generation table error: %v\n", err)
		os.Exit(1)
	}

	// initialize upgrade transaction.
	tx, err := db.Begin()
	if err != nil {
		log.Printf("begin db transaction error: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		err := tx.Rollback()
		if !errors.Is(err, sql.ErrTxDone) {
			log.Printf("rollback db transaction error: %v\n", err)
		}
	}()

	row = tx.QueryRow(`SELECT generation FROM db_generation ORDER BY generation DESC LIMIT 1`)
	var genString string
	err = row.Scan(&genString)
	if errors.Is(err, sql.ErrNoRows) {
		genString = "0.5.10"
	} else if err != nil {
		log.Printf("selecting from db_generation error: %v\n", err)
		os.Exit(1)
	}

	dbGeneration, err := semver.NewVersion(genString)
	if err != nil {
		log.Printf("parsing generation from db error: %v\n", err)
		os.Exit(1)
	}
	log.Printf("current database generation: %v\n", dbGeneration)

	upgraders := []generationUpgrader{
		{
			version: semver.MustParse("0.6.0"),
			logic:   updateGeneration_0_6_0,
		},
		{
			version: semver.MustParse("0.7.1"),
			logic:   updateGeneration_0_7_1,
		},
		{
			version: semver.MustParse("0.7.3"),
			logic:   updateGeneration_0_7_3,
		},
		{
			version: semver.MustParse("0.7.5"),
			logic:   updateGeneration_0_7_5,
		},
	}

	for _, upgrader := range upgraders {
		// check if version needs upgrading
		if !upgrader.version.GreaterThan(dbGeneration) {
			continue
		}
		log.Printf("updating to generation %s\n", upgrader.version)
		err = upgrader.logic(tx)
		if err != nil {
			log.Printf("running upgrader version: %s, error: %v\n", upgrader.version, err)
			os.Exit(1)
		}
		_, err = db.Exec(fmt.Sprintf(`INSERT INTO db_generation(generation) VALUES('%s')`, upgrader.version))
		if err != nil {
			log.Printf("inserting in db_generation error: %v\n", err)
			os.Exit(1)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("committing db transaction error: %v\n", err)
		os.Exit(1)
	}
}

type generationUpgrader struct {
	version *semver.Version
	logic   func(tx *sql.Tx) error
}

func updateGeneration_0_7_5(db *sql.Tx) error {
	queries := []string{}

	for k, v := range map[string][]string{
		"annotations": {"annotations_workflows_annotations", "annotations_inodes_annotations"},
		"instances":   {"instances_workflow_instances", "instances_revisions_instances"},
		"events":      {"events_workflows_wfevents"},
		"var_refs":    {"var_refs_workflows_vars"},
	} {
		for i := range v {
			queries = append(queries, fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;", k, v[i]))
		}
	}

	queries = append(queries, "DROP TABLE log_msgs;")

	for i := range queries {
		_, err := db.Exec(queries[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func updateGeneration_0_7_3(db *sql.Tx) error {
	// old is id, name, data, new one has namespace
	_, err := db.Exec("DROP INDEX services_name_key")
	return err
}

func updateGeneration_0_7_1(db *sql.Tx) error {
	// old is id, name, data, new one has namespace
	_, err := db.Exec("drop table services")
	return err
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
