import { JsonExprUnaryOperators, strictSingleKeyObject } from "../utils";
import { z } from "zod";

const UnaryArgumentSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z.object({ arg: jsonExprSchema }).strict();

// when { !context.something }; / when { -1 }; / when { [1, 2].isEmpty() };
export const UnaryJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z.union(
    JsonExprUnaryOperators.map((operator) =>
      strictSingleKeyObject(operator, UnaryArgumentSchema(jsonExprSchema))
    ) as unknown as [z.ZodTypeAny, z.ZodTypeAny, ...z.ZodTypeAny[]]
  );
