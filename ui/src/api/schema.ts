import { z } from "zod";

export const PageinfoSchema = z.object({
  order: z.array(
    z.object({
      field: z.string(), // f.e. "NAME"
      direction: z.string(),
    })
  ),
  filter: z.array(
    z.object({
      field: z.string(), // f.e. "NAME"
      type: z.string(), // f.e. CONTAINS
      val: z.string(), // f.e. "something"
    })
  ),
  limit: z.number(),
  offset: z.number(),
  total: z.number(),
});
