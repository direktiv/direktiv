import { JsonExprBinaryOperators } from "../constants";
import { strictSingleKeyObject, unionFromArray } from "../utils";
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
export const BinaryJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) => {
  return unionFromArray(
    JsonExprBinaryOperators.map((operator) =>
      BinaryOperatorSchema(operator, jsonExprSchema)
    )
  );
};

type BinaryJsonExprSchemaType = z.infer<
  ReturnType<typeof BinaryJsonExprSchema>
>;
