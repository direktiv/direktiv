import { TemplateString } from "../primitives/templateString";
import { z } from "zod";

export const headlineLevels = ["h1", "h2", "h3"] as const;

export const Headline = z.object({
  type: z.literal("headline"),
  level: z.enum(headlineLevels),
  label: TemplateString,
});

export type HeadlineType = z.infer<typeof Headline>;
