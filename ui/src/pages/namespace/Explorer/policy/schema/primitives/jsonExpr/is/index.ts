import { z } from "zod";

// when { principal is User in Group::"friends" };
export const IsJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z
    .object({
      is: z
        .object({
          left: jsonExprSchema,
          entity_type: z.string(),
          in: jsonExprSchema.optional(),
        })
        .strict(),
    })
    .strict();

export type IsJsonExprSchemaType = z.infer<ReturnType<typeof IsJsonExprSchema>>;
