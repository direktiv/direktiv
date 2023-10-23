import { z } from "zod";

/**
 * example:
 * 
  {
    "api": "4d1cc3a",
    "flow": "4d1cc3a",
    "functions": "4d1cc3a"
  }
 */
export const VersionSchema = z.object({
  api: z.string(),
});
