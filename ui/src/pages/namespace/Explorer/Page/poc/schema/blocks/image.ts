import { TemplateString } from "../primitives/templateString";
import { z } from "zod";

export const Image = z.object({
  type: z.literal("image"),
  src: TemplateString,
  width: z.number(),
  height: z.number(),
});

export type ImageType = z.infer<typeof Image>;
