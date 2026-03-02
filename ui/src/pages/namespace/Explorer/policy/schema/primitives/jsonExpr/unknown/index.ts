import { z } from "zod";

// when { Unknown("x") };
export const UnknownJsonExprSchema = z
  .object({
    Unknown: z.record(z.string()),
  })
  .strict()
  .refine((value) => Object.keys(value.Unknown).length === 1);
