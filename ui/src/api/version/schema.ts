import { z } from "zod";

/**
 * example:
 * 
  {
    "data": {
      "version": "latest-04b8cb1a0",
      "isEnterprise": false,
      "requiresAuth": false
    }
  }
 */
export const VersionSchema = z.object({
  data: z.object({
    version: z.string(),
    isEnterprise: z.boolean(),
    requiresAuth: z.boolean(),
  }),
});
