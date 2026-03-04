import { z } from "zod";

export const JsonExprUnaryOperators = ["!", "neg", "isEmpty"] as const;

export const JsonExprBinaryOperators = [
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

export const JsonExprReservedKeys = new Set([
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
  ...JsonExprUnaryOperators,
  ...JsonExprBinaryOperators,
]);

export const strictSingleKeyObject = <
  Key extends string,
  Schema extends z.ZodTypeAny,
>(
  key: Key,
  valueSchema: Schema
) => z.object({ [key]: valueSchema } as Record<Key, Schema>).strict();
