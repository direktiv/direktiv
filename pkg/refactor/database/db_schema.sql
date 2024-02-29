CREATE TABLE IF NOT EXISTS  "namespaces" (
    "id" uuid,
    "name" text NOT NULL UNIQUE,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id")
);

CREATE TABLE IF NOT EXISTS  "filesystem_roots" (
    "id" uuid,
    "namespace" text UNIQUE,

    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY ("id"),
    CONSTRAINT "fk_namespaces_filesystem_roots"
    FOREIGN KEY ("namespace") REFERENCES "namespaces"("name") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "filesystem_files" (
    "id" uuid,
    "root_id" uuid NOT NULL,
    "path" text NOT NULL,
    "depth" integer NOT NULL,
    "typ" text NOT NULL,

    "data" bytea,
    "checksum" text,

    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mime_type" text NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "file_path_no_dup_check" UNIQUE ("root_id","path"),
    CONSTRAINT "fk_filesystem_roots_filesystem_files"
    FOREIGN KEY ("root_id") REFERENCES "filesystem_roots"("id") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "mirror_configs" (
    "namespace" text,
    "url" text NOT NULL,
    "git_ref" text NOT NULL,
    "git_commit_hash" text,
    "public_key" text,
    "private_key" text,
    "private_key_passphrase" text,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "insecure" boolean NOT NULL,
    PRIMARY KEY ("namespace"),
    CONSTRAINT "fk_namespaces_mirror_configs"
    FOREIGN KEY ("namespace") REFERENCES "namespaces"("name") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "mirror_processes" (
    "id" uuid,
    "namespace" text NOT NULL,
    "status" text NOT NULL,
    "typ" 	 text NOT NULL,
    "ended_at" timestamptz,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id"),
    CONSTRAINT "fk_namespaces_mirror_processes"
    FOREIGN KEY ("namespace") REFERENCES "namespaces"("name") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "secrets" (
    "namespace" text NOT NULL,
    "name" text NOT NULL,
    "data" 	 bytea,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("namespace", "name"),
    CONSTRAINT "fk_namespaces_secrets"
    FOREIGN KEY ("namespace") REFERENCES "namespaces"("name") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "instances_v2" (
    "id" uuid,
    "namespace_id" uuid NOT NULL,
    "namespace" text NOT NULL,
    "root_instance_id" uuid NOT NULL,
    "server" uuid,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "ended_at" timestamptz,
    "deadline" timestamptz,
    "status" integer NOT NULL,
    "workflow_path" text NOT NULL,
    "error_code" text NOT NULL,
    "invoker" text NOT NULL,
    "definition" bytea NOT NULL,
    "settings" bytea NOT NULL,
    "descent_info" bytea NOT NULL,
    "telemetry_info" bytea NOT NULL,
    "runtime_info" bytea NOT NULL,
    "children_info" bytea NOT NULL,
    "input" bytea NOT NULL,
    "live_data" bytea NOT NULL,
    "state_memory" bytea NOT NULL,
    "output" bytea,
    "error_message" bytea,
    "metadata" bytea,
    PRIMARY KEY ("id"),
    CONSTRAINT "fk_namespaces_instances"
    FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("id") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "instance_messages" (
    "id" uuid NOT NULL,
    "instance_id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "payload" bytea NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "fk_instances_v2_instance_messages"
    FOREIGN KEY ("instance_id") REFERENCES "instances_v2"("id") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "runtime_variables" (
    "id" uuid,
    "namespace" text NOT NULL,

    "workflow_path" text,
    "instance_id" uuid,

    "name"  text NOT NULL,
    "mime_type"  text NOT NULL,
    "data"  bytea NOT NULL,

    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY ("id"),

    CONSTRAINT "fk_namespaces_runtime_variables"
    FOREIGN KEY ("namespace") REFERENCES "namespaces"("name") ON DELETE CASCADE ON UPDATE CASCADE,

    -- TODO: Find a way to clean up runtime vars for workflows when they get deleted.
    CONSTRAINT "fk_instances_v2_runtime_variables"
    FOREIGN KEY ("instance_id") REFERENCES "instances_v2"("id") ON DELETE CASCADE ON UPDATE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS "runtime_variables_unique" ON runtime_variables(namespace, COALESCE(workflow_path, instance_id::text), name);


CREATE TABLE IF NOT EXISTS "engine_messages" (
    "id" uuid,
    "timestamp" timestamptz NOT NULL,
    "topic" text,
    "source" uuid,
    "level" integer,
    "log_instance_call_path" text,
    "entry" bytea NOT NULL,
    PRIMARY KEY ("id")
);

-- partitioning the logtable to speeds up pagination and queries
CREATE INDEX IF NOT EXISTS "engine_messages_topic" ON "engine_messages" USING hash("topic");

CREATE TABLE IF NOT EXISTS "staging_events" (
    "id" uuid NOT NULL,
    "event_id" text,
    "source" text NOT NULL,
    "type" text NOT NULL,
    "cloudevent" text NOT NULL,
    "namespace_id" uuid NOT NULL,
    "namespace_name" text,
    "received_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "created_at" timestamptz NOT NULL,
    "delayed_until" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("id") ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT "no_dup_stag_check" UNIQUE ("source","event_id", "namespace_id"),
    PRIMARY KEY ("id")
);

CREATE TABLE IF NOT EXISTS "events_history" (
    "id" text,
    "type" text NOT NULL,
    "source" text NOT NULL,
    "cloudevent" text NOT NULL,
    "namespace_id" uuid NOT NULL,
    "received_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "created_at" timestamptz NOT NULL,
    FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("id") ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT "no_dup_check" UNIQUE ("source","id", "namespace_id")
);

-- for cursor style pagination
CREATE INDEX IF NOT EXISTS "events_history_sorted" ON "events_history" ("namespace_id", "created_at" DESC);

CREATE TABLE IF NOT EXISTS "event_listeners" (
    "id" uuid UNIQUE,
    "namespace_id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted" boolean NOT NULL,
    "received_events" text,
    "trigger_type" integer NOT NULL,
    "events_lifespan" integer NOT NULL DEFAULT 0,
    "glob_gates" text, 
    "event_types" text NOT NULL, -- lets keep it for the ui just in case
    "trigger_info" text NOT NULL,
    "metadata" text,
    PRIMARY KEY ("id"),
    FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("id") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "event_topics" (
    "id" uuid,
    "event_listener_id" uuid NOT NULL,
    "namespace_id" uuid NOT NULL,
    "topic" text NOT NULL,
    "filter" text,
    PRIMARY KEY ("id"),
    CONSTRAINT "no_dup_topics_check" UNIQUE ("event_listener_id", "topic", "filter"),
    FOREIGN KEY ("event_listener_id") REFERENCES "event_listeners"("id") ON DELETE CASCADE ON UPDATE CASCADE
);

-- for processing the events with minimal latency, we assume that the topic 
-- is a compound like this: "namespace-id:event-type"
CREATE INDEX IF NOT EXISTS "event_topic_bucket" ON "event_topics" USING hash("topic");

CREATE TABLE IF NOT EXISTS "metrics" (
    "id" serial,
    "namespace" text,
    "workflow" text,
    "instance" text,
    "state" text,
    "timestamp" timestamptz DEFAULT CURRENT_TIMESTAMP,
    "workflow_ms" integer,
    "isolate_ms" integer,
    "error_code" text,
    "invoker" text,
    "next" integer,
    "transition" text,
    PRIMARY KEY ("id")
);


-- Remove file_annotations.
DROP TABLE IF EXISTS "file_annotations";

-- Remove filesystem_revisions table and move its columns to filesystem_file table.
ALTER TABLE "instances_v2" DROP COLUMN IF EXISTS "revision_id";
ALTER TABLE "instances_v2" ADD COLUMN IF NOT EXISTS "server" uuid;
ALTER TABLE "metrics" DROP COLUMN IF EXISTS "revision";
DROP TABLE IF EXISTS "filesystem_revisions";
ALTER TABLE "filesystem_files" ADD COLUMN IF NOT EXISTS "data" bytea;
ALTER TABLE "filesystem_files" ADD COLUMN IF NOT EXISTS "checksum" text;
ALTER TABLE "event_topics" ADD COLUMN IF NOT EXISTS "filter" text;

ALTER TABLE "namespaces" DROP COLUMN IF EXISTS "config";
