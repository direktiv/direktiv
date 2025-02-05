import { z } from "zod";

export const BaseFileFormSchema = z.object({
  openapi: z.string(),
  info: z
    .object({
      title: z.string(),
      version: z.string(),
      description: z.string(),
    })
    .passthrough(),
}); // TODO: Add the rest of the fields

export type BaseFileFormSchemaType = z.infer<typeof BaseFileFormSchema>;
