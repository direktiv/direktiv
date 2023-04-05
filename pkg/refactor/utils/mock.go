package utils

import (
	"log"
	"os"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewMockGorm() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold: time.Second,   // Slow SQL threshold
				LogLevel:      logger.Silent, // Log level
			},
		),
	})
	if err != nil {
		return nil, err
	}

	res := db.Exec(`
	 CREATE TABLE IF NOT EXISTS "roots"
			(
				"id" text,
				"created_at" datetime,
				"updated_at" datetime,
				PRIMARY KEY ("id")
				);
	 CREATE TABLE IF NOT EXISTS "files"
			(
				"id" text,
				"path" text,
				"depth" integer,
				"typ" text,
				"root_id" text,
				"created_at" datetime,
				"updated_at" datetime,
				PRIMARY KEY ("id"),
				CONSTRAINT "fk_roots_files"
				FOREIGN KEY ("root_id") REFERENCES "roots"("id") ON DELETE CASCADE ON UPDATE CASCADE
				);
	 CREATE TABLE IF NOT EXISTS "revisions"
			(
				"id" text,
				"tags" text,
				"is_current" numeric,
				"data" blob,
				"checksum" text,
				"file_id" text,
				"created_at" datetime,
				"updated_at" datetime,
				PRIMARY KEY ("id"),
				CONSTRAINT "fk_files_revisions"
				FOREIGN KEY ("file_id") REFERENCES "files"("id") ON DELETE CASCADE ON UPDATE CASCADE
				);
	 CREATE TABLE IF NOT EXISTS "file_annotations"
			(
				"file_id" text,
				"data" text,
				"created_at" datetime,
				"updated_at" datetime,
				PRIMARY KEY ("file_id"),
				CONSTRAINT "fk_files_file_annotations"
				FOREIGN KEY ("file_id") REFERENCES "files"("id") ON DELETE CASCADE ON UPDATE CASCADE
				);
	 CREATE TABLE IF NOT EXISTS "mirror_configs" 
	 		(
	 		    "id" text,
	 		    "url" text,
	 		    "git_ref" text,
	 		    "git_commit_hash" text,
	 		    "public_key" text,
	 		    "private_key" text,
	 		    "private_key_passphrase" text,
	 		    "created_at" datetime,
	 		    "updated_at" datetime,
	 		    PRIMARY KEY ("id")
	     );
	 CREATE TABLE IF NOT EXISTS "mirror_processes" 
	 		(
	 		    "id" text,
	 		    "config_id" text,
	 		    "status" text,
	 		    "ended_at" datetime,
	 		    "created_at" datetime,
	 		    "updated_at" datetime,
	 		    PRIMARY KEY ("id"),
	 		    CONSTRAINT "fk_mirror_configs_mirror_processes"
				FOREIGN KEY ("config_id") REFERENCES "files"("id") ON DELETE CASCADE ON UPDATE CASCADE
	 		)
`)

	if res.Error != nil {
		return nil, err
	}

	return db, nil
}
