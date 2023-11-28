import { z } from "zod";

/**
 * example:
 * 
  {
    "api": "d7403237",
    "flow":"d7403237"
  }
 */
export const VersionSchema = z.object({
  api: z.string(),
  flow: z.string(),
});
