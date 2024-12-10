import { DynamicString } from "../primitives/dynamicString";
import { z } from "zod";

export const Text = z.object({
  type: z.literal("text"),
  data: z.object({
    label: DynamicString,
  }),
});

export type TextType = z.infer<typeof Text>;
