import { z } from "zod";

export const envVariablesSchema = z.object({
  VITE_DEV_API_DOMAIN: z.string().optional(),
  VITE_APP_VERSION: z
    .string()
    .optional()
    .transform((value) => {
      if (value && `${value}`.length === 0) {
        return undefined;
      }
      return value;
    }),
  VITE_IS_ENTERPRISE: z
    .string()
    .optional()
    .transform((value) => `${value}`.toLocaleLowerCase() === "true"),
  VITE_LEGACY_DESIGN: z
    .string()
    .optional()
    .transform((value) => `${value}`.toLocaleLowerCase() === "true"),
  VITE_RQ_DEV_TOOLS: z
    .string()
    .optional()
    .transform((value) => `${value}`.toLocaleLowerCase() === "true"),
});
