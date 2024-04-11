import { z } from "zod";

const MirrorSchema = z.object({
  url: z.string(),
  gitRef: z.string(),
  authToken: z.string().optional(),
  publicKey: z.string().optional(),
  privateKey: z.string().optional(),
  privateKeyPassphrase: z.string().optional(),
  insecure: z.boolean(),
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
export type MirrorSchemaType = z.infer<typeof MirrorSchema>;
