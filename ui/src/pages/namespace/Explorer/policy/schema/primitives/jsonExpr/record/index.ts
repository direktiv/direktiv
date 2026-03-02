import { strictSingleKeyObject } from "../utils";
import { z } from "zod";

// when { { foo: "spam", somethingelse: false } };
export const RecordJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) =>
  strictSingleKeyObject("Record", z.record(jsonExprSchema));

export type RecordJsonExprSchemaType = z.infer<
  ReturnType<typeof RecordJsonExprSchema>
>;
