import { strictSingleKeyObject } from "../utils";
import { z } from "zod";

const AttributeArgumentsSchema = (expressionSchema: z.ZodTypeAny) =>
  z
    .object({
      left: expressionSchema,
      attr: z.string(),
    })
    .strict();

// when { context.tls_version }; / when { principal has "email" };
export const AttributeExpressionSchema = (expressionSchema: z.ZodTypeAny) =>
  z.union([
    strictSingleKeyObject(".", AttributeArgumentsSchema(expressionSchema)),
    strictSingleKeyObject("has", AttributeArgumentsSchema(expressionSchema)),
  ]);
