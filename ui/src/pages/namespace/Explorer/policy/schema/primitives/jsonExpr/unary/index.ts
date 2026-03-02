import { z } from "zod";

const UnaryArgumentSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z
    .object({
      arg: jsonExprSchema,
    })
    .strict();

// when { !context.something };
const NotJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z
    .object({
      "!": UnaryArgumentSchema(jsonExprSchema),
    })
    .strict();

// when { -1 };
const NegJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z
    .object({
      neg: UnaryArgumentSchema(jsonExprSchema),
    })
    .strict();

// when { [1, 2].isEmpty() };
const IsEmptyJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z
    .object({
      isEmpty: UnaryArgumentSchema(jsonExprSchema),
    })
    .strict();

export const UnaryJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z.union([
    NotJsonExprSchema(jsonExprSchema),
    NegJsonExprSchema(jsonExprSchema),
    IsEmptyJsonExprSchema(jsonExprSchema),
  ]);

export type UnaryJsonExprSchemaType = z.infer<
  ReturnType<typeof UnaryJsonExprSchema>
>;
