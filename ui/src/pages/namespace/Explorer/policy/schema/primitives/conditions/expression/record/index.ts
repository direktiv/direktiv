import { strictSingleKeyObject } from "../utils";
import { z } from "zod";

// when { {"user": {"role": "admin", "mfa": true}}.user.mfa };
export const RecordExpressionSchema = (expressionSchema: z.ZodTypeAny) =>
  strictSingleKeyObject("Record", z.record(expressionSchema));
