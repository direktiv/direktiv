import { PageinfoSchema } from "../../../schema";
import { gitUrlSchema } from "./validation";
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
    "passphrase":  "",
    "insecure": false
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

// In the current API implementation, for secret values, "-" means a value exists.
export const MirrorInfoInfoSchema = z.object({
  url: z.string(),
  ref: z.string(),
  lastSync: z.string().or(z.null()),
  publicKey: z.string(),
  privateKey: z.enum(["-", ""]),
  passphrase: z.enum(["-", ""]),
  insecure: z.boolean(),
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

export const MirrorSyncResponseSchema = z.null();

export const UpdateMirrorResponseSchema = z.null();

// note: in the current API implementation, a mirror is created
// by creating a namespace with the mirror object in the payload.
export const MirrorPublicPostSchema = z.object({
  url: z.string().url().nonempty(),
  ref: z.string().nonempty(),
  insecure: z.boolean(),
});

// When Token auth is used, token is submitted as "passphrase"
export const MirrorTokenPostSchema = z.object({
  url: z.string().url().nonempty(),
  ref: z.string().nonempty(),
  passphrase: z
    .string()
    .nonempty({ message: "Required when using token auth" }),
  insecure: z.boolean(),
});

export const MirrorSshPostSchema = z.object({
  url: gitUrlSchema.nonempty({
    message: "format must be git@host:path when using SSH",
  }),
  ref: z.string().nonempty(),
  passphrase: z.string().optional(),
  privateKey: z.string().nonempty({ message: "Required when using SSH" }),
  publicKey: z.string().nonempty({ message: "Required when using SSH" }),
  insecure: z.boolean(),
});

export const MirrorPostSchema = MirrorPublicPostSchema.or(
  MirrorTokenPostSchema
).or(MirrorSshPostSchema);

const mirrorFormType = z.enum([
  "public",
  "ssh",
  "token",
  "keep-ssh",
  "keep-token",
]);

export type MirrorActivitySchemaType = z.infer<typeof MirrorActivitySchema>;

export type MirrorActivityTypeSchemaType = z.infer<
  typeof MirrorActivityTypeSchema
>;
export type MirrorActivityStatusSchemaType = z.infer<
  typeof MirrorActivityStatusSchema
>;
export type MirrorSyncResponseSchemaType = z.infer<
  typeof MirrorSyncResponseSchema
>;
export type MirrorInfoSchemaType = z.infer<typeof MirrorInfoSchema>;
export type MirrorPostSchemaType = z.infer<typeof MirrorPostSchema>;
export type UpdateMirrorResponseSchemaType = z.infer<
  typeof UpdateMirrorResponseSchema
>;
export type MirrorFormType = z.infer<typeof mirrorFormType>;
