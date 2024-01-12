import { z } from "zod";

export const SessionSchema = z.object({
  response: z.string(),
});
