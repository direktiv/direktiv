import { strictSingleKeyObject } from "../utils";
import { z } from "zod";

// when { action in [Action::"viewReport", Action::"downloadReport"] };
export const SetExpressionSchema = (expressionSchema: z.ZodTypeAny) =>
  strictSingleKeyObject("Set", z.array(expressionSchema));
