import { z } from "zod";

export type LowercaseLetter =
  | "a"
  | "b"
  | "c"
  | "d"
  | "e"
  | "f"
  | "g"
  | "h"
  | "i"
  | "j"
  | "k"
  | "l"
  | "m"
  | "n"
  | "o"
  | "p"
  | "q"
  | "r"
  | "s"
  | "t"
  | "u"
  | "v"
  | "w"
  | "x"
  | "y"
  | "z";

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
