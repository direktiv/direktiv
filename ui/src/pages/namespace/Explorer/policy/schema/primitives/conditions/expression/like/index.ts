import { strictSingleKeyObject } from "../utils";
import { z } from "zod";

const PatternElementSchema = z.union([
  z.literal("Wildcard"),
  z
    .object({
      Literal: z.string(),
    })
    .strict(),
]);

// when { resource.email like "*@amazon.com" };
export const LikeExpressionSchema = (expressionSchema: z.ZodTypeAny) =>
  strictSingleKeyObject(
    "like",
    z
      .object({
        left: expressionSchema,
        pattern: z.array(PatternElementSchema),
      })
      .strict()
  );
