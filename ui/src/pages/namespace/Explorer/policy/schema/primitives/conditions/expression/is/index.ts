import { strictSingleKeyObject } from "../utils";
import { z } from "zod";

// when { principal is User in Group::"friends" };
export const IsExpressionSchema = (expressionSchema: z.ZodTypeAny) =>
  strictSingleKeyObject(
    "is",
    z
      .object({
        left: expressionSchema,
        entity_type: z.string(),
        in: expressionSchema.optional(),
      })
      .strict()
  );
