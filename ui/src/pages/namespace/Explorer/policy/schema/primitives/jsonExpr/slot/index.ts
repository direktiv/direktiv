import { z } from "zod";

// when { ?principal };
export const SlotJsonExprSchema = z
  .object({
    Slot: z.enum(["?principal", "?resource"]),
  })
  .strict();

type SlotJsonExprSchemaType = z.infer<typeof SlotJsonExprSchema>;
