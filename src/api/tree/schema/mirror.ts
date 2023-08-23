import { LogLevelSchema, PageinfoSchema } from "../../schema";

import { z } from "zod";

/**
 * Example for a mirror-info response
 * {
  "namespace":  "examples",
  "info":  {
    "url":  "https://github.com/direktiv/direktiv-examples",
    "ref":  "main",
    "cron":  "",
    "publicKey":  "",
    "commitId":  "",
    "lastSync":  null,
    "privateKey":  "",
    "passphrase":  ""
  },
  "activities":  {
    "pageInfo":  null,
    "results":  [
      {
        "id":  "29f1c217-2f2a-447d-8730-23f519634755",
        "type":  "init",
        "status":  "complete",
        "createdAt":  "2023-08-04T12:26:18.271385Z",
        "updatedAt":  "2023-08-04T12:26:18.968351Z"
      }
    ]
  }
}
*/

export const MirrorInfoInfoSchema = z.object({
  url: z.string(),
  ref: z.string(),
  lastSync: z.string().or(z.null()),
});

// According to API spec, but currently dry_run isn't used in the API.
const MirrorActivityTypeSchema = z.enum(["init", "sync", "dry_run"]);

// According to the API spec, but currently cancelled isn't used in the API.
const MirrorActivityStatusSchema = z.enum([
  "pending",
  "executing",
  "complete",
  "failed",
  "cancelled",
]);

export const MirrorActivitySchema = z.object({
  id: z.string(),
  type: MirrorActivityTypeSchema,
  status: MirrorActivityStatusSchema,
  createdAt: z.string(),
  updatedAt: z.string(),
});

export const MirrorInfoSchema = z.object({
  namespace: z.string(),
  info: MirrorInfoInfoSchema,
  activities: z.object({
    pageInfo: PageinfoSchema.or(z.null()),
    results: z.array(MirrorActivitySchema),
  }),
});

/**
 * Example for mirror activity log response (streaming only)
 {
  "pageInfo": {
    "order": [],
    "filter": [],
    "limit": 0,
    "offset": 0,
    "total": 136
  },
  "namespace": "examples",
  "activity": "2d92ecec-1f88-4fcd-a525-4e8c8594e6cc",
  "results": [
    {
      "t": "2023-08-22T08:57:10.581391Z",
      "level": "info",
      "msg": "starting mirroring process, type = sync, process_id = 2d92ecec-1f88-4fcd-a525-4e8c8594e6cc",
      "tags": {
        "level": "info",
        "mirror-id": "2d92ecec-1f88-4fcd-a525-4e8c8594e6cc",
        "recipientType": "mirror",
        "source": "2d92ecec-1f88-4fcd-a525-4e8c8594e6cc",
        "trace": "00000000000000000000000000000000",
        "type": "mirror"
      }
    },
  }
 */

export const MirrorActivityLogItemSchema = z.object({
  t: z.string(),
  level: LogLevelSchema,
  msg: z.string(),
});

export const MirrorActivityLogSchema = z.object({
  pageInfo: PageinfoSchema,
  namespace: z.string(),
  activity: z.string(),
  results: z.array(MirrorActivityLogItemSchema),
});

export type MirrorActivitySchemaType = z.infer<typeof MirrorActivitySchema>;
export type MirrorActivityLogSchemaType = z.infer<
  typeof MirrorActivityLogSchema
>;
export type MirrorActivityTypeSchemaType = z.infer<
  typeof MirrorActivityTypeSchema
>;
export type MirrorActivityStatusSchemaType = z.infer<
  typeof MirrorActivityStatusSchema
>;
