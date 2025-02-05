import { z } from "zod";

export const OpenApiBaseFileFormSchema = z
  .object({
    openapi: z.string(),
    info: z
      .object({
        title: z.string(),
        version: z.string(),
        description: z.string(),
      })
      .passthrough(),
  })
  .passthrough(); // TODO: Add the rest of the fields

export type OpenApiBaseFileFormSchemaType = z.infer<
  typeof OpenApiBaseFileFormSchema
>;
