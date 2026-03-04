import { ExpressionUnaryOperators, strictSingleKeyObject } from "../utils";
import { z } from "zod";

const UnaryArgumentSchema = (expressionSchema: z.ZodTypeAny) =>
  z.object({ arg: expressionSchema }).strict();

// when { !context.something }; / when { -1 }; / when { [1, 2].isEmpty() };
export const UnaryExpressionSchema = (expressionSchema: z.ZodTypeAny) =>
  z.union(
    ExpressionUnaryOperators.map((operator) =>
      strictSingleKeyObject(operator, UnaryArgumentSchema(expressionSchema))
    ) as unknown as [z.ZodTypeAny, z.ZodTypeAny, ...z.ZodTypeAny[]]
  );
