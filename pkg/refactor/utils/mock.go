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
	 CREATE TABLE IF NOT EXISTS "filesystem_roots"
			(
				"id" text,
				"created_at" datetime,
				"updated_at" datetime,
				PRIMARY KEY ("id"),
				CONSTRAINT "fk_namespaces_filesystem_roots"
				FOREIGN KEY ("id") REFERENCES "namespaces"("oid") ON DELETE CASCADE ON UPDATE CASCADE
				);
	 CREATE TABLE IF NOT EXISTS "filesystem_files"
			(
				"id" text,
				"path" text,
				"depth" integer,
				"typ" text,
				"root_id" text,
				"created_at" datetime,
				"updated_at" datetime,
				PRIMARY KEY ("id"),
				CONSTRAINT "fk_filesystem_roots_filesystem_files"
				FOREIGN KEY ("root_id") REFERENCES "filesystem_roots"("id") ON DELETE CASCADE ON UPDATE CASCADE
				);
	 CREATE TABLE IF NOT EXISTS "filesystem_revisions"
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
				CONSTRAINT "fk_filesystem_files_filesystem_revisions"
				FOREIGN KEY ("file_id") REFERENCES "filesystem_files"("id") ON DELETE CASCADE ON UPDATE CASCADE
				);
	 CREATE TABLE IF NOT EXISTS "file_annotations"
			(
				"file_id" text,
				"data" text,
				"created_at" datetime,
				"updated_at" datetime,
				PRIMARY KEY ("file_id"),
				CONSTRAINT "fk_filesystem_files_file_annotations"
				FOREIGN KEY ("file_id") REFERENCES "filesystem_files"("id") ON DELETE CASCADE ON UPDATE CASCADE
				);
	 CREATE TABLE IF NOT EXISTS "mirror_configs" 
	 		(
	 		    "namespace_id" text,
	 		    "url" text,
	 		    "git_ref" text,
	 		    "git_commit_hash" text,
	 		    "public_key" text,
	 		    "private_key" text,
	 		    "private_key_passphrase" text,
	 		    "created_at" datetime,
	 		    "updated_at" datetime,
	 		    PRIMARY KEY ("namespace_id"),
				CONSTRAINT "fk_namespaces_mirror_configs"
				FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("oid") ON DELETE CASCADE ON UPDATE CASCADE
	     );
	 CREATE TABLE IF NOT EXISTS "mirror_processes"
	 		(
	 		    "id" text,
	 		    "namespace_id" text,
	 		    "status" text,
	            "typ" text,
	 		    "ended_at" datetime,
	 		    "created_at" datetime,
	 		    "updated_at" datetime,
	 		    PRIMARY KEY ("id"),
	 		    CONSTRAINT "fk_namespaces_mirror_processes"
				FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("oid") ON DELETE CASCADE ON UPDATE CASCADE
	 		);
	CREATE TABLE IF NOT EXISTS "log_msgs" 
			 (
				 "oid" text,
				 "t" datetime,
				 "msg" text,
				 "level" text,
				 "root_instance_id" text,
				 "log_instance_call_path" text,
				 "tags" jsonb,
				 "workflow_id" text,
				 "mirror_activity_id" text,
				 "instance_logs" text,
				 "namespace_logs" text,
				 PRIMARY KEY ("oid")
			 );
`)

	if res.Error != nil {
		return nil, res.Error
	}

	return db, nil
}
