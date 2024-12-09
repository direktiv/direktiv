import { z } from "zod";

export const Text = z.object({
  type: z.literal("text"),
  data: z.object({
    label: z.string().min(1),
  }),
});
