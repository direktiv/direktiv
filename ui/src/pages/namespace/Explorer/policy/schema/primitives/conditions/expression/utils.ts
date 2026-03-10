import { z } from "zod";

export const ExpressionUnaryOperators = ["!", "neg", "isEmpty"] as const;

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

export const ExpressionReservedKeys = new Set([
  "Value",
  "Var",
  "Slot",
  "Unknown",
  ".",
  "has",
  "is",
  "like",
  "if-then-else",
  "Set",
  "Record",
  ...ExpressionUnaryOperators,
  ...ExpressionBinaryOperators,
]);

export const strictSingleKeyObject = <
  Key extends string,
  Schema extends z.ZodTypeAny,
>(
  key: Key,
  valueSchema: Schema
) => z.object({ [key]: valueSchema } as Record<Key, Schema>).strict();
