import { FormBase } from "./utils";
import { z } from "zod";

export const FormNumberInput = FormBase.extend({
  type: z.literal("form-number-input"),
  defaultValue: z.number(),
});

export type FormNumberInputType = z.infer<typeof FormNumberInput>;
