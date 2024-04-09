import { z } from "zod";

const MirrorSchema = z.object({
  url: z.string(),
  gitRef: z.string(),
  gitCommitHash: z.string().optional(), // null?
  publicKey: z.string().optional(), // null?,
  insecure: z.boolean(), // true
  createdAt: z.string(),
  updatedAt: z.string(),
});

export const NamespaceSchema = z.object({
  createdAt: z.string(),
  updatedAt: z.string(),
  name: z.string(),
  mirror: MirrorSchema.nullable(),
});

export const NamespaceListSchema = z.object({
  data: z.array(NamespaceSchema),
});

export const NamespaceCreatedSchema = z.object({
  data: NamespaceSchema.omit({ mirror: true }),
});

export const NamespaceDeletedSchema = z.null();

export type NamespaceListSchemaType = z.infer<typeof NamespaceListSchema>;
