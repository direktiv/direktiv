import { strictSingleKeyObject } from "../utils";
import { z } from "zod";

const AttributeArgumentsSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z
    .object({
      left: jsonExprSchema,
      attr: z.string(),
    })
    .strict();

// when { context.tls_version }; / when { principal has "email" };
export const AttributeJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z.union([
    strictSingleKeyObject(".", AttributeArgumentsSchema(jsonExprSchema)),
    strictSingleKeyObject("has", AttributeArgumentsSchema(jsonExprSchema)),
  ]);
