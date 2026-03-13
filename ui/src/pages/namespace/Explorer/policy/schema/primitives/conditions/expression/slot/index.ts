import { z } from "zod";

// when { principal == ?principal };
export const SlotExpressionSchema = z
  .object({
    Slot: z.enum(["?principal", "?resource"]),
  })
  .strict();

export type SlotExpression = z.infer<typeof SlotExpressionSchema>;
export type SlotExpressionInput = z.input<typeof SlotExpressionSchema>;
