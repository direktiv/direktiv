import { strictSingleKeyObject } from "../utils";
import { z } from "zod";

// when { { foo: "spam", somethingelse: false } };
export const RecordExpressionSchema = (expressionSchema: z.ZodTypeAny) =>
  strictSingleKeyObject("Record", z.record(expressionSchema));
