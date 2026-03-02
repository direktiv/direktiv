import { z } from "zod";

const BinaryOperators = [
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

const BinaryArgumentsSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z
    .object({
      left: jsonExprSchema,
      right: jsonExprSchema,
    })
    .strict();

const BinaryOperatorSchema = (
  operator: (typeof BinaryOperators)[number],
  jsonExprSchema: z.ZodTypeAny
) =>
  z
    .object({
      [operator]: BinaryArgumentsSchema(jsonExprSchema),
    })
    .strict();

// when { principal == action };
export const BinaryJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) => {
  const schemas = BinaryOperators.map((operator) =>
    BinaryOperatorSchema(operator, jsonExprSchema)
  );

  const [first, second, ...rest] = schemas;

  return z.union([first!, second!, ...rest]);
};

export type BinaryJsonExprSchemaType = z.infer<
  ReturnType<typeof BinaryJsonExprSchema>
>;
