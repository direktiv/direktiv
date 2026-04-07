import { z } from "zod";

export const ExpressionUnaryOperators = ["!", "neg", "isEmpty"] as const;
export type UnaryOperator = (typeof ExpressionUnaryOperators)[number];

const _ExpressionAttributeOperators = [".", "has"] as const;
export type AttributeOperator = (typeof _ExpressionAttributeOperators)[number];

export const ExpressionBinaryOperators = [
  "==",
  "!=",
  "in",
  "<",
  "<=",
  ">",
  ">=",
  "&&",
  "||",
  "+",
  "-",
  "*",
  "contains",
  "containsAll",
  "containsAny",
  "hasTag",
  "getTag",
] as const;

export type BinaryOperator = (typeof ExpressionBinaryOperators)[number];

export type BooleanOperator = Extract<BinaryOperator, "&&" | "||">;

export const strictSingleKeyObject = <
  Key extends string,
  Schema extends z.ZodTypeAny,
>(
  key: Key,
  valueSchema: Schema
) => z.object({ [key]: valueSchema } as Record<Key, Schema>).strict();
