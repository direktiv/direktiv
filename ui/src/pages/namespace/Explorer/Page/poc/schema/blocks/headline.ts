import { TemplateString } from "../primitives/templateString";
import { z } from "zod";

export const Headline = z.object({
  type: z.literal("headline"),
  level: z.enum(["h1", "h2", "h3"]),
  label: TemplateString,
});

export type HeadlineType = z.infer<typeof Headline>;
