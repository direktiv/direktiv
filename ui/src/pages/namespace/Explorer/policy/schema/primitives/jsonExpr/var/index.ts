import { z } from "zod";

// when { context };
export const VarJsonExprSchema = z
  .object({
    Var: z.enum(["principal", "action", "resource", "context"]),
  })
  .strict();

export type VarJsonExprSchemaType = z.infer<typeof VarJsonExprSchema>;
