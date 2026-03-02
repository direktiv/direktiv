import { strictSingleKeyObject } from "../utils";
import { z } from "zod";

// when { principal is User in Group::"friends" };
export const IsJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) =>
  strictSingleKeyObject(
    "is",
    z
      .object({
        left: jsonExprSchema,
        entity_type: z.string(),
        in: jsonExprSchema.optional(),
      })
      .strict()
  );

export type IsJsonExprSchemaType = z.infer<ReturnType<typeof IsJsonExprSchema>>;
