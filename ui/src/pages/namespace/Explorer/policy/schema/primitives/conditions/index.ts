import { ExpressionSchema } from "./expression";
import { z } from "zod";

// when { ... } / unless { ... }
const ConditionSchema = z
  .object({
    kind: z.enum(["when", "unless"]),
    body: ExpressionSchema,
  })
  .strict();

// ... when { ... } unless { ... }
export const ConditionsSchema = z.array(ConditionSchema);
