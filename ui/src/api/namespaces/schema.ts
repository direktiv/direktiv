import { z } from "zod";

export const NamespaceSchema = z.object({
  createdAt: z.string(),
  updatedAt: z.string(),
  name: z.string(),
});

export const NamespaceListSchema = z.object({
  results: z.array(NamespaceSchema),
});

export const NamespaceCreatedSchema = z.object({
  namespace: NamespaceSchema,
});

export const NamespaceDeletedSchema = z.null();

export type NamespaceListSchemaType = z.infer<typeof NamespaceListSchema>;
