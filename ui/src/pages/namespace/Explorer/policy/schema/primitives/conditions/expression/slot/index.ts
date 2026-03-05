import { z } from "zod";

// when { principal == ?principal };
export const SlotExpressionSchema = z
  .object({
    Slot: z.enum(["?principal", "?resource"]),
  })
  .strict();
