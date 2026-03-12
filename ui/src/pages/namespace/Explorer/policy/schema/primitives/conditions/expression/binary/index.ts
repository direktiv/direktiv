import { ExpressionBinaryOperators, strictSingleKeyObject } from "../utils";
import type { ExpressionSchemaType } from "..";
import { z } from "zod";

const BinaryArgumentsSchema = (expressionSchema: ExpressionSchemaType) =>
  z
    .object({
      left: expressionSchema,
      right: expressionSchema,
    })
    .strict();

const BinaryOperatorSchema = (
  operator: (typeof ExpressionBinaryOperators)[number],
  expressionSchema: ExpressionSchemaType
) => strictSingleKeyObject(operator, BinaryArgumentsSchema(expressionSchema));

type BinaryOperatorSchemaType = ReturnType<typeof BinaryOperatorSchema>;

// when { principal == User::"alice" };
export const BinaryExpressionSchema = (
  expressionSchema: ExpressionSchemaType
) =>
  z.union(
    ExpressionBinaryOperators.map((operator) =>
      BinaryOperatorSchema(operator, expressionSchema)
    ) as [
      BinaryOperatorSchemaType,
      BinaryOperatorSchemaType,
      ...BinaryOperatorSchemaType[],
    ]
  );
