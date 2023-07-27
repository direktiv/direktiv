import { z } from "zod";

export const LogListSchema = z.object({
  namespace: z.string(),
});
