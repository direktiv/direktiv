import { z } from "zod";

export const PageinfoSchema = z.object({
  order: z.array(z.string()),
  filter: z.array(z.string()),
  limit: z.number(),
  offset: z.number(),
  total: z.number(),
});
