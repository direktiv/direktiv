import { z } from "zod";

export const strictSingleKeyObject = (key: string, valueSchema: z.ZodTypeAny) =>
  z.object({ [key]: valueSchema }).strict();

export const unionFromArray = (schemas: z.ZodTypeAny[]) => {
  const [first, second, ...rest] = schemas;
  return z.union([first!, second!, ...rest]);
};
