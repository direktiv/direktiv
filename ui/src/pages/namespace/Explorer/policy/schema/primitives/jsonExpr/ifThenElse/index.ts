import { z } from "zod";

// when { if context.something then principal has "email" else false };
export const IfThenElseJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z
    .object({
      "if-then-else": z
        .object({
          if: jsonExprSchema,
          then: jsonExprSchema,
          else: jsonExprSchema,
        })
        .strict(),
    })
    .strict();

export type IfThenElseJsonExprSchemaType = z.infer<
  ReturnType<typeof IfThenElseJsonExprSchema>
>;
