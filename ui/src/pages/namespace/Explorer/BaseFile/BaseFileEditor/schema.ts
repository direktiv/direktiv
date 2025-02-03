import { z } from "zod";

export const BaseFileFormSchema = z.object({
  data: z.object({
    spec: z
      .object({
        openapi: z.string(),
        info: z
          .object({
            title: z.string(),
            version: z.string(),
            description: z.string().optional(),
          })
          .passthrough(),
        paths: z.record(z.any()),
      })
      .passthrough(),
    file_path: z.string(),
    errors: z.array(z.unknown()),
  }),
}); // TODO: Add the rest of the fields

export type BaseFileFormSchemaType = z.infer<typeof BaseFileFormSchema>;
