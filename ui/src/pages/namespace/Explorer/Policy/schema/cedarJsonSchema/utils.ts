import { z } from "zod";

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
