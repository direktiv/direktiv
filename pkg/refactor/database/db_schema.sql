CREATE TABLE IF NOT EXISTS  "namespaces" (
    "id" uuid,
    "name" text NOT NULL UNIQUE,
    "config" text NOT NULL,
    "roots_info" text NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id")
);

CREATE TABLE IF NOT EXISTS  "filesystem_roots" (
    "id" uuid,
    "namespace_id" uuid,
    "name" text NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id"),
    CONSTRAINT "root_no_dup_check" UNIQUE ("namespace_id","name"),
    CONSTRAINT "fk_namespaces_filesystem_roots"
    FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("id") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "filesystem_files" (
    "id" uuid,
    "root_id" uuid NOT NULL,
    "path" text NOT NULL,
    "depth" integer NOT NULL,
    "typ" text NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "api_id" text NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "file_path_no_dup_check" UNIQUE ("root_id","path"),
    CONSTRAINT "file_api_id_no_dup_check" UNIQUE ("root_id","api_id"),
    CONSTRAINT "fk_filesystem_roots_filesystem_files"
    FOREIGN KEY ("root_id") REFERENCES "filesystem_roots"("id") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "filesystem_revisions" (
    "id" uuid,
    "tags" text,
    "is_current" boolean NOT NULL,
    "data" bytea NOT NULL,
    "checksum" text NOT NULL,
    "file_id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id"),
    CONSTRAINT "fk_filesystem_files_filesystem_revisions"
    FOREIGN KEY ("file_id") REFERENCES "filesystem_files"("id") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "file_annotations" (
    "file_id" uuid,
    "data" text,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("file_id"),
    CONSTRAINT "fk_filesystem_files_file_annotations"
    FOREIGN KEY ("file_id") REFERENCES "filesystem_files"("id") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "mirror_configs" (
    "namespace_id" uuid,
    "root_name" text NOT NULL,
    "url" text NOT NULL,
    "git_ref" text NOT NULL,
    "git_commit_hash" text,
    "public_key" text,
    "private_key" text,
    "private_key_passphrase" text,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("namespace_id"),
    CONSTRAINT "fk_namespaces_mirror_configs"
    FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("id") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "mirror_processes" (
    "id" uuid,
    "namespace_id" uuid NOT NULL,
    "root_id" uuid NOT NULL,
    "status" text NOT NULL,
    "typ" 	 text NOT NULL,
    "ended_at" timestamptz,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id"),
    CONSTRAINT "fk_namespaces_mirror_processes"
    FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("id") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "secrets" (
    "id" uuid,
    "namespace_id" uuid NOT NULL,
    "name" text NOT NULL,
    "data" 	 text,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id"),
    CONSTRAINT "fk_namespaces_secrets"
    FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("id") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "services" (
    "id" uuid,
    "namespace_id" uuid NOT NULL,
    "name" text NOT NULL,
    "url" text NOT NULL,
    "data" 	 text NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id"),
    CONSTRAINT "fk_namespaces_services"
    FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("id") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "instances_v2" (
    "id" uuid,
    "namespace_id" uuid NOT NULL,
    "revision_id" uuid NOT NULL,
    "root_instance_id" uuid NOT NULL,
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

CREATE TABLE IF NOT EXISTS "runtime_variables" (
    "id" uuid,
    "namespace_id" uuid,
    "workflow_path" text,
    "instance_id" uuid,

    "name"  text NOT NULL,
    "mime_type"  text NOT NULL,
    "data"  bytea NOT NULL,

    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY ("id"),
    UNIQUE(namespace_id, workflow_path, instance_id, name),

    CONSTRAINT "fk_namespaces_runtime_variables"
    FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("id") ON DELETE CASCADE ON UPDATE CASCADE,

    -- TODO: Find a way to clean up runtime vars for workflows when they get deleted.
    CONSTRAINT "fk_instances_v2_runtime_variables"
    FOREIGN KEY ("instance_id") REFERENCES "instances_v2"("id") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "log_entries" (
    "id" uuid,
    "timestamp" timestamptz NOT NULL,
    "level" integer,
    "root_instance_id" uuid,
    "source" uuid,
    "type" text,
    "log_instance_call_path" text,
    "entry" bytea NOT NULL,
    PRIMARY KEY ("id")
);

-- speeds up pagination
CREATE INDEX  IF NOT EXISTS "log_entries_sorted" ON log_entries ("timestamp" ASC);

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
    PRIMARY KEY ("id"),
    CONSTRAINT "no_dup_topics_check" UNIQUE ("event_listener_id","topic"),
    FOREIGN KEY ("event_listener_id") REFERENCES "event_listeners"("id") ON DELETE CASCADE ON UPDATE CASCADE
);

-- for processing the events with minimal latency, we assume that the topic 
-- is a compound like this: "namespace-id:event-type"
CREATE INDEX IF NOT EXISTS "event_topic_bucket" ON "event_topics" USING hash("topic");

CREATE TABLE IF NOT EXISTS "events_filters" (
    "id" uuid,
    "namespace_id" uuid NOT NULL,
    "name" text NOT NULL,
    "js_code" text NOT NULL,
    PRIMARY KEY ("id"),
    FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("id") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "metrics" (
    "id" serial,
    "namespace" text,
    "workflow" text,
    "revision" text,
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
