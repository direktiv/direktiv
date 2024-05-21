import { z } from "zod";

export const envVariablesSchema = z.object({
  VITE_DEV_API_DOMAIN: z.string().optional(),
  VITE_RQ_DEV_TOOLS: z
    .string()
    .optional()
    .transform((value) => `${value}`.toLocaleLowerCase() === "true"),
});
