import { z } from "zod";

const boolean = z
  .string()
  .optional()
  .transform((value) => `${value}`.toLocaleLowerCase() === "true");

export const envVariablesSchema = z.object({
  VITE_DEV_API_DOMAIN: z.string().optional(),
  VITE_RQ_DEV_TOOLS: boolean,
  VITE_ENABLE_TS_WORKFLOWS: boolean,
});
