import { FormBase } from "./utils";
import { z } from "zod";

export const inputTypes = [
  "text",
  "date",
  "password",
  "email",
  "url",
  "number",
] as const;

export const FormInput = FormBase.extend({
  type: z.literal("form-input"),
  variant: z.enum(inputTypes),
  // TODO: no binding to variant yet (alternatively, we could have a form-text-input and form-number-input)
  defaultValue: z.union([z.string(), z.number()]),
});

export type FormInputType = z.infer<typeof FormInput>;
