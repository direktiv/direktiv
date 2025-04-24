import { TemplateString } from "../primitives/templateString";
import { z } from "zod";

export const Text = z.object({
  type: z.literal("text"),
  label: TemplateString,
});

export type TextType = z.infer<typeof Text>;
