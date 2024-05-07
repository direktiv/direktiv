import { MirrorSchema } from "./mirror";
import { z } from "zod";

export const NamespaceSchema = z.object({
  createdAt: z.string(),
  updatedAt: z.string(),
  name: z.string(),
  mirror: MirrorSchema.nullable(),
  isSystemNamespace: z.boolean(),
});

export const NamespaceListSchema = z.object({
  data: z.array(NamespaceSchema),
});

export const NamespaceCreatedEditedSchema = z.object({
  data: NamespaceSchema,
});

export const NamespaceDeletedSchema = z.null();

export type NamespaceListSchemaType = z.infer<typeof NamespaceListSchema>;
export type NamespaceCreatedEditedSchemaType = z.infer<
  typeof NamespaceCreatedEditedSchema
>;
