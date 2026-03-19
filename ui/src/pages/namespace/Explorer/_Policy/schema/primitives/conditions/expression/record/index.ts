import type { ExpressionSchemaType } from "../types";
import { strictSingleKeyObject } from "../utils";
import { z } from "zod";

// when { {"user": {"role": "admin", "mfa": true}}.user.mfa };
export const RecordExpressionSchema = (
  expressionSchema: ExpressionSchemaType
) => strictSingleKeyObject("Record", z.record(expressionSchema));
