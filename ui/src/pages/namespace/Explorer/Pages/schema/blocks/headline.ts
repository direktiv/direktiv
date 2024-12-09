import { z } from "zod";

export const Headline = z.object({
  type: z.literal("headline"),
  data: z.object({
    label: z.string().min(1),
    description: z.string().optional(),
  }),
});

export type HeadlineType = z.infer<typeof Headline>;
