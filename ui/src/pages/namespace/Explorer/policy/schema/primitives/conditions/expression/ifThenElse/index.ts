import { strictSingleKeyObject } from "../utils";
import { z } from "zod";

// when { if context.uses_mfa then principal has "mfa_device_id" else false };
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
