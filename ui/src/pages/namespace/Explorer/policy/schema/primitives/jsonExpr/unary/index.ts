import { JsonExprUnaryOperators } from "../constants";
import { strictSingleKeyObject, unionFromArray } from "../utils";
import { z } from "zod";

const UnaryArgumentSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z
    .object({
      arg: jsonExprSchema,
    })
    .strict();

// when { !context.something }; / when { -1 }; / when { [1, 2].isEmpty() };
export const UnaryJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) =>
  unionFromArray(
    JsonExprUnaryOperators.map((operator) =>
      strictSingleKeyObject(operator, UnaryArgumentSchema(jsonExprSchema))
    )
  );

type UnaryJsonExprSchemaType = z.infer<
  ReturnType<typeof UnaryJsonExprSchema>
>;
