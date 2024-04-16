import { gitUrlSchema } from "./validation";
import { z } from "zod";

export const MirrorAuthTypeSchema = z.enum(["public", "ssh", "token"]);

// this is part of a namespace record response
export const MirrorSchema = z.object({
  authType: MirrorAuthTypeSchema,
  url: z.string(),
  gitRef: z.string(),
  authToken: z.string().optional(),
  publicKey: z.string().optional(),
  privateKey: z.string().optional(),
  privateKeyPassphrase: z.string().optional(),
  insecure: z.boolean(),
});

// the schemas below are used in POST/PATCH payloads
export const MirrorPublicPostSchema = z.object({
  authType: z.literal("public"),
  url: z.string().url().nonempty(),
  gitRef: z.string().nonempty(),
  insecure: z.boolean(),
});

export const MirrorTokenPostSchema = z.object({
  authType: z.literal("token"),
  url: z.string().url().nonempty(),
  gitRef: z.string().nonempty(),
  authToken: z.string().nonempty({ message: "Required when using token auth" }),
  insecure: z.boolean(),
});

export const MirrorSshPostSchema = z.object({
  authType: z.literal("ssh"),
  url: gitUrlSchema.nonempty({
    message: "format must be git@host:path when using SSH",
  }),
  gitRef: z.string().nonempty(),
  publicKey: z.string().nonempty({ message: "Required when using SSH" }),
  privateKey: z.string().nonempty({ message: "Required when using SSH" }),
  privateKeyPassphrase: z.string().optional(),
  insecure: z.boolean(),
});

export const MirrorKeepTokenPatchSchema = z.object({
  authType: z.undefined(),
  gitRef: z.string().nonempty(),
  url: z.string().url().nonempty(),
  insecure: z.boolean(),
});

export const MirrorKeepSshPatchSchema = z.object({
  authType: z.undefined(),
  gitRef: z.string().nonempty(),
  url: gitUrlSchema.nonempty({
    message: "format must be git@host:path when using SSH",
  }),
  insecure: z.boolean(),
});

export const MirrorPostPatchSchema = z.union([
  MirrorPublicPostSchema,
  MirrorTokenPostSchema,
  MirrorSshPostSchema,
  MirrorKeepTokenPatchSchema,
  MirrorKeepSshPatchSchema,
]);

export type MirrorSchemaType = z.infer<typeof MirrorSchema>;
export type MirrorPostPatchSchemaType = z.infer<typeof MirrorPostPatchSchema>;
