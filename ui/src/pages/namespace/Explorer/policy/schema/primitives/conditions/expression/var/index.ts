import { z } from "zod";

// when { context };
export const VarExpressionSchema = z
  .object({
    Var: z.enum(["principal", "action", "resource", "context"]),
  })
  .strict();
