import { FormBase } from "./utils";
import { TemplateString } from "../../primitives/templateString";
import { VariableSchema } from "../../primitives/extendedKeyValue";
import { z } from "zod";

const ArraySchema = z.object({
  type: z.literal("array"),
  value: z.array(z.string()),
});

const ValuesSchema = z.discriminatedUnion("type", [
  VariableSchema,
  ArraySchema,
]);

export const allowedValuesTypes = ["variable", "array"] as const;

export const ValuesTypeSchema = z.enum(allowedValuesTypes);

export const FormSelect = FormBase.extend({
  type: z.literal("form-select"),
  values: ValuesSchema,
  defaultValue: TemplateString,
});

export type FormSelectType = z.infer<typeof FormSelect>;
