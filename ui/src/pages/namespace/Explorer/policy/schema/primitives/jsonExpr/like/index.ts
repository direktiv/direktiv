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
export const LikeJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) =>
  strictSingleKeyObject(
    "like",
    z
      .object({
        left: jsonExprSchema,
        pattern: z.array(PatternElementSchema),
      })
      .strict()
  );
