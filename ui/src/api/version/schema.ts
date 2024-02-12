import { z } from "zod";

/**
 * example:
 * 
  {
    "data": "c898514fa"
  }
 */
export const VersionSchema = z.object({
  data: z.string(),
});
