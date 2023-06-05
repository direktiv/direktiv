CREATE TABLE IF NOT EXISTS  "filesystem_roots" (
    "id" uuid,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id"),
    CONSTRAINT "fk_namespaces_filesystem_roots"
    FOREIGN KEY ("id") REFERENCES "namespaces"("oid") ON DELETE CASCADE ON UPDATE CASCADE
);


CREATE TABLE IF NOT EXISTS "filesystem_files" (
    "id" uuid,
    "path" text NOT NULL,
    "depth" integer NOT NULL,
    "typ" text NOT NULL,
    "root_id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id"),
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
    FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("oid") ON DELETE CASCADE ON UPDATE CASCADE
);


CREATE TABLE IF NOT EXISTS "mirror_processes" (
    "id" uuid,
    "namespace_id" uuid NOT NULL,
    "status" text NOT NULL,
    "typ" 	 text NOT NULL,
    "ended_at" timestamptz,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id"),
    CONSTRAINT "fk_namespaces_mirror_processes"
    FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("oid") ON DELETE CASCADE ON UPDATE CASCADE
);


CREATE TABLE IF NOT EXISTS "secrets" (
    "id" uuid,
    "namespace_id" uuid NOT NULL,
    "name" text NOT NULL,
    "data" 	 text NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id"),
    CONSTRAINT "fk_namespaces_secrets"
    FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("oid") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "runtime_variables" (
    "id" uuid,
    "namespace_id" uuid,
    "workflow_id" uuid,
    "instance_id" uuid,

    "name"  text NOT NULL,
    "mime_type"  text NOT NULL,
    "data"  bytea NOT NULL,

    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY ("id"),
    UNIQUE(namespace_id, workflow_id, instance_id, name),

    CONSTRAINT "fk_namespaces_runtime_variables"
    FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("oid") ON DELETE CASCADE ON UPDATE CASCADE,

    CONSTRAINT "fk_filesystem_files_runtime_variables"
    FOREIGN KEY ("workflow_id") REFERENCES "filesystem_files"("id") ON DELETE CASCADE ON UPDATE CASCADE

    -- TODO: alan, please add instance_id FOREIGN KEY.
);


-- TODO: alex this table schema need have not null modifiers.
CREATE TABLE IF NOT EXISTS "log_msgs" (
    "oid" uuid,
    "t" timestamptz,
    "msg" text,
    "level" integer,
    "root_instance_id" uuid,
    "log_instance_call_path" text,
    "tags" jsonb,
    "workflow_id" uuid,
    "mirror_activity_id" uuid,
    "instance_logs" text,
    "namespace_logs" text,
    PRIMARY KEY ("oid")
);


-- TODO: alan please fix id and other fields types for postgres.
CREATE TABLE IF NOT EXISTS "instances_v2" (
    "id" uuid,
    "namespace_id" uuid NOT NULL,
    "workflow_id" uuid NOT NULL,
    "revision_id" uuid NOT NULL,
    "root_instance_id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "ended_at" timestamptz,
    "deadline" timestamptz,
    "status" integer NOT NULL,
    "called_as" text NOT NULL,
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
    FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("oid") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "events_history" (
    "id" uuid,
    "cloudevent" jsonb NOT NULL,
    "namespace_id" uuid NOT NULL,
    "received_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id"),
);

CREATE TABLE IF NOT EXISTS "event_subscribers" (
    "id" uuid,
    "namespace_id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted" boolean NOT NULL DEFAULT 0,
    "received_events" jsonb NOT NULL,
    "events_lifespan" integer NOT NULL DEFAULT 0,
    "needs_all"  boolean NOT NULL,
    "workflow_id" uuid,
    "instance_id" uuid,
    "step" int,
    PRIMARY KEY ("id"),
    FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("oid") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "event_filters" (
    "id" uuid,
    "namespace_id" uuid NOT NULL,
    "name" text NOT NULL,
    "jscode" text NOT NULL,
    PRIMARY KEY ("id"),
    FOREIGN KEY ("namespace_id") REFERENCES "namespaces"("oid") ON DELETE CASCADE ON UPDATE CASCADE
);
