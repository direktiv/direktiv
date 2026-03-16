import { ExpressionSchema } from "./expression";
import { z } from "zod";

/*
  when { principal == User::"alice" }
  unless { resource in Folder::"quarantine" }
*/
const ConditionSchema = z
  .object({
    kind: z.enum(["when", "unless"]),
    body: ExpressionSchema,
  })
  .strict();

export const ConditionsSchema = z.array(ConditionSchema);

export type ConditionType = z.input<typeof ConditionSchema>;

export type ConditionsType = z.input<typeof ConditionsSchema>;
