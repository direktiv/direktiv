import { z } from "zod";

// when { unknown("department") == context.expectedDepartment };
export const UnknownExpressionSchema = z
  .object({
    Unknown: z.record(z.string()),
  })
  .strict()
  .refine((value) => Object.keys(value.Unknown).length === 1);

export type UnknownExpression = z.infer<typeof UnknownExpressionSchema>;
export type UnknownExpressionInput = z.input<typeof UnknownExpressionSchema>;
