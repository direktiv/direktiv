import { type AttributeOperator, strictSingleKeyObject } from "../utils";
import type { ExpressionSchemaType } from "..";
import { z } from "zod";

const AttributeArgumentsSchema = (expressionSchema: ExpressionSchemaType) =>
  z
    .object({
      left: expressionSchema,
      attr: z.string(),
    })
    .strict();

/*
  when { context.tls_version == "1.3" };
  when { principal has "email" };
*/
export const AttributeExpressionSchema = (
  expressionSchema: ExpressionSchemaType
) => {
  const attributeSchemas = {
    ".": strictSingleKeyObject(".", AttributeArgumentsSchema(expressionSchema)),
    has: strictSingleKeyObject(
      "has",
      AttributeArgumentsSchema(expressionSchema)
    ),
  } satisfies Record<AttributeOperator, ExpressionSchemaType>;

  return z.union([attributeSchemas["."], attributeSchemas.has]);
};
