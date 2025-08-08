import { Mutation } from "../procedures/mutation";
import { TemplateString } from "../primitives/templateString";
import { z } from "zod";

export const Button = z.object({
  type: z.literal("button"),
  label: TemplateString.min(1),
  submit: Mutation.optional(),
});

export type ButtonType = z.infer<typeof Button>;
