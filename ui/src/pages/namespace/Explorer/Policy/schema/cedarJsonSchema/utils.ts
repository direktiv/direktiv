import { z } from "zod";

// Zod's `z.record()` validates values but not semantic constraints on the keys.
// Cedar needs both, for example ensuring `entityTypes` keys are valid identifiers.
export const recordWithValidatedKeys = <Schema extends z.ZodTypeAny>(
  valueSchema: Schema,
  isValidKey: (key: string) => boolean,
  message: string
) =>
  z.record(valueSchema).superRefine((value, ctx) => {
    Object.keys(value).forEach((key) => {
      if (!isValidKey(key)) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message,
          path: [key],
        });
      }
    });
  });
