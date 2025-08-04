import { FormBase } from "./utils";
import { z } from "zod";

export const FormSelect = FormBase.extend({
  type: z.literal("form-select"),
  values: z.array(z.string()).min(1),
  defaultValue: z.string(),
});

export type FormSelectType = z.infer<typeof FormSelect>;
