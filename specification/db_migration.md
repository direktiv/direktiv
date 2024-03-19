# 0.8.3

## DB Schema Change

Exec the following sql queries:

```sql
ALTER TABLE "instances_v2" ADD COLUMN IF NOT EXISTS "server" uuid;
```

# 0.8.2

## DB Schema Change

Exec the following sql queries:

```sql
ALTER TABLE "filesystem_roots" ADD COLUMN IF NOT EXISTS "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE "filesystem_roots" ADD COLUMN IF NOT EXISTS "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE "filesystem_files" ADD COLUMN IF NOT EXISTS "data" bytea;
ALTER TABLE "filesystem_files" ADD COLUMN IF NOT EXISTS "checksum" text;


ALTER TABLE "instances_v2" DROP COLUMN IF EXISTS "revision_id";
ALTER TABLE "metrics" DROP COLUMN IF EXISTS "revision";
DROP TABLE IF EXISTS "filesystem_revisions";


ALTER TABLE "event_topics" ADD COLUMN IF NOT EXISTS "filter" text;
ALTER TABLE "event_topics" DROP CONSTRAINT "no_dup_topics_check";
ALTER TABLE "event_topics" ADD  CONSTRAINT "no_dup_topics_check" UNIQUE ("event_listener_id", "topic", "filter");
```
