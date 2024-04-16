import { gitUrlSchema } from "./validation";
import { z } from "zod";

export const MirrorPublicPostSchema = z.object({
  url: z.string().url().nonempty(),
  gitRef: z.string().nonempty(),
  insecure: z.boolean(),
});

export const MirrorTokenPostSchema = z.object({
  url: z.string().url().nonempty(),
  gitRef: z.string().nonempty(),
  authToken: z.string().nonempty({ message: "Required when using token auth" }),
  insecure: z.boolean(),
});

export const MirrorSshPostSchema = z.object({
  url: gitUrlSchema.nonempty({
    message: "format must be git@host:path when using SSH",
  }),
  gitRef: z.string().nonempty(),
  publicKey: z.string().nonempty({ message: "Required when using SSH" }),
  privateKey: z.string().nonempty({ message: "Required when using SSH" }),
  privateKeyPassphrase: z.string().optional(),
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

export type MirrorPostSchemaType = z.infer<typeof MirrorPostSchema>;
export type MirrorFormType = z.infer<typeof mirrorFormType>;
