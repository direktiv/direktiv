import { JsonExprSchema } from "./jsonExpr";
import { z } from "zod";

// when { ... } / unless { ... }
const ConditionSchema = z
  .object({
    kind: z.enum(["when", "unless"]),
    body: JsonExprSchema,
  })
  .strict();

// ... when { ... } unless { ... }
export const ConditionsSchema = z.array(ConditionSchema);
