import { FormBase } from "./utils";
import { StringArraySchema } from "../../primitives/extendedKeyValue/array";
import { TemplateString } from "../../primitives/templateString";
import { VariableSchema } from "../../primitives/extendedKeyValue";
import { z } from "zod";

const ValuesSchema = z.discriminatedUnion("type", [
  VariableSchema,
  StringArraySchema,
]);

export const allowedValuesTypes = ["variable", "string-array"] as const;

export const ValuesTypeSchema = z.enum(allowedValuesTypes);

export const FormSelect = FormBase.extend({
  type: z.literal("form-select"),
  values: ValuesSchema,
  defaultValue: TemplateString,
});

export type FormSelectType = z.infer<typeof FormSelect>;
