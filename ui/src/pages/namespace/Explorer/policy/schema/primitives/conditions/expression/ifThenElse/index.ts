import { strictSingleKeyObject } from "../utils";
import { z } from "zod";

// when { if context.something then principal has "email" else false };
export const IfThenElseExpressionSchema = (expressionSchema: z.ZodTypeAny) =>
  strictSingleKeyObject(
    "if-then-else",
    z
      .object({
        if: expressionSchema,
        then: expressionSchema,
        else: expressionSchema,
      })
      .strict()
  );
