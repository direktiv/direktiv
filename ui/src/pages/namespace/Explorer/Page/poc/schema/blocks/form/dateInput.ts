import { FormBase } from "./utils";
import { z } from "zod";

export const FormDateInput = FormBase.extend({
  type: z.literal("form-date-input"),
  defaultValue: z.string(),
});

export type FormDateInputType = z.infer<typeof FormDateInput>;
