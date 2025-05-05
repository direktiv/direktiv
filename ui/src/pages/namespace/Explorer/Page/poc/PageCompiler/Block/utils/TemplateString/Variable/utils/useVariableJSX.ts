import { UseVariableFailure, useVariable } from "./useVariable";

import { VariableType } from "../../../../../../schema/primitives/variable";
import { z } from "zod";

export const JSXValueSchema = z.union([
  z.string(),
  z.number(),
  z.boolean(),
  z.null(),
  z.undefined(),
]);

export type JSXValueType = z.infer<typeof JSXValueSchema>;

type UseVariableJSXSuccess = [JSXValueType, undefined];
type JSXFailure = [undefined, "couldNotStringify"];
type UseVariableJSXFailure = UseVariableFailure | JSXFailure;

/**
 * useVariableJSX does the same as useVariable, but it will add a validation
 * layer on top to ensure that the value returned is JSX compatible.
 */
export const useVariableJSX = (
  value: VariableType
): UseVariableJSXSuccess | UseVariableJSXFailure => {
  const [data, error] = useVariable(value);

  if (error) {
    return [undefined, error];
  }

  const dataParsed = JSXValueSchema.safeParse(data);
  if (!dataParsed.success) {
    return [undefined, "couldNotStringify"];
  }

  return [dataParsed.data, undefined];
};
