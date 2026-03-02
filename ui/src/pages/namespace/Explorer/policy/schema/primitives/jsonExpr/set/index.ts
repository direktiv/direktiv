import { z } from "zod";

// when { [1, 2, "something"] };
export const SetJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z
    .object({
      Set: z.array(jsonExprSchema),
    })
    .strict();

export type SetJsonExprSchemaType = z.infer<
  ReturnType<typeof SetJsonExprSchema>
>;
