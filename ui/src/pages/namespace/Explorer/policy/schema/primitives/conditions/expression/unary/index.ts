import { ExpressionUnaryOperators, strictSingleKeyObject } from "../utils";
import type { ExpressionSchemaType } from "../types";
import { z } from "zod";

const UnaryArgumentSchema = (expressionSchema: ExpressionSchemaType) =>
  z.object({ arg: expressionSchema }).strict();

/*
  when { !context.mfa_verified };
  when { -context.risk_score };
  when { isEmpty(context.session.tags) };
*/
export const UnaryExpressionSchema = (expressionSchema: ExpressionSchemaType) =>
  z.union(
    ExpressionUnaryOperators.map(
      (operator) =>
        strictSingleKeyObject(operator, UnaryArgumentSchema(expressionSchema))
      // z.union() expects at least two schemas.
    ) as unknown as [
      ExpressionSchemaType,
      ExpressionSchemaType,
      ...ExpressionSchemaType[],
    ]
  );
