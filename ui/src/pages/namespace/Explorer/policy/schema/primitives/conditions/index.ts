import { z } from "zod";

// JsonExpr object placeholder: one top-level key, implemented in detail later.
const JsonExprSchema = z
  .record(z.unknown())
  .refine((expr) => Object.keys(expr).length === 1);

// when { ... } / unless { ... }
const ConditionSchema = z
  .object({
    kind: z.enum(["when", "unless"]),
    body: JsonExprSchema,
  })
  .strict();

// ... when { ... } unless { ... }
export const ConditionsSchema = z.array(ConditionSchema);

export type ConditionsSchemaType = z.infer<typeof ConditionsSchema>;
