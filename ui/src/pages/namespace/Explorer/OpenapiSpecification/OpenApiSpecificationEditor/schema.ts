import { z } from "zod";

export const OpenapiSpecificationFormSchema = z
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
  .passthrough();
