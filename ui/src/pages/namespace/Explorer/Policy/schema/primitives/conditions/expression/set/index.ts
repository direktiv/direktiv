import type { ExpressionSchemaType } from "../types";
import { strictSingleKeyObject } from "../utils";
import { z } from "zod";

// when { action in [Action::"viewReport", Action::"downloadReport"] };
export const SetExpressionSchema = (expressionSchema: ExpressionSchemaType) =>
  strictSingleKeyObject("Set", z.array(expressionSchema));
