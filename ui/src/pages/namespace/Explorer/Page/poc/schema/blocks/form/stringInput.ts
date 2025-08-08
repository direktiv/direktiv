import { FormBase } from "./utils";
import { TemplateString } from "../../primitives/templateString";
import { z } from "zod";

export const stringInputTypes = [
  "text",
  "date",
  "password",
  "email",
  "url",
] as const;

export const FormStringInput = FormBase.extend({
  type: z.literal("form-string-input"),
  variant: z.enum(stringInputTypes),
  defaultValue: TemplateString.min(1),
});

export type FormStringInputType = z.infer<typeof FormStringInput>;
