import { DynamicString } from "../primitives/dynamicString";
import { z } from "zod";

export const Text = z.object({
  type: z.literal("text"),
  label: DynamicString,
});

export type TextType = z.infer<typeof Text>;
