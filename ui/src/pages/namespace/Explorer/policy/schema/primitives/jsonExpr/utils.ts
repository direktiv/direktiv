import { z } from "zod";

export const strictSingleKeyObject = <
  Key extends string,
  Schema extends z.ZodTypeAny,
>(
  key: Key,
  valueSchema: Schema
) => z.object({ [key]: valueSchema } as Record<Key, Schema>).strict();

export const unionFromArray = (schemas: z.ZodTypeAny[]) => {
  const [first, second, ...rest] = schemas;
  return z.union([first!, second!, ...rest]);
};
