import { MirrorSchema } from "./mirror";
import { z } from "zod";

export const NamespaceSchema = z.object({
  createdAt: z.string(),
  updatedAt: z.string(),
  name: z.string(),
  mirror: MirrorSchema.nullable(),
});

export const NamespaceListSchema = z.object({
  data: z.array(NamespaceSchema),
});

export const NamespaceCreatedEditedSchema = z.object({
  data: NamespaceSchema.omit({ mirror: true }),
});

export const NamespaceDeletedSchema = z.null();

export type NamespaceListSchemaType = z.infer<typeof NamespaceListSchema>;
