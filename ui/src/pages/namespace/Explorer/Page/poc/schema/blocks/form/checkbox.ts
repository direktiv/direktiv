import { FormBase } from "./utils";
import { z } from "zod";

export const FormCheckbox = FormBase.extend({
  type: z.literal("form-checkbox"),
  defaultValue: z.boolean(),
});

export type FormCheckboxType = z.infer<typeof FormCheckbox>;
