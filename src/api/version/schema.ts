import { z } from "zod";

export const VersionSchema = z.object({
  api: z.string(),
  flow: z.string(),
  functions: z.string(),
});
