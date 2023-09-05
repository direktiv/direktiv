import { z } from "zod";

export const PolicySchema = z.object({
  body: z.string(),
});

export const PolicyCreatedSchema = z.object({});
