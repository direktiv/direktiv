import { JsonExprBinaryOperators, strictSingleKeyObject } from "../utils";
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
) => strictSingleKeyObject(operator, BinaryArgumentsSchema(jsonExprSchema));

// when { principal == action };
export const BinaryJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z.union(
    JsonExprBinaryOperators.map((operator) =>
      BinaryOperatorSchema(operator, jsonExprSchema)
    ) as unknown as [z.ZodTypeAny, z.ZodTypeAny, ...z.ZodTypeAny[]]
  );
