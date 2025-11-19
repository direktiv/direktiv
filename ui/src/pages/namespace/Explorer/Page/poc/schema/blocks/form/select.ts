import { FormBase } from "./utils";
import { TemplateString } from "../../primitives/templateString";
import { Variable } from "../../primitives/variable";
import { z } from "zod";

const SelectOption = z.object({
  label: TemplateString,
  value: TemplateString,
});

const StaticSelectOptions = z.object({
  type: z.literal("static-select-options"),
  value: z.array(SelectOption),
});

const VariableSelectOptions = z.object({
  type: z.literal("variable-select-options"),
  data: Variable,
  label: Variable,
  value: Variable,
});

const ValuesSchema = z.discriminatedUnion("type", [
  VariableSelectOptions,
  StaticSelectOptions,
]);

export const allowedValuesTypes = [
  "variable-select-options",
  "static-select-options",
] as const;

export const ValuesTypeSchema = z.enum(allowedValuesTypes);

export const FormSelect = FormBase.extend({
  type: z.literal("form-select"),
  values: ValuesSchema,
  defaultValue: TemplateString,
});

export type FormSelectType = z.infer<typeof FormSelect>;
