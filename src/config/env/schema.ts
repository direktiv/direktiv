import { z } from "zod";

export const envVariablesSchema = z.object({
  VITE_DEV_API_DOMAIN: z.string().optional(),
  VITE_IS_ENTERPRISE: z
    .string()
    .optional()
    .transform((value) => `${value}`.toLocaleLowerCase() === "true"),
  VITE_E2E_UI_HOST: z.string().optional(),
  VITE_E2E_UI_PORT: z.string().optional(),
  VITE_RQ_DEV_TOOLS: z
    .string()
    .optional()
    .transform((value) => `${value}`.toLocaleLowerCase() === "true"),
  VITE_BASE: z.string().optional(),
});
