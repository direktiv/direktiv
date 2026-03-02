import { z } from "zod";

const AttributeArgumentsSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z
    .object({
      left: jsonExprSchema,
      attr: z.string(),
    })
    .strict();

// when { context.tls_version };
const DotJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z
    .object({
      ".": AttributeArgumentsSchema(jsonExprSchema),
    })
    .strict();

// when { principal has "email" };
const HasJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z
    .object({
      has: AttributeArgumentsSchema(jsonExprSchema),
    })
    .strict();

export const AttributeJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z.union([
    DotJsonExprSchema(jsonExprSchema),
    HasJsonExprSchema(jsonExprSchema),
  ]);

export type AttributeJsonExprSchemaType = z.infer<
  ReturnType<typeof AttributeJsonExprSchema>
>;
