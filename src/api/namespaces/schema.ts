import { PageinfoSchema } from "../schema";
import { z } from "zod";

export const NamespaceSchema = z.object({
  createdAt: z.string(),
  updatedAt: z.string(),
  name: z.string(),
  oid: z.string(),
});

export const NamespaceListSchema = z.object({
  pageInfo: PageinfoSchema,
  results: z.array(NamespaceSchema),
});

export const NamespaceCreatedSchema = z.object({
  namespace: NamespaceSchema,
});

// note: in the current API implementation, a mirror is created
// by creating a namespace with this in the payload.
export const MirrorSchema = z.object({
  passphrase: z.string().optional(),
  privateKey: z.string().optional(),
  publicKey: z.string().optional(),
  ref: z.string().nonempty(),
  url: z.string().url().nonempty(),
});

export type NamespaceListSchemaType = z.infer<typeof NamespaceListSchema>;
export type MirrorSchemaType = z.infer<typeof MirrorSchema>;
