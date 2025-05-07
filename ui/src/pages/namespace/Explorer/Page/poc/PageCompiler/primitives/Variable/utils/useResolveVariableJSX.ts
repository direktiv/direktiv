import { ResolveVariableJSXError } from "./errors";
import { Result } from "./types";
import { VariableType } from "../../../../schema/primitives/variable";
import { useResolveVariable } from "./useResolveVariable";
import { z } from "zod";

export const JSXValueSchema = z.union([
  z.string(),
  z.number(),
  z.boolean(),
  z.null(),
  z.undefined(),
]);

export type JSXValueType = z.infer<typeof JSXValueSchema>;

export const useResolveVariableJSX = (
  value: VariableType
): Result<JSXValueType, ResolveVariableJSXError> => {
  const variableResult = useResolveVariable(value);

  if (!variableResult.success) {
    return { success: false, error: variableResult.error };
  }

  const dataParsed = JSXValueSchema.safeParse(variableResult.data);
  if (!dataParsed.success) {
    return { success: false, error: "couldNotStringify" };
  }

  return { success: true, data: dataParsed.data };
};
