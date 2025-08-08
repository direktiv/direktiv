import {
  NumberSchema,
  VariableSchema,
} from "../../primitives/extendedKeyValue";

import { FormBase } from "./utils";
import { z } from "zod";

// default value is either of a number input is either a static number or a pointer to a variable
const DefaultValue = z.discriminatedUnion("type", [
  VariableSchema,
  NumberSchema,
]);

export const FormNumberInput = FormBase.extend({
  type: z.literal("form-number-input"),
  defaultValue: DefaultValue,
});

export type FormNumberInputType = z.infer<typeof FormNumberInput>;
