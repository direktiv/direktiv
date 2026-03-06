import { ExpressionBinaryOperators, strictSingleKeyObject } from "../utils";
import { z } from "zod";

const BinaryArgumentsSchema = (expressionSchema: z.ZodTypeAny) =>
  z
    .object({
      left: expressionSchema,
      right: expressionSchema,
    })
    .strict();

const BinaryOperatorSchema = (
  operator: (typeof ExpressionBinaryOperators)[number],
  expressionSchema: z.ZodTypeAny
) => strictSingleKeyObject(operator, BinaryArgumentsSchema(expressionSchema));

type BinaryOperatorSchemaType = ReturnType<typeof BinaryOperatorSchema>;

// when { principal == User::"alice" };
export const BinaryExpressionSchema = (expressionSchema: z.ZodTypeAny) =>
  z.union(
    ExpressionBinaryOperators.map((operator) =>
      BinaryOperatorSchema(operator, expressionSchema)
    ) as [
      BinaryOperatorSchemaType,
      BinaryOperatorSchemaType,
      ...BinaryOperatorSchemaType[],
    ]
  );
