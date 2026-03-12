import type { ExpressionSchemaType } from "../types";
import { strictSingleKeyObject } from "../utils";
import { z } from "zod";

export const PatternElementSchema = z.union([
  z.literal("Wildcard"),
  z.object({ Literal: z.string() }).strict(),
]);

export type PatternElement = z.infer<typeof PatternElementSchema>;

// when { context.requesterEmail like "*@company.com" };
export const LikeExpressionSchema = (expressionSchema: ExpressionSchemaType) =>
  strictSingleKeyObject(
    "like",
    z
      .object({
        left: expressionSchema,
        pattern: z.array(PatternElementSchema),
      })
      .strict()
  );
