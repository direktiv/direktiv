CREATE TABLE IF NOT EXISTS  "namespaces" (
    "name" text NOT NULL UNIQUE,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("name")
);

CREATE TABLE IF NOT EXISTS  "system_heart_beats" (
    "group" text,
    "key" text,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("group", "key")
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
    "root_id" uuid NOT NULL,
    "path" text NOT NULL,
    "depth" integer NOT NULL,
    "typ" text NOT NULL,

    "data" bytea,
    "checksum" text,

    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mime_type" text NOT NULL,
    PRIMARY KEY ("root_id", "path"),
    CONSTRAINT "fk_filesystem_roots_filesystem_files"
    FOREIGN KEY ("root_id") REFERENCES "filesystem_roots"("id") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "mirror_configs" (
    "namespace" text,
    "url" text NOT NULL,
    "git_ref" text NOT NULL,
    "auth_type" text NOT NULL,
    "auth_token" text,
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

    CONSTRAINT "runtime_variables_unique_2"
    UNIQUE NULLS NOT DISTINCT (namespace, name, workflow_path, instance_id)
);
DROP INDEX IF EXISTS "runtime_variables_unique";

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

CREATE TABLE IF NOT EXISTS "traces" (
    "trace_id" text PRIMARY KEY,
    "span_id" text NOT NULL,
    "parent_span_id" text,
    "start_time" timestamptz NOT NULL,
    "end_time" timestamptz,
    "metadata" JSONB
);
