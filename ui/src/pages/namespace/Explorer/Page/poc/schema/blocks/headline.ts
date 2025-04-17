import { DynamicString } from "../primitives/dynamicString";
import { z } from "zod";

export const Headline = z.object({
  type: z.literal("headline"),
  label: DynamicString,
  description: DynamicString.optional(),
});

export type HeadlineType = z.infer<typeof Headline>;
