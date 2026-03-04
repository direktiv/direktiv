import { strictSingleKeyObject } from "../utils";
import { z } from "zod";

// when { [1, 2, "something"] };
export const SetExpressionSchema = (expressionSchema: z.ZodTypeAny) =>
  strictSingleKeyObject("Set", z.array(expressionSchema));
