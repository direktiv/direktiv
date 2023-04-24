import { z } from "zod";

// at the moment, we only get a type check on build- and server-start-time
// and we break the app early if any of the env variables are missing or set
// to an invalid value.
// If you use import.meta.env.VITE_APP_VERSION in a component, it will
// not be typed yet.
const envVariablesSchema = z.object({
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
    .boolean()
    .optional()
    .transform((value) => `${value}`.toLocaleLowerCase() === "true"),
  VITE_LEGACY_DESIGN: z
    .boolean()
    .optional()
    .transform((value) => `${value}`.toLocaleLowerCase() === "true"),
});

export default envVariablesSchema;
