import { FormBase } from "./utils";
import { TemplateString } from "../../primitives/templateString";
import { z } from "zod";

const ArraySchema = z.object({
  type: z.literal("array"),
  value: z.array(z.string()),
});

export const VariableSelectOptions = z.object({
  type: z.literal("variable-select-options"),
  arrayPath: z.string().min(1),
  labelPath: z.string().min(1),
  valuePath: z.string().min(1),
});

const ValuesSchema = z.discriminatedUnion("type", [
  VariableSelectOptions,
  ArraySchema,
]);

export const allowedValuesTypes = ["select-options", "array"] as const;

export const ValuesTypeSchema = z.enum(allowedValuesTypes);

export const FormSelect = FormBase.extend({
  type: z.literal("form-select"),
  values: ValuesSchema,
  defaultValue: TemplateString,
});

export type FormSelectType = z.infer<typeof FormSelect>;
