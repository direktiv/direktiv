import {
  BooleanSchema,
  VariableSchema,
} from "../../primitives/extendedKeyValue";

import { FormBase } from "./utils";
import { TemplateString } from "../../primitives/templateString";
import { z } from "zod";

// default value is either of a checkbox is either a static boolean or a pointer to a variable
const DefaultValueSchema = z.discriminatedUnion("type", [
  VariableSchema,
  BooleanSchema,
]);

type DefaultValueSchemaType = z.infer<typeof DefaultValueSchema>;

export const allowedDefaultValueTypes = [
  "boolean",
  "variable",
] as const satisfies DefaultValueSchemaType["type"][];

export const DefaultValueTypeSchema = z.enum(allowedDefaultValueTypes);

export const FormCheckbox = FormBase.extend({
  type: z.literal("form-checkbox"),
  description: TemplateString.min(1), // overwrite description from base to be required
  defaultValue: DefaultValueSchema,
});

export type FormCheckboxType = z.infer<typeof FormCheckbox>;
