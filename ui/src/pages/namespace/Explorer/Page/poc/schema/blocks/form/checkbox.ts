import { FormBase } from "./utils";
import { TemplateString } from "../../primitives/templateString";
import { z } from "zod";

export const FormCheckbox = FormBase.extend({
  type: z.literal("form-checkbox"),
  description: TemplateString.min(1), // overwrite description from base to be required
  defaultValue: z.boolean(),
});

export type FormCheckboxType = z.infer<typeof FormCheckbox>;
