import { z } from "zod";

// when { context.tls_version == "1.3" };
export const VarExpressionSchema = z
  .object({
    Var: z.enum(["principal", "action", "resource", "context"]),
  })
  .strict();
