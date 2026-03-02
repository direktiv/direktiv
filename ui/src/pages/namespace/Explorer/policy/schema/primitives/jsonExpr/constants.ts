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
  ...JsonExprUnaryOperators,
  ...JsonExprBinaryOperators,
  ".",
  "has",
  "is",
  "like",
  "if-then-else",
  "Set",
  "Record",
]);
