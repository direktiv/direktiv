import type { ExpressionSchemaType } from "..";
import { strictSingleKeyObject } from "../utils";
import { z } from "zod";

// when { principal is User in Group::"engineering" };
export const IsExpressionSchema = (expressionSchema: ExpressionSchemaType) =>
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
