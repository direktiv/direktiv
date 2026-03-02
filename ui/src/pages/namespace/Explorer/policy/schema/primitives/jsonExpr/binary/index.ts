import { JsonExprBinaryOperators } from "../constants";
import { z } from "zod";

const BinaryArgumentsSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z
    .object({
      left: jsonExprSchema,
      right: jsonExprSchema,
    })
    .strict();

const BinaryOperatorSchema = (
  operator: (typeof JsonExprBinaryOperators)[number],
  jsonExprSchema: z.ZodTypeAny
) =>
  z
    .object({
      [operator]: BinaryArgumentsSchema(jsonExprSchema),
    })
    .strict();

// when { principal == action };
export const BinaryJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) => {
  const schemas = JsonExprBinaryOperators.map((operator) =>
    BinaryOperatorSchema(operator, jsonExprSchema)
  );

  const [first, second, ...rest] = schemas;

  return z.union([first!, second!, ...rest]);
};

export type BinaryJsonExprSchemaType = z.infer<
  ReturnType<typeof BinaryJsonExprSchema>
>;
