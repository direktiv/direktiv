import { FormBase } from "./utils";
import { TemplateString } from "../../primitives/templateString";
import { z } from "zod";

export const FormSelect = FormBase.extend({
  type: z.literal("form-select"),
  values: z.array(z.string()).min(1),
  defaultValue: TemplateString.min(1),
});

export type FormSelectType = z.infer<typeof FormSelect>;
