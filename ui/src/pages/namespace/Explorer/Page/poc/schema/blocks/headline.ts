import { TemplateString } from "../primitives/templateString";
import { z } from "zod";

export const Headline = z.object({
  type: z.literal("headline"),
  label: TemplateString,
  description: TemplateString.optional(),
});

export type HeadlineType = z.infer<typeof Headline>;
